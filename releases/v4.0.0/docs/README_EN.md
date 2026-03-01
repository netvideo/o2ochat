# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)**

## Pure P2P Instant Messaging Software

O2OChat is a pure peer-to-peer (P2P) instant messaging software that does not rely on central servers to store messages. All communications occur directly between users.

### Core Features

- 🔒 **End-to-End Encryption** - All messages use AES-256-GCM encryption
- 🌐 **Pure P2P Architecture** - No central server, direct communication
- 📱 **Multi-Platform Support** - Android, iOS, Windows, Linux, macOS, HarmonyOS
- 📁 **File Transfer** - Resume broken transfers, multi-source download, Merkle tree verification
- 🌍 **16 Languages** - Chinese, English, Japanese, Korean, German, French, Spanish, Russian, Malay, Hebrew, Arabic, Tibetan, Mongolian, Uyghur, Traditional Chinese

### Quick Start

```bash
# Clone the project
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# Build
go build -o o2ochat ./cmd/o2ochat

# Run
./o2ochat
```

### Project Structure

```
o2ochat/
├── cmd/              # Entry points
├── pkg/              # Core libraries
│   ├── identity/     # Identity management
│   ├── transport/    # Network transport
│   ├── signaling/    # Signaling service
│   ├── crypto/       # Encryption module
│   ├── storage/      # Data storage
│   ├── filetransfer/ # File transfer
│   └── media/        # Audio/video processing
├── ui/               # User interface
├── cli/              # Command line tools
├── tests/            # Tests
├── docs/             # Documentation
└── scripts/          # Build scripts
```

### Technology Stack

- **Go 1.21+** - Backend core
- **Protocol Buffers** - Serialization
- **QUIC/WebRTC** - P2P transport
- **SQLite** - Local storage
- **Fyne** - Desktop GUI
- **Jetpack Compose** - Android UI
- **SwiftUI** - iOS UI
- **ArkTS** - HarmonyOS UI

### Contributing

Contributions are welcome! Please read the [Contributing Guide](CONTRIBUTING.md).

### License

MIT License - See [LICENSE](LICENSE) file for details.

### Contact Us

- Project Homepage: https://o2ochat.io
- Issue Tracker: https://github.com/yourusername/o2ochat/issues
- Email: support@o2ochat.io

---

### 🤖 AI Development Statement

**This project is primarily developed autonomously by AI, all code is AI-generated**

This is a groundbreaking AI-autonomous development project, from architecture design to code implementation, all completed through collaboration of multiple advanced AI models:

- 🧠 **Project Planning**: DeepSeek3.2 Web Version
- 💻 **Code and Documentation**: 
  - DeepSeek3.2-Chat
  - MiniMax M2.5
  - Qwen3.5 Plus (current)
  - Kimi K2.5
- 🛠️ **Development Tool**: OpenCode

**Project Features**:
- ✅ AI autonomous decision-making and execution
- ✅ Multi-model collaborative development
- ✅ Humans only propose requirements, AI completes all implementation
- ✅ Demonstrates AI's potential in software development

**Tech Stack**:
- Go (Backend Core)
- Solidity (Smart Contracts)
- Kotlin (Android)
- Swift (iOS)
- ArkTS (HarmonyOS)
- Fyne (Cross-platform Desktop)

**Development Date**: February 28, 2026  
**Development Mode**: AI Autonomous  
**Human Involvement**: Requirement proposal, AI full implementation

---

### ⚠️ Legal Risk Warning

**Important Notice: This project is for educational purposes only**

- 📚 **Educational Purpose** - This project demonstrates P2P communication, end-to-end encryption technologies
- ⚖️ **Compliance with Laws** - Users must comply with local laws and regulations
- 🚫 **No Illegal Use** - Prohibited for illegal activities or content
- 📝 **User Responsibility** - Users bear full legal responsibility
- 🔒 **Technology Neutrality** - Technology itself is neutral

**By using this project, you agree to:**
1. Use only for legal purposes
2. Not engage in illegal activities
3. Accept technical risks
4. Comply with [Terms](TERMS_OF_SERVICE.md) and [Privacy](PRIVACY.md)

See: [Security Notice](SECURITY_NOTICE.md)

---

<p align="center">
  <b>Pure P2P | End-to-End Encryption | Free Communication</b>
</p>
