# Transport Module - 网络传输模块

## 功能概述
负责P2P网络连接的建立、管理和数据传输，支持IPv6/QUIC和IPv4/WebRTC双栈传输。

## 核心功能
1. **连接管理**：建立、维护和关闭P2P连接
2. **协议选择**：IPv6/QUIC优先，IPv4/WebRTC降级
3. **流管理**：多路复用数据流
4. **NAT穿透**：STUN/TURN辅助连接
5. **连接迁移**：网络切换时保持连接
6. **拥塞控制**：自适应带宽调整

## 接口定义

### 类型定义
```go
// 连接类型
type ConnectionType string

const (
    ConnectionTypeQUIC   ConnectionType = "quic"
    ConnectionTypeWebRTC ConnectionType = "webrtc"
)

// 连接状态
type ConnectionState string

const (
    StateDisconnected ConnectionState = "disconnected"
    StateConnecting   ConnectionState = "connecting"
    StateConnected    ConnectionState = "connected"
    StateFailed       ConnectionState = "failed"
    StateClosing      ConnectionState = "closing"
)

// 连接配置
type ConnectionConfig struct {
    PeerID        string            // 目标Peer ID
    IPv6Addresses []string          // IPv6地址列表
    IPv4Addresses []string          // IPv4地址列表
    Priority      []ConnectionType  // 连接优先级
    Timeout       time.Duration     // 连接超时
    RetryCount    int               // 重试次数
}

// 连接信息
type ConnectionInfo struct {
    ID            string            // 连接ID
    PeerID        string            // 对端Peer ID
    Type          ConnectionType    // 连接类型
    LocalAddr     net.Addr          // 本地地址
    RemoteAddr    net.Addr          // 远程地址
    State         ConnectionState   // 连接状态
    EstablishedAt time.Time         // 建立时间
    Stats         ConnectionStats   // 连接统计
}

// 数据流配置
type StreamConfig struct {
    Reliable    bool    // 是否可靠传输
    Ordered     bool    // 是否有序传输
    MaxRetries  int     // 最大重试次数
    BufferSize  int     // 缓冲区大小
}
```

### 主要接口
```go
// 传输管理器接口
type TransportManager interface {
    // 建立连接
    Connect(config *ConnectionConfig) (Connection, error)
    
    // 接受连接
    Accept() (Connection, error)
    
    // 关闭所有连接
    Close() error
    
    // 获取连接列表
    GetConnections() ([]ConnectionInfo, error)
    
    // 查找连接
    FindConnection(peerID string) (Connection, error)
    
    // 监听地址
    Listen(addr string) error
}

// 连接接口
type Connection interface {
    // 打开数据流
    OpenStream(config *StreamConfig) (Stream, error)
    
    // 接受数据流
    AcceptStream() (Stream, error)
    
    // 关闭连接
    Close() error
    
    // 获取连接信息
    GetInfo() ConnectionInfo
    
    // 发送控制消息
    SendControlMessage(msg []byte) error
    
    // 接收控制消息
    ReceiveControlMessage() ([]byte, error)
}

// 数据流接口
type Stream interface {
    // 读取数据
    Read(p []byte) (n int, err error)
    
    // 写入数据
    Write(p []byte) (n int, err error)
    
    // 关闭流
    Close() error
    
    // 获取流ID
    GetStreamID() uint32
    
    // 设置超时
    SetDeadline(t time.Time) error
    SetReadDeadline(t time.Time) error
    SetWriteDeadline(t time.Time) error
}

// NAT穿透接口
type NATTraversal interface {
    // 获取公网地址
    GetPublicAddresses() ([]string, error)
    
    // 创建打洞连接
    CreateHolePunching(localAddr, remoteAddr string) (net.Conn, error)
    
    // 使用TURN中继
    CreateRelayConnection(relayServer string) (net.Conn, error)
}
```

## 实现要求

### 1. QUIC传输实现
- 使用quic-go库实现QUIC连接
- 支持0-RTT连接建立
- 实现连接迁移
- 支持多路复用流

### 2. WebRTC传输实现
- 使用Pion WebRTC库
- 支持DataChannel可靠/不可靠传输
- 实现ICE候选收集和选择
- 支持STUN/TURN服务器

### 3. 连接策略
1. **优先级顺序**：
   - IPv6 + QUIC直连
   - IPv4 + WebRTC直连（STUN打洞）
   - IPv4 + TURN中继

2. **降级机制**：
   - 连接失败自动尝试下一优先级
   - 网络切换时保持连接
   - 实现优雅的重连机制

### 4. 性能优化
- 连接池管理
- 流控和背压机制
- 缓冲区优化
- 零拷贝传输（可选）

## 测试要求

### 单元测试
```bash
# 运行传输模块测试
go test ./transport -v

# 测试特定功能
go test ./transport -run TestQUICConnection
go test ./transport -run TestWebRTCConnection
go test ./transport -run TestConnectionManager
```

### 集成测试
```bash
# 需要实际网络环境
go test ./transport -tags=integration

# 测试NAT穿透
go test ./transport -tags=nattest
```

### 测试用例
1. **连接建立测试**：测试各种连接场景
2. **数据传输测试**：测试可靠/不可靠传输
3. **降级测试**：测试协议降级机制
4. **并发测试**：测试多连接并发
5. **错误恢复测试**：测试连接断开和恢复

### 性能测试
```bash
# 基准测试
go test ./transport -bench=.
go test ./transport -bench=BenchmarkQUICThroughput
go test ./transport -bench=BenchmarkWebRTCLatency
```

## 依赖关系
- identity模块：用于连接身份验证
- signaling模块：用于信令交换
- crypto模块：用于传输层加密

## 使用示例

```go
// 创建传输管理器
manager := NewTransportManager()

// 配置连接
config := &ConnectionConfig{
    PeerID:        "QmPeer456",
    IPv6Addresses: []string{"[2001:db8::1]:4242"},
    IPv4Addresses: []string{"192.168.1.100:4242"},
    Priority:      []ConnectionType{ConnectionTypeQUIC, ConnectionTypeWebRTC},
    Timeout:       10 * time.Second,
    RetryCount:    3,
}

// 建立连接
conn, err := manager.Connect(config)
if err != nil {
    log.Printf("连接失败: %v", err)
    return
}

// 打开数据流
streamConfig := &StreamConfig{
    Reliable:   true,
    Ordered:    true,
    BufferSize: 64 * 1024,
}
stream, err := conn.OpenStream(streamConfig)

// 发送数据
data := []byte("Hello, P2P!")
n, err := stream.Write(data)

// 接收数据
buffer := make([]byte, 1024)
n, err = stream.Read(buffer)
```

## QUIC配置示例

```go
quicConfig := &quic.Config{
    MaxIncomingStreams:    100,
    MaxIncomingUniStreams: 100,
    KeepAlive:             true,
    HandshakeIdleTimeout:  30 * time.Second,
    MaxIdleTimeout:        5 * time.Minute,
}
```

## WebRTC配置示例

```go
webrtcConfig := webrtc.Configuration{
    ICEServers: []webrtc.ICEServer{
        {
            URLs: []string{"stun:stun.l.google.com:19302"},
        },
        {
            URLs:           []string{"turn:turn.example.com:3478"},
            Username:       "username",
            Credential:     "password",
            CredentialType: webrtc.ICECredentialTypePassword,
        },
    },
}
```

## 错误处理
- 连接失败必须尝试降级协议
- 流错误必须关闭并清理资源
- 网络切换必须实现平滑迁移

## 监控指标
1. 连接成功率
2. 数据传输速率
3. 延迟和抖动
4. 丢包率
5. 重传率

## 安全注意事项
1. 验证对端身份
2. 使用前向安全加密
3. 防止连接泛洪攻击
4. 实现速率限制