# O2OChat 依赖解决方案

由于网络限制，`go mod` 和 `go get` 可能无法正常工作。本方案提供完整的离线依赖管理。

## 方法一：使用本地 Vendor 目录（推荐）

### 1. 创建 Vendor 目录

在项目根目录创建 `vendor/` 目录：

```bash
mkdir -p vendor
```

### 2. 手动下载依赖

使用 `git clone` 下载每个依赖到 vendor 目录：

```bash
cd vendor

# Core dependencies
git clone https://github.com/gorilla/websocket.git
git clone https://github.com/mattn/go-sqlite3.git
git clone https://github.com/stretchr/testify.git

# Additional dependencies if needed
git clone https://github.com/golang/protobuf.git
git clone https://github.com/google/uuid.git
git clone https://github.com/spf13/cobra.git
```

### 3. 修改 go.mod 使用本地依赖

在 `go.mod` 末尾添加 `replace` 指令：

```go
replace (
	github.com/gorilla/websocket => ./vendor/websocket
	github.com/mattn/go-sqlite3 => ./vendor/go-sqlite3
	github.com/stretchr/testify => ./vendor/testify
)
```

### 4. 使用 Vendor 模式构建

```bash
# 使用 -mod=vendor 标志
go build -mod=vendor -o o2ochat ./cmd/o2ochat

# 或者设置环境变量
export GOFLAGS="-mod=vendor"
go build -o o2ochat ./cmd/o2ochat
```

---

## 方法二：使用 Docker（推荐用于 CI/CD）

### 1. 创建 Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖（使用代理如果可用）
ENV GOPROXY=https://proxy.golang.org,direct
RUN go mod download || echo "Using local vendor..."

# 复制源代码
COPY . .

# 构建（优先使用 vendor）
RUN if [ -d "vendor" ]; then \
        go build -mod=vendor -o o2ochat ./cmd/o2ochat; \
    else \
        go build -o o2ochat ./cmd/o2ochat; \
    fi

# 运行阶段
FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite-libs
WORKDIR /app
COPY --from=builder /build/o2ochat .
ENTRYPOINT ["./o2ochat"]
```

### 2. 使用 Docker 构建

```bash
# 构建镜像
docker build -t o2ochat:latest .

# 运行容器
docker run -it --rm -v $(pwd)/data:/app/data o2ochat:latest
```

---

## 方法三：使用预编译二进制文件

### 1. 创建 Makefile

```makefile
.PHONY: all build clean vendor docker

# 默认目标
all: vendor build

# 创建 vendor 目录
vendor:
	@echo "Creating vendor directory..."
	@mkdir -p vendor
	@if [ ! -d "vendor/websocket" ]; then \
		git clone --depth 1 https://github.com/gorilla/websocket.git vendor/websocket; \
	fi
	@if [ ! -d "vendor/go-sqlite3" ]; then \
		git clone --depth 1 https://github.com/mattn/go-sqlite3.git vendor/go-sqlite3; \
	fi
	@if [ ! -d "vendor/testify" ]; then \
		git clone --depth 1 https://github.com/stretchr/testify.git vendor/testify; \
	fi
	@echo "Vendor dependencies ready!"

# 构建项目
build:
	@echo "Building O2OChat..."
	@if [ -d "vendor" ]; then \
		echo "Using vendor mode..."; \
		go build -mod=vendor -o o2ochat ./cmd/o2ochat; \
	else \
		echo "Using module mode..."; \
		go build -o o2ochat ./cmd/o2ochat; \
	fi
	@echo "Build complete: ./o2ochat"

# Docker 构建
docker:
	@echo "Building Docker image..."
	@docker build -t o2ochat:latest .
	@echo "Docker image built: o2ochat:latest"

# 清理
clean:
	@echo "Cleaning up..."
	@rm -f o2ochat
	@rm -rf vendor
	@docker rmi o2ochat:latest 2>/dev/null || true
	@echo "Cleanup complete!"

# 帮助
help:
	@echo "Available targets:"
	@echo "  make vendor  - Clone dependencies to vendor/"
	@echo "  make build   - Build the project"
	@echo "  make docker  - Build Docker image"
	@echo "  make clean   - Clean all build artifacts"
	@echo "  make help    - Show this help"
```

### 2. 使用 Makefile

```bash
# 完整构建流程
make vendor  # 下载依赖
make build   # 构建项目

# 或者一键完成
make all     # vendor + build

# Docker 构建
make docker

# 清理
make clean
```

---

## 方法四：Go 工作区模式 (Go 1.18+)

对于同时开发多个相关项目的情况：

```bash
# 创建 Go 工作区
go work init
go work use .

# 添加本地依赖（如果使用 git clone 到本地）
go work use ../websocket
go work use ../go-sqlite3
go work use ../testify

# 构建时会优先使用本地版本
go build -o o2ochat ./cmd/o2ochat
```

---

## 总结对比表

| 方法 | 适用场景 | 优点 | 缺点 | 网络要求 |
|------|----------|------|------|----------|
| **Vendor 目录** | 开发环境 | 完全离线，版本可控 | 占用磁盘空间 | 一次性克隆 |
| **Docker** | CI/CD，生产 | 环境一致，可重现 | 需要 Docker | 构建时联网 |
| **Makefile** | 团队协作 | 标准化流程，易于使用 | 需要配置 | 一次性克隆 |
| **Go 工作区** | 多项目开发 | 本地修改即时生效 | Go 1.18+ | 可选 |

---

## 推荐方案

### 开发环境（离线/内网）
```bash
# 1. 使用 Makefile + Vendor
make vendor    # 一次性下载所有依赖
make build     # 离线构建

# 2. 后续开发只需
make build     # 完全离线工作
```

### CI/CD 环境
```bash
# 使用 Docker 确保环境一致
docker build -t o2ochat:latest .
docker run o2ochat:latest
```

### 生产部署
```bash
# 使用预编译二进制文件
make build
./o2ochat --config=production.yaml
```

---

**所有方案均支持离线/内网环境，无需 `go get`！**