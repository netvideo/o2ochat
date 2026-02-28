async function main() {
  console.log("🚀 Starting O2O Token deployment...");
  
  // Get deployer account
  const [deployer] = await ethers.getSigners();
  console.log("Deploying with account:", deployer.address);
  
  // Get balance
  const balance = await ethers.provider.getBalance(deployer.address);
  console.log("Account balance:", ethers.formatEther(balance), "ETH");
  
  // Deploy O2O Token
  console.log("\n📝 Deploying O2OToken...");
  const O2OToken = await ethers.getContractFactory("O2OToken");
  const o2o = await O2OToken.deploy();
  await o2o.waitForDeployment();
  
  const o2oAddress = await o2o.getAddress();
  console.log("✅ O2OToken deployed to:", o2oAddress);
  
  // Get token info
  const name = await o2o.name();
  const symbol = await o2o.symbol();
  const totalSupply = await o2o.totalSupply();
  const decimals = await o2o.decimals();
  
  console.log("\n📊 Token Information:");
  console.log("  Name:", name);
  console.log("  Symbol:", symbol);
  console.log("  Decimals:", decimals);
  console.log("  Total Supply:", ethers.formatEther(totalSupply), "O2O");
  console.log("  Deployer Balance:", ethers.formatEther(await o2o.balanceOf(deployer.address)), "O2O");
  
  // Deploy Staking Contract
  console.log("\n📝 Deploying O2OStaking...");
  const O2OStaking = await ethers.getContractFactory("O2OStaking");
  const staking = await O2OStaking.deploy(o2oAddress, o2oAddress);
  await staking.waitForDeployment();
  
  const stakingAddress = await staking.getAddress();
  console.log("✅ O2OStaking deployed to:", stakingAddress);
  
  // Get staking info
  const poolInfo = await staking.poolInfo();
  console.log("\n📊 Staking Information:");
  console.log("  Reward Per Second:", ethers.formatEther(poolInfo.rewardPerSecond), "O2O");
  console.log("  Min Staking Time:", poolInfo.minStakingTime / 86400, "days");
  console.log("  Lock Period:", poolInfo.lockPeriod / 86400, "days");
  
  // Set up roles (optional - transfer to multi-sig)
  console.log("\n🔐 Setting up roles...");
  const MINTER_ROLE = await o2o.MINTER_ROLE();
  const PAUSER_ROLE = await o2o.PAUSER_ROLE();
  
  console.log("  MINTER_ROLE:", MINTER_ROLE);
  console.log("  PAUSER_ROLE:", PAUSER_ROLE);
  console.log("  Admin:", deployer.address);
  
  console.log("\n✅ Deployment complete!");
  console.log("\n📝 Next Steps:");
  console.log("1. Verify contracts on Etherscan");
  console.log("2. Transfer admin to multi-sig wallet");
  console.log("3. Fund staking contract with reward tokens");
  console.log("4. Run integration tests");
  
  return {
    o2oAddress,
    stakingAddress,
  };
}

// Run deployment
main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
