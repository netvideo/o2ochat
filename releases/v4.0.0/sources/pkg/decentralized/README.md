# Decentralized Network Module (pkg/decentralized)

去中心化网络模块实现完全分布式的 P2P 通信，无需任何中央服务器。

## 功能特性

- 🌐 **完全去中心化** - 无中央服务器，所有节点平等
- 🗄️ **DHT 分布式哈希表** - Kademlia 协议实现
- 🔍 **节点发现** - 自动发现和连接其他节点
- 🕳️ **NAT 穿透** - STUN/TURN, UDP 打洞
- 📡 **IPv6 优先** - 优先使用 IPv6 直连
- 🔗 **中继支持** - 无法直连时自动中继

## 核心组件

### dht.go
DHT 实现，包括节点管理、PeerStore、连接管理

## 网络架构

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

## 使用示例

```go
config := &decentralized.DHTConfig{
    NodeID: decentralized.GenerateNodeID("public-key"),
    ListenAddresses: []string{":8080"},
    BootstrapNodes: []string{"bootstrap.o2ochat.io:8080"},
    EnableRelay: true,
}

dht := decentralized.NewDHT(config)
dht.Start()

// 发送消息
dht.SendToPeer(ctx, "target-node-id", []byte("Hello P2P!"))
```

## 文档

- [去中心化网络指南](../../docs/DECENTRALIZED_NETWORK.md)

**版本**: v1.0.0
