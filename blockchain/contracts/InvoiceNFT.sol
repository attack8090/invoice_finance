// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

contract InvoiceNFT is ERC721, ERC721URIStorage, Ownable {
    using Counters for Counters.Counter;

    Counters.Counter private _tokenIdCounter;

    struct Invoice {
        uint256 tokenId;
        string invoiceNumber;
        address smeAddress;
        uint256 invoiceAmount;
        uint256 dueDate;
        uint256 issueDate;
        string customerName;
        bool isVerified;
        bool isFinanced;
        uint8 riskLevel; // 0: Low, 1: Medium, 2: High
    }

    mapping(uint256 => Invoice) public invoices;
    mapping(string => uint256) public invoiceNumberToTokenId;
    
    event InvoiceTokenized(
        uint256 indexed tokenId,
        string invoiceNumber,
        address indexed smeAddress,
        uint256 invoiceAmount
    );
    
    event InvoiceVerified(uint256 indexed tokenId, bool verified);
    event InvoiceFinanced(uint256 indexed tokenId, address indexed financier);

    constructor(address initialOwner) 
        ERC721("Invoice Finance NFT", "IFNFT") 
        Ownable(initialOwner) 
    {}

    function tokenizeInvoice(
        string memory invoiceNumber,
        address smeAddress,
        uint256 invoiceAmount,
        uint256 dueDate,
        uint256 issueDate,
        string memory customerName,
        string memory tokenURI,
        uint8 riskLevel
    ) public onlyOwner returns (uint256) {
        require(invoiceNumberToTokenId[invoiceNumber] == 0, "Invoice already tokenized");
        require(invoiceAmount > 0, "Invoice amount must be greater than 0");
        require(dueDate > block.timestamp, "Due date must be in the future");

        _tokenIdCounter.increment();
        uint256 tokenId = _tokenIdCounter.current();

        _safeMint(smeAddress, tokenId);
        _setTokenURI(tokenId, tokenURI);

        invoices[tokenId] = Invoice({
            tokenId: tokenId,
            invoiceNumber: invoiceNumber,
            smeAddress: smeAddress,
            invoiceAmount: invoiceAmount,
            dueDate: dueDate,
            issueDate: issueDate,
            customerName: customerName,
            isVerified: false,
            isFinanced: false,
            riskLevel: riskLevel
        });

        invoiceNumberToTokenId[invoiceNumber] = tokenId;

        emit InvoiceTokenized(tokenId, invoiceNumber, smeAddress, invoiceAmount);
        return tokenId;
    }

    function verifyInvoice(uint256 tokenId, bool verified) public onlyOwner {
        require(_exists(tokenId), "Invoice does not exist");
        invoices[tokenId].isVerified = verified;
        emit InvoiceVerified(tokenId, verified);
    }

    function markAsFinanced(uint256 tokenId, address financier) public onlyOwner {
        require(_exists(tokenId), "Invoice does not exist");
        require(invoices[tokenId].isVerified, "Invoice must be verified first");
        
        invoices[tokenId].isFinanced = true;
        emit InvoiceFinanced(tokenId, financier);
    }

    function getInvoice(uint256 tokenId) public view returns (Invoice memory) {
        require(_exists(tokenId), "Invoice does not exist");
        return invoices[tokenId];
    }

    function getInvoiceByNumber(string memory invoiceNumber) public view returns (Invoice memory) {
        uint256 tokenId = invoiceNumberToTokenId[invoiceNumber];
        require(tokenId != 0, "Invoice not found");
        return invoices[tokenId];
    }

    function totalSupply() public view returns (uint256) {
        return _tokenIdCounter.current();
    }

    // Override required functions
    function _burn(uint256 tokenId) internal override(ERC721, ERC721URIStorage) {
        super._burn(tokenId);
    }

    function tokenURI(uint256 tokenId) public view override(ERC721, ERC721URIStorage) returns (string memory) {
        return super.tokenURI(tokenId);
    }

    function supportsInterface(bytes4 interfaceId) public view override(ERC721, ERC721URIStorage) returns (bool) {
        return super.supportsInterface(interfaceId);
    }
}
