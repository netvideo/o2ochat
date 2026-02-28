# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[繁體中文](README_ZH_TW.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)** | **[Português](README_PT_BR.md)** | **[Italiano](README_IT.md)**

## Software di Messaggistica Istantanea P2P Puro

O2OChat è un software di messaggistica istantanea peer-to-peer (P2P) puro che non dipende da server centrali per archiviare i messaggi. Tutte le comunicazioni avvengono direttamente tra gli utenti.

### Caratteristiche Principali

- 🔒 **Crittografia End-to-End** - Tutti i messaggi utilizzano crittografia AES-256-GCM
- 🌐 **Architettura P2P Pura** - Nessun server centrale, comunicazione diretta
- 📱 **Supporto Multipiattaforma** - Android, iOS, Windows, Linux, macOS, HarmonyOS
- 📁 **Trasferimento File** - Riprendi trasferimenti interrotti, download da più fonti, verifica albero di Merkle
- 🌍 **16 Lingue** - Cinese, Inglese, Giapponese, Coreano, Tedesco, Francese, Spagnolo, Russo, Malese, Ebraico, Arabo, Tibetano, Mongolo, Uiguro, Cinese Tradizionale, Italiano

### Supporto per Sistemi Operativi Multipli

O2OChat supporta tutti i principali sistemi operativi, fornendo applicazioni native e un'esperienza utente unificata:

| Sistema Operativo | Tipo di Applicazione | Stack Tecnologico | Stato |
|------------------|---------------------|------------------|-------|
| **Android** | Applicazione Nativa | Kotlin + Jetpack Compose | ✅ Disponibile |
| **iOS** | Applicazione Nativa | Swift + SwiftUI | ✅ Disponibile |
| **HarmonyOS** | Applicazione Nativa | ArkTS + ArkUI | ✅ Disponibile |
| **Windows** | Applicazione Desktop | Go + Fyne | ✅ Disponibile |
| **macOS** | Applicazione Desktop | Go + Fyne/SwiftUI | ✅ Disponibile |
| **Linux** | Applicazione Desktop | Go + Fyne | ✅ Disponibile |

#### Caratteristiche della Piattaforma

- **Mobile** (Android/iOS/HarmonyOS): Esperienza mobile completa, supporto notifiche push, esecuzione in background, messaggi offline
- **Desktop** (Windows/macOS/Linux): Esperienza desktop completa, supporto multi-finestra, trascina e rilascia file, scorciatoie da tastiera
- **Architettura Unificata**: Tutte le piattaforme condividono la stessa libreria core P2P, garantendo esperienza di comunicazione coerente
- **Sincronizzazione Dati**: Lo stesso account può accedere da più dispositivi, i messaggi si sincronizzano automaticamente

### Avvio Rapido

```bash
# Clona il progetto
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# Compila
go build -o o2ochat ./cmd/o2ochat

# Esegui
./o2ochat
```

### Struttura del Progetto

```
o2ochat/
├── cmd/              # Punti di ingresso
├── pkg/              # Librerie core
│   ├── identity/     # Gestione identità
│   ├── transport/    # Trasporto di rete
│   ├── signaling/    # Servizio di segnalazione
│   ├── crypto/       # Modulo crittografia
│   ├── storage/      # Archiviazione dati
│   ├── filetransfer/ # Trasferimento file
│   └── media/        # Elaborazione audio/video
├── ui/               # Interfaccia utente
├── cli/              # Strumenti da riga di comando
├── tests/            # Test
├── docs/             # Documentazione
└── scripts/          # Script di compilazione
```

### Stack Tecnologico

- **Go 1.21+** - Core backend
- **Protocol Buffers** - Serializzazione
- **QUIC/WebRTC** - Trasporto P2P
- **SQLite** - Archiviazione locale
- **Fyne** - GUI desktop
- **Jetpack Compose** - UI Android
- **SwiftUI** - UI iOS
- **ArkTS** - UI HarmonyOS

### Contribuire

I contributi sono benvenuti! Per favore leggi la [Guida ai Contributi](CONTRIBUTING.md).

### Licenza

Licenza MIT - Vedi il file [LICENSE](LICENSE) per i dettagli.

### Contatti

- Pagina del Progetto: https://o2ochat.io
- Issue Tracker: https://github.com/yourusername/o2ochat/issues
- Email: support@o2ochat.io

---

### Documenti Correlati

- [Informativa sulla Privacy](PRIVACY.md)
- [Termini di Servizio](TERMS_OF_SERVICE.md)
- [Istruzioni di Sicurezza](SECURITY_NOTICE.md)
- [Guida Rapida](QUICKSTART.md)
- [Documentazione Architettura](ARCHITECTURE.md)
- [Guida allo Sviluppo](DEVELOPMENT_GUIDE.md)

---


---

### ⚠️ Avviso di Rischio Legale

**Avviso Importante: Questo progetto è solo per scopi educativi**

- 📚 **Scopo Educativo** - Dimostrazione di comunicazione P2P e crittografia
- ⚖️ **Conformità Legale** - Gli utenti devono rispettare le leggi locali
- 🚫 **Nessun Uso Illegale** - Vietato per attività illegali
- 📝 **Responsabilità Utente** - Gli utenti assumono responsabilità legale
- 🔒 **Neutralità Tecnologica** - La tecnologia è neutrale

**Utilizzando questo progetto, accetti:**
1. Usare solo per scopi legali
2. Non impegnarsi in attività illegali
3. Accettare i rischi tecnici
4. Rispettare [Termini](TERMS_OF_SERVICE.md) e [Privacy](PRIVACY.md)

Vedi: [Avviso di Sicurezza](SECURITY_NOTICE.md)

---

<p align="center">
  <b>P2P Puro | Crittografia End-to-End | Comunicazione Libera</b>
</p>

---

**Versione**: v1.0.0  
**Ultimo Aggiornamento**: 28 Febbraio 2026  
**Stato**: ✅ Completato
