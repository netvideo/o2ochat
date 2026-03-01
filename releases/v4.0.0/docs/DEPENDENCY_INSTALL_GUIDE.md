# O2OChat 依赖安装指南

本指南介绍如何在不使用 `go get` 的情况下安装项目依赖。

## 方法一：使用 git clone（推荐）

### 1. 克隆主项目
```bash
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat
```

### 2. 使用 go mod 自动下载依赖
```bash
# 下载所有依赖（使用 go mod）
go mod download

# 或者使用 vendor 模式
go mod vendor
```

### 3. 手动克隆特定依赖（如果需要）
```bash
# 示例：如果需要修改某个依赖
mkdir -p ../deps
cd ../deps

# 克隆常用依赖
git clone https://github.com/gorilla/websocket.git
git clone https://github.com/mattn/go-sqlite3.git
git clone https://github.com/stretchr/testify.git

# 然后在 go.mod 中使用 replace 指向本地版本
```

## 方法二：使用 go install（Go 1.17+）

```bash
# 安装命令行工具
go install github.com/some/tool@latest

# 示例：安装 protobuf 编译器
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

## 方法三：使用包管理器

### macOS (Homebrew)
```bash
brew install protobuf
brew install sqlite3
```

### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install -y protobuf-compiler libsqlite3-dev
```

### Windows (Chocolatey)
```powershell
choco install protobuf
choco install sqlite
```

## 方法四：使用 Docker

```bash
# 拉取 Go 开发环境镜像
docker pull golang:1.22

# 运行容器并挂载项目目录
docker run -it --rm -v $(pwd):/workspace -w /workspace golang:1.22 bash

# 在容器内运行
root@container:/workspace# go mod download
root@container:/workspace# go build -o o2ochat ./cmd/o2ochat
```

## 方法五：手动下载预编译二进制文件

对于某些工具，可以直接下载预编译的二进制文件：

```bash
# 示例：下载 protoc
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v3.20.1/protoc-3.20.1-linux-x86_64.zip
unzip protoc-3.20.1-linux-x86_64.zip -d /usr/local
```

## 推荐的工作流程

### 开发环境设置
```bash
# 1. 克隆项目
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# 2. 下载依赖（使用 go mod，无需 go get）
go mod download

# 3. 验证依赖
ls -la go.sum
go list -m all

# 4. 构建项目
go build -o o2ochat ./cmd/o2ochat

# 5. 运行测试
go test ./...
```

### 添加新依赖
```bash
# 方法1: 直接在 go.mod 中添加，然后运行
go mod tidy

# 方法2: 使用 go edit（需要安装）
go mod edit -require=github.com/example/package@v1.0.0
go mod tidy

# 方法3: 手动编辑 go.mod，然后运行
go mod download
go mod verify
```

## 常见问题

### Q: 为什么避免使用 `go get`？
A: 从 Go 1.17 开始，`go get` 的行为发生了变化。对于依赖管理，推荐使用 `go mod download` 和 `go mod tidy`。

### Q: 如何在没有网络的情况下构建？
A: 使用 `go mod vendor` 创建 vendor 目录，然后使用 `-mod=vendor` 标志构建。

### Q: 依赖版本冲突怎么办？
A: 使用 `go mod graph` 查看依赖图，然后使用 `replace` 指令替换特定版本。

## 参考资源

- [Go Modules Reference](https://golang.org/ref/mod)
- [Migrating to Go Modules](https://blog.golang.org/migrating-to-go-modules)
- [Go Command Documentation](https://golang.org/cmd/go/)

---

**注意**: 本文档假设您使用的是 Go 1.18 或更高版本。对于旧版本，某些命令可能略有不同。
