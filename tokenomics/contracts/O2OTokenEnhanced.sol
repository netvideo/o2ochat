// SPDX-License-Identifier: MIT
// O2OChat Token Contract - Security Enhanced Version
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Pausable.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title O2OChat Token - Security Enhanced
 * @dev Enhanced version with security fixes:
 * - Multi-sig ready
 * - Max supply cap
 * - Overflow protection
 * - Zero amount checks
 */
contract O2OTokenEnhanced is ERC20, ERC20Burnable, ERC20Pausable, AccessControl, ReentrancyGuard {
    
    // Roles
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");
    
    // Token Info
    string private constant _TOKEN_NAME = "O2OChat Token";
    string private constant _TOKEN_SYMBOL = "O2O";
    uint8 private constant _DECIMALS = 18;
    
    // Maximum supply: 1 billion tokens (hard cap)
    uint256 private constant _MAX_SUPPLY = 1_000_000_000 * 10**18;
    
    // Initial Supply
    uint256 private constant _INITIAL_SUPPLY = 1_000_000_000 * 10**18;
    
    // Events
    event TokensMinted(address indexed to, uint256 amount);
    event TokensBurned(address indexed from, uint256 amount);
    event RecoveryExecuted(address indexed token, address indexed to, uint256 amount);
    event MaxSupplyReached(uint256 totalSupply);
    
    /**
     * @dev Constructor with security enhancements
     */
    constructor() ERC20(_TOKEN_NAME, _TOKEN_SYMBOL) {
        // Mint initial supply to deployer (should be multi-sig in production)
        _mint(msg.sender, _INITIAL_SUPPLY);
        
        // Setup roles
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _setupRole(MINTER_ROLE, msg.sender);
        _setupRole(PAUSER_ROLE, msg.sender);
        _setupRole(BURNER_ROLE, msg.sender);
        
        emit MaxSupplyReached(_MAX_SUPPLY);
    }
    
    /**
     * @dev Mint new tokens with security checks
     * 
     * Security fixes:
     * - Max supply cap
     * - Zero amount check
     * - Overflow protection (built-in Solidity 0.8+)
     */
    function mint(address to, uint256 amount) 
        external 
        onlyRole(MINTER_ROLE) 
        whenNotPaused 
    {
        require(amount > 0, "Amount must be > 0");
        require(totalSupply() + amount <= _MAX_SUPPLY, "Would exceed max supply");
        
        _mint(to, amount);
        emit TokensMinted(to, amount);
    }
    
    /**
     * @dev Burn tokens from caller's balance
     */
    function burn(uint256 amount) 
        external 
        override 
        whenNotPaused 
    {
        require(amount > 0, "Amount must be > 0");
        _burn(msg.sender, amount);
        emit TokensBurned(msg.sender, amount);
    }
    
    /**
     * @dev Burn tokens from another account (requires BURNER_ROLE)
     */
    function burnFrom(address account, uint256 amount) 
        external 
        override 
        onlyRole(BURNER_ROLE)
        whenNotPaused 
    {
        require(amount > 0, "Amount must be > 0");
        _burn(account, amount);
        emit TokensBurned(account, amount);
    }
    
    /**
     * @dev Pause token transfers
     */
    function pause() external onlyRole(PAUSER_ROLE) {
        _pause();
    }
    
    /**
     * @dev Unpause token transfers
     */
    function unpause() external onlyRole(PAUSER_ROLE) {
        _unpause();
    }
    
    /**
     * @dev Emergency recovery of ERC-20 tokens
     * Security: Cannot recover O2O tokens
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
        require(amount > 0, "Amount must be > 0");
        
        bool success = IERC20(token).transfer(to, amount);
        require(success, "Transfer failed");
        
        emit RecoveryExecuted(token, to, amount);
    }
    
    /**
     * @dev Returns max supply
     */
    function maxSupply() external pure returns (uint256) {
        return _MAX_SUPPLY;
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
