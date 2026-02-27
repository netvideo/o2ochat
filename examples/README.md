# Examples Module - 示例代码模块

## 功能概述
提供各种使用示例和演示代码，帮助开发者理解和使用各个模块。

## 核心示例
1. **基础示例**：模块基本用法
2. **集成示例**：多模块协同工作
3. **高级示例**：复杂场景实现
4. **测试示例**：测试代码示例
5. **性能示例**：性能优化示例

## 示例分类

### 1. 身份管理示例
```go
// examples/identity/basic.go
package main

import (
    "fmt"
    "log"
    "github.com/netvideo/identity"
)

func main() {
    // 创建身份管理器
    manager := identity.NewIdentityManager()
    
    // 创建新身份
    config := &identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    }
    
    id, err := manager.CreateIdentity(config)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("创建身份成功:\n")
    fmt.Printf("  Peer ID: %s\n", id.PeerID)
    fmt.Printf("  公钥长度: %d bytes\n", len(id.PublicKey))
    fmt.Printf("  创建时间: %s\n", id.CreatedAt.Format("2006-01-02 15:04:05"))
    
    // 签名消息
    message := []byte("Hello, O2OChat!")
    signature, err := manager.SignMessage(message)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("\n消息签名成功:\n")
    fmt.Printf("  消息: %s\n", string(message))
    fmt.Printf("  签名长度: %d bytes\n", len(signature))
    
    // 验证签名
    valid := manager.VerifySignature(id.PeerID, message, signature)
    fmt.Printf("  签名验证: %v\n", valid)
    
    // 导出身份
    exportData, err := manager.ExportIdentity("secure-password")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("\n身份导出成功:\n")
    fmt.Printf("  导出数据长度: %d bytes\n", len(exportData))
}
```

### 2. 信令交换示例
```go
// examples/signaling/client_server.go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/netvideo/signaling"
)

func main() {
    // 启动信令服务器
    server := signaling.NewWebSocketServer()
    go func() {
        if err := server.Start(":8080"); err != nil {
            log.Fatal("服务器启动失败:", err)
        }
    }()
    
    time.Sleep(100 * time.Millisecond)
    
    // 客户端A
    clientA := signaling.NewWebSocketClient()
    if err := clientA.Connect("ws://localhost:8080"); err != nil {
        log.Fatal("客户端A连接失败:", err)
    }
    
    peerInfoA := &signaling.PeerInfo{
        PeerID:    "QmPeerA",
        IPv6Addrs: []string{"[2001:db8::1]:4242"},
        PublicKey: []byte("public-key-a"),
    }
    
    if err := clientA.Register(peerInfoA); err != nil {
        log.Fatal("客户端A注册失败:", err)
    }
    
    // 客户端B
    clientB := signaling.NewWebSocketClient()
    if err := clientB.Connect("ws://localhost:8080"); err != nil {
        log.Fatal("客户端B连接失败:", err)
    }
    
    peerInfoB := &signaling.PeerInfo{
        PeerID:    "QmPeerB",
        IPv6Addrs: []string{"[2001:db8::2]:4242"},
        PublicKey: []byte("public-key-b"),
    }
    
    if err := clientB.Register(peerInfoB); err != nil {
        log.Fatal("客户端B注册失败:", err)
    }
    
    // 客户端A发送连接请求
    offerMsg := &signaling.SignalingMessage{
        Type:      signaling.MessageTypeOffer,
        From:      "QmPeerA",
        To:        "QmPeerB",
        Data:      signaling.SDPInfo{Type: "offer", SDP: "v=0\r\no=..."},
        Timestamp: time.Now(),
    }
    
    if err := clientA.SendMessage(offerMsg); err != nil {
        log.Fatal("发送offer失败:", err)
    }
    
    fmt.Println("客户端A发送offer给客户端B")
    
    // 客户端B接收消息
    go func() {
        msg, err := clientB.ReceiveMessage()
        if err != nil {
            log.Fatal("接收消息失败:", err)
        }
        
        fmt.Printf("客户端B收到消息: %s from %s\n", msg.Type, msg.From)
        
        // 发送answer
        answerMsg := &signaling.SignalingMessage{
            Type:      signaling.MessageTypeAnswer,
            From:      "QmPeerB",
            To:        "QmPeerA",
            Data:      signaling.SDPInfo{Type: "answer", SDP: "v=0\r\no=..."},
            Timestamp: time.Now(),
        }
        
        if err := clientB.SendMessage(answerMsg); err != nil {
            log.Fatal("发送answer失败:", err)
        }
        
        fmt.Println("客户端B发送answer给客户端A")
    }()
    
    time.Sleep(2 * time.Second)
    
    // 清理
    clientA.Close()
    clientB.Close()
    server.Stop()
}
```

### 3. 文件传输示例
```go
// examples/filetransfer/basic.go
package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"
    "github.com/netvideo/filetransfer"
)

func main() {
    // 创建临时测试文件
    tempDir := "./test_data"
    os.MkdirAll(tempDir, 0755)
    
    testFile := filepath.Join(tempDir, "test_file.bin")
    createTestFile(testFile, 10*1024*1024) // 10MB测试文件
    
    // 创建文件传输管理器
    manager := filetransfer.NewFileTransferManager()
    
    // 分块文件
    metadata, err := manager.ChunkFile(testFile, 1024*1024) // 1MB块
    if err != nil {
        log.Fatal("分块失败:", err)
    }
    
    fmt.Printf("文件分块完成:\n")
    fmt.Printf("  文件ID: %s\n", metadata.FileID)
    fmt.Printf("  文件名: %s\n", metadata.FileName)
    fmt.Printf("  文件大小: %d bytes\n", metadata.FileSize)
    fmt.Printf("  总块数: %d\n", metadata.TotalChunks)
    fmt.Printf("  块大小: %d bytes\n", metadata.ChunkSize)
    fmt.Printf("  Merkle根哈希: %x\n", metadata.MerkleRoot[:16])
    
    // 创建下载任务（模拟）
    taskID, err := manager.CreateDownloadTask(
        metadata.FileID,
        filepath.Join(tempDir, "download"),
        []string{"QmPeer123", "QmPeer456"},
    )
    if err != nil {
        log.Fatal("创建下载任务失败:", err)
    }
    
    fmt.Printf("\n创建下载任务: %s\n", taskID)
    
    // 模拟下载进度
    go func() {
        for i := 0; i <= 100; i += 10 {
            time.Sleep(500 * time.Millisecond)
            
            // 更新进度
            task, _ := manager.GetTaskStatus(taskID)
            if task != nil {
                task.Progress.Completed = i
                task.Progress.BytesTransferred = int64(float64(metadata.FileSize) * float64(i) / 100.0)
                task.Progress.Speed = 1024.0 * float64(i) / 5.0 // 模拟速度
                
                fmt.Printf("下载进度: %d%%, 速度: %.2f KB/s\n", 
                    i, task.Progress.Speed/1024.0)
            }
        }
        
        // 完成下载
        task, _ := manager.GetTaskStatus(taskID)
        if task != nil {
            task.Status = filetransfer.StatusCompleted
            task.Progress.Completed = 100
            task.Progress.BytesTransferred = metadata.FileSize
            
            fmt.Println("\n下载完成!")
        }
    }()
    
    // 验证文件完整性
    time.Sleep(3 * time.Second)
    
    valid, err := manager.VerifyFile(metadata.FileID)
    if err != nil {
        log.Fatal("验证失败:", err)
    }
    
    fmt.Printf("文件完整性验证: %v\n", valid)
    
    // 清理
    os.RemoveAll(tempDir)
}

func createTestFile(path string, size int64) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // 写入测试数据
    data := make([]byte, size)
    for i := range data {
        data[i] = byte(i % 256)
    }
    
    _, err = file.Write(data)
    return err
}
```

### 4. 音视频通话示例
```go
// examples/media/call_demo.go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/netvideo/media"
)

func main() {
    // 创建媒体管理器
    manager := media.NewMediaManager()
    if err := manager.Initialize(); err != nil {
        log.Fatal("媒体管理器初始化失败:", err)
    }
    
    defer manager.Destroy()
    
    // 获取可用设备
    audioDevices, err := manager.GetDevices(media.MediaTypeAudio)
    if err != nil {
        log.Fatal("获取音频设备失败:", err)
    }
    
    videoDevices, err := manager.GetDevices(media.MediaTypeVideo)
    if err != nil {
        log.Fatal("获取视频设备失败:", err)
    }
    
    fmt.Println("可用音频设备:")
    for _, device := range audioDevices {
        fmt.Printf("  %s: %s (默认: %v)\n", 
            device.ID, device.Name, device.Default)
    }
    
    fmt.Println("\n可用视频设备:")
    for _, device := range videoDevices {
        fmt.Printf("  %s: %s (默认: %v)\n", 
            device.ID, device.Name, device.Default)
    }
    
    // 配置通话
    config := &media.CallConfig{
        AudioConfig: &media.MediaConfig{
            MediaType:  media.MediaTypeAudio,
            Enabled:    true,
            Codec:      "opus",
            Bitrate:    64000,
            SampleRate: 48000,
            Channels:   2,
        },
        VideoConfig: &media.MediaConfig{
            MediaType:  media.MediaTypeVideo,
            Enabled:    true,
            Codec:      "vp8",
            Bitrate:    500000,
            Width:      640,
            Height:     480,
            FrameRate:  30,
        },
        MaxBitrate:   1000000,
        MinBitrate:   100000,
        StartBitrate: 500000,
        UseFEC:       true,
        UseNACK:      true,
    }
    
    // 创建通话会话
    session, err := manager.CreateCallSession(config)
    if err != nil {
        log.Fatal("创建通话会话失败:", err)
    }
    
    fmt.Printf("\n创建通话会话: %s\n", session.GetSessionID())
    
    // 模拟通话过程
    fmt.Println("\n开始模拟通话...")
    
    // 启动通话
    if err := session.Start(); err != nil {
        log.Fatal("启动通话失败:", err)
    }
    
    // 模拟通话统计
    go func() {
        for i := 0; i < 10; i++ {
            time.Sleep(1 * time.Second)
            
            stats, err := manager.GetCallStats(session.GetSessionID())
            if err != nil {
                continue
            }
            
            fmt.Printf("通话统计 [%ds]:\n", i+1)
            fmt.Printf("  音频码率: %d bps\n", stats.AudioStats.Bitrate)
            fmt.Printf("  视频码率: %d bps\n", stats.VideoStats.Bitrate)
            fmt.Printf("  网络丢包: %.2f%%\n", stats.NetworkStats.PacketLoss*100)
            fmt.Printf("  通话质量: %.2f\n", stats.Quality)
        }
        
        // 结束通话
        session.Stop()
        fmt.Println("\n通话结束")
    }()
    
    // 等待通话结束
    time.Sleep(12 * time.Second)
    
    // 关闭会话
    session.Close()
}
```

### 5. 完整应用示例
```go
// examples/full_app/main.go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "github.com/netvideo/identity"
    "github.com/netvideo/signaling"
    "github.com/netvideo/transport"
    "github.com/netvideo/filetransfer"
    "github.com/netvideo/storage"
)

type O2OChatApp struct {
    identityManager    *identity.IdentityManager
    signalingClient    *signaling.SignalingClient
    transportManager   *transport.TransportManager
    fileTransferManager *filetransfer.FileTransferManager
    storageManager     *storage.StorageManager
    
    peerID            string
    configPath        string
}

func NewO2OChatApp(configPath string) (*O2OChatApp, error) {
    app := &O2OChatApp{
        configPath: configPath,
    }
    
    // 初始化存储
    storageConfig := &storage.StorageConfig{
        Type:        storage.StorageTypeSQLite,
        Path:        "./data",
        MaxSize:     10 * 1024 * 1024 * 1024, // 10GB
        Compression: true,
        Encryption:  true,
        CacheSize:   256,
    }
    
    app.storageManager = storage.NewStorageManager()
    if err := app.storageManager.Initialize(storageConfig); err != nil {
        return nil, fmt.Errorf("存储初始化失败: %v", err)
    }
    
    // 初始化身份管理
    app.identityManager = identity.NewIdentityManager()
    
    // 加载或创建身份
    identities, _ := app.storageManager.List("identity:")
    if len(identities) > 0 {
        // 加载现有身份
        identityData, err := app.storageManager.Get(identities[0])
        if err == nil {
            id, err := app.identityManager.ImportIdentity(identityData, "")
            if err == nil {
                app.peerID = id.PeerID
            }
        }
    }
    
    if app.peerID == "" {
        // 创建新身份
        config := &identity.IdentityConfig{
            KeyType:   "ed25519",
            KeyLength: 256,
        }
        
        id, err := app.identityManager.CreateIdentity(config)
        if err != nil {
            return nil, fmt.Errorf("创建身份失败: %v", err)
        }
        
        app.peerID = id.PeerID
        
        // 保存身份
        exportData, err := app.identityManager.ExportIdentity("")
        if err == nil {
            app.storageManager.Put("identity:"+app.peerID, exportData, nil)
        }
    }
    
    fmt.Printf("身份初始化完成: %s\n", app.peerID)
    
    // 初始化信令客户端
    app.signalingClient = signaling.NewWebSocketClient()
    
    // 初始化传输管理器
    app.transportManager = transport.NewTransportManager()
    
    // 初始化文件传输管理器
    app.fileTransferManager = filetransfer.NewFileTransferManager()
    
    return app, nil
}

func (app *O2OChatApp) Start() error {
    fmt.Println("启动O2OChat应用程序...")
    
    // 连接到信令服务器
    if err := app.signalingClient.Connect("ws://signaling.o2ochat.example.com:8080"); err != nil {
        return fmt.Errorf("连接信令服务器失败: %v", err)
    }
    
    // 注册在线状态
    peerInfo := &signaling.PeerInfo{
        PeerID:    app.peerID,
        PublicKey: app.identityManager.GetPublicKey(app.peerID),
    }
    
    if err := app.signalingClient.Register(peerInfo); err != nil {
        return fmt.Errorf("注册失败: %v", err)
    }
    
    fmt.Println("已连接到信令服务器")
    
    // 开始监听连接
    if err := app.transportManager.Listen("[::]:4242"); err != nil {
        return fmt.Errorf("监听端口失败: %v", err)
    }
    
    fmt.Println("开始监听连接...")
    
    // 启动消息处理循环
    go app.handleSignalingMessages()
    
    // 启动连接接受循环
    go app.acceptConnections()
    
    return nil
}

func (app *O2OChatApp) handleSignalingMessages() {
    for {
        msg, err := app.signalingClient.ReceiveMessage()
        if err != nil {
            log.Printf("接收信令消息失败: %v", err)
            continue
        }
        
        switch msg.Type {
        case signaling.MessageTypeOffer:
            app.handleOffer(msg)
        case signaling.MessageTypeAnswer:
            app.handleAnswer(msg)
        case signaling.MessageTypeCandidate:
            app.handleCandidate(msg)
        case signaling.MessageTypeInvite:
            app.handleInvite(msg)
        }
    }
}

func (app *O2OChatApp) acceptConnections() {
    for {
        conn, err := app.transportManager.Accept()
        if err != nil {
            log.Printf("接受连接失败: %v", err)
            continue
        }
        
        go app.handleConnection(conn)
    }
}

func (app *O2OChatApp) handleConnection(conn transport.Connection) {
    peerID := conn.GetInfo().PeerID
    fmt.Printf("新连接: %s (%s)\n", peerID, conn.GetInfo().Type)
    
    // 处理连接...
    
    conn.Close()
}

func (app *O2OChatApp) handleOffer(msg *signaling.SignalingMessage) {
    fmt.Printf("收到连接请求 from %s\n", msg.From)
    
    // 处理offer并发送answer...
}

func (app *O2OChatApp) handleAnswer(msg *signaling.SignalingMessage) {
    fmt.Printf("收到连接应答 from %s\n", msg.From)
    
    // 处理answer...
}

func (app *O2OChatApp) handleCandidate(msg *signaling.SignalingMessage) {
    // 处理ICE候选...
}

func (app *O2OChatApp) handleInvite(msg *signaling.SignalingMessage) {
    fmt.Printf("收到邀请 from %s\n", msg.From)
    
    // 处理邀请...
}

func (app *O2OChatApp) Stop() {
    fmt.Println("停止O2OChat应用程序...")
    
    if app.signalingClient != nil {
        app.signalingClient.Close()
    }
    
    if app.transportManager != nil {
        app.transportManager.Close()
    }
    
    if app.storageManager != nil {
        app.storageManager.Close()
    }
    
    fmt.Println("应用程序已停止")
}

func main() {
    // 创建应用程序
    app, err := NewO2OChatApp("./config.yaml")
    if err != nil {
        log.Fatal("创建应用程序失败:", err)
    }
    
    // 启动应用程序
    if err := app.Start(); err != nil {
        log.Fatal("启动应用程序失败:", err)
    }
    
    // 等待中断信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    fmt.Println("\nO2OChat 正在运行...")
    fmt.Println("按 Ctrl+C 停止")
    
    <-sigChan
    
    // 停止应用程序
    app.Stop()
}
```

## 运行示例

```bash
# 运行基础示例
go run examples/identity/basic.go

# 运行信令示例
go run examples/signaling/client_server.go

# 运行文件传输示例
go run examples/filetransfer/basic.go

# 运行完整应用示例
go run examples/full_app/main.go
```

## 测试示例

```go
// examples/tests/integration_test.go
package tests

import (
    "testing"
    "github.com/netvideo/identity"
    "github.com/netvideo/signaling"
)

func TestIdentityAndSignalingIntegration(t *testing.T) {
    // 创建身份
    identityManager := identity.NewIdentityManager()
    id, err := identityManager.CreateIdentity(&identity.IdentityConfig{
        KeyType:   "ed25519",
        KeyLength: 256,
    })
    if err != nil {
        t.Fatalf("创建身份失败: %v", err)
    }
    
    // 创建信令消息
    message := &signaling.SignalingMessage{
        Type:      signaling.MessageTypeOffer,
        From:      id.PeerID,
        To:        "QmPeerTest",
        Timestamp: time.Now(),
    }
    
    // 签名消息
    signature, err := identityManager.SignMessage([]byte(message.String()))
    if err != nil {
        t.Fatalf("签名失败: %v", err)
    }
    
    message.Signature = signature
    
    // 验证签名
    valid := identityManager.VerifySignature(id.PeerID, []byte(message.String()), signature)
    if !valid {
        t.Fatal("签名验证失败")
    }
    
    t.Logf("集成测试通过: PeerID=%s", id.PeerID)
}
```

## 性能示例

```go
// examples/performance/benchmark.go
package main

import (
    "fmt"
    "time"
    "github.com/netvideo/crypto"
)

func benchmarkEncryption() {
    manager := crypto.NewCryptoManager()
    
    config := &crypto.EncryptionConfig{
        Algorithm: crypto.AlgorithmAESGCM,
        KeySize:   32,
        NonceSize: 12,
        TagSize:   16,
    }
    
    key := manager.RandomBytes(32)
    data := make([]byte, 1024*1024) // 1MB数据
    
    // 基准测试
    start := time.Now()
    iterations := 100
    
    for i := 0; i < iterations; i++ {
        encrypted, err := manager.Encrypt(data, key, config)
        if err != nil {
            fmt.Printf("加密失败: %v\n", err)
            return
        }
        
        _, err = manager.Decrypt(encrypted, key)
        if err != nil {
            fmt.Printf("解密失败: %v\n", err)
            return
        }
    }
    
    elapsed := time.Since(start)
    throughput := float64(len(data)*iterations*2) / elapsed.Seconds() / 1024 / 1024
    
    fmt.Printf("加密解密性能测试:\n")
    fmt.Printf("  数据大小: %d MB\n", len(data)/1024/1024)
    fmt.Printf("  迭代次数: %d\n", iterations)
    fmt.Printf("  总时间: %v\n", elapsed)
    fmt.Printf("  吞吐量: %.2f MB/s\n", throughput)
}

func main() {
    fmt.Println("性能测试示例")
    fmt.Println("=============")
    
    benchmarkEncryption()
}
```

## 使用说明

1. **环境准备**：
   ```bash
   # 安装Go
   go version
   
   # 下载依赖
   go mod download
   
   # 构建示例
   go build ./examples/...
   ```

2. **运行顺序**：
   - 先运行基础示例了解模块功能
   - 再运行集成示例理解模块协作
   - 最后运行完整应用示例

3. **调试技巧**：
   - 使用`-v`参数查看详细输出
   - 设置环境变量控制日志级别
   - 使用`pprof`进行性能分析

4. **扩展示例**：
   - 根据需要修改示例代码
   - 添加新的使用场景示例
   - 贡献测试用例和性能示例