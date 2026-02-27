# Crypto Module - 加密安全模块

## 功能概述
负责所有加密相关操作，包括密钥交换、消息加密、数字签名、哈希计算等安全功能。

## 核心功能
1. **密钥交换**：前向安全的密钥协商
2. **消息加密**：对称加密保护数据
3. **数字签名**：身份验证和消息完整性
4. **哈希计算**：文件完整性验证
5. **随机数生成**：安全的随机数生成
6. **密钥管理**：密钥生命周期管理

## 接口定义

### 类型定义
```go
// 加密算法类型
type AlgorithmType string

const (
    AlgorithmAESGCM    AlgorithmType = "aes-gcm"
    AlgorithmChaCha20  AlgorithmType = "chacha20-poly1305"
    AlgorithmEd25519   AlgorithmType = "ed25519"
    AlgorithmX25519    AlgorithmType = "x25519"
    AlgorithmSHA256    AlgorithmType = "sha256"
    AlgorithmSHA3_256  AlgorithmType = "sha3-256"
    AlgorithmBLAKE2b   AlgorithmType = "blake2b"
)

// 密钥对
type KeyPair struct {
    PublicKey  []byte    `json:"public_key"`  // 公钥
    PrivateKey []byte    `json:"private_key"` // 私钥（加密存储）
    Algorithm  AlgorithmType `json:"algorithm"` // 算法类型
    CreatedAt  time.Time `json:"created_at"`  // 创建时间
    ExpiresAt  time.Time `json:"expires_at"`  // 过期时间
}

// 加密配置
type EncryptionConfig struct {
    Algorithm     AlgorithmType `json:"algorithm"`      // 加密算法
    KeySize       int           `json:"key_size"`       // 密钥大小
    NonceSize     int           `json:"nonce_size"`     // Nonce大小
    TagSize       int           `json:"tag_size"`       // 认证标签大小
    UseHKDF       bool          `json:"use_hkdf"`       // 是否使用HKDF
    SaltSize      int           `json:"salt_size"`      // 盐值大小
}

// 加密消息
type EncryptedMessage struct {
    Ciphertext   []byte    `json:"ciphertext"`   // 密文
    Nonce        []byte    `json:"nonce"`        // Nonce
    Tag          []byte    `json:"tag"`          // 认证标签
    Algorithm    AlgorithmType `json:"algorithm"`    // 算法标识
    Version      string    `json:"version"`      // 协议版本
    Timestamp    time.Time `json:"timestamp"`    // 时间戳
}

// 签名消息
type SignedMessage struct {
    Message     []byte    `json:"message"`      // 原始消息
    Signature   []byte    `json:"signature"`    // 签名
    PublicKey   []byte    `json:"public_key"`   // 公钥（可选）
    Algorithm   AlgorithmType `json:"algorithm"`   // 算法标识
}

// 密钥交换结果
type KeyExchangeResult struct {
    SharedSecret []byte    `json:"shared_secret"` // 共享密钥
    PublicKey    []byte    `json:"public_key"`    // 临时公钥
    SessionID    string    `json:"session_id"`    // 会话ID
    ExpiresAt    time.Time `json:"expires_at"`    // 过期时间
}
```

### 主要接口
```go
// 加密管理器接口
type CryptoManager interface {
    // 生成密钥对
    GenerateKeyPair(algorithm AlgorithmType) (*KeyPair, error)
    
    // 加密数据
    Encrypt(plaintext []byte, key []byte, config *EncryptionConfig) (*EncryptedMessage, error)
    
    // 解密数据
    Decrypt(msg *EncryptedMessage, key []byte) ([]byte, error)
    
    // 签名数据
    Sign(message []byte, privateKey []byte) (*SignedMessage, error)
    
    // 验证签名
    Verify(signedMsg *SignedMessage, publicKey []byte) (bool, error)
    
    // 计算哈希
    Hash(data []byte, algorithm AlgorithmType) ([]byte, error)
    
    // 生成随机数
    RandomBytes(size int) ([]byte, error)
    
    // 密钥派生
    DeriveKey(secret []byte, salt []byte, info []byte, size int) ([]byte, error)
    
    // 清理敏感数据
    Cleanup() error
}

// 密钥交换接口
type KeyExchange interface {
    // 初始化密钥交换
    Initiate() (*KeyExchangeResult, error)
    
    // 响应密钥交换
    Respond(peerPublicKey []byte) (*KeyExchangeResult, error)
    
    // 完成密钥交换
    Finalize(peerPublicKey []byte, sessionID string) ([]byte, error)
    
    // 验证密钥交换
    VerifyExchange(sharedSecret []byte, proof []byte) (bool, error)
    
    // 轮换会话密钥
    RotateKey(sessionID string) ([]byte, error)
    
    // 销毁会话密钥
    DestroySession(sessionID string) error
}

// 密钥存储接口
type KeyStorage interface {
    // 存储密钥
    StoreKey(keyID string, key []byte, metadata map[string]string) error
    
    // 获取密钥
    GetKey(keyID string) ([]byte, map[string]string, error)
    
    // 删除密钥
    DeleteKey(keyID string) error
    
    // 列出所有密钥
    ListKeys() ([]string, error)
    
    // 密钥是否存在
    KeyExists(keyID string) (bool, error)
    
    // 清理过期密钥
    CleanupExpiredKeys() error
}

// 密码学工具接口
type CryptoUtil interface {
    // 常量时间比较
    ConstantTimeCompare(a, b []byte) bool
    
    // 安全内存清零
    SecureZeroMemory(data []byte)
    
    // 生成安全随机数
    SecureRandomInt(min, max int) (int, error)
    
    // 密码哈希（Argon2）
    HashPassword(password string) ([]byte, error)
    
    // 验证密码
    VerifyPassword(password string, hash []byte) (bool, error)
    
    // 生成密钥ID
    GenerateKeyID(publicKey []byte) string
}
```

## 实现要求

### 1. 算法选择
- **对称加密**：AES-GCM-256 或 ChaCha20-Poly1305
- **非对称加密**：X25519（密钥交换），Ed25519（签名）
- **哈希算法**：SHA256 或 BLAKE2b
- **密钥派生**：HKDF-SHA256

### 2. 前向安全
- 使用临时密钥交换（ECDHE）
- 定期轮换会话密钥
- 实现完美的前向保密

### 3. 密钥管理
- 安全存储私钥（加密存储）
- 实现密钥生命周期管理
- 支持密钥备份和恢复
- 定期密钥轮换

### 4. 随机数生成
- 使用密码学安全的随机数生成器
- 确保足够的熵源
- 防止随机数预测攻击

## 测试要求

### 单元测试
```bash
# 运行加密模块测试
go test ./crypto -v

# 测试特定功能
go test ./crypto -run TestEncryption
go test ./crypto -run TestSignature
go test ./crypto -run TestKeyExchange
```

### 安全测试
```bash
# 运行安全测试
go test ./crypto -tags=security

# 测试侧信道攻击防护
go test ./crypto -tags=sidechannel
```

### 测试用例
1. **加密解密测试**：测试对称加密功能
2. **签名验证测试**：测试数字签名功能
3. **密钥交换测试**：测试前向安全密钥交换
4. **随机数测试**：测试随机数质量
5. **边界测试**：测试各种边界条件

### 性能测试
```bash
# 基准测试
go test ./crypto -bench=.
go test ./crypto -bench=BenchmarkEncrypt
go test ./crypto -bench=BenchmarkSign
```

## 依赖关系
- identity模块：使用身份密钥对
- storage模块：用于密钥存储

## 使用示例

```go
// 创建加密管理器
manager := NewCryptoManager()

// 生成密钥对
keyPair, err := manager.GenerateKeyPair(AlgorithmEd25519)

// 加密数据
config := &EncryptionConfig{
    Algorithm: AlgorithmAESGCM,
    KeySize:   32, // 256位
    NonceSize: 12,
    TagSize:   16,
    UseHKDF:   true,
    SaltSize:  32,
}

plaintext := []byte("敏感数据")
encryptionKey := manager.RandomBytes(32)

encryptedMsg, err := manager.Encrypt(plaintext, encryptionKey, config)

// 解密数据
decrypted, err := manager.Decrypt(encryptedMsg, encryptionKey)

// 签名消息
message := []byte("重要消息")
signedMsg, err := manager.Sign(message, keyPair.PrivateKey)

// 验证签名
valid, err := manager.Verify(signedMsg, keyPair.PublicKey)

// 密钥交换
keyExchange := NewX25519KeyExchange()

// 发起方
initiatorResult, err := keyExchange.Initiate()

// 响应方
responderResult, err := keyExchange.Respond(initiatorResult.PublicKey)

// 双方计算共享密钥
sharedSecret1, err := keyExchange.Finalize(responderResult.PublicKey, initiatorResult.SessionID)
sharedSecret2, err := keyExchange.Finalize(initiatorResult.PublicKey, responderResult.SessionID)

// 派生会话密钥
sessionKey, err := manager.DeriveKey(
    sharedSecret1,
    []byte("session-key"),
    []byte("o2ochat-v1"),
    32,
)
```

## 密钥交换协议示例

```go
// 前向安全密钥交换协议
type ForwardSecrecyProtocol struct {
    cryptoManager CryptoManager
    keyExchange   KeyExchange
}

func (p *ForwardSecrecyProtocol) Handshake(peerPublicKey []byte) (*SessionKeys, error) {
    // 1. 生成临时密钥对
    ephemeralKeyPair, err := p.cryptoManager.GenerateKeyPair(AlgorithmX25519)
    
    // 2. 计算共享密钥
    sharedSecret, err := p.keyExchange.Respond(peerPublicKey)
    
    // 3. 派生会话密钥
    sessionKey := p.cryptoManager.DeriveKey(
        sharedSecret.SharedSecret,
        []byte("session-key"),
        []byte("o2ochat-handshake"),
        32,
    )
    
    // 4. 生成认证数据
    authData := append(ephemeralKeyPair.PublicKey, peerPublicKey...)
    signature, err := p.cryptoManager.Sign(authData, localPrivateKey)
    
    return &SessionKeys{
        EncryptionKey: sessionKey,
        AuthTag:       signature.Signature,
        SessionID:     sharedSecret.SessionID,
    }, nil
}
```

## 安全配置示例

```go
// 安全配置
securityConfig := &SecurityConfig{
    // 加密算法
    EncryptionAlgorithm: AlgorithmAESGCM,
    EncryptionKeySize:   32, // 256位
    
    // 签名算法
    SignatureAlgorithm: AlgorithmEd25519,
    
    // 密钥交换
    KeyExchangeAlgorithm: AlgorithmX25519,
    KeyRotationInterval:  24 * time.Hour,
    
    // 哈希算法
    HashAlgorithm: AlgorithmSHA256,
    
    // 随机数
    MinEntropyBits: 256,
    
    // 安全参数
    MinPasswordLength: 12,
    MaxFailedAttempts: 5,
    LockoutDuration:   15 * time.Minute,
}
```

## 错误处理
- 加密失败必须返回错误（不返回部分数据）
- 密钥操作失败必须清理内存
- 随机数生成失败必须中止操作
- 验证失败必须记录安全事件

## 安全最佳实践
1. **使用常量时间比较**：防止时序攻击
2. **安全内存清零**：防止内存泄露
3. **验证所有输入**：防止注入攻击
4. **定期密钥轮换**：减少密钥暴露风险
5. **安全审计日志**：记录所有安全相关操作

## 密码学库选择
- Go: crypto/包（标准库），golang.org/x/crypto
- Rust: ring, libsodium, openssl
- 避免使用不安全的算法（如RC4, MD5, SHA1）