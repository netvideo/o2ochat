# O2OChat 鸿蒙应用开发指南

本文档介绍 O2OChat 鸿蒙（HarmonyOS）应用的开发和使用。

## 项目结构

```
harmony/
└── O2OChat/
    └── entry/
        └── src/
            └── main/
                ├── ets/
                │   ├── entryability/
                │   │   └── EntryAbility.ts
                │   ├── models/
                │   │   ├── Identity.ts
                │   │   ├── Contact.ts
                │   │   └── Message.ts
                │   ├── services/
                │   │   ├── IdentityService.ts
                │   │   ├── SignalingService.ts
                │   │   └── ContactService.ts
                │   └── pages/
                │       ├── Index.ets
                │       └── Chat.ets
                └── module.json5
```

## 技术栈

- **语言**: ArkTS (TypeScript)
- **最低 SDK**: API 9 (HarmonyOS 4.0)
- **架构**: MVVM
- **UI 框架**: ArkUI
- **存储**: Preferences

## 核心模块

### 1. 身份模块 (IdentityService)

负责用户身份管理：
- 生成 RSA 密钥对
- 创建 Peer ID
- 身份数据持久化存储

主要方法：
- `createIdentity(displayName, password)` - 创建新身份
- `hasIdentity()` - 检查是否已存在身份
- `getIdentity()` - 获取当前身份

### 2. 信令模块 (SignalingService)

WebSocket 信令通信：
- 连接信令服务器 `wss://signal.o2ochat.io`
- 交换 SDP 和 ICE 候选
- 处理在线状态

消息类型：
- `offer` - WebRTC offer
- `answer` - WebRTC answer
- `ice` - ICE 候选
- `online` / `offline` - 在线状态

### 3. 联系人模块 (ContactService)

管理联系人列表：
- 添加/删除联系人
- 在线状态管理
- 联系人数据持久化

## 构建与运行

### 环境要求

- DevEco Studio 4.0+
- SDK 4.0.0 (API 10)
- Node.js 18+

### 构建步骤

1. **打开项目**

使用 DevEco Studio 打开 `harmony/O2OChat` 目录

2. **同步项目**

点击 File > Sync Project，等待 Gradle 同步完成

3. **运行应用**

选择目标设备，点击 Run (Shift + F10)

### 或使用命令行

```bash
cd harmony/O2OChat

# 构建 Debug 版本
./gradlew assembleDebug

# 构建 Release 版本
./gradlew assembleRelease
```

## 页面说明

### Index.ets - 主页面

显示联系人列表：
- 顶部显示应用名称和连接状态
- 中间显示联系人列表
- 右下角添加联系人按钮

### Chat.ets - 聊天页面

聊天功能：
- 顶部显示联系人名称和返回按钮
- 中间消息列表显示对话
- 底部输入框和发送按钮

## 功能说明

### 首次启动

1. 应用检测是否已存在身份
2. 如无身份，弹出设置对话框
3. 输入显示名称创建身份
4. 自动连接信令服务器

### 添加联系人

1. 点击右下角 "+" 按钮
2. 输入对方的 Peer ID
3. 联系人添加到列表

### 发送消息

1. 点击联系人进入聊天
2. 在输入框编写消息
3. 点击发送按钮

## 配置说明

### 信令服务器

默认服务器: `wss://signal.o2ochat.io`

可在 `SignalingService.ts` 中修改 `SERVER_URL` 常量。

## 安全特性

- 端到端加密（AES）
- 私钥安全存储
- 安全随机数生成

## 权限说明

需要在 `module.json5` 中声明的权限：

| 权限 | 用途 |
|------|------|
| ohos.permission.INTERNET | 网络访问 |
| ohos.permission.GET_NETWORK_INFO | 网络状态 |

## 常见问题

### 无法连接

- 检查网络连接
- 确认信令服务器可用
- 检查设备时间

### 应用无响应

- 重启应用
- 清除应用数据

## 开发指南

### 添加新页面

1. 在 `pages` 目录创建新 `.ets` 文件
2. 在 `module.json5` 中注册路由

### 代码示例

```typescript
// 路由跳转
import router from '@ohos.router';

router.push({
  url: 'pages/Chat',
  params: {
    peerId: 'xxx',
    displayName: 'User'
  }
});

// 接收参数
const params = router.getParams();
```

## 许可证

MIT License - 见项目根目录 LICENSE 文件
