# O2OChat v4.0.0 Build Instructions

## Prerequisites

- Go 1.22+
- Git
- Make (optional)

## Build Commands

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -o o2ochat-linux ./cmd/o2ochat
```

### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o o2ochat-windows.exe ./cmd/o2ochat
```

### macOS
```bash
GOOS=darwin GOARCH=amd64 go build -o o2ochat-macos ./cmd/o2ochat
```

### Docker
```bash
docker build -t o2ochat:4.0.0 .
docker-compose up -d
```

### Kubernetes
```bash
kubectl apply -f k8s/o2ochat.yaml
```

## Verify Build

```bash
./o2ochat --version
```

## Installation

See INSTALL.md for detailed installation instructions.
