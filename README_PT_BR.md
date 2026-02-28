# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[繁體中文](README_ZH_TW.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)** | **[Português](README_PT_BR.md)** | **[Italiano](README_IT.md)**

## Software de Mensagens Instantâneas P2P Puro

O2OChat é um software de mensagens instantâneas peer-to-peer (P2P) puro que não depende de servidores centrais para armazenar mensagens. Todas as comunicações ocorrem diretamente entre os usuários.

### Recursos Principais

- 🔒 **Criptografia de Ponta a Ponta** - Todas as mensagens usam criptografia AES-256-GCM
- 🌐 **Arquitetura P2P Pura** - Sem servidor central, comunicação direta
- 📱 **Suporte Multiplataforma** - Android, iOS, Windows, Linux, macOS, HarmonyOS
- 📁 **Transferência de Arquivos** - Retomar transferências interrompidas, download de múltiplas fontes, verificação de árvore Merkle
- 🌍 **16 Idiomas** - Chinês, Inglês, Japonês, Coreano, Alemão, Francês, Espanhol, Russo, Malaio, Hebraico, Árabe, Tibetano, Mongol, Uigur, Chinês Tradicional, Português

### Suporte a Múltiplos Sistemas Operacionais

O2OChat suporta todos os principais sistemas operacionais, fornecendo aplicativos nativos e uma experiência de usuário unificada:

| Sistema Operacional | Tipo de Aplicativo | Stack Tecnológico | Status |
|--------------------|-------------------|------------------|--------|
| **Android** | Aplicativo Nativo | Kotlin + Jetpack Compose | ✅ Disponível |
| **iOS** | Aplicativo Nativo | Swift + SwiftUI | ✅ Disponível |
| **HarmonyOS** | Aplicativo Nativo | ArkTS + ArkUI | ✅ Disponível |
| **Windows** | Aplicativo Desktop | Go + Fyne | ✅ Disponível |
| **macOS** | Aplicativo Desktop | Go + Fyne/SwiftUI | ✅ Disponível |
| **Linux** | Aplicativo Desktop | Go + Fyne | ✅ Disponível |

#### Recursos da Plataforma

- **Mobile** (Android/iOS/HarmonyOS): Experiência mobile completa, suporte a notificações push, execução em segundo plano, mensagens offline
- **Desktop** (Windows/macOS/Linux): Experiência desktop completa, suporte a múltiplas janelas, arrastar e soltar arquivos, atalhos de teclado
- **Arquitetura Unificada**: Todas as plataformas compartilham a mesma biblioteca core P2P, garantindo experiência de comunicação consistente
- **Sincronização de Dados**: Mesma conta pode fazer login em múltiplos dispositivos, mensagens sincronizam automaticamente

### Início Rápido

```bash
# Clonar o projeto
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# Compilar
go build -o o2ochat ./cmd/o2ochat

# Executar
./o2ochat
```

### Estrutura do Projeto

```
o2ochat/
├── cmd/              # Pontos de entrada
├── pkg/              # Bibliotecas core
│   ├── identity/     # Gerenciamento de identidade
│   ├── transport/    # Transporte de rede
│   ├── signaling/    # Serviço de sinalização
│   ├── crypto/       # Módulo de criptografia
│   ├── storage/      # Armazenamento de dados
│   ├── filetransfer/ # Transferência de arquivos
│   └── media/        # Processamento de áudio/vídeo
├── ui/               # Interface do usuário
├── cli/              # Ferramentas de linha de comando
├── tests/            # Testes
├── docs/             # Documentação
└── scripts/          # Scripts de build
```

### Stack Tecnológico

- **Go 1.21+** - Core backend
- **Protocol Buffers** - Serialização
- **QUIC/WebRTC** - Transporte P2P
- **SQLite** - Armazenamento local
- **Fyne** - GUI desktop
- **Jetpack Compose** - UI Android
- **SwiftUI** - UI iOS
- **ArkTS** - UI HarmonyOS

### Contribuindo

Contribuições são bem-vindas! Por favor leia o [Guia de Contribuição](CONTRIBUTING.md).

### Licença

Licença MIT - Veja o arquivo [LICENSE](LICENSE) para detalhes.

### Contato

- Página do Projeto: https://o2ochat.io
- Rastreador de Issues: https://github.com/yourusername/o2ochat/issues
- Email: support@o2ochat.io

---

### Documentos Relacionados

- [Política de Privacidade](PRIVACY.md)
- [Termos de Serviço](TERMS_OF_SERVICE.md)
- [Instruções de Segurança](SECURITY_NOTICE.md)
- [Guia de Início Rápido](QUICKSTART.md)
- [Documentação de Arquitetura](ARCHITECTURE.md)
- [Guia de Desenvolvimento](DEVELOPMENT_GUIDE.md)

---

<p align="center">
  <b>P2P Puro | Criptografia de Ponta a Ponta | Comunicação Livre</b>
</p>

---

**Versão**: v1.0.0  
**Última Atualização**: 28 de Fevereiro de 2026  
**Status**: ✅ Concluído
