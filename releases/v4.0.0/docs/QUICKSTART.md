# O2OChat 快速开始指南

## 5 分钟快速体验

### 1. 环境检查

```bash
# 检查 Go 版本
go version
# 需要 Go 1.18+

# 检查网络连接
ping -6 ipv6.google.com  # IPv6 可选
```

### 2. 获取代码

```bash
git clone https://github.com/netvideo/o2ochat.git
cd o2ochat
```

### 3. 安装依赖

```bash
go mod download
```

### 4. 运行测试

```bash
# 运行所有测试
go test ./... -v

# 预期输出：600+ 测试用例通过
```

### 5. 构建应用

```bash
# 构建 CLI 客户端
go build -o o2ochat ./cmd/cli

# 构建信令服务器
go build -o signaling-server ./cmd/signaling
```

## 使用示例

### CLI 客户端

```bash
# 启动 CLI
./o2ochat

# 可用命令
> help              # 显示帮助
> identity create   # 创建身份
> connect peer123   # 连接对等节点
> send message      # 发送消息
> file send         # 发送文件
> call audio        # 语音通话
> call video        # 视频通话
```

### 信令服务器

```bash
# 启动信令服务器
./signaling-server --port 8080

# 查看在线用户
curl http://localhost:8080/health
```

## 模块使用

### 身份模块

```go
package main

import (
    "github.com/netvideo/identity"
)

func main() {
    // 创建身份管理器
    manager := identity.NewIdentityManager(nil)
    
    // 创建新身份
    identity, err := manager.CreateIdentity("user123")
    if err != nil {
        panic(err)
    }
    
    // 获取 Peer ID
    peerID := identity.PeerID
    println("Peer ID:", peerID)
    
    // 签名消息
    message := []byte("Hello, World!")
    signature, err := manager.SignMessage("user123", message)
    
    // 验证签名
    valid, err := manager.VerifySignature("user123", message, signature)
    println("Valid:", valid)
}
```

### 信令模块

```go
package main

import (
    "github.com/netvideo/signaling"
)

func main() {
    // 创建服务器
    server := signaling.NewWebSocketServer(nil)
    server.Start(":8080")
    
    // 创建客户端
    client := signaling.NewWebSocketClient(nil)
    client.Connect("ws://localhost:8080/ws")
    
    // 注册在线
    client.Register(&signaling.PeerInfo{
        PeerID: "my-peer-id",
    })
    
    // 发送消息
    client.SendMessage(&signaling.SignalingMessage{
        Type: signaling.MessageTypeOffer,
        To:   "target-peer",
        Data: map[string]interface{}{"sdp": "offer"},
    })
}
```

### 文件传输

```go
package main

import (
    "github.com/netvideo/filetransfer"
)

func main() {
    // 创建管理器
    chunkManager := filetransfer.NewChunkManager()
    scheduler := filetransfer.NewScheduler()
    manager := filetransfer.NewFileTransferManager(
        chunkManager, scheduler, 4,
    )
    
    // 创建下载任务
    taskID, err := manager.CreateDownloadTask(
        "file-hash-123",
        "/tmp/download",
        []string{"peer1", "peer2"},
    )
    
    // 开始传输
    manager.StartTransfer(taskID)
    
    // 查看进度
    task, _ := manager.GetTaskStatus(taskID)
    println("Progress:", task.Progress.GetPercentage(), "%")
}
```

## 常见问题

### Q: 如何创建第一个身份？

```bash
./o2ochat
> identity create --name Alice
> identity list
```

### Q: 如何连接到其他用户？

```bash
> connect QmPeerID123
> status
```

### Q: 如何发送文件？

```bash
> file send /path/to/file.pdf to QmPeerID123
> file status
```

### Q: 如何开始语音通话？

```bash
> call audio QmPeerID123
> call end
```

## 下一步

- 📖 阅读 [开发指南](DEVELOPMENT_GUIDE.md)
- 🔧 查看 [模块文档](./identity/README.md)
- 🧪 运行 [集成测试](tests/integration/README.md)
- 📊 查看 [项目进度](PROGRESS_TRACKING.md)

## 获取帮助

- **文档**: 查看各模块 README.md
- **问题**: https://github.com/netvideo/o2ochat/issues
- **讨论**: https://github.com/netvideo/o2ochat/discussions

---

**开始构建你的 P2P 应用吧！** 🚀
