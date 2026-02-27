# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)**

## Logiciel de Messagerie Instantanée Pure P2P

O2OChat est un logiciel de messagerie instantanée purement peer-to-peer (P2P) qui ne dépend pas de serveurs centraux pour stocker les messages. Toutes les communications se font directement entre les utilisateurs.

### Fonctionnalités Principales

- 🔒 **Chiffrement de Bout en Bout** - Tous les messages utilisent le chiffrement AES-256-GCM
- 🌐 **Architecture Pure P2P** - Pas de serveur central, communication directe
- 📱 **Support Multi-Plateforme** - Android, iOS, Windows, Linux, macOS, HarmonyOS
- 📁 **Transfert de Fichiers** - Reprise de transfert interrompu, téléchargement multi-source, vérification par arbre de Merkle
- 🌍 **16 Langues** - Chinois, Anglais, Japonais, Coréen, Allemand, Français, Espagnol, Russe, Malais, Hébreu, Arabe, Tibétain, Mongol, Ouïghour, Chinois (Traditionnel)

### Démarrage Rapide

```bash
# Cloner le projet
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# Construire
go build -o o2ochat ./cmd/o2ochat

# Exécuter
./o2ochat
```

### Structure du Projet

```
o2ochat/
├── cmd/              # Points d'entrée
├── pkg/              # Bibliothèques principales
│   ├── identity/     # Gestion des identités
│   ├── transport/    # Transport réseau
│   ├── signaling/    # Service de signalisation
│   ├── crypto/       # Module de chiffrement
│   ├── storage/      # Stockage de données
│   ├── filetransfer/ # Transfert de fichiers
│   └── media/        # Traitement audio/vidéo
├── ui/               # Interface utilisateur
├── cli/              # Outils en ligne de commande
├── tests/            # Tests
├── docs/             # Documentation
└── scripts/          # Scripts de construction
```

### Stack Technologique

- **Go 1.21+** - Cœur backend
- **Protocol Buffers** - Sérialisation
- **QUIC/WebRTC** - Transport P2P
- **SQLite** - Stockage local
- **Fyne** - GUI bureau
- **Jetpack Compose** - UI Android
- **SwiftUI** - UI iOS
- **ArkTS** - UI HarmonyOS

### Contribuer

Les contributions sont les bienvenues ! Veuillez lire le [Guide de Contribution](CONTRIBUTING.md).

### Licence

Licence MIT - Voir le fichier [LICENSE](LICENSE) pour plus de détails.

### Contact

- Page d'accueil du projet : https://o2ochat.io
- Suivi des problèmes : https://github.com/yourusername/o2ochat/issues
- Email : support@o2ochat.io

---

<p align="center">
  <b>Pur P2P | Chiffrement de Bout en Bout | Communication Libre</b>
</p>
