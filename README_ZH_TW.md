# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[繁體中文](README_ZH_TW.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)**

## 純 P2P 即時通訊軟體

O2OChat 是一個純點對點（P2P）即時通訊軟體，不依賴中央伺服器儲存訊息，所有通訊直接在使用者之間進行。

### 核心特性

- 🔒 **端到端加密** - 所有訊息使用 AES-256-GCM 加密
- 🌐 **純 P2P 架構** - 無中央伺服器，直接通訊
- 📱 **多平台支援** - Android、iOS、Windows、Linux、macOS、HarmonyOS
- 📁 **檔案傳輸** - 斷點續傳、多源下載、Merkle 樹驗證
- 🌍 **16 種語言** - 中文、英文、日文、韓文、德文、法文、西班牙文、俄文、馬來文、希伯來文、阿拉伯文、藏文、蒙文、維吾爾文、繁體中文

### 多作業系統支援

O2OChat 支援所有主流作業系統，提供原生應用程式和統一的使用者體驗：

| 作業系統 | 應用程式類型 | 技術堆疊 | 狀態 |
|---------|------------|---------|------|
| **Android** | 原生應用程式 | Kotlin + Jetpack Compose | ✅ 可用 |
| **iOS** | 原生應用程式 | Swift + SwiftUI | ✅ 可用 |
| **HarmonyOS** | 原生應用程式 | ArkTS + ArkUI | ✅ 可用 |
| **Windows** | 桌面應用程式 | Go + Fyne | ✅ 可用 |
| **macOS** | 桌面應用程式 | Go + Fyne/SwiftUI | ✅ 可用 |
| **Linux** | 桌面應用程式 | Go + Fyne | ✅ 可用 |

#### 平台特性

- **行動端** (Android/iOS/HarmonyOS)：完整的行動體驗，支援推播通知、背景運行、離線訊息
- **桌面端** (Windows/macOS/Linux)：完整的桌面體驗，支援多視窗、檔案拖放、快速鍵
- **統一架構**：所有平台共用相同的 P2P 核心庫，確保一致的通訊體驗
- **資料同步**：同一帳號可在多個裝置登入，訊息自動同步

### 快速開始

```bash
# 克隆專案
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# 建構
go build -o o2ochat ./cmd/o2ochat

# 運行
./o2ochat
```

### 專案結構

```
o2ochat/
├── cmd/              # 入口點
├── pkg/              # 核心函式庫
│   ├── identity/     # 身份管理
│   ├── transport/    # 網路傳輸
│   ├── signaling/    # 信令服務
│   ├── crypto/       # 加密模組
│   ├── storage/      # 資料儲存
│   ├── filetransfer/ # 檔案傳輸
│   └── media/        # 音訊/視訊處理
├── ui/               # 使用者介面
├── cli/              # 命令列工具
├── tests/            # 測試
├── docs/             # 文件
└── scripts/          # 建構腳本
```

### 技術堆疊

- **Go 1.21+** - 後端核心
- **Protocol Buffers** - 序列化
- **QUIC/WebRTC** - P2P 傳輸
- **SQLite** - 本機儲存
- **Fyne** - 桌面 GUI
- **Jetpack Compose** - Android UI
- **SwiftUI** - iOS UI
- **ArkTS** - HarmonyOS UI

### 貢獻

歡迎貢獻！請閱讀 [貢獻指南](CONTRIBUTING.md)。

### 授權

MIT License - 詳見 [LICENSE](LICENSE) 檔案。

### 聯絡我們

- 專案首頁：https://o2ochat.io
- 問題追蹤：https://github.com/yourusername/o2ochat/issues
- 電子郵件：support@o2ochat.io

---

### ⚠️ 法律風險警告

**重要提示：本專案僅供學習和研究使用**

- 📚 **學習目的** - 本專案旨在展示 P2P 通訊、端到端加密等技術的實現
- ⚖️ **遵守法律** - 使用者務必遵守所在國家/地區的法律法規
- 🚫 **禁止濫用** - 嚴禁將本專案用於任何非法活動或傳播非法內容
- 📝 **使用者責任** - 使用者應對自己的通訊內容和使用行為承擔全部法律責任
- 🔒 **技術中立** - 加密技術和 P2P 架構本身是中立的，善惡在於使用者

**使用本專案即表示您同意：**
1. 僅用於合法通訊目的
2. 不從事任何違法活動
3. 了解並接受相關技術風險
4. 遵守 [使用者服務條款](TERMS_OF_SERVICE.md) 和 [隱私政策](PRIVACY.md)

詳見：[安全使用說明](SECURITY_NOTICE.md)

---

### 相關文件

- [隱私政策](PRIVACY.md)
- [使用者服務條款](TERMS_OF_SERVICE.md)
- [安全使用說明](SECURITY_NOTICE.md)
- [快速開始指南](QUICKSTART.md)
- [架構文件](ARCHITECTURE.md)
- [開發指南](DEVELOPMENT_GUIDE.md)

---

<p align="center">
  <b>純 P2P | 端到端加密 | 自由通訊</b>
</p>

---

**版本**: v1.0.0  
**最後更新**: 2026 年 2 月 28 日  
**狀態**: ✅ 完成
