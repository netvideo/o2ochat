# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)**

## Software de mensajería instantánea P2P puro

O2OChat es un software de mensajería instantánea puramente peer-to-peer (P2P) que no depende de servidores centrales para almacenar mensajes. Todas las comunicaciones ocurren directamente entre los usuarios.

### Características principales

- 🔒 **Cifrado de extremo a extremo** - Todos los mensajes usan cifrado AES-256-GCM
- 🌐 **Arquitectura P2P pura** - Sin servidor central, comunicación directa
- 📱 **Soporte multiplataforma** - Android, iOS, Windows, Linux, macOS, HarmonyOS
- 📁 **Transferencia de archivos** - Reanudar transferencias interrumpidas, descarga de múltiples fuentes, verificación de árbol Merkle
- 🌍 **16 idiomas** - Chino, Inglés, Japonés, Coreano, Alemán, Francés, Español, Ruso, Malayo, Hebreo, Árabe, Tibetano, Mongol, Uigur, Chino tradicional

### Inicio rápido

```bash
# Clonar el proyecto
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# Compilar
go build -o o2ochat ./cmd/o2ochat

# Ejecutar
./o2ochat
```

### Estructura del proyecto

```
o2ochat/
├── cmd/              # Puntos de entrada
├── pkg/              # Bibliotecas principales
│   ├── identity/     # Gestión de identidad
│   ├── transport/    # Transporte de red
│   ├── signaling/    # Servicio de señalización
│   ├── crypto/       # Módulo de cifrado
│   ├── storage/      # Almacenamiento de datos
│   ├── filetransfer/ # Transferencia de archivos
│   └── media/        # Procesamiento de audio/video
├── ui/               # Interfaz de usuario
├── cli/              # Herramientas de línea de comandos
├── tests/            # Pruebas
├── docs/             # Documentación
└── scripts/          # Scripts de compilación
```

### Stack tecnológico

- **Go 1.21+** - Núcleo del backend
- **Protocol Buffers** - Serialización
- **QUIC/WebRTC** - Transporte P2P
- **SQLite** - Almacenamiento local
- **Fyne** - GUI de escritorio
- **Jetpack Compose** - UI de Android
- **SwiftUI** - UI de iOS
- **ArkTS** - UI de HarmonyOS

### Contribuir

¡Las contribuciones son bienvenidas! Por favor lee la [Guía de contribución](CONTRIBUTING.md).

### Licencia

Licencia MIT - Ver el archivo [LICENSE](LICENSE) para más detalles.

### Contacto

- Página del proyecto: https://o2ochat.io
- Rastreador de problemas: https://github.com/yourusername/o2ochat/issues
- Correo electrónico: support@o2ochat.io

---


---

### ⚠️ Advertencia de Riesgo Legal

**Aviso Importante: Este proyecto es solo para fines educativos**

- 📚 **Propósito Educativo** - Demostración de comunicación P2P y cifrado
- ⚖️ **Cumplimiento Legal** - Debe cumplir con leyes locales
- 🚫 **Prohibido Uso Ilegal** - No para actividades ilegales
- 📝 **Responsabilidad** - Usuario tiene responsabilidad legal
- 🔒 **Neutralidad Tecnológica** - La tecnología es neutral

**Al usar este proyecto, aceptas:**
1. Usar solo para propósitos legales
2. No realizar actividades ilegales
3. Aceptar riesgos técnicos
4. Cumplir [Términos](TERMS_OF_SERVICE.md) y [Privacidad](PRIVACY.md)

Ver: [Aviso de Seguridad](SECURITY_NOTICE.md)

---

<p align="center">
  <b>P2P puro | Cifrado de extremo a extremo | Comunicación libre</b>
</p>
