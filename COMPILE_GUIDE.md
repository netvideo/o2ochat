# O2OChat v4.0.0 编译指南

## 环境要求

### 必需
- Go 1.22+
- Git
- 网络连接 (下载依赖)

### 可选
- Docker (容器化编译)
- Make (自动化编译)

---

## 快速编译

### 1. Linux (AMD64)

```bash
cd /mnt/f/o2ochat
GOOS=linux GOARCH=amd64 go build -o o2ochat-linux ./cmd/o2ochat
```

### 2. Windows (AMD64)

```bash
cd /mnt/f/o2ochat
GOOS=windows GOARCH=amd64 go build -o o2ochat-windows.exe ./cmd/o2ochat
```

### 3. macOS (AMD64)

```bash
cd /mnt/f/o2ochat
GOOS=darwin GOARCH=amd64 go build -o o2ochat-macos ./cmd/o2ochat
```

### 4. macOS (ARM64/Apple Silicon)

```bash
cd /mnt/f/o2ochat
GOOS=darwin GOARCH=arm64 go build -o o2ochat-macos-arm ./cmd/o2ochat
```

---

## 完整编译 (所有平台)

### 使用编译脚本

```bash
cd /mnt/f/o2ochat
./build_release.sh
```

### 手动编译所有平台

```bash
cd /mnt/f/o2ochat
mkdir -p releases/v4.0.0/binaries

# Linux
GOOS=linux GOARCH=amd64 go build -o releases/v4.0.0/binaries/o2ochat-linux-amd64 ./cmd/o2ochat
GOOS=linux GOARCH=arm64 go build -o releases/v4.0.0/binaries/o2ochat-linux-arm64 ./cmd/o2ochat

# Windows
GOOS=windows GOARCH=amd64 go build -o releases/v4.0.0/binaries/o2ochat-windows-amd64.exe ./cmd/o2ochat
GOOS=windows GOARCH=386 go build -o releases/v4.0.0/binaries/o2ochat-windows-386.exe ./cmd/o2ochat

# macOS
GOOS=darwin GOARCH=amd64 go build -o releases/v4.0.0/binaries/o2ochat-macos-amd64 ./cmd/o2ochat
GOOS=darwin GOARCH=arm64 go build -o releases/v4.0.0/binaries/o2ochat-macos-arm64 ./cmd/o2ochat

# ARM (Raspberry Pi 等)
GOOS=linux GOARCH=arm GOARM=7 go build -o releases/v4.0.0/binaries/o2ochat-linux-armv7 ./cmd/o2ochat
```

---

## Docker 编译

### 使用 Docker 编译

```bash
cd /mnt/f/o2ochat
docker build -t o2ochat:4.0.0 .
```

### 使用 Docker Compose

```bash
cd /mnt/f/o2ochat
docker-compose up --build
```

---

## 交叉编译

### 从 Linux 编译 Windows

```bash
GOOS=windows GOARCH=amd64 go build -o o2ochat.exe ./cmd/o2ochat
```

### 从 macOS 编译 Linux

```bash
GOOS=linux GOARCH=amd64 go build -o o2ochat ./cmd/o2ochat
```

---

## 优化编译

### 减小二进制大小

```bash
# 去掉调试信息
go build -ldflags="-s -w" -o o2ochat ./cmd/o2ochat

# 去掉符号表
go build -ldflags="-s -w -extldflags=-static" -o o2ochat ./cmd/o2ochat
```

### 启用优化

```bash
# 启用编译器优化
go build -gcflags="-O" -o o2ochat ./cmd/o2ochat
```

---

## 验证编译

### 检查版本

```bash
./o2ochat --version
```

### 运行测试

```bash
go test ./...
```

### 检查二进制

```bash
file o2ochat
# 输出示例: o2ochat: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, not stripped
```

---

## 常见问题

### Q: 编译时提示找不到包？

```bash
# 下载依赖
go mod download
go mod tidy
```

### Q: 编译速度慢？

```bash
# 使用国内代理
export GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# 启用编译缓存
export GOCACHE=/tmp/go-build-cache
```

### Q: 编译后文件太大？

```bash
# 使用 UPX 压缩
upx --best o2ochat

# 或使用 ldflags 去掉调试信息
go build -ldflags="-s -w" -o o2ochat ./cmd/o2ochat
```

---

## 编译输出

编译成功后，将在以下位置生成文件：

```
releases/v4.0.0/binaries/
├── o2ochat-linux-amd64        # Linux 64-bit
├── o2ochat-linux-arm64        # Linux ARM 64-bit
├── o2ochat-windows-amd64.exe  # Windows 64-bit
├── o2ochat-windows-386.exe    # Windows 32-bit
├── o2ochat-macos-amd64        # macOS Intel
├── o2ochat-macos-arm64        # macOS Apple Silicon
└── o2ochat-linux-armv7        # ARM v7 (Raspberry Pi)
```

---

## 下一步

编译完成后：

1. 运行测试：`go test ./...`
2. 生成校验和：`sha256sum binaries/* > SHA256SUMS.txt`
3. 创建 Release：在 GitHub 上创建 Release
4. 上传文件：上传编译后的二进制文件

---

**编译准备就绪！**
