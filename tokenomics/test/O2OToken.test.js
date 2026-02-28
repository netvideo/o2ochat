const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("O2OToken", function () {
  let O2OToken;
  let o2o;
  let owner;
  let addr1;
  let addr2;
  let addrs;

  const INITIAL_SUPPLY = ethers.parseEther("1000000000"); // 1 billion tokens

  beforeEach(async function () {
    // Get signers
    [owner, addr1, addr2, ...addrs] = await ethers.getSigners();

    // Deploy contract
    O2OToken = await ethers.getContractFactory("O2OToken");
    o2o = await O2OToken.deploy();
    await o2o.waitForDeployment();
  });

  describe("Deployment", function () {
    it("Should set the correct token name and symbol", async function () {
      expect(await o2o.name()).to.equal("O2OChat Token");
      expect(await o2o.symbol()).to.equal("O2O");
    });

    it("Should mint initial supply to deployer", async function () {
      const balance = await o2o.balanceOf(owner.address);
      expect(balance).to.equal(INITIAL_SUPPLY);
    });

    it("Should set correct decimals", async function () {
      expect(await o2o.decimals()).to.equal(18);
    });

    it("Should assign DEFAULT_ADMIN_ROLE to deployer", async function () {
      expect(await o2o.hasRole(await o2o.DEFAULT_ADMIN_ROLE(), owner.address)).to.be.true;
    });
  });

  describe("Transfers", function () {
    it("Should transfer tokens between accounts", async function () {
      const transferAmount = ethers.parseEther("1000");
      
      await o2o.transfer(addr1.address, transferAmount);
      expect(await o2o.balanceOf(addr1.address)).to.equal(transferAmount);
      
      await o2o.connect(addr1).transfer(addr2.address, ethers.parseEther("500"));
      expect(await o2o.balanceOf(addr2.address)).to.equal(ethers.parseEther("500"));
    });

    it("Should fail if sender doesn't have enough tokens", async function () {
      const transferAmount = ethers.parseEther("1001");
      
      await expect(
        o2o.connect(addr1).transfer(addr2.address, transferAmount)
      ).to.be.reverted;
    });
  });

  describe("Minting", function () {
    it("Should allow minter to mint tokens", async function () {
      const MINTER_ROLE = await o2o.MINTER_ROLE();
      await o2o.grantRole(MINTER_ROLE, addr1.address);
      
      const mintAmount = ethers.parseEther("10000");
      await o2o.connect(addr1).mint(addr2.address, mintAmount);
      
      expect(await o2o.balanceOf(addr2.address)).to.equal(mintAmount);
    });

    it("Should not allow non-minter to mint", async function () {
      const mintAmount = ethers.parseEther("10000");
      
      await expect(
        o2o.connect(addr1).mint(addr2.address, mintAmount)
      ).to.be.reverted;
    });

    it("Should emit TokensMinted event", async function () {
      const MINTER_ROLE = await o2o.MINTER_ROLE();
      await o2o.grantRole(MINTER_ROLE, addr1.address);
      
      const mintAmount = ethers.parseEther("1000");
      
      await expect(
        o2o.connect(addr1).mint(addr2.address, mintAmount)
      ).to.emit(o2o, "TokensMinted")
        .withArgs(addr2.address, mintAmount);
    });
  });

  describe("Burning", function () {
    beforeEach(async function () {
      const transferAmount = ethers.parseEther("10000");
      await o2o.transfer(addr1.address, transferAmount);
    });

    it("Should allow users to burn their tokens", async function () {
      const burnAmount = ethers.parseEther("1000");
      const initialBalance = await o2o.balanceOf(addr1.address);
      
      await o2o.connect(addr1).burn(burnAmount);
      
      expect(await o2o.balanceOf(addr1.address)).to.equal(initialBalance - burnAmount);
    });

    it("Should emit TokensBurned event", async function () {
      const burnAmount = ethers.parseEther("500");
      
      await expect(
        o2o.connect(addr1).burn(burnAmount)
      ).to.emit(o2o, "TokensBurned")
        .withArgs(addr1.address, burnAmount);
    });

    it("Should allow burner to burn from another account", async function () {
      const BURNER_ROLE = await o2o.BURNER_ROLE();
      await o2o.grantRole(BURNER_ROLE, addr2.address);
      
      const burnAmount = ethers.parseEther("100");
      const initialBalance = await o2o.balanceOf(addr1.address);
      
      await o2o.connect(addr2).burnFrom(addr1.address, burnAmount);
      
      expect(await o2o.balanceOf(addr1.address)).to.equal(initialBalance - burnAmount);
    });
  });

  describe("Pausing", function () {
    it("Should allow pauser to pause transfers", async function () {
      const PAUSER_ROLE = await o2o.PAUSER_ROLE();
      await o2o.grantRole(PAUSER_ROLE, addr1.address);
      
      await o2o.connect(addr1).pause();
      
      await expect(
        o2o.transfer(addr1.address, ethers.parseEther("100"))
      ).to.be.reverted;
    });

    it("Should allow pauser to unpause transfers", async function () {
      const PAUSER_ROLE = await o2o.PAUSER_ROLE();
      await o2o.grantRole(PAUSER_ROLE, addr1.address);
      
      await o2o.connect(addr1).pause();
      await o2o.connect(addr1).unpause();
      
      // Should work after unpause
      await o2o.transfer(addr1.address, ethers.parseEther("100"));
      expect(await o2o.balanceOf(addr1.address)).to.equal(ethers.parseEther("100"));
    });
  });

  describe("Recovery", function () {
    it("Should allow admin to recover ERC20 tokens", async function () {
      // Deploy another token
      const TestToken = await ethers.getContractFactory("O2OToken");
      const testToken = await TestToken.deploy();
      
      // Send some tokens to O2O contract
      await testToken.transfer(await o2o.getAddress(), ethers.parseEther("1000"));
      
      // Recover them
      const adminBalanceBefore = await testToken.balanceOf(owner.address);
      await o2o.recoverERC20(await testToken.getAddress(), owner.address, ethers.parseEther("1000"));
      const adminBalanceAfter = await testToken.balanceOf(owner.address);
      
      expect(adminBalanceAfter - adminBalanceBefore).to.equal(ethers.parseEther("1000"));
    });

    it("Should not allow recovering O2O tokens", async function () {
      await expect(
        o2o.recoverERC22(await o2o.getAddress(), owner.address, ethers.parseEther("100"))
      ).to.be.revertedWith("Cannot recover O2O tokens");
    });

    it("Should not allow non-admin to recover", async function () {
      const TestToken = await ethers.getContractFactory("O2OToken");
      const testToken = await TestToken.deploy();
      
      await expect(
        o2o.connect(addr1).recoverERC20(await testToken.getAddress(), addr1.address, ethers.parseEther("1000"))
      ).to.be.reverted;
    });
  });

  describe("Access Control", function () {
    it("Should allow admin to grant and revoke roles", async function () {
      const MINTER_ROLE = await o2o.MINTER_ROLE();
      
      // Grant role
      await o2o.grantRole(MINTER_ROLE, addr1.address);
      expect(await o2o.hasRole(MINTER_ROLE, addr1.address)).to.be.true;
      
      // Revoke role
      await o2o.revokeRole(MINTER_ROLE, addr1.address);
      expect(await o2o.hasRole(MINTER_ROLE, addr1.address)).to.be.false;
    });

    it("Should not allow non-admin to grant roles", async function () {
      const MINTER_ROLE = await o2o.MINTER_ROLE();
      
      await expect(
        o2o.connect(addr1).grantRole(MINTER_ROLE, addr2.address)
      ).to.be.reverted;
    });
  });
});
