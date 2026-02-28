# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)**

## 純粋なP2Pインスタントメッセージングソフトウェア

O2OChatは、メッセージを中央サーバーに保存しない純粋なピアツーピア（P2P）インスタントメッセージングソフトウェアです。すべての通信はユーザー間で直接行われます。

### コア機能

- 🔒 **エンドツーエンド暗号化** - すべてのメッセージはAES-256-GCM暗号化を使用
- 🌐 **純粋なP2Pアーキテクチャ** - 中央サーバーなし、直接通信
- 📱 **マルチプラットフォームサポート** - Android、iOS、Windows、Linux、macOS、HarmonyOS
- 📁 **ファイル転送** - 中断された転送の再開、マルチソースダウンロード、Merkleツリー検証
- 🌍 **16言語** - 中国語、英語、日本語、韓国語、ドイツ語、フランス語、スペイン語、ロシア語、マレー語、ヘブライ語、アラビア語、チベット語、モンゴル語、ウイグル語、中国語（繁体字）

### クイックスタート

```bash
# プロジェクトをクローン
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# ビルド
go build -o o2ochat ./cmd/o2ochat

# 実行
./o2ochat
```

### プロジェクト構造

```
o2ochat/
├── cmd/              # エントリーポイント
├── pkg/              # コアライブラリ
│   ├── identity/     # 身元管理
│   ├── transport/    # ネットワーク転送
│   ├── signaling/    # シグナリングサービス
│   ├── crypto/       # 暗号化モジュール
│   ├── storage/      # データ保存
│   ├── filetransfer/ # ファイル転送
│   └── media/        # オーディオ/ビデオ処理
├── ui/               # ユーザーインターフェイス
├── cli/              # コマンドラインクーツ
├── tests/            # テスト
├── docs/             # ドキュメント
└── scripts/          # ビルドスクリプト
```

### テクノロジースタック

- **Go 1.21+** - バックエンドコア
- **Protocol Buffers** - シリアライゼーション
- **QUIC/WebRTC** - P2P転送
- **SQLite** - ローカル保存
- **Fyne** - デスクトップGUI
- **Jetpack Compose** - Android UI
- **SwiftUI** - iOS UI
- **ArkTS** - HarmonyOS UI

### コントリビューション

コントリビューションを歓迎します！[Contributing Guide](CONTRIBUTING.md)をお読みください。

### ライセンス

MIT License - 詳細は[LICENSE](LICENSE)ファイルを参照してください。

### お問い合わせ

- プロジェクトホームページ: https://o2ochat.io
- Issueトラッカー: https://github.com/yourusername/o2ochat/issues
- メール: support@o2ochat.io

---


---

### ⚠️ 法的リスク警告

**重要なお知らせ：このプロジェクトは教育目的のみです**

- 📚 **教育目的** - P2P 通信、エンドツーエンド暗号化技術の実演
- ⚖️ **法令遵守** - 現地の法律を遵守する必要があります
- 🚫 **違法使用禁止** - 違法活動に使用しないでください
- 📝 **利用者責任** - 利用者は法的責任を負います
- 🔒 **技術的中立性** - 技術自体は中立的です

**本プロジェクトを使用することで同意します：**
1. 合法的目的のみに使用する
2. 違法活動に従事しない
3. 技術的リスクを受け入れる
4. [利用規約](TERMS_OF_SERVICE.md) と [プライバシーポリシー](PRIVACY.md) を遵守する

詳細：[セキュリティ通知](SECURITY_NOTICE.md)

---

<p align="center">
  <b>純粋なP2P | エンドツーエンド暗号化 | 自由なコミュニケーション</b>
</p>
