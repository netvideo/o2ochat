# Signaling Module - 信令交换模块

## 功能概述
负责P2P连接建立前的信令交换，包括连接请求、SDP/ICE候选交换、在线状态管理等。

## 核心功能
1. **信令服务器**：中心化信令交换（初期实现）
2. **DHT信令**：去中心化信令交换（可选扩展）
3. **连接协商**：交换SDP和ICE候选信息
4. **在线状态管理**：维护用户在线状态
5. **消息路由**：转发信令消息给目标用户
6. **离线消息**：存储和转发离线消息

## 接口定义

### 类型定义
```go
// 信令消息类型
type MessageType string

const (
    MessageTypeOffer     MessageType = "offer"
    MessageTypeAnswer    MessageType = "answer"
    MessageTypeCandidate MessageType = "candidate"
    MessageTypeInvite    MessageType = "invite"
    MessageTypeBye       MessageType = "bye"
    MessageTypePing      MessageType = "ping"
    MessageTypePong      MessageType = "pong"
)

// 信令消息
type SignalingMessage struct {
    Type      MessageType           // 消息类型
    From      string                // 发送方Peer ID
    To        string                // 接收方Peer ID
    Data      interface{}           // 消息数据（SDP/ICE候选等）
    Timestamp time.Time             // 时间戳
    Signature []byte                // 发送方签名
    Nonce     string                // 防重放随机数
}

// SDP信息
type SDPInfo struct {
    Type string `json:"type"`  // "offer" 或 "answer"
    SDP  string `json:"sdp"`   // SDP字符串
}

// ICE候选
type ICECandidate struct {
    Candidate     string `json:"candidate"`
    SDPMid        string `json:"sdpMid"`
    SDPMLineIndex int    `json:"sdpMLineIndex"`
}

// 用户在线信息
type PeerInfo struct {
    PeerID    string   `json:"peer_id"`
    IPv6Addrs []string `json:"ipv6_addrs,omitempty"`
    IPv4Addrs []string `json:"ipv4_addrs,omitempty"`
    PublicKey []byte   `json:"public_key"`
    LastSeen  time.Time `json:"last_seen"`
    Online    bool     `json:"online"`
}
```

### 主要接口
```go
// 信令客户端接口
type SignalingClient interface {
    // 连接到信令服务器
    Connect(serverURL string) error
    
    // 发送信令消息
    SendMessage(msg *SignalingMessage) error
    
    // 接收信令消息
    ReceiveMessage() (*SignalingMessage, error)
    
    // 注册在线状态
    Register(peerInfo *PeerInfo) error
    
    // 注销
    Unregister() error
    
    // 查询用户信息
    LookupPeer(peerID string) (*PeerInfo, error)
    
    // 关闭连接
    Close() error
}

// 信令服务器接口
type SignalingServer interface {
    // 启动服务器
    Start(addr string) error
    
    // 停止服务器
    Stop() error
    
    // 处理客户端连接
    HandleConnection(conn net.Conn)
    
    // 广播消息
    Broadcast(msg *SignalingMessage) error
    
    // 获取在线用户列表
    GetOnlinePeers() ([]*PeerInfo, error)
}

// DHT信令接口（可选）
type DHTSignaling interface {
    // 加入DHT网络
    Join(bootstrapNodes []string) error
    
    // 发布在线信息
    Publish(peerInfo *PeerInfo) error
    
    // 查找用户信息
    FindPeer(peerID string) (*PeerInfo, error)
    
    // 离开网络
    Leave() error
}
```

## 实现要求

### 1. 信令协议
- 使用WebSocket协议进行通信
- 消息格式：JSON + 签名
- 支持消息压缩（可选）

### 2. 安全性
- 所有消息必须签名验证
- 实现防重放攻击机制
- 支持消息加密（可选）

### 3. 可靠性
- 实现心跳机制保持连接
- 支持断线重连
- 实现消息确认机制

### 4. 性能优化
- 支持连接池
- 实现消息队列
- 优化内存使用

## 测试要求

### 单元测试
```bash
# 运行信令模块测试
go test ./signaling -v

# 测试特定功能
go test ./signaling -run TestSignalingMessage
go test ./signaling -run TestClientServer
go test ./signaling -run TestMessageRouting
```

### 集成测试
```bash
# 启动测试服务器
go run ./signaling/test_server.go

# 运行客户端测试
go test ./signaling -tags=integration
```

### 测试用例
1. **消息序列化测试**：测试消息编解码
2. **客户端连接测试**：测试连接建立和断开
3. **消息路由测试**：测试消息转发功能
4. **并发测试**：测试多客户端并发连接
5. **错误恢复测试**：测试断线重连机制

### 性能测试
```bash
# 基准测试
go test ./signaling -bench=.
go test ./signaling -bench=BenchmarkMessageSend
go test ./signaling -bench=BenchmarkConcurrentClients
```

## 依赖关系
- identity模块：用于身份验证和签名
- crypto模块：用于消息加密（可选）
- transport模块：用于底层网络连接

## 使用示例

```go
// 客户端使用
client := NewWebSocketClient()
err := client.Connect("ws://signaling.example.com:8080")

// 注册在线状态
peerInfo := &PeerInfo{
    PeerID:    "QmPeer123",
    IPv6Addrs: []string{"2001:db8::1"},
    PublicKey: publicKey,
}
err = client.Register(peerInfo)

// 发送连接请求
offerMsg := &SignalingMessage{
    Type:      MessageTypeOffer,
    From:      "QmPeer123",
    To:        "QmPeer456",
    Data:      sdpInfo,
    Timestamp: time.Now(),
}
err = client.SendMessage(offerMsg)

// 接收消息
msg, err := client.ReceiveMessage()
```

## 服务器配置示例

```go
// 启动信令服务器
server := NewWebSocketServer()
err := server.Start(":8080")

// 配置选项
config := &ServerConfig{
    MaxConnections:    1000,
    HeartbeatInterval: 30 * time.Second,
    MessageTimeout:    10 * time.Second,
    EnableCompression: true,
}
```

## 使用示例

### 服务器端

```go
package main

import (
    "log"
    "github.com/netvideo/signaling"
)

func main() {
    // 创建服务器配置
    config := &signaling.ServerConfig{
        Port:              8080,
        HeartbeatInterval: 30 * time.Second,
        MessageTimeout:    10 * time.Second,
        MaxConnections:    1000,
    }

    // 创建并启动服务器
    server := signaling.NewWebSocketServer(config)
    err := server.Start(":8080")
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
    defer server.Stop()

    // 获取在线用户
    peers, _ := server.GetOnlinePeers()
    log.Printf("Online peers: %d", len(peers))

    // 广播消息
    msg := &signaling.SignalingMessage{
        Type: signaling.MessageTypeInvite,
        Data: map[string]interface{}{"message": "Hello"},
    }
    server.Broadcast(msg)
}
```

### 客户端

```go
package main

import (
    "log"
    "github.com/netvideo/signaling"
)

func main() {
    // 创建客户端配置
    config := &signaling.ClientConfig{
        ServerURL:          "ws://localhost:8080/ws",
        HeartbeatInterval:  30 * time.Second,
        MessageTimeout:     10 * time.Second,
        MaxReconnectAttempts: 5,
        ReconnectInterval:  5 * time.Second,
    }

    // 创建客户端
    client := signaling.NewWebSocketClient(config)

    // 设置消息处理器
    client.SetMessageHandler(func(msg *signaling.SignalingMessage) {
        log.Printf("Received message: %s from %s", msg.Type, msg.From)
    })

    // 设置错误处理器
    client.SetErrorHandler(func(err error) {
        log.Printf("Error: %v", err)
    })

    // 连接服务器
    err := client.Connect("ws://localhost:8080/ws")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()

    // 注册在线状态
    peerInfo := &signaling.PeerInfo{
        PeerID:    "QmMyPeer123",
        IPv6Addrs: []string{"2001:db8::1"},
        IPv4Addrs: []string{"192.168.1.100"},
        PublicKey: myPublicKey,
    }
    err = client.Register(peerInfo)
    if err != nil {
        log.Printf("Failed to register: %v", err)
    }

    // 发送连接请求
    offerMsg := &signaling.SignalingMessage{
        Type: signaling.MessageTypeOffer,
        To:   "QmTargetPeer456",
        Data: signaling.SDPInfo{
            Type: "offer",
            SDP:  "v=0\r\no=- 0 0 IN IP4 127.0.0.1",
        },
    }
    err = client.SendMessage(offerMsg)
    if err != nil {
        log.Printf("Failed to send message: %v", err)
    }

    // 接收消息
    for {
        msg, err := client.ReceiveMessage()
        if err != nil {
            log.Printf("Receive error: %v", err)
            continue
        }
        log.Printf("Received: %+v", msg)
    }
}
```

### 消息签名

```go
import (
    "crypto/ed25519"
    "crypto/rand"
    "github.com/netvideo/signaling"
)

// 生成密钥对
publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)

// 创建签名器
signer := signaling.NewMessageSigner(privateKey, "QmMyPeer123")

// 创建消息
msg := &signaling.SignalingMessage{
    Type: signaling.MessageTypeOffer,
    From: "QmMyPeer123",
    To:   "QmTargetPeer456",
    Data: map[string]interface{}{"sdp": "offer"},
}

// 签名消息
signature, _ := signer.SignMessage([]byte("message data"))
msg.Signature = signature

// 验证消息
valid := ed25519.Verify(publicKey, []byte("message data"), signature)
if !valid {
    log.Fatal("Invalid signature")
}
```

## 错误处理
- 连接失败必须重试（带退避策略）
- 消息发送失败必须记录日志
- 无效消息必须拒绝并通知发送方

## 安全注意事项
1. 验证所有消息签名
2. 限制连接频率防止 DoS 攻击
3. 实现消息大小限制
4. 定期清理过期会话
5. 使用 TLS 加密连接
6. 实现 Nonce 防重放攻击