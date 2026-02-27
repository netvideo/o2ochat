# O2OChat Windows 使用指南

本指南介绍 O2OChat Windows 应用程序的使用方法。

## 目录

- [程序简介](#程序简介)
- [安装运行](#安装运行)
- [界面介绍](#界面介绍)
- [功能使用](#功能使用)
- [配置说明](#配置说明)
- [常见问题](#常见问题)

## 程序简介

O2OChat 是一款纯 P2P 架构的即时通讯软件，支持以下核心功能：

- **文本聊天** - 点对点加密消息
- **文件传输** - 分块传输、多源下载、断点续传
- **语音通话** - 点对点语音通信
- **视频通话** - 点对点视频通信
- **端到端加密** - 所有通信均采用加密保护

## 安装运行

### 运行程序

双击 `o2ochat.exe` 启动应用程序。首次运行时会自动创建数据目录：

```
%APPDATA%\O2OChat\
```

程序支持以下命令行参数：

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--config` | 配置文件路径 | `%APPDATA%\O2OChat\config.json` |
| `--data-dir` | 数据存储目录 | `%APPDATA%\O2OChat` |
| `--debug` | 启用调试模式 | false |
| `--version` | 显示版本信息 | - |
| `--help` | 显示帮助信息 | - |

### 版本信息

```bash
o2ochat.exe --version
```

输出示例：
```
O2OChat v1.0.0
Build Time: 2024-01-01
P2P Instant Messaging Application

Features:
  - End-to-end encryption
  - P2P file transfer
  - Voice and video calls
  - Multi-platform support
```

## 界面介绍

### 主界面布局

```
┌─────────────────────────────────────────────┐
│  O2OChat                              [_][□][X] │
├──────────┬──────────────────────────────────┤
│          │                                    │
│  联系人   │         聊天窗口                   │
│  列表     │                                    │
│          │                                    │
│          │                                    │
│          ├──────────────────────────────────┤
│          │  输入框                      [发送] │
└──────────┴──────────────────────────────────┘
```

### 主要区域

1. **标题栏** - 程序名称和窗口控制按钮
2. **联系人列表** - 显示已添加的好友和群组
3. **聊天窗口** - 显示消息对话内容
4. **输入区域** - 编写和发送消息

## 功能使用

### 添加联系人

1. 点击联系人列表的"添加好友"按钮
2. 输入对方的 Peer ID 或用户名
3. 等待对方确认连接

### 发送消息

1. 在联系人列表中选择好友
2. 在输入框中编写消息
3. 点击发送按钮或按 Enter 发送

### 发送文件

1. 点击聊天窗口的"发送文件"按钮
2. 选择要传输的文件
3. 等待传输完成

文件传输特点：
- 自动分块传输（大文件自动拆分）
- 多源下载（可同时从多个节点下载）
- 断点续传（支持中断后继续传输）
- Merkle 树验证（确保文件完整性）

### 语音通话

1. 点击聊天窗口的"语音通话"按钮
2. 等待对方接听
3. 通话结束后点击"结束通话"

### 视频通话

1. 点击聊天窗口的"视频通话"按钮
2. 等待对方接听
3. 通话结束后点击"结束通话"

## 配置说明

### 配置文件位置

Windows 配置文件默认位于：`%APPDATA%\O2OChat\config.json`

### 配置文件格式

```json
{
  "app": {
    "debug": false,
    "data_dir": "%APPDATA%\\O2OChat"
  },
  "network": {
    "listen_port": 0,
    "use_ipv6": true,
    "stun_servers": [
      "stun:stun.l.google.com:19302"
    ],
    "turn_servers": []
  },
  "signaling": {
    "servers": [
      "wss://signal.o2ochat.io"
    ]
  },
  "storage": {
    "message_retention_days": 30
  }
}
```

### 配置项说明

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| app.debug | 启用调试模式 | true/false |
| app.data_dir | 数据存储目录 | 路径 |
| network.listen_port | 监听端口 | 0=随机端口 |
| network.use_ipv6 | 启用 IPv6 | true/false |
| network.stun_servers | STUN 服务器列表 | URL 列表 |
| network.turn_servers | TURN 服务器列表 | URL 列表 |
| signaling.servers | 信令服务器列表 | URL 列表 |
| storage.message_retention_days | 消息保留天数 | 数字 |

### 数据目录结构

```
%APPDATA%\O2OChat\
├── config.json          # 配置文件
├── identity/            # 身份密钥
│   ├── private.key      # 私钥
│   └── public.key       # 公钥
├── messages/            # 消息数据库
│   └── messages.db
├── files/               # 接收的文件
└── logs/                # 日志文件
    └── o2ochat.log
```

## 常见问题

### 无法连接网络

- 检查防火墙是否阻止程序访问网络
- 确认网络环境支持 IPv6（如不支持，修改配置关闭 IPv6）
- 尝试更换信令服务器

### 文件传输失败

- 确保双方都在线
- 检查网络连接稳定性
- 如果位于 NAT 后，可能需要配置 TURN 服务器

### 语音/视频通话质量差

- 检查网络带宽
- 关闭其他占用带宽的应用
- 确保防火墙允许 RTP 协议

### 程序无响应

- 按 Ctrl+C 强制退出
- 检查数据目录中的日志文件
- 删除数据目录后重新运行（会丢失本地数据）

## 技术支持

- 问题反馈：https://github.com/netvideo/o2ochat/issues
- 讨论区：https://github.com/netvideo/o2ochat/discussions

---

**最后更新**：2024年
