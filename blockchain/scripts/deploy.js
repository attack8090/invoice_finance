async function main() {
  const [deployer] = await ethers.getSigners();

  console.log("Deploying contracts with the account:", deployer.address);
  console.log("Account balance:", (await deployer.getBalance()).toString());

  // Deploy InvoiceNFT contract
  const InvoiceNFT = await ethers.getContractFactory("InvoiceNFT");
  const invoiceNFT = await InvoiceNFT.deploy(deployer.address);
  await invoiceNFT.deployed();

  console.log("InvoiceNFT contract deployed to:", invoiceNFT.address);

  // Deploy FinancingEscrow contract
  const FinancingEscrow = await ethers.getContractFactory("FinancingEscrow");
  const financingEscrow = await FinancingEscrow.deploy(
    invoiceNFT.address,
    deployer.address, // Platform wallet
    deployer.address  // Initial owner
  );
  await financingEscrow.deployed();

  console.log("FinancingEscrow contract deployed to:", financingEscrow.address);

  // Save contract addresses to a file
  const fs = require("fs");
  const contractAddresses = {
    InvoiceNFT: invoiceNFT.address,
    FinancingEscrow: financingEscrow.address,
    deployer: deployer.address
  };

  fs.writeFileSync(
    "./contract-addresses.json",
    JSON.stringify(contractAddresses, null, 2)
  );

  console.log("Contract addresses saved to contract-addresses.json");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
