# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)**

## Perisian Pemesejan Segera P2P Tulen

O2OChat adalah perisian pemesejan segera peer-to-peer (P2P) tulen yang tidak bergantung pada pelayan pusat untuk menyimpan mesej. Semua komunikasi berlaku terus antara pengguna.

### Ciri-ciri Utama

- 🔒 **Penyulitan Hujung-ke-Hujung** - Semua mesej menggunakan penyulitan AES-256-GCM
- 🌐 **Senibina P2P Tulen** - Tiada pelayan pusat, komunikasi terus
- 📱 **Sokongan Pelbagai Platform** - Android, iOS, Windows, Linux, macOS, HarmonyOS
- 📁 **Pemindahan Fail** - Sambung semula pemindahan yang terputus, muat turun pelbagai sumber, pengesahan pokok Merkle
- 🌍 **16 Bahasa** - Cina, Inggeris, Jepun, Korea, Jerman, Perancis, Sepanyol, Rusia, Melayu, Ibrani, Arab, Tibet, Mongolia, Uyghur, Cina (Tradisional)

### Permulaan Pantas

```bash
# Klon projek
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# Bina
go build -o o2ochat ./cmd/o2ochat

# Jalankan
./o2ochat
```

### Struktur Projek

```
o2ochat/
├── cmd/              # Titik masuk
├── pkg/              # Perpustakaan teras
│   ├── identity/     # Pengurusan identiti
│   ├── transport/    # Pengangkutan rangkaian
│   ├── signaling/    # Perkhidmatan isyarat
│   ├── crypto/       # Modul penyulitan
│   ├── storage/      # Penyimpanan data
│   ├── filetransfer/ # Pemindahan fail
│   └── media/        # Pemprosesan audio/video
├── ui/               # Antara muka pengguna
├── cli/              # Alat baris perintah
├── tests/            # Ujian
├── docs/             # Dokumentasi
└── scripts/          # Skrip pembinaan
```

### Tumpukan Teknologi

- **Go 1.21+** - Teras backend
- **Protocol Buffers** - Penserialan
- **QUIC/WebRTC** - Pengangkutan P2P
- **SQLite** - Penyimpanan setempat
- **Fyne** - GUI desktop
- **Jetpack Compose** - UI Android
- **SwiftUI** - UI iOS
- **ArkTS** - UI HarmonyOS

### Menyumbang

Sumbangan dialu-alukan! Sila baca [Panduan Menyumbang](CONTRIBUTING.md).

### Lesen

Lesen MIT - Lihat fail [LICENSE](LICENSE) untuk butiran.

### Hubungi Kami

- Laman utama projek: https://o2ochat.io
- Penjejak isu: https://github.com/yourusername/o2ochat/issues
- Emel: support@o2ochat.io

---


---

### ⚠️ Amaran Risiko Undang-undang

**Pemberitahuan Penting: Projek ini untuk tujuan pendidikan sahaja**

- 📚 **Tujuan Pendidikan** - Demonstrasi komunikasi P2P dan penyulitan
- ⚖️ **Pematuhan Undang-undang** - Pengguna mesti mematuhi undang-undang tempatan
- 🚫 **Tiada Penggunaan Haram** - Dilarang untuk aktiviti haram
- 📝 **Tanggungjawab Pengguna** - Pengguna memikul tanggungjawab undang-undang
- 🔒 **Keberkecualian Teknologi** - Teknologi adalah neutral

**Dengan menggunakan projek ini, anda bersetuju:**
1. Hanya untuk tujuan yang sah
2. Tidak terlibat dalam aktiviti haram
3. Menerima risiko teknikal
4. Mematuhi [Terma](TERMS_OF_SERVICE.md) dan [Privasi](PRIVACY.md)

Lihat: [Notis Keselamatan](SECURITY_NOTICE.md)

---

<p align="center">
  <b>P2P Tulen | Penyulitan Hujung-ke-Hujung | Komunikasi Bebas</b>
</p>
