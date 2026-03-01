# O2OChat v4.0.0 Installation Guide

## Quick Install

### Linux
```bash
# Download
wget https://github.com/netvideo/o2ochat/releases/download/v4.0.0/o2ochat-linux-amd64

# Make executable
chmod +x o2ochat-linux-amd64

# Run
./o2ochat-linux-amd64
```

### Windows
1. Download o2ochat-windows-amd64.exe
2. Run the executable
3. (Optional) Add to PATH

### macOS
```bash
# Download
curl -L -o o2ochat-macos https://github.com/netvideo/o2ochat/releases/download/v4.0.0/o2ochat-macos-amd64

# Make executable
chmod +x o2ochat-macos

# Run
./o2ochat-macos
```

### Docker
```bash
docker pull ghcr.io/netvideo/o2ochat:4.0.0
docker run -d -p 8080:8080 ghcr.io/netvideo/o2ochat:4.0.0
```

### Kubernetes
```bash
kubectl apply -f https://raw.githubusercontent.com/netvideo/o2ochat/v4.0.0/k8s/o2ochat.yaml
```

## Post-Installation

1. Configure settings in config.yaml
2. Start the application
3. Access web interface at http://localhost:8080

## Troubleshooting

See TROUBLESHOOTING.md for common issues.
