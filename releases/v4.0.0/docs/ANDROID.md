# O2OChat Android 应用开发指南

本文档介绍 O2OChat Android 应用的开发和使用。

## 项目结构

```
android/
├── app/
│   ├── src/main/
│   │   ├── java/com/o2ochat/
│   │   │   ├── O2OChatApp.kt           # 应用入口
│   │   │   ├── domain/model/           # 领域模型
│   │   │   │   ├── Identity.kt
│   │   │   │   ├── Contact.kt
│   │   │   │   ├── Message.kt
│   │   │   │   └── Call.kt
│   │   │   ├── data/
│   │   │   │   ├── local/              # 本地数据库
│   │   │   │   │   ├── Database.kt
│   │   │   │   │   ├── MessageDao.kt
│   │   │   │   │   ├── ContactDao.kt
│   │   │   │   │   └── IdentityDao.kt
│   │   │   │   ├── remote/             # 远程API
│   │   │   │   │   └── SignalingApi.kt
│   │   │   │   └── repository/         # 数据仓库
│   │   │   │       ├── IdentityRepository.kt
│   │   │   │       ├── ContactRepository.kt
│   │   │   │       └── MessageRepository.kt
│   │   │   ├── ui/                     # UI层
│   │   │   │   ├── main/
│   │   │   │   │   ├── MainActivity.kt
│   │   │   │   │   └── ContactAdapter.kt
│   │   │   │   ├── chat/
│   │   │   │   ├── contacts/
│   │   │   │   ├── call/
│   │   │   │   └── settings/
│   │   │   └── service/                # 后台服务
│   │   │       ├── SignalingService.kt
│   │   │       └── P2PService.kt
│   │   ├── res/
│   │   │   ├── layout/                 # 布局文件
│   │   │   ├── values/                 # 资源文件
│   │   │   └── drawable/
│   │   └── AndroidManifest.xml
│   └── build.gradle
├── build.gradle
└── settings.gradle
```

## 技术栈

- **语言**: Kotlin 1.9.22
- **最低 SDK**: 24 (Android 7.0)
- **目标 SDK**: 34 (Android 14)
- **架构**: MVVM + Clean Architecture
- **依赖注入**: 手动注入 (简化版)
- **数据库**: Room
- **网络**: OkHttp + WebSocket
- **UI**: Material Design 3

## 核心模块

### 1. 身份模块 (Identity)

负责用户身份管理，包括：
- 生成 Ed25519 密钥对
- 创建 Peer ID
- 加密存储私钥
- 身份验证

主要类：
- `IdentityRepository` - 身份数据操作
- `Identity` - 身份领域模型
- `IdentityEntity` - 身份数据库实体

### 2. 联系人模块 (Contact)

管理好友列表：
- 添加/删除联系人
- 在线状态管理
- 联系人信息存储

主要类：
- `ContactRepository` - 联系人数据操作
- `Contact` - 联系人领域模型

### 3. 消息模块 (Message)

处理聊天消息：
- 发送/接收消息
- 消息状态跟踪
- 消息类型支持（文本、图片、文件、语音、视频）

主要类：
- `MessageRepository` - 消息数据操作
- `Message` - 消息领域模型

### 4. 信令模块 (Signaling)

WebSocket 信令通信：
- 连接信令服务器
- 交换 SDP 和 ICE 候选
- 处理在线状态

主要类：
- `SignalingApi` - WebSocket API
- `SignalingMessage` - 信令消息类型

### 5. 通话模块 (Call)

P2P 音视频通话：
- WebRTC 连接管理
- 音视频轨道处理
- 通话状态控制

## 构建与运行

### 环境要求

- Android Studio Arctic Fox 或更高版本
- JDK 17
- Android SDK 34

### 构建命令

```bash
cd android

# 调试构建
./gradlew assembleDebug

# 发布构建
./gradlew assembleRelease

# 运行测试
./gradlew test

# 清理构建
./gradlew clean
```

### 运行应用

1. 在 Android Studio 中打开 `android` 目录
2. 等待 Gradle 同步完成
3. 连接 Android 设备或启动模拟器
4. 点击 Run 按钮运行应用

## 功能说明

### 首次启动

1. 应用启动后提示创建身份
2. 输入显示名称
3. 设置密码（用于加密私钥）
4. 系统生成 Peer ID

### 添加联系人

1. 点击主界面右下角 "+" 按钮
2. 输入对方的 Peer ID
3. 等待连接建立

### 发送消息

1. 在联系人列表点击好友
2. 在输入框编写消息
3. 点击发送按钮

### 语音/视频通话

1. 在聊天界面点击电话图标发起语音通话
2. 点击摄像头图标发起视频通话
3. 对方接听后开始通话

## 配置说明

### 配置文件

应用配置存储在：
- 加密配置: `%ENCRYPTED_PREFS%`
- 数据库: `o2ochat.db`

### 网络配置

默认信令服务器: `wss://signal.o2ochat.io`

可在代码中修改 `SignalingApi.kt` 定制服务器地址。

## 安全特性

- 端到端加密（AES-256-GCM）
- 私钥加密存储（PBKDF2 + AES）
- 安全随机数生成
- 网络传输加密（TLS 1.3）

## 常见问题

### 无法连接

- 检查网络连接
- 确认信令服务器可用
- 检查防火墙设置

### 通话失败

- 确认双方都在线
- 检查麦克风和摄像头权限
- 尝试更换网络环境

### 应用无响应

- 强制停止应用
- 清除应用数据
- 重新启动应用

## 开发指南

### 添加新功能

1. 在 `domain/model` 创建领域模型
2. 在 `data/local` 创建数据库相关类
3. 在 `data/repository` 创建仓库类
4. 在 `ui` 创建对应的 Activity/Fragment
5. 更新 `AndroidManifest.xml`

### 运行测试

```bash
# 单元测试
./gradlew test

# UI 测试
./gradlew connectedAndroidTest
```

## 许可证

MIT License - 见项目根目录 LICENSE 文件
