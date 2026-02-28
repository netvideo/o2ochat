// SPDX-License-Identifier: MIT
// O2OChat Token Contract
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Pausable.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title O2OChat Token
 * @dev ERC20 Token for O2OChat decentralized messaging platform
 * 
 * Features:
 * - Minting capability (restricted)
 * - Burning capability
 * - Pausable transfers
 * - Role-based access control
 * - Reentrancy protection
 * 
 * Security:
 * - OpenZeppelin standard contracts
 * - Multiple audits required
 * - Multi-sig wallet for admin functions
 */
contract O2OToken is ERC20, ERC20Burnable, ERC20Pausable, AccessControl, ReentrancyGuard {
    
    // Roles
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");
    
    // Token Info
    string private constant _TOKEN_NAME = "O2OChat Token";
    string private constant _TOKEN_SYMBOL = "O2O";
    uint8 private constant _DECIMALS = 18;
    
    // Initial Supply: 1 billion tokens
    uint256 private constant _INITIAL_SUPPLY = 1_000_000_000 * 10**18;
    
    // Events
    event TokensMinted(address indexed to, uint256 amount);
    event TokensBurned(address indexed from, uint256 amount);
    event RecoveryExecuted(address indexed token, address indexed to, uint256 amount);
    
    /**
     * @dev Constructor - mints initial supply to deployer
     * Sets up roles and permissions
     */
    constructor() ERC20(_TOKEN_NAME, _TOKEN_SYMBOL) {
        // Mint initial supply to deployer (will be distributed according to tokenomics)
        _mint(msg.sender, _INITIAL_SUPPLY);
        
        // Setup roles
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _setupRole(MINTER_ROLE, msg.sender);
        _setupRole(PAUSER_ROLE, msg.sender);
        _setupRole(BURNER_ROLE, msg.sender);
    }
    
    /**
     * @dev Mint new tokens
     * @param to Address to mint tokens to
     * @param amount Amount of tokens to mint
     * 
     * Requirements:
     * - Caller must have MINTER_ROLE
     * - Cannot mint when paused
     */
    function mint(address to, uint256 amount) 
        external 
        onlyRole(MINTER_ROLE) 
        whenNotPaused 
    {
        _mint(to, amount);
        emit TokensMinted(to, amount);
    }
    
    /**
     * @dev Burn tokens from caller's balance
     * @param amount Amount of tokens to burn
     */
    function burn(uint256 amount) 
        external 
        override 
        whenNotPaused 
    {
        _burn(msg.sender, amount);
        emit TokensBurned(msg.sender, amount);
    }
    
    /**
     * @dev Burn tokens from another account (requires BURNER_ROLE)
     * @param account Account to burn tokens from
     * @param amount Amount of tokens to burn
     */
    function burnFrom(address account, uint256 amount) 
        external 
        override 
        onlyRole(BURNER_ROLE)
        whenNotPaused 
    {
        _burn(account, amount);
        emit TokensBurned(account, amount);
    }
    
    /**
     * @dev Pause token transfers
     * 
     * Requirements:
     * - Caller must have PAUSER_ROLE
     */
    function pause() external onlyRole(PAUSER_ROLE) {
        _pause();
    }
    
    /**
     * @dev Unpause token transfers
     * 
     * Requirements:
     * - Caller must have PAUSER_ROLE
     */
    function unpause() external onlyRole(PAUSER_ROLE) {
        _unpause();
    }
    
    /**
     * @dev Emergency recovery of ERC-20 tokens sent to contract by mistake
     * @param token ERC-20 token contract address
     * @param to Address to send recovered tokens to
     * @param amount Amount of tokens to recover
     * 
     * Requirements:
     * - Caller must have DEFAULT_ADMIN_ROLE
     * - Cannot recover O2O tokens (use burn instead)
     */
    function recoverERC20(
        address token,
        address to,
        uint256 amount
    ) 
        external 
        onlyRole(DEFAULT_ADMIN_ROLE) 
        nonReentrant 
    {
        require(token != address(this), "Cannot recover O2O tokens");
        
        bool success = IERC20(token).transfer(to, amount);
        require(success, "Transfer failed");
        
        emit RecoveryExecuted(token, to, amount);
    }
    
    /**
     * @dev Override _update to add pause check
     */
    function _update(
        address from,
        address to,
        uint256 amount
    ) internal override(ERC20, ERC20Pausable) {
        super._update(from, to, amount);
    }
    
    /**
     * @dev Returns total supply of tokens
     */
    function totalSupply() public view override returns (uint256) {
        return super.totalSupply();
    }
    
    /**
     * @dev Returns decimals for display purposes
     */
    function decimals() public view override returns (uint8) {
        return _DECIMALS;
    }
}
