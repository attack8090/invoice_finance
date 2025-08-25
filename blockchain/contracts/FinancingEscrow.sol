// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "./InvoiceNFT.sol";

contract FinancingEscrow is ReentrancyGuard, Ownable, Pausable {
    struct FinancingRequest {
        uint256 invoiceTokenId;
        address smeAddress;
        uint256 requestedAmount;
        uint256 interestRate; // Basis points (e.g., 500 = 5%)
        uint256 repaymentAmount;
        uint256 dueDate;
        bool isActive;
        bool isCompleted;
        uint256 createdAt;
    }

    struct Investment {
        address investor;
        uint256 amount;
        uint256 expectedReturn;
        bool isWithdrawn;
        uint256 investedAt;
    }

    InvoiceNFT public invoiceNFT;
    uint256 public platformFeeRate = 100; // 1% in basis points
    address public platformWallet;

    mapping(uint256 => FinancingRequest) public financingRequests;
    mapping(uint256 => mapping(address => Investment)) public investments;
    mapping(uint256 => address[]) public requestInvestors;
    mapping(uint256 => uint256) public totalInvested;

    uint256 public requestCounter;

    event FinancingRequestCreated(
        uint256 indexed requestId,
        uint256 indexed invoiceTokenId,
        address indexed sme,
        uint256 requestedAmount,
        uint256 interestRate
    );

    event InvestmentMade(
        uint256 indexed requestId,
        address indexed investor,
        uint256 amount,
        uint256 expectedReturn
    );

    event RequestFulfilled(
        uint256 indexed requestId,
        uint256 totalAmount
    );

    event RepaymentMade(
        uint256 indexed requestId,
        uint256 amount,
        uint256 platformFee
    );

    event InvestorPayout(
        uint256 indexed requestId,
        address indexed investor,
        uint256 amount
    );

    constructor(
        address _invoiceNFT,
        address _platformWallet,
        address initialOwner
    ) Ownable(initialOwner) {
        invoiceNFT = InvoiceNFT(_invoiceNFT);
        platformWallet = _platformWallet;
    }

    function createFinancingRequest(
        uint256 invoiceTokenId,
        uint256 requestedAmount,
        uint256 interestRate
    ) external nonReentrant whenNotPaused returns (uint256) {
        require(invoiceNFT.ownerOf(invoiceTokenId) == msg.sender, "Not invoice owner");
        
        InvoiceNFT.Invoice memory invoice = invoiceNFT.getInvoice(invoiceTokenId);
        require(invoice.isVerified, "Invoice must be verified");
        require(!invoice.isFinanced, "Invoice already financed");
        require(requestedAmount <= invoice.invoiceAmount, "Requested amount exceeds invoice amount");

        requestCounter++;
        uint256 requestId = requestCounter;

        uint256 repaymentAmount = requestedAmount + (requestedAmount * interestRate / 10000);

        financingRequests[requestId] = FinancingRequest({
            invoiceTokenId: invoiceTokenId,
            smeAddress: msg.sender,
            requestedAmount: requestedAmount,
            interestRate: interestRate,
            repaymentAmount: repaymentAmount,
            dueDate: invoice.dueDate,
            isActive: true,
            isCompleted: false,
            createdAt: block.timestamp
        });

        emit FinancingRequestCreated(
            requestId,
            invoiceTokenId,
            msg.sender,
            requestedAmount,
            interestRate
        );

        return requestId;
    }

    function invest(uint256 requestId) external payable nonReentrant whenNotPaused {
        FinancingRequest storage request = financingRequests[requestId];
        require(request.isActive, "Request not active");
        require(!request.isCompleted, "Request already completed");
        require(msg.value > 0, "Investment amount must be greater than 0");
        require(
            totalInvested[requestId] + msg.value <= request.requestedAmount,
            "Investment exceeds requested amount"
        );

        uint256 proportion = (msg.value * 10000) / request.requestedAmount;
        uint256 expectedReturn = (request.repaymentAmount * proportion) / 10000;

        if (investments[requestId][msg.sender].amount == 0) {
            requestInvestors[requestId].push(msg.sender);
        }

        investments[requestId][msg.sender].investor = msg.sender;
        investments[requestId][msg.sender].amount += msg.value;
        investments[requestId][msg.sender].expectedReturn += expectedReturn;
        investments[requestId][msg.sender].investedAt = block.timestamp;

        totalInvested[requestId] += msg.value;

        emit InvestmentMade(requestId, msg.sender, msg.value, expectedReturn);

        // Check if request is fully funded
        if (totalInvested[requestId] >= request.requestedAmount) {
            _fulfillRequest(requestId);
        }
    }

    function _fulfillRequest(uint256 requestId) internal {
        FinancingRequest storage request = financingRequests[requestId];
        request.isActive = false;

        // Transfer funds to SME
        uint256 platformFee = (request.requestedAmount * platformFeeRate) / 10000;
        uint256 smeAmount = request.requestedAmount - platformFee;

        payable(request.smeAddress).transfer(smeAmount);
        payable(platformWallet).transfer(platformFee);

        // Mark invoice as financed
        invoiceNFT.markAsFinanced(request.invoiceTokenId, address(this));

        emit RequestFulfilled(requestId, request.requestedAmount);
    }

    function repayLoan(uint256 requestId) external payable nonReentrant {
        FinancingRequest storage request = financingRequests[requestId];
        require(msg.sender == request.smeAddress, "Only SME can repay");
        require(!request.isActive, "Request still active");
        require(!request.isCompleted, "Loan already repaid");
        require(msg.value >= request.repaymentAmount, "Insufficient repayment amount");

        request.isCompleted = true;

        uint256 platformFee = (msg.value * platformFeeRate) / 10000;
        uint256 investorPayout = msg.value - platformFee;

        payable(platformWallet).transfer(platformFee);

        // Distribute returns to investors
        _distributeReturns(requestId, investorPayout);

        emit RepaymentMade(requestId, msg.value, platformFee);
    }

    function _distributeReturns(uint256 requestId, uint256 totalPayout) internal {
        address[] memory investors = requestInvestors[requestId];
        uint256 totalInvestment = totalInvested[requestId];

        for (uint256 i = 0; i < investors.length; i++) {
            address investor = investors[i];
            Investment storage investment = investments[requestId][investor];
            
            if (!investment.isWithdrawn) {
                uint256 proportion = (investment.amount * 10000) / totalInvestment;
                uint256 payout = (totalPayout * proportion) / 10000;
                
                investment.isWithdrawn = true;
                payable(investor).transfer(payout);

                emit InvestorPayout(requestId, investor, payout);
            }
        }
    }

    function getFinancingRequest(uint256 requestId) external view returns (FinancingRequest memory) {
        return financingRequests[requestId];
    }

    function getInvestment(uint256 requestId, address investor) external view returns (Investment memory) {
        return investments[requestId][investor];
    }

    function getRequestInvestors(uint256 requestId) external view returns (address[] memory) {
        return requestInvestors[requestId];
    }

    function setPlatformFeeRate(uint256 _feeRate) external onlyOwner {
        require(_feeRate <= 1000, "Fee rate cannot exceed 10%");
        platformFeeRate = _feeRate;
    }

    function setPlatformWallet(address _platformWallet) external onlyOwner {
        platformWallet = _platformWallet;
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    function emergencyWithdraw() external onlyOwner whenPaused {
        payable(owner()).transfer(address(this).balance);
    }

    receive() external payable {}
}
