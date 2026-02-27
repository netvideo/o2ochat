# O2OChat - P2P 即时通讯系统

<div align="center">

![Status](https://img.shields.io/badge/status-development%20complete-brightgreen)
![Progress](https://img.shields.io/badge/progress-81%25-blue)
![Modules](https://img.shields.io/badge/modules-9%2F9-brightgreen)
![Tests](https://img.shields.io/badge/tests-612%2B-brightgreen)
![Coverage](https://img.shields.io/badge/coverage-81%25-blue)

**纯 P2P 架构 · 端到端加密 · 全球可用**

</div>

## 项目简介

O2OChat 是一个跨平台 P2P 即时通讯系统，支持文本、语音、视频聊天和文件传输功能。采用纯 P2P 架构，最小化中心化服务器使用，确保用户隐私和数据安全。

## 核心特性

### 🔒 安全隐私
- **端到端加密** - 所有通信使用 AES-GCM 和 Ed25519 加密
- **前向安全性** - 密钥轮换确保历史消息安全
- **去中心化** - 最小化中心化服务器依赖

### 🚀 高性能
- **快速文件传输** - 分块 + 多源下载 + Merkle 树验证
- **低延迟通信** - QUIC/WebRTC 双栈传输
- **智能降级** - IPv6 优先，自动降级到 IPv4

### 🌍 全球可用
- **IPv6 优先** - 支持下一代互联网协议
- **NAT 穿透** - STUN/TURN 辅助连接
- **离线消息** - 存储转发机制

## 模块架构

```
┌─────────────────────────────────────────────────────────────┐
│                        O2OChat 系统                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐       │
│  │  UI 模块 │  │ CLI 模块 │  │ 媒体模块 │  │文件传输 │       │
│  │  (80%)  │  │  (80%)  │  │  (90%)  │  │ (80%)   │       │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘       │
│       │           │           │           │               │
│  ┌────┴───────────┴──────┬────┴───────────┴────┐           │
│  │      传输模块 (75%)    │   信令模块 (85%)    │           │
│  └───────────┬───────────┴───────────┬─────────┘           │
│              │                       │                     │
│  ┌───────────┴───────────┐  ┌────────┴────────┐           │
│  │    加密模块 (80%)      │  │   身份模块 (80%) │           │
│  └───────────┬───────────┘  └────────┬────────┘           │
│              │                       │                     │
│  ┌───────────┴───────────────────────┴─────────┐           │
│  │          存储模块 (85%)                      │           │
│  └─────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## 快速开始

### 环境要求

- Go 1.18+
- 支持 IPv6/IPv4 的网络环境
- （可选）音频/视频设备

### 安装

```bash
# 克隆项目
git clone https://github.com/netvideo/o2ochat.git
cd o2ochat

# 安装依赖
go mod download

# 构建
go build ./cmd/o2ochat
```

### 运行

```bash
# 启动 CLI 客户端
./o2ochat cli

# 启动 UI 界面
./o2ochat ui

# 运行测试
go test ./... -v
```

## 模块说明

| 模块 | 进度 | 功能 | 文档 |
|------|------|------|------|
| 身份模块 | 80% | 密钥管理、签名验证、Peer ID 生成 | [README](identity/README.md) |
| 信令模块 | 85% | WebSocket 服务器/客户端、消息路由 | [README](signaling/README.md) |
| 传输模块 | 75% | QUIC/WebRTC 双栈传输 | [README](transport/README.md) |
| 文件传输 | 80% | Merkle 树、多源下载、断点续传 | [README](filetransfer/README.md) |
| 媒体模块 | 90% | 音视频编解码、RTP 传输 | [README](media/README.md) |
| 加密模块 | 80% | AES-GCM、Ed25519、X25519 | [README](crypto/README.md) |
| 存储模块 | 85% | SQLite 存储、缓存管理 | [README](storage/README.md) |
| UI 模块 | 80% | 聊天界面、通话界面 | [README](ui/README.md) |
| CLI 模块 | 80% | 命令行管理、调试工具 | [README](cli/README.md) |

## 项目状态

### 开发进度

- **总体进度**: 81.1%
- **模块完成**: 9/9 (100%)
- **测试覆盖**: 81%
- **文档完整**: 100%

### 里程碑

- ✅ M1: 接口定义完成
- ✅ M2: 核心功能完成
- ✅ M3: 模块开发完成
- ✅ M4: 优化和文档完成
- ⏳ M5: 集成测试准备

## 测试

```bash
# 运行所有测试
go test ./... -v

# 运行单元测试
go test ./identity/... ./crypto/... ./storage/... -v

# 运行集成测试
go test ./tests/integration/... -v

# 运行性能测试
go test ./... -bench=. -benchmem
```

## 文档

| 文档 | 说明 |
|------|------|
| [开发指南](DEVELOPMENT_GUIDE.md) | 并行开发指南 |
| [进度跟踪](PROGRESS_TRACKING.md) | 实时进度跟踪 |
| [测试报告](tests/integration/README.md) | 集成测试文档 |
| [完成总结](PROJECT_COMPLETION_SUMMARY.md) | 项目完成报告 |

## 技术栈

- **语言**: Go 1.18+
- **加密**: AES-GCM, Ed25519, X25519, SHA256
- **传输**: QUIC (quic-go), WebRTC (pion)
- **存储**: SQLite, 内存缓存
- **信令**: WebSocket (gorilla/websocket)

## 贡献

欢迎贡献代码、报告问题或提出建议！

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- **项目主页**: https://github.com/netvideo
- **问题反馈**: https://github.com/netvideo/o2ochat/issues
- **讨论区**: https://github.com/netvideo/o2ochat/discussions

---

<div align="center">

**O2OChat** - 安全、私密、去中心化的即时通讯

Made with ❤️ by the O2OChat Team

</div>
