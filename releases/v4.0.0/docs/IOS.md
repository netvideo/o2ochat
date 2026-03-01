# O2OChat iOS 应用开发指南

本文档介绍 O2OChat iOS 应用的开发和使用。

## 项目结构

```
ios/
├── O2OChat/
│   ├── Sources/
│   │   ├── App/
│   │   │   ├── AppDelegate.swift
│   │   │   └── SceneDelegate.swift
│   │   ├── Models/
│   │   │   ├── Identity.swift
│   │   │   ├── Contact.swift
│   │   │   ├── Message.swift
│   │   │   └── Call.swift
│   │   ├── Services/
│   │   │   ├── IdentityService.swift
│   │   │   ├── SignalingService.swift
│   │   │   └── ContactService.swift
│   │   └── Views/
│   │       ├── MainViewController.swift
│   │       ├── ChatViewController.swift
│   │       ├── ContactCell.swift
│   │       └── MessageCell.swift
│   └── Resources/
│       └── Info.plist
├── project.yml
└── Podfile (待添加)
```

## 技术栈

- **语言**: Swift 5.9
- **最低 iOS 版本**: 15.0
- **架构**: MVVM + Combine
- **UI 框架**: UIKit
- **存储**: UserDefaults + Keychain

## 核心模块

### 1. 身份模块 (IdentityService)

负责用户身份管理：
- 生成 RSA 密钥对
- 创建 Peer ID
- 加密存储私钥到 Keychain
- 密码派生 (PBKDF2)

主要功能：
- `createIdentity(displayName:password:)` - 创建新身份
- `hasIdentity()` - 检查是否已存在身份
- `getPrivateKey(password:)` - 获取私钥

### 2. 信令模块 (SignalingService)

WebSocket 信令通信：
- 连接信令服务器 `wss://signal.o2ochat.io`
- 交换 SDP 和 ICE 候选
- 处理在线状态通知

消息类型：
- `offer` - WebRTC offer
- `answer` - WebRTC answer
- `iceCandidate` - ICE 候选
- `online` / `offline` - 在线状态

### 3. 联系人模块 (ContactService)

管理好友列表：
- 添加/删除联系人
- 在线状态管理
- 与信令服务联动更新状态

### 4. 视图层

- **MainViewController** - 主界面，显示联系人列表
- **ChatViewController** - 聊天界面
- **ContactCell** - 联系人列表项
- **MessageCell** - 消息气泡

## 构建与运行

### 环境要求

- Xcode 15.0+
- Swift 5.9+
- iOS 15.0+ 设备或模拟器

### 构建步骤

1. **使用 XcodeGen 生成项目**

```bash
cd ios
xcodegen generate
```

2. **打开项目**

```bash
open O2OChat.xcodeproj
```

3. **运行应用**

在 Xcode 中选择目标设备，点击 Run (⌘R)

### 或使用命令行

```bash
xcodebuild -project O2OChat.xcodeproj -scheme O2OChat -configuration Debug -destination 'platform=iOS Simulator,name=iPhone 15' build
```

## 功能说明

### 首次启动

1. 应用启动后弹出设置身份对话框
2. 输入显示名称和密码
3. 系统生成 RSA 密钥对和 Peer ID
4. 自动连接信令服务器

### 添加联系人

1. 点击主界面右下角 "+" 按钮
2. 输入对方的 Peer ID 和显示名称
3. 联系人添加到列表

### 发送消息

1. 在联系人列表点击好友
2. 在输入框编写消息
3. 点击发送按钮或按回车

### 语音/视频通话

1. 在聊天界面点击电话图标发起语音通话
2. 点击摄像头图标发起视频通话
3. (需要实现 WebRTC)

## 配置说明

### 信令服务器

默认服务器: `wss://signal.o2ochat.io`

可在 `SignalingService.swift` 中修改。

### 安全特性

- 端到端加密（AES-256）
- 私钥 Keychain 存储
- 密码派生 (PBKDF2 + SHA256)
- 安全随机数生成

## 权限说明

应用需要以下权限：

| 权限 | 用途 |
|------|------|
| Camera | 视频通话 |
| Microphone | 语音通话/语音消息 |
| Photo Library | 发送图片 |

## 常见问题

### 无法连接

- 检查网络连接
- 确认信令服务器可用
- 检查设备时间是否正确

### 应用无响应

- 强制退出应用
- 清理应用数据
- 重新安装

## 开发指南

### 添加新功能

1. 在 `Models` 创建数据模型
2. 在 `Services` 创建服务类
3. 在 `Views` 创建视图控制器
4. 使用 Combine 进行数据绑定

### 代码示例

```swift
// 监听联系人变化
ContactService.shared.$contacts
    .sink { contacts in
        // 更新 UI
    }
    .store(in: &cancellables)

// 监听信令消息
SignalingService.shared.messagePublisher
    .sink { message in
        // 处理信令消息
    }
    .store(in: &cancellables)
```

## 许可证

MIT License - 见项目根目录 LICENSE 文件
