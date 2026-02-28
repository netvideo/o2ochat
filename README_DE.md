# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)**

## Reine P2P-Instant-Messaging-Software

O2OChat ist eine reine Peer-to-Peer (P2P) Instant-Messaging-Software, die nicht auf zentralen Servern zur Speicherung von Nachrichten angewiesen ist. Alle Kommunikationen finden direkt zwischen den Benutzern statt.

### Kernfunktionen

- 🔒 **Ende-zu-Ende-Verschlüsselung** - Alle Nachrichten verwenden AES-256-GCM-Verschlüsselung
- 🌐 **Reine P2P-Architektur** - Kein zentraler Server, direkte Kommunikation
- 📱 **Multi-Plattform-Unterstützung** - Android, iOS, Windows, Linux, macOS, HarmonyOS
- 📁 **Dateiübertragung** - Unterbrechungsfreie Übertragung, Multi-Source-Download, Merkle-Tree-Verifizierung
- 🌍 **16 Sprachen** - Chinesisch, Englisch, Japanisch, Koreanisch, Deutsch, Französisch, Spanisch, Russisch, Malaiisch, Hebräisch, Arabisch, Tibetisch, Mongolisch, Uigurisch, Chinesisch (Traditionell)

### Schnellstart

```bash
# Projekt klonen
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# Bauen
go build -o o2ochat ./cmd/o2ochat

# Ausführen
./o2ochat
```

### Projektstruktur

```
o2ochat/
├── cmd/              # Einstiegspunkte
├── pkg/              # Kernbibliotheken
│   ├── identity/     # Identitätsmanagement
│   ├── transport/    # Netzwerktransport
│   ├── signaling/    # Signalisierungsdienst
│   ├── crypto/       # Verschlüsselungsmodul
│   ├── storage/      # Datenspeicherung
│   ├── filetransfer/ # Dateiübertragung
│   └── media/        # Audio/Video-Verarbeitung
├── ui/               # Benutzeroberfläche
├── cli/              # Befehlszeilentools
├── tests/            # Tests
├── docs/             # Dokumentation
└── scripts/          # Build-Skripte
```

### Technologie-Stack

- **Go 1.21+** - Backend-Kern
- **Protocol Buffers** - Serialisierung
- **QUIC/WebRTC** - P2P-Transport
- **SQLite** - Lokale Speicherung
- **Fyne** - Desktop-GUI
- **Jetpack Compose** - Android-UI
- **SwiftUI** - iOS-UI
- **ArkTS** - HarmonyOS-UI

### Mitwirken

Beiträge sind willkommen! Bitte lesen Sie den [Contributing Guide](CONTRIBUTING.md).

### Lizenz

MIT License - Siehe [LICENSE](LICENSE) Datei für Details.

### Kontakt

- Projekt-Homepage: https://o2ochat.io
- Issue-Tracker: https://github.com/yourusername/o2ochat/issues
- E-Mail: support@o2ochat.io

---


---

### ⚠️ Rechtlicher Warnhinweis

**Wichtiger Hinweis: Dieses Projekt dient nur zu Bildungszwecken**

- 📚 **Bildungszweck** - Demonstration von P2P-Kommunikation und Verschlüsselung
- ⚖️ **Einhaltung der Gesetze** - Benutzer müssen lokale Gesetze befolgen
- 🚫 **Keine illegale Nutzung** - Verboten für illegale Aktivitäten
- 📝 **Benutzerverantwortung** - Benutzer tragen rechtliche Verantwortung
- 🔒 **Technologische Neutralität** - Technologie ist neutral

**Mit der Nutzung stimmen Sie zu:**
1. Nur für legale Zwecke verwenden
2. Keine illegalen Aktivitäten durchführen
3. Technische Risiken akzeptieren
4. [Nutzungsbedingungen](TERMS_OF_SERVICE.md) und [Datenschutz](PRIVACY.md) einhalten

Siehe: [Sicherheitshinweis](SECURITY_NOTICE.md)

---

<p align="center">
  <b>Reines P2P | Ende-zu-Ende-Verschlüsselung | Freie Kommunikation</b>
</p>
