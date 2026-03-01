# O2OChat Linux Desktop Application

A cross-platform P2P instant messaging desktop client for Linux built with Go and Fyne.

## Features

- P2P messaging with end-to-end encryption
- Contact management
- Real-time messaging via WebSocket signaling
- File transfer support
- Voice and video calls (planned)
- Cross-platform GUI using Fyne

## Prerequisites

- Go 1.22 or higher
- GCC or C compiler (for Fyne dependencies)
- Linux desktop environment (GTK3)

### Install Go

```bash
# Download and install Go 1.22
wget https://go.dev/dl/go1.22.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Install Fyne Dependencies

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y libgl1-mesa-glx libglib2.0-dev libgtk-3-dev

# Fedora
sudo dnf install -y mesa-libGLU glib2-devel gtk3-devel

# Arch Linux
sudo pacman -S --noconfirm mesa libglvnd gtk3
```

## Build

```bash
cd linux/O2OChat
go mod tidy
go build -o o2ochat .
```

## Run

```bash
./o2ochat
```

Or run directly:

```bash
cd linux/O2OChat
go run main.go
```

## Project Structure

```
linux/O2OChat/
├── go.mod          # Go module definition
└── main.go         # Main application code
```

## Configuration

The application stores data in:
- `~/.local/share/o2ochat/` (data files)
- `~/.config/o2ochat/` (configuration)

## Usage

1. **First Launch**: The app automatically creates an identity with a unique Peer ID
2. **Add Contacts**: Click "Add Contact" and enter the other user's Peer ID
3. **Send Messages**: Select a contact and type a message
4. **Settings**: Change your nickname

## Architecture

```
┌─────────────────────────────────────────┐
│              O2OChat App                │
├─────────────────────────────────────────┤
│  UI Layer (Fyne)                        │
│  - Main window, chat window, dialogs    │
├─────────────────────────────────────────┤
│  Service Layer                           │
│  - Identity management                  │
│  - Signaling client (WebSocket)         │
│  - Contact management                   │
├─────────────────────────────────────────┤
│  Core Modules                            │
│  - identity: Key generation, Peer ID     │
│  - signaling: WebSocket communication  │
│  - storage: Local data persistence      │
│  - transport: P2P connections           │
└─────────────────────────────────────────┘
```

## Signalling Server

Default: `wss://signal.o2ochat.io`

## Development

### Run with hot reload

```bash
go run main.go
```

### Build release

```bash
GOOS=linux GOARCH=amd64 go build -o o2ochat-linux-amd64 .
```

### Cross-compile for other platforms

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o o2ochat.exe .

# macOS
GOOS=darwin GOARCH=amd64 go build -o o2ochat-darwin .
```

## Troubleshooting

### GUI not showing

Make sure you have a desktop environment and X11/Wayland running:
```bash
export WAYLAND_DISPLAY=wayland-0  # For Wayland
export DISPLAY=:0                 # For X11
```

### Build errors

Ensure all dependencies are installed:
```bash
go mod download
go mod tidy
```

## License

Same as main O2OChat project.
