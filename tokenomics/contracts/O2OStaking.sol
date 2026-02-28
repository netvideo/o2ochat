// SPDX-License-Identifier: MIT
// O2OChat Staking Contract
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

/**
 * @title O2OChat Staking
 * @dev Staking contract for O2O token holders to earn rewards
 */
contract O2OStaking is ReentrancyGuard, AccessControl {
    using SafeERC20 for IERC20;
    
    // Reward token (O2O)
    IERC20 public stakingToken;
    IERC20 public rewardToken;
    
    // Staking info
    struct StakeInfo {
        uint256 amount;
        uint256 rewardDebt;
        uint256 pendingRewards;
        uint256 lastStakeTime;
        bool isActive;
    }
    
    // Pool info
    struct PoolInfo {
        uint256 totalStaked;
        uint256 rewardPerSecond;
        uint256 lastRewardTime;
        uint256 accRewardPerShare;
        uint256 minStakingTime;
        uint256 lockPeriod;
    }
    
    // State variables
    mapping(address => StakeInfo) public stakes;
    PoolInfo public poolInfo;
    
    uint256 public totalStaked;
    uint256 public totalRewardsDistributed;
    
    // Constants
    uint256 public constant REWARD_MULTIPLIER = 1e18;
    uint256 public constant SECONDS_PER_YEAR = 31536000;
    
    // Events
    event Staked(address indexed user, uint256 amount);
    event Withdrawn(address indexed user, uint256 amount);
    event RewardPaid(address indexed user, uint256 amount);
    event RewardRateUpdated(uint256 newRate);
    
    /**
     * @dev Constructor
     * @param _stakingToken Staking token address
     * @param _rewardToken Reward token address
     */
    constructor(address _stakingToken, address _rewardToken) {
        stakingToken = IERC20(_stakingToken);
        rewardToken = IERC20(_rewardToken);
        
        // Setup admin role
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
        
        // Initialize pool
        poolInfo = PoolInfo({
            totalStaked: 0,
            rewardPerSecond: 10 * 1e18, // 10 tokens per second initially
            lastRewardTime: block.timestamp,
            accRewardPerShare: 0,
            minStakingTime: 1 days,
            lockPeriod: 7 days
        });
    }
    
    /**
     * @dev Stake tokens
     * @param amount Amount to stake
     */
    function stake(uint256 amount) 
        external 
        nonReentrant 
    {
        require(amount > 0, "Amount must be > 0");
        
        StakeInfo storage stake = stakes[msg.sender];
        
        // Update pending rewards
        if (stake.isActive) {
            updatePool();
            uint256 pending = pendingRewards(msg.sender);
            if (pending > 0) {
                safeRewardTransfer(msg.sender, pending);
                totalRewardsDistributed += pending;
            }
        }
        
        // Stake new tokens
        stakingToken.safeTransferFrom(msg.sender, address(this), amount);
        
        if (!stake.isActive) {
            stake.isActive = true;
            stake.lastStakeTime = block.timestamp;
        }
        
        stake.amount += amount;
        stake.rewardDebt = stake.amount * poolInfo.accRewardPerShare / REWARD_MULTIPLIER;
        poolInfo.totalStaked += amount;
        totalStaked += amount;
        
        emit Staked(msg.sender, amount);
    }
    
    /**
     * @dev Withdraw staked tokens
     * @param amount Amount to withdraw
     */
    function withdraw(uint256 amount) 
        external 
        nonReentrant 
    {
        StakeInfo storage stake = stakes[msg.sender];
        require(stake.isActive, "No active stake");
        require(amount > 0, "Amount must be > 0");
        require(amount <= stake.amount, "Insufficient staked balance");
        
        // Check lock period
        require(
            block.timestamp >= stake.lastStakeTime + poolInfo.lockPeriod,
            "Tokens still locked"
        );
        
        // Claim rewards
        updatePool();
        uint256 pending = pendingRewards(msg.sender);
        if (pending > 0) {
            safeRewardTransfer(msg.sender, pending);
            totalRewardsDistributed += pending;
        }
        
        // Withdraw principal
        stake.amount -= amount;
        stake.rewardDebt = stake.amount * poolInfo.accRewardPerShare / REWARD_MULTIPLIER;
        poolInfo.totalStaked -= amount;
        totalStaked -= amount;
        
        if (stake.amount == 0) {
            stake.isActive = false;
            stake.lastStakeTime = 0;
        }
        
        stakingToken.safeTransfer(msg.sender, amount);
        
        emit Withdrawn(msg.sender, amount);
    }
    
    /**
     * @dev Claim pending rewards without withdrawing stake
     */
    function claimRewards() external nonReentrant {
        StakeInfo storage stake = stakes[msg.sender];
        require(stake.isActive, "No active stake");
        
        updatePool();
        uint256 pending = pendingRewards(msg.sender);
        require(pending > 0, "No pending rewards");
        
        stake.rewardDebt = stake.amount * poolInfo.accRewardPerShare / REWARD_MULTIPLIER;
        
        safeRewardTransfer(msg.sender, pending);
        totalRewardsDistributed += pending;
        
        emit RewardPaid(msg.sender, pending);
    }
    
    /**
     * @dev Update reward variables for pool
     */
    function updatePool() public {
        if (block.timestamp <= poolInfo.lastRewardTime) {
            return;
        }
        
        if (poolInfo.totalStaked == 0) {
            poolInfo.lastRewardTime = block.timestamp;
            return;
        }
        
        uint256 timeDelta = block.timestamp - poolInfo.lastRewardTime;
        uint256 reward = timeDelta * poolInfo.rewardPerSecond;
        
        poolInfo.accRewardPerShare += (reward * REWARD_MULTIPLIER) / poolInfo.totalStaked;
        poolInfo.lastRewardTime = block.timestamp;
    }
    
    /**
     * @dev Calculate pending rewards for user
     * @param user User address
     * @return Pending reward amount
     */
    function pendingRewards(address user) external view returns (uint256) {
        StakeInfo storage stake = stakes[user];
        if (!stake.isActive) return 0;
        
        uint256 accRewardPerShare = poolInfo.accRewardPerShare;
        if (block.timestamp > poolInfo.lastRewardTime && poolInfo.totalStaked > 0) {
            uint256 timeDelta = block.timestamp - poolInfo.lastRewardTime;
            uint256 reward = timeDelta * poolInfo.rewardPerSecond;
            accRewardPerShare += (reward * REWARD_MULTIPLIER) / poolInfo.totalStaked;
        }
        
        return (stake.amount * accRewardPerShare) / REWARD_MULTIPLIER - stake.rewardDebt;
    }
    
    /**
     * @dev Set reward rate (admin only)
     * @param newRate New reward per second
     */
    function setRewardRate(uint256 newRate) external onlyRole(DEFAULT_ADMIN_ROLE) {
        updatePool();
        poolInfo.rewardPerSecond = newRate;
        emit RewardRateUpdated(newRate);
    }
    
    /**
     * @dev Emergency withdrawal (admin only)
     * @param token Token address
     * @param to Recipient address
     * @param amount Amount to transfer
     */
    function emergencyWithdraw(
        address token,
        address to,
        uint256 amount
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        IERC20(token).safeTransfer(to, amount);
    }
    
    /**
     * @dev Safe reward transfer
     */
    function safeRewardTransfer(address to, uint256 amount) internal {
        uint256 balance = rewardToken.balanceOf(address(this));
        if (amount > balance) {
            rewardToken.safeTransfer(to, balance);
        } else {
            rewardToken.safeTransfer(to, amount);
        }
    }
    
    /**
     * @dev Get staking info for user
     * @param user User address
     * @return amount, rewardDebt, pendingRewards, lastStakeTime, isActive
     */
    function getStakeInfo(address user) 
        external 
        view 
        returns (
            uint256 amount,
            uint256 rewardDebt,
            uint256 pending,
            uint256 lastStakeTime,
            bool isActive
        ) 
    {
        StakeInfo storage stake = stakes[user];
        return (
            stake.amount,
            stake.rewardDebt,
            pendingRewards(user),
            stake.lastStakeTime,
            stake.isActive
        );
    }
}
