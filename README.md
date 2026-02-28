# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[繁體中文](README_ZH_TW.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)** | **[Português](README_PT_BR.md)** | **[Italiano](README_IT.md)**

## 纯 P2P 即时通讯软件

O2OChat 是一个纯点对点（P2P）即时通讯软件，不依赖中央服务器存储消息，所有通信直接在用户之间进行。

### 核心特性

- 🔒 **端到端加密** - 所有消息使用 AES-256-GCM 加密
- 🌐 **纯 P2P 架构** - 无中央服务器，直接通信
- 📱 **多平台支持** - Android、iOS、Windows、Linux、macOS、HarmonyOS
- 📁 **文件传输** - 断点续传、多源下载、Merkle 树验证
- 🌍 **16 种语言** - 中文、英文、日文、韩文、德文、法文、西班牙文、俄文、马来文、希伯来文、阿拉伯文、藏文、蒙文、维吾尔文、繁体中文

### 多操作系统支持

O2OChat 支持所有主流操作系统，提供原生应用和统一的用户体验：

| 操作系统 | 应用类型 | 技术栈 | 状态 |
|---------|---------|--------|------|
| **Android** | 原生应用 | Kotlin + Jetpack Compose | ✅ 可用 |
| **iOS** | 原生应用 | Swift + SwiftUI | ✅ 可用 |
| **HarmonyOS** | 原生应用 | ArkTS + ArkUI | ✅ 可用 |
| **Windows** | 桌面应用 | Go + Fyne | ✅ 可用 |
| **macOS** | 桌面应用 | Go + Fyne/SwiftUI | ✅ 可用 |
| **Linux** | 桌面应用 | Go + Fyne | ✅ 可用 |

#### 平台特性

- **移动端** (Android/iOS/HarmonyOS): 完整的移动体验，支持推送通知、后台运行、离线消息
- **桌面端** (Windows/macOS/Linux): 完整的桌面体验，支持多窗口、文件拖放、快捷键
- **统一架构**: 所有平台共享相同的 P2P 核心库，确保一致的通信体验
- **数据同步**: 同一账号可在多个设备登录，消息自动同步

### 快速开始

```bash
# 克隆项目
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# 构建
go build -o o2ochat ./cmd/o2ochat

# 运行
./o2ochat
```

### 项目结构

```
o2ochat/
├── cmd/              # 入口点
├── pkg/              # 核心库
│   ├── identity/     # 身份管理
│   ├── transport/    # 网络传输
│   ├── signaling/    # 信令服务
│   ├── crypto/       # 加密模块
│   ├── storage/      # 数据存储
│   ├── filetransfer/ # 文件传输
│   └── media/        # 音视频处理
├── ui/               # 用户界面
├── cli/              # 命令行工具
├── tests/            # 测试
├── docs/             # 文档
└── scripts/          # 构建脚本
```

### 技术栈

- **Go 1.21+** - 后端核心
- **Protocol Buffers** - 序列化
- **QUIC/WebRTC** - P2P 传输
- **SQLite** - 本地存储
- **Fyne** - 桌面 GUI
- **Jetpack Compose** - Android UI
- **SwiftUI** - iOS UI
- **ArkTS** - HarmonyOS UI

### 贡献

欢迎贡献！请阅读 [贡献指南](CONTRIBUTING.md)。

### 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。

### 联系我们

- 项目主页：https://github.com/netvideo/o2ochat
- 问题反馈：https://github.com/netvideo/o2ochat/issues
- 邮件：netvideo1@sina.com

---

### 相关文档

- [隐私政策](PRIVACY.md)
- [用户协议](TERMS_OF_SERVICE.md)
- [安全使用说明](SECURITY_NOTICE.md)
- [测试报告](TEST_REPORT.md)
- [修复报告](FIX_REPORT.md)
- [完善报告](IMPROVEMENT_COMPLETION_REPORT.md)
- [多语言报告](MULTILINGUAL_COMPLETION_REPORT.md)

---

### ⚠️ 法律风险警告

**重要提示：本项目仅供学习和研究使用**

- 📚 **学习目的** - 本项目旨在展示 P2P 通信、端到端加密等技术的实现
- ⚖️ **遵守法律** - 使用者务必遵守所在国家/地区的法律法规
- 🚫 **禁止滥用** - 严禁将本项目用于任何非法活动或传播非法内容
- 📝 **用户责任** - 用户应对自己的通信内容和使用行为承担全部法律责任
- 🔒 **技术中立** - 加密技术和 P2P 架构本身是中立的，善恶在于使用者

**使用本项目即表示您同意：**
1. 仅用于合法通信目的
2. 不从事任何违法活动
3. 了解并接受相关技术风险
4. 遵守 [用户协议](TERMS_OF_SERVICE.md) 和 [隐私政策](PRIVACY.md)

详见：[安全使用说明](SECURITY_NOTICE.md) | [法律风险审核报告](docs/LEGAL_RISK_AUDIT.md)

---

### 项目统计

- **版本**: v1.0.0
- **代码**: 5,791+ 行
- **测试**: 2,500+ 行
- **文档**: 203,550+ 行
- **语言**: 15 种 README
- **界面**: 16 种语言
- **平台**: 6 个
- **许可**: MIT
- **状态**: ✅ 完成

---

### 发布信息

- **GitHub**: https://github.com/netvideo/o2ochat
- **发布**: v1.0.0
- **时间**: 2026 年 2 月 28 日
- **作者**: netvideo <netvideo1@sina.com>
- **状态**: ✅ 已发布

---

<p align="center">
  <b>纯 P2P | 端到端加密 | 自由通信</b>
</p>

---

**版本**: v1.0.0  
**最后更新**: 2026 年 2 月 28 日  
**状态**: ✅ 完成
