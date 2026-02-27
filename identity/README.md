# Identity Module - 身份管理模块

## 功能概述
负责生成、管理和验证用户身份，包括密钥对生成、Peer ID计算、身份验证等核心功能。

## 核心功能
1. **密钥对生成**：使用Ed25519算法生成公私钥对
2. **Peer ID计算**：基于公钥哈希生成唯一标识符
3. **身份验证**：验证消息签名和挑战响应
4. **密钥存储**：安全存储私钥和身份信息
5. **身份导入/导出**：支持身份备份和恢复

## 接口定义

### 类型定义
```go
// 身份信息
type Identity struct {
    PeerID     string          // Base58编码的Peer ID
    PublicKey  ed25519.PublicKey  // 公钥
    PrivateKey ed25519.PrivateKey // 私钥（加密存储）
    CreatedAt  time.Time       // 创建时间
}

// 身份配置
type IdentityConfig struct {
    KeyType    string        // 密钥类型："ed25519"
    KeyLength  int           // 密钥长度：256
    PeerIDEncoding string    // Peer ID编码："base58"
}
```

### 主要接口
```go
// 身份管理器接口
type IdentityManager interface {
    // 创建新身份
    CreateIdentity(config *IdentityConfig) (*Identity, error)
    
    // 加载现有身份
    LoadIdentity(peerID string) (*Identity, error)
    
    // 验证身份
    VerifyIdentity(peerID string, publicKey []byte) bool
    
    // 签名消息
    SignMessage(message []byte) ([]byte, error)
    
    // 验证签名
    VerifySignature(peerID string, message, signature []byte) bool
    
    // 导出身份（加密）
    ExportIdentity(password string) ([]byte, error)
    
    // 导入身份
    ImportIdentity(data []byte, password string) (*Identity, error)
}
```

### 辅助接口
```go
// Peer ID工具
type PeerIDUtil interface {
    // 从公钥生成Peer ID
    GeneratePeerID(publicKey []byte) string
    
    // 验证Peer ID格式
    ValidatePeerID(peerID string) bool
    
    // 从Peer ID提取公钥哈希
    ExtractPublicKeyHash(peerID string) ([]byte, error)
}

// 密钥存储接口
type KeyStorage interface {
    SavePrivateKey(peerID string, encryptedKey []byte) error
    LoadPrivateKey(peerID string) ([]byte, error)
    DeletePrivateKey(peerID string) error
}
```

## 实现要求

### 1. 密钥生成
- 使用Ed25519算法生成密钥对
- 私钥必须加密存储
- 支持密钥轮换机制

### 2. Peer ID生成
- Peer ID = Base58(SHA256(公钥)[0:16])
- 确保全局唯一性
- 支持Peer ID格式验证

### 3. 签名验证
- 所有信令消息必须签名
- 支持挑战-响应验证
- 实现前向安全签名

### 4. 安全存储
- 私钥使用AES-GCM加密
- 支持硬件安全模块（可选）
- 实现密钥备份和恢复

## 测试要求

### 单元测试
```bash
# 运行身份模块测试
go test ./identity -v

# 测试特定功能
go test ./identity -run TestCreateIdentity
go test ./identity -run TestSignVerify
go test ./identity -run TestPeerIDGeneration
```

### 测试用例
1. **身份创建测试**：验证密钥对生成和Peer ID计算
2. **签名验证测试**：测试消息签名和验证功能
3. **导入导出测试**：验证身份备份和恢复
4. **并发安全测试**：测试多线程环境下的安全性
5. **错误处理测试**：测试各种错误场景

### 性能测试
```bash
# 基准测试
go test ./identity -bench=.
go test ./identity -bench=BenchmarkSignMessage
go test ./identity -bench=BenchmarkVerifySignature
```

## 依赖关系
- crypto模块：用于加密操作
- storage模块：用于密钥存储

## 使用示例

```go
import "github.com/netvideo/identity"

// 创建内存存储管理器
store := identity.NewMemoryIdentityStore()
keyStore := identity.NewMemoryKeyStorage()

// 创建身份管理器
manager, err := identity.NewIdentityManager(store, keyStore)
if err != nil {
    log.Fatal(err)
}

// 创建新身份
config := &identity.IdentityConfig{
    KeyType:        identity.KeyTypeEd25519,
    KeyLength:     256,
    PeerIDEncoding: identity.PeerIDEncodingBase58,
}
identity, err := manager.CreateIdentity(config)

// 签名消息
message := []byte("Hello, World!")
signature, err := manager.SignMessage(message)

// 验证签名
valid := manager.VerifySignature(identity.PeerID, message, signature)

// 获取身份元数据
metadata, err := manager.GetMetadata(identity.PeerID)
metadata.DisplayName = "My Name"
manager.UpdateMetadata(identity.PeerID, metadata)

// 挑战-响应验证
challenge, err := manager.GenerateChallenge(identity.PeerID)
// ... 使用私钥签名 challenge.Challenge 得到 response
response := &identity.ChallengeResponse{
    Challenge: challenge.Challenge,
    Response:  base64.StdEncoding.EncodeToString(signature),
    PeerID:    identity.PeerID,
}
valid, err = manager.VerifyChallenge(identity.PeerID, challenge, response)

// 导出身份（加密备份）
exportData, err := manager.ExportIdentity("secure-password")

// 导入身份
importedIdentity, err := manager.ImportIdentity(exportData, "secure-password")

// 列出所有身份
identities, err := manager.ListIdentities()

// 删除身份
err = manager.DeleteIdentity(identity.PeerID)
```

## 错误处理
- 所有函数必须返回明确的错误信息
- 密钥操作失败必须清理临时数据
- 身份验证失败必须记录日志

## 安全注意事项
1. 私钥绝不能以明文形式存储或传输
2. 所有加密操作必须使用安全的随机数生成器
3. 实现防暴力破解机制
4. 定期审计密钥使用情况