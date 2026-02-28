# O2OChat 去中心化无服务器支持

## 概述

O2OChat 实现完全去中心化的 P2P 通信架构，无需任何中央服务器。所有节点平等，消息直接在用户之间传输。

## 核心特性

### 完全去中心化

- ✅ **无中央服务器** - 无需任何服务器基础设施
- ✅ **分布式架构** - 所有节点平等，无主从关系
- ✅ **自我组织** - 节点自动发现和连接
- ✅ **抗审查** - 无法被单点关闭或审查

### DHT 分布式哈希表

- ✅ **Kademlia 协议** - 高效的分布式查找
- ✅ **节点发现** - 自动发现网络中的其他节点
- ✅ **数据存储** - 分布式存储和检索
- ✅ **容错性** - 节点故障不影响网络

### NAT 穿透

- ✅ **STUN/TURN** - 自动 NAT 穿透
- ✅ **UDP 打洞** - 直接 P2P 连接
- ✅ **中继支持** - 无法直连时自动中继
- ✅ **IPv6 优先** - 优先使用 IPv6 直连

## 架构设计

### 网络层

```
┌─────────────────────────────────────────┐
│          应用层 (O2OChat)               │
├─────────────────────────────────────────┤
│          消息路由层                      │
├─────────────────────────────────────────┤
│          DHT (Kademlia)                 │
├─────────────────────────────────────────┤
│   传输层 (QUIC/TCP/WebRTC)              │
├─────────────────────────────────────────┤
│   网络层 (IPv4/IPv6)                    │
└─────────────────────────────────────────┘
```

### 节点结构

每个节点包含：
- **Node ID** - 唯一标识（从公钥生成）
- **公钥/私钥** - 身份认证和加密
- **地址列表** - 可连接的地址（IP:PORT, multiaddr）
- **能力列表** - 支持的功能（消息、文件、语音、视频）
- **状态信息** - 在线/离线、延迟等

## 使用示例

### 基本使用

```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/netvideo/o2ochat/pkg/decentralized"
)

func main() {
	// 创建 DHT 配置
	config := &decentralized.DHTConfig{
		NodeID:              decentralized.GenerateNodeID("my-public-key"),
		PublicKey:           "my-public-key",
		PrivateKey:          "my-private-key",
		ListenAddresses:     []string{":8080", ":8081"},
		BootstrapNodes:      []string{"bootstrap1.o2ochat.io:8080"},
		MaxPeers:            100,
		ConnectionTimeout:   10 * time.Second,
		DiscoveryInterval:   30 * time.Second,
		EnableRelay:         true,
		EnableHolePunching:  true,
	}
	
	// 创建 DHT 实例
	dht := decentralized.NewDHT(config)
	
	// 设置回调
	dht.SetOnPeerDiscovered(func(node *decentralized.Node) {
		fmt.Printf("发现新节点：%s\n", node.ID)
	})
	
	dht.SetOnPeerConnected(func(node *decentralized.Node) {
		fmt.Printf("连接到节点：%s\n", node.ID)
	})
	
	// 启动 DHT
	err := dht.Start()
	if err != nil {
		panic(err)
	}
	
	// 等待节点连接
	time.Sleep(5 * time.Second)
	
	// 获取在线节点
	peers := dht.GetActivePeers()
	fmt.Printf("在线节点数：%d\n", len(peers))
	
	// 发送消息给特定节点
	ctx := context.Background()
	err = dht.SendToPeer(ctx, "target-node-id", []byte("Hello P2P!"))
	
	// 广播消息
	err = dht.Broadcast(ctx, []byte("Hello everyone!"))
	
	// 停止 DHT
	dht.Stop()
}
```

### 节点发现

```go
// 查找特定节点
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

node, err := dht.FindNode(ctx, "target-node-id")
if err != nil {
	fmt.Printf("未找到节点：%v\n", err)
} else {
	fmt.Printf("找到节点：%s\n", node.ID)
}

// 存储数据到 DHT
err = dht.Store(ctx, "message-key", []byte("message-value"))

// 从 DHT 检索数据
value, err := dht.FindValue(ctx, "message-key")
```

### 自定义 PeerStore

```go
peerStore := dht.GetPeerStore()

// 添加节点
peerStore.Add(&decentralized.Node{
	ID:     "node-id",
	Status: decentralized.NodeOnline,
})

// 获取节点
node, err := peerStore.Get("node-id")

// 更新状态
peerStore.UpdateStatus("node-id", decentralized.NodeOffline)

// 列出所有节点
nodes := peerStore.List()

// 获取在线数量
onlineCount := peerStore.GetOnlineCount()
```

## 配置选项

### DHTConfig

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `NodeID` | NodeID | 节点唯一标识 | 必填 |
| `PrivateKey` | string | 私钥 | 必填 |
| `PublicKey` | string | 公钥 | 必填 |
| `ListenAddresses` | []string | 监听地址列表 | []string{":8080"} |
| `BootstrapNodes` | []string | 引导节点列表 | []string{} |
| `MaxPeers` | int | 最大连接数 | 100 |
| `ConnectionTimeout` | time.Duration | 连接超时 | 10 秒 |
| `DiscoveryInterval` | time.Duration | 发现间隔 | 30 秒 |
| `EnableRelay` | bool | 启用中继 | true |
| `EnableHolePunching` | bool | 启用打洞 | true |

### NodeStatus

| 状态 | 值 | 说明 |
|------|-----|------|
| `NodeOnline` | 0 | 节点在线，可连接 |
| `NodeOffline` | 1 | 节点离线 |
| `NodeUnknown` | 2 | 状态未知 |

## 网络拓扑

### Kademlia DHT

O2OChat 使用 Kademlia 协议的变种：

```
节点 ID 空间：2^256
距离计算：XOR 距离
路由表：K-Buckets
查找复杂度：O(log N)
```

### 节点连接

```
每个节点维护：
- 近距离节点（高优先级）
- 中距离节点（中优先级）
- 远距离节点（低优先级）

自动维护连接：
- 定期 ping 检查活跃性
- 移除超时节点
- 添加新发现节点
```

## 安全性

### 身份认证

- ✅ **公钥基础设施** - 每个节点有唯一密钥对
- ✅ **数字签名** - 所有消息签名验证
- ✅ **身份绑定** - NodeID 从公钥生成

### 通信安全

- ✅ **端到端加密** - 消息内容加密
- ✅ **传输加密** - TLS/QUIC 加密传输
- ✅ **前向保密** - 会话密钥定期更换

### 防攻击

- ✅ **节点封禁** - 恶意节点可被封禁
- ✅ **速率限制** - 防止洪水攻击
- ✅ **连接限制** - 限制单节点连接数

## 性能优化

### 连接池

- 维护到常用节点的热连接
- 延迟自动选择最优路径
- 连接复用减少建立开销

### 消息路由

- 智能路由选择最短路径
- 批量发送减少网络往返
- 压缩减少带宽使用

### 缓存

- 节点信息缓存
- 路由表缓存
- 消息去重缓存

## 故障恢复

### 节点失效

- 自动检测节点离线
- 自动重连或寻找替代节点
- 消息队列等待重发

### 网络分区

- 检测网络分区
- 本地操作继续
- 分区恢复后同步

### 数据一致性

- 最终一致性模型
- 冲突解决策略
- 版本控制

## 部署模式

### 纯 P2P 模式

完全去中心化，无需任何服务器：
```go
config := &DHTConfig{
	BootstrapNodes: []string{}, // 空引导节点
	// 其他配置...
}
```

### 混合模式

使用少量引导节点加速发现：
```go
config := &DHTConfig{
	BootstrapNodes: []string{
		"bootstrap1.o2ochat.io:8080",
		"bootstrap2.o2ochat.io:8080",
	},
}
```

### 私有网络

创建私有 P2P 网络：
```go
config := &DHTConfig{
	BootstrapNodes: []string{"private-node:8080"},
	// 使用私有密钥和证书
}
```

## 监控和调试

### 节点状态监控

```go
// 获取统计信息
peers := dht.GetActivePeers()
onlineCount := peerStore.GetOnlineCount()

// 监控连接质量
for _, peer := range peers {
	fmt.Printf("节点：%s, 延迟：%dms\n", peer.ID, peer.Latency)
}
```

### 日志记录

```go
// 启用详细日志
config.VerboseLogging = true

// 设置日志级别
config.LogLevel = "debug"
```

### 诊断工具

```go
// 网络诊断
dht.Diagnose()

// 路由表导出
routingTable := dht.ExportRoutingTable()

// 性能统计
stats := dht.GetStats()
```

## 限制和注意事项

### NAT 限制

- 对称 NAT 可能无法穿透
- 需要中继作为后备
- IPv6 可避免大部分 NAT 问题

### 可扩展性

- 单节点连接数有限制（默认 100）
- 大网络需要分层路由
- 消息广播有延迟

### 隐私考虑

- IP 地址对连接节点可见
- 使用 Tor/VPN 可隐藏 IP
- 消息内容端到端加密

## 最佳实践

### 节点配置

1. **监听多个地址** - 同时监听 IPv4 和 IPv6
2. **使用多个引导节点** - 提高可靠性
3. **合理设置 MaxPeers** - 根据资源调整

### 连接管理

1. **定期清理离线节点** - 保持连接质量
2. **优先直连** - 减少中继使用
3. **监控延迟** - 自动选择最优路径

### 消息处理

1. **批量发送** - 减少网络往返
2. **压缩大消息** - 节省带宽
3. **设置超时** - 避免无限等待

## 相关文档

- [DHT 代码](../pkg/decentralized/dht.go)
- [P2P 连接模块](../pkg/p2p/)
- [消息路由](../pkg/message/)

---

**版本**: v1.0.0  
**更新时间**: 2026 年 2 月 28 日
