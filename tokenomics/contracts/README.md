# O2OChat Token 智能合约

## 合约概述

O2OChat Token (O2O) 是基于 ERC-20 标准的功能型代币合约，支持多链部署。

## 合约地址

### 测试网

| 网络 | 合约地址 | 状态 |
|------|---------|------|
| Ethereum Goerli | TBD | ⏳ 待部署 |
| Polygon Mumbai | TBD | ⏳ 待部署 |
| BSC Testnet | TBD | ⏳ 待部署 |

### 主网

| 网络 | 合约地址 | 状态 |
|------|---------|------|
| Ethereum Mainnet | TBD | ⏳ 待部署 |
| Polygon | TBD | ⏳ 待部署 |
| BSC | TBD | ⏳ 待部署 |

## 合约特性

### 核心功能

- ✅ **ERC-20 标准** - 完全兼容 ERC-20 接口
- ✅ **可升级合约** - 支持代理升级模式
- ✅ **多签钱包** - 团队资金多签管理
- ✅ **时间锁** - 重大操作有时间延迟
- ✅ **权限控制** - 基于角色的访问控制

### 代币经济学

| 属性 | 值 |
|------|-----|
| **名称** | O2OChat Token |
| **符号** | O2O |
| **精度** | 18 |
| **总供应量** | 1,000,000,000 O2O |
| **可铸造** | 是 (受限) |
| **可燃烧** | 是 |

### 安全特性

- ✅ **重入保护** - 防止重入攻击
- ✅ **溢出检查** - SafeMath 库
- ✅ **权限分离** - Owner/Operator 分离
- ✅ **紧急暂停** - Emergency Stop 功能
- ✅ **审计** - 多次第三方审计

## 合约架构

```
contracts/
├── O2OToken.sol          # 主代币合约
├── O2OTokenUpgradeable.sol  # 可升级版本
├── Staking.sol           # 质押合约
├── Vesting.sol           # 解锁合约
├── Governance.sol        # 治理合约
├── Treasury.sol          # 金库合约
└── interfaces/
    ├── IO2OToken.sol     # 代币接口
    ├── IStaking.sol      # 质押接口
    └── IGovernance.sol   # 治理接口
```

## 部署指南

### 环境准备

```bash
# 安装依赖
npm install

# 编译合约
npm run compile

# 运行测试
npm run test
```

### 部署到测试网

```bash
# 部署到 Goerli
npx hardhat run scripts/deploy.js --network goerli

# 验证合约
npx hardhat verify --network goerli DEPLOYED_CONTRACT_ADDRESS
```

### 部署到主网

```bash
# ⚠️ 需要多重签名确认
npx hardhat run scripts/deploy.js --network mainnet

# 验证合约
npx hardhat verify --network mainnet DEPLOYED_CONTRACT_ADDRESS
```

## 使用说明

### 转账

```javascript
// 使用 MetaMask 或其他钱包
const tx = await token.transfer(recipientAddress, amount);
await tx.wait();
```

### 授权

```javascript
// 授权spender 使用代币
const tx = await token.approve(spenderAddress, amount);
await tx.wait();
```

### 查询余额

```javascript
const balance = await token.balanceOf(walletAddress);
console.log(`Balance: ${ethers.utils.formatEther(balance)} O2O`);
```

## 安全考虑

### 已实施的安全措施

1. **代码审计**
   - ✅ OpenZeppelin 标准合约
   - ✅ 多次第三方审计
   - ✅ Bug Bounty 计划

2. **权限控制**
   - ✅ 多签钱包管理
   - ✅ 时间锁延迟
   - ✅ 角色分离

3. **应急措施**
   - ✅ 紧急暂停功能
   - ✅ 合约升级机制
   - ✅ 资金追回机制

### 已知限制

- ⚠️ ERC-20 标准限制
- ⚠️ Gas 费用波动
- ⚠️ 跨链桥风险

## 升级流程

```
1. 提案阶段 (7 天)
   └→ 社区讨论

2. 投票阶段 (7 天)
   └→ DAO 投票

3. 时间锁 (48 小时)
   └→ 延迟执行

4. 部署新版本
   └→ 代理升级

5. 迁移数据
   └→ 状态迁移
```

## 监控和报告

### 链上监控

- Etherscan/Polygonscan
- Dune Analytics Dashboard
- The Graph Subgraphs

### 事件日志

```javascript
// 监听转账事件
token.on("Transfer", (from, to, amount, event) => {
  console.log(`Transfer: ${from} -> ${to}, Amount: ${amount}`);
});

// 监听授权事件
token.on("Approval", (owner, spender, amount, event) => {
  console.log(`Approval: ${owner} -> ${spender}, Amount: ${amount}`);
});
```

## 相关文档

- [代币经济学白皮书](../WHITEPAPER.md)
- [治理系统](./GOVERNANCE.md)
- [质押指南](./STAKING.md)

## 免责声明

⚠️ **智能合约存在风险**

- 代码已通过审计，但不保证 100% 安全
- 使用合约前请自行评估风险
- 投资需谨慎，自行承担风险

---

**版本**: v1.0.0  
**最后更新**: 2026 年 2 月 28 日  
**状态**: 开发中
