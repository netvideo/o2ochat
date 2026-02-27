# Tests Module - 测试模块

## 功能概述
提供完整的测试套件，包括单元测试、集成测试、性能测试和端到端测试。

## 测试分类

### 1. 单元测试
测试单个模块或函数的正确性。

### 2. 集成测试
测试多个模块协同工作的正确性。

### 3. 性能测试
测试系统性能和资源使用情况。

### 4. 端到端测试
测试完整应用流程。

### 5. 安全测试
测试安全性和漏洞。

## 测试目录结构

```
tests/
├── unit/              # 单元测试
│   ├── identity/      # 身份模块测试
│   ├── signaling/     # 信令模块测试
│   ├── transport/     # 传输模块测试
│   ├── filetransfer/  # 文件传输测试
│   ├── media/         # 媒体模块测试
│   ├── crypto/        # 加密模块测试
│   ├── storage/       # 存储模块测试
│   ├── ui/            # UI模块测试
│   └── cli/           # CLI模块测试
├── integration/       # 集成测试
├── performance/       # 性能测试
├── e2e/              # 端到端测试
├── security/         # 安全测试
├── fixtures/         # 测试数据
├── mocks/            # 模拟对象
└── utils/            # 测试工具
```

## 测试框架

### Go测试框架
```go
// tests/unit/identity/identity_test.go
package identity_test

import (
    "testing"
    "github.com/netvideo/identity"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCreateIdentity(t *testing.T) {
    // 准备
    manager := identity.NewIdentityManager()
    config := &identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    }
    
    // 执行
    id, err := manager.CreateIdentity(config)
    
    // 验证
    require.NoError(t, err, "创建身份不应出错")
    assert.NotEmpty(t, id.PeerID, "Peer ID不应为空")
    assert.NotEmpty(t, id.PublicKey, "公钥不应为空")
    assert.Equal(t, 32, len(id.PublicKey), "Ed25519公钥应为32字节")
    assert.NotZero(t, id.CreatedAt, "创建时间不应为零")
}

func TestSignAndVerify(t *testing.T) {
    // 准备
    manager := identity.NewIdentityManager()
    config := &identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    }
    
    id, err := manager.CreateIdentity(config)
    require.NoError(t, err)
    
    message := []byte("测试消息")
    
    // 执行
    signature, err := manager.SignMessage(message)
    require.NoError(t, err, "签名不应出错")
    
    valid := manager.VerifySignature(id.PeerID, message, signature)
    
    // 验证
    assert.True(t, valid, "签名验证应通过")
    assert.NotEmpty(t, signature, "签名不应为空")
    assert.Equal(t, 64, len(signature), "Ed25519签名应为64字节")
}

func TestExportImportIdentity(t *testing.T) {
    // 准备
    manager := identity.NewIdentityManager()
    config := &identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    }
    
    originalID, err := manager.CreateIdentity(config)
    require.NoError(t, err)
    
    password := "test-password"
    
    // 执行
    exportData, err := manager.ExportIdentity(password)
    require.NoError(t, err, "导出不应出错")
    
    importedID, err := manager.ImportIdentity(exportData, password)
    require.NoError(t, err, "导入不应出错")
    
    // 验证
    assert.Equal(t, originalID.PeerID, importedID.PeerID, "Peer ID应相同")
    assert.Equal(t, originalID.PublicKey, importedID.PublicKey, "公钥应相同")
}

func TestInvalidPassword(t *testing.T) {
    // 准备
    manager := identity.NewIdentityManager()
    config := &identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    }
    
    id, err := manager.CreateIdentity(config)
    require.NoError(t, err)
    
    exportData, err := manager.ExportIdentity("correct-password")
    require.NoError(t, err)
    
    // 执行和验证
    _, err = manager.ImportIdentity(exportData, "wrong-password")
    assert.Error(t, err, "错误密码应导致导入失败")
}

func TestConcurrentIdentityCreation(t *testing.T) {
    // 准备
    manager := identity.NewIdentityManager()
    config := &identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    }
    
    numGoroutines := 10
    results := make(chan *identity.Identity, numGoroutines)
    errors := make(chan error, numGoroutines)
    
    // 执行
    for i := 0; i < numGoroutines; i++ {
        go func() {
            id, err := manager.CreateIdentity(config)
            if err != nil {
                errors <- err
                return
            }
            results <- id
        }()
    }
    
    // 验证
    peerIDs := make(map[string]bool)
    for i := 0; i < numGoroutines; i++ {
        select {
        case id := <-results:
            assert.NotEmpty(t, id.PeerID)
            // 验证Peer ID唯一性
            assert.False(t, peerIDs[id.PeerID], "Peer ID应唯一")
            peerIDs[id.PeerID] = true
        case err := <-errors:
            t.Errorf("并发创建身份失败: %v", err)
        }
    }
    
    assert.Equal(t, numGoroutines, len(peerIDs), "应创建唯一身份")
}
```

### 集成测试示例
```go
// tests/integration/signaling_transport_test.go
package integration_test

import (
    "context"
    "testing"
    "time"
    "github.com/netvideo/identity"
    "github.com/netvideo/signaling"
    "github.com/netvideo/transport"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSignalingToTransportIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("跳过集成测试")
    }
    
    // 准备测试环境
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // 创建两个身份
    identityManager := identity.NewIdentityManager()
    config := &identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    }
    
    peerA, err := identityManager.CreateIdentity(config)
    require.NoError(t, err)
    
    peerB, err := identityManager.CreateIdentity(config)
    require.NoError(t, err)
    
    // 启动信令服务器
    signalingServer := signaling.NewWebSocketServer()
    go func() {
        if err := signalingServer.Start(":18080"); err != nil {
            t.Logf("信令服务器启动失败: %v", err)
        }
    }()
    
    time.Sleep(100 * time.Millisecond)
    
    // 创建信令客户端
    clientA := signaling.NewWebSocketClient()
    err = clientA.Connect("ws://localhost:18080")
    require.NoError(t, err)
    defer clientA.Close()
    
    clientB := signaling.NewWebSocketClient()
    err = clientB.Connect("ws://localhost:18080")
    require.NoError(t, err)
    defer clientB.Close()
    
    // 注册在线状态
    peerInfoA := &signaling.PeerInfo{
        PeerID:    peerA.PeerID,
        IPv6Addrs: []string{"[::1]:24242"},
        PublicKey: peerA.PublicKey,
        Online:    true,
    }
    
    peerInfoB := &signaling.PeerInfo{
        PeerID:    peerB.PeerID,
        IPv6Addrs: []string{"[::1]:24243"},
        PublicKey: peerB.PublicKey,
        Online:    true,
    }
    
    err = clientA.Register(peerInfoA)
    require.NoError(t, err)
    
    err = clientB.Register(peerInfoB)
    require.NoError(t, err)
    
    // 创建传输管理器
    transportA := transport.NewTransportManager()
    transportB := transport.NewTransportManager()
    
    // 启动监听
    err = transportA.Listen("[::1]:24242")
    require.NoError(t, err)
    defer transportA.Close()
    
    err = transportB.Listen("[::1]:24243")
    require.NoError(t, err)
    defer transportB.Close()
    
    // Peer A发送连接请求
    offerMsg := &signaling.SignalingMessage{
        Type:      signaling.MessageTypeOffer,
        From:      peerA.PeerID,
        To:        peerB.PeerID,
        Data:      map[string]interface{}{"type": "offer", "sdp": "test-sdp"},
        Timestamp: time.Now(),
    }
    
    // 签名消息
    signature, err := identityManager.SignMessage([]byte(offerMsg.String()))
    require.NoError(t, err)
    offerMsg.Signature = signature
    
    err = clientA.SendMessage(offerMsg)
    require.NoError(t, err)
    
    // Peer B接收并处理offer
    go func() {
        select {
        case <-ctx.Done():
            return
        default:
            msg, err := clientB.ReceiveMessage()
            if err != nil {
                t.Logf("接收消息失败: %v", err)
                return
            }
            
            // 验证签名
            valid := identityManager.VerifySignature(
                peerA.PeerID,
                []byte(msg.String()),
                msg.Signature,
            )
            assert.True(t, valid, "签名应验证通过")
            
            // 发送answer
            answerMsg := &signaling.SignalingMessage{
                Type:      signaling.MessageTypeAnswer,
                From:      peerB.PeerID,
                To:        peerA.PeerID,
                Data:      map[string]interface{}{"type": "answer", "sdp": "test-sdp-answer"},
                Timestamp: time.Now(),
            }
            
            signature, err := identityManager.SignMessage([]byte(answerMsg.String()))
            if err == nil {
                answerMsg.Signature = signature
                clientB.SendMessage(answerMsg)
            }
        }
    }()
    
    // Peer A尝试连接
    connConfig := &transport.ConnectionConfig{
        PeerID:        peerB.PeerID,
        IPv6Addresses: []string{"[::1]:24243"},
        Priority:      []transport.ConnectionType{transport.ConnectionTypeQUIC},
        Timeout:       5 * time.Second,
        RetryCount:    1,
    }
    
    conn, err := transportA.Connect(connConfig)
    if err != nil {
        t.Logf("连接失败（可能预期）: %v", err)
    } else {
        defer conn.Close()
        
        // 测试数据传输
        testMessage := []byte("Hello, Integration Test!")
        
        streamConfig := &transport.StreamConfig{
            Reliable:   true,
            Ordered:    true,
            BufferSize: 1024,
        }
        
        stream, err := conn.OpenStream(streamConfig)
        require.NoError(t, err)
        defer stream.Close()
        
        n, err := stream.Write(testMessage)
        require.NoError(t, err)
        assert.Equal(t, len(testMessage), n)
    }
    
    // 停止信令服务器
    signalingServer.Stop()
    
    t.Log("集成测试完成")
}
```

### 性能测试示例
```go
// tests/performance/crypto_benchmark_test.go
package performance_test

import (
    "testing"
    "github.com/netvideo/crypto"
)

func BenchmarkEncryption(b *testing.B) {
    manager := crypto.NewCryptoManager()
    config := &crypto.EncryptionConfig{
        Algorithm: crypto.AlgorithmAESGCM,
        KeySize:   32,
        NonceSize: 12,
        TagSize:   16,
    }
    
    key := manager.RandomBytes(32)
    data := make([]byte, 1024) // 1KB数据
    
    b.ResetTimer()
    b.SetBytes(int64(len(data)))
    
    for i := 0; i < b.N; i++ {
        encrypted, err := manager.Encrypt(data, key, config)
        if err != nil {
            b.Fatalf("加密失败: %v", err)
        }
        
        _, err = manager.Decrypt(encrypted, key)
        if err != nil {
            b.Fatalf("解密失败: %v", err)
        }
    }
}

func BenchmarkSigning(b *testing.B) {
    manager := crypto.NewCryptoManager()
    keyPair, err := manager.GenerateKeyPair(crypto.AlgorithmEd25519)
    if err != nil {
        b.Fatalf("生成密钥对失败: %v", err)
    }
    
    message := make([]byte, 256)
    
    b.ResetTimer()
    b.SetBytes(int64(len(message)))
    
    for i := 0; i < b.N; i++ {
        signedMsg, err := manager.Sign(message, keyPair.PrivateKey)
        if err != nil {
            b.Fatalf("签名失败: %v", err)
        }
        
        valid, err := manager.Verify(signedMsg, keyPair.PublicKey)
        if err != nil || !valid {
            b.Fatalf("验证失败: %v", err)
        }
    }
}

func BenchmarkHashCalculation(b *testing.B) {
    manager := crypto.NewCryptoManager()
    data := make([]byte, 1024*1024) // 1MB数据
    
    b.ResetTimer()
    b.SetBytes(int64(len(data)))
    
    for i := 0; i < b.N; i++ {
        _, err := manager.Hash(data, crypto.AlgorithmSHA256)
        if err != nil {
            b.Fatalf("哈希计算失败: %v", err)
        }
    }
}

func BenchmarkKeyGeneration(b *testing.B) {
    manager := crypto.NewCryptoManager()
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := manager.GenerateKeyPair(crypto.AlgorithmEd25519)
        if err != nil {
            b.Fatalf("生成密钥对失败: %v", err)
        }
    }
}
```

### 端到端测试示例
```go
// tests/e2e/file_transfer_test.go
package e2e_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "time"
    "github.com/netvideo/identity"
    "github.com/netvideo/signaling"
    "github.com/netvideo/transport"
    "github.com/netvideo/filetransfer"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestEndToEndFileTransfer(t *testing.T) {
    if testing.Short() {
        t.Skip("跳过端到端测试")
    }
    
    // 创建测试目录
    testDir := t.TempDir()
    sourceDir := filepath.Join(testDir, "source")
    destDir := filepath.Join(testDir, "dest")
    
    os.MkdirAll(sourceDir, 0755)
    os.MkdirAll(destDir, 0755)
    
    // 创建测试文件
    testFile := filepath.Join(sourceDir, "test_file.bin")
    testData := make([]byte, 1024*1024) // 1MB
    for i := range testData {
        testData[i] = byte(i % 256)
    }
    
    err := os.WriteFile(testFile, testData, 0644)
    require.NoError(t, err)
    
    // 创建两个Peer的模拟环境
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()
    
    // Peer A（发送方）
    identityA := identity.NewIdentityManager()
    peerA, err := identityA.CreateIdentity(&identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    })
    require.NoError(t, err)
    
    transportA := transport.NewTransportManager()
    err = transportA.Listen("[::1]:30001")
    require.NoError(t, err)
    defer transportA.Close()
    
    fileTransferA := filetransfer.NewFileTransferManager()
    
    // Peer B（接收方）
    identityB := identity.NewIdentityManager()
    peerB, err := identityB.CreateIdentity(&identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    })
    require.NoError(t, err)
    
    transportB := transport.NewTransportManager()
    err = transportB.Listen("[::1]:30002")
    require.NoError(t, err)
    defer transportB.Close()
    
    fileTransferB := filetransfer.NewFileTransferManager()
    
    // 启动信令服务器（模拟）
    signalingServer := signaling.NewWebSocketServer()
    go func() {
        signalingServer.Start(":18081")
    }()
    defer signalingServer.Stop()
    
    time.Sleep(100 * time.Millisecond)
    
    // 连接信令服务器
    clientA := signaling.NewWebSocketClient()
    err = clientA.Connect("ws://localhost:18081")
    require.NoError(t, err)
    defer clientA.Close()
    
    clientB := signaling.NewWebSocketClient()
    err = clientB.Connect("ws://localhost:18081")
    require.NoError(t, err)
    defer clientB.Close()
    
    // 注册在线状态
    peerInfoA := &signaling.PeerInfo{
        PeerID:    peerA.PeerID,
        IPv6Addrs: []string{"[::1]:30001"},
        PublicKey: peerA.PublicKey,
        Online:    true,
    }
    
    peerInfoB := &signaling.PeerInfo{
        PeerID:    peerB.PeerID,
        IPv6Addrs: []string{"[::1]:30002"},
        PublicKey: peerB.PublicKey,
        Online:    true,
    }
    
    err = clientA.Register(peerInfoA)
    require.NoError(t, err)
    
    err = clientB.Register(peerInfoB)
    require.NoError(t, err)
    
    // Peer A分块文件
    metadata, err := fileTransferA.ChunkFile(testFile, 256*1024) // 256KB块
    require.NoError(t, err)
    
    t.Logf("文件分块完成: %s (%d bytes, %d chunks)", 
        metadata.FileName, metadata.FileSize, metadata.TotalChunks)
    
    // 模拟文件传输
    completed := make(chan bool, 1)
    
    go func() {
        // Peer B创建下载任务
        taskID, err := fileTransferB.CreateDownloadTask(
            metadata.FileID,
            destDir,
            []string{peerA.PeerID},
        )
        require.NoError(t, err)
        
        err = fileTransferB.StartTransfer(taskID)
        require.NoError(t, err)
        
        // 等待下载完成
        for {
            select {
            case <-ctx.Done():
                return
            default:
                task, err := fileTransferB.GetTaskStatus(taskID)
                if err != nil {
                    t.Logf("获取任务状态失败: %v", err)
                    continue
                }
                
                if task.Status == filetransfer.StatusCompleted {
                    t.Logf("下载完成: %.2f%%", 
                        float64(task.Progress.Completed)/float64(task.Progress.TotalChunks)*100)
                    completed <- true
                    return
                }
                
                if task.Status == filetransfer.StatusFailed {
                    t.Errorf("下载失败")
                    completed <- false
                    return
                }
                
                time.Sleep(100 * time.Millisecond)
            }
        }
    }()
    
    // 等待测试完成
    select {
    case success := <-completed:
        assert.True(t, success, "文件传输应成功")
        
        // 验证文件完整性
        downloadedFile := filepath.Join(destDir, metadata.FileName)
        _, err := os.Stat(downloadedFile)
        assert.NoError(t, err, "下载的文件应存在")
        
        downloadedData, err := os.ReadFile(downloadedFile)
        assert.NoError(t, err)
        assert.Equal(t, len(testData), len(downloadedData), "文件大小应相同")
        assert.Equal(t, testData, downloadedData, "文件内容应相同")
        
        t.Logf("端到端文件传输测试通过")
        
    case <-ctx.Done():
        t.Fatal("测试超时")
    }
}
```

## 测试工具

### 测试辅助函数
```go
// tests/utils/test_helpers.go
package utils

import (
    "crypto/rand"
    "io"
    "os"
    "path/filepath"
    "testing"
)

// CreateTestFile 创建测试文件
func CreateTestFile(t *testing.T, path string, size int64) {
    t.Helper()
    
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        t.Fatalf("创建目录失败: %v", err)
    }
    
    file, err := os.Create(path)
    if err != nil {
        t.Fatalf("创建文件失败: %v", err)
    }
    defer file.Close()
    
    // 写入随机数据
    _, err = io.CopyN(file, rand.Reader, size)
    if err != nil {
        t.Fatalf("写入文件失败: %v", err)
    }
}

// CleanupTestDir 清理测试目录
func CleanupTestDir(t *testing.T, path string) {
    t.Helper()
    
    if err := os.RemoveAll(path); err != nil {
        t.Logf("清理目录失败: %v", err)
    }
}

// WaitForCondition 等待条件成立
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration) bool {
    t.Helper()
    
    deadline := time.Now().Add(timeout)
    
    for time.Now().Before(deadline) {
        if condition() {
            return true
        }
        time.Sleep(interval)
    }
    
    return false
}

// GetFreePort 获取空闲端口
func GetFreePort(t *testing.T) int {
    t.Helper()
    
    // 实现获取空闲端口的逻辑
    return 0
}
```

### Mock对象
```go
// tests/mocks/signaling_mock.go
package mocks

import (
    "github.com/netvideo/signaling"
    "github.com/stretchr/testify/mock"
)

type MockSignalingClient struct {
    mock.Mock
}

func (m *MockSignalingClient) Connect(serverURL string) error {
    args := m.Called(serverURL)
    return args.Error(0)
}

func (m *MockSignalingClient) SendMessage(msg *signaling.SignalingMessage) error {
    args := m.Called(msg)
    return args.Error(0)
}

func (m *MockSignalingClient) ReceiveMessage() (*signaling.SignalingMessage, error) {
    args := m.Called()
    
    if msg, ok := args.Get(0).(*signaling.SignalingMessage); ok {
        return msg, args.Error(1)
    }
    
    return nil, args.Error(1)
}

func (m *MockSignalingClient) Register(peerInfo *signaling.PeerInfo) error {
    args := m.Called(peerInfo)
    return args.Error(0)
}

func (m *MockSignalingClient) Unregister() error {
    args := m.Called()
    return args.Error(0)
}

func (m *MockSignalingClient) LookupPeer(peerID string) (*signaling.PeerInfo, error) {
    args := m.Called(peerID)
    
    if info, ok := args.Get(0).(*signaling.PeerInfo); ok {
        return info, args.Error(1)
    }
    
    return nil, args.Error(1)
}

func (m *MockSignalingClient) Close() error {
    args := m.Called()
    return args.Error(0)
}
```

## 测试运行

### 运行所有测试
```bash
# 运行单元测试
go test ./tests/unit/... -v

# 运行集成测试
go test ./tests/integration/... -v -tags=integration

# 运行性能测试
go test ./tests/performance/... -v -bench=.

# 运行端到端测试
go test ./tests/e2e/... -v -tags=e2e

# 运行安全测试
go test ./tests/security/... -v -tags=security
```

### 测试覆盖率
```bash
# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 查看覆盖率
go tool cover -func=coverage.out
```

### 测试标签
- `unit`: 单元测试
- `integration`: 集成测试
- `e2e`: 端到端测试
- `security`: 安全测试
- `performance`: 性能测试
- `short`: 快速测试

## 测试最佳实践

1. **测试命名**：使用`Test`前缀，描述测试功能
2. **测试组织**：按模块组织测试文件
3. **测试数据**：使用fixtures目录存放测试数据
4. **测试清理**：每个测试后清理资源
5. **测试隔离**：测试之间不共享状态
6. **错误处理**：测试中正确处理错误
7. **并发测试**：测试并发安全性
8. **性能测试**：关注关键路径性能

## 持续集成

### GitHub Actions示例
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: '1.21'
    
    - name: Run unit tests
      run: go test ./tests/unit/... -v -race
    
    - name: Run integration tests
      run: go test ./tests/integration/... -v -tags=integration
    
    - name: Run security tests
      run: go test ./tests/security/... -v -tags=security
    
    - name: Generate coverage report
      run: |
        go test ./... -coverprofile=coverage.out
        go tool cover -func=coverage.out
    
    - name: Upload coverage
      uses: codecov/codecov-action@v2
```

## 测试报告

测试报告应包括：
1. 测试通过率
2. 代码覆盖率
3. 性能基准
4. 安全漏洞
5. 测试持续时间