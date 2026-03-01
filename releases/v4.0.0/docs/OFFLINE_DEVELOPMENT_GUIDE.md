# O2OChat - 离线依赖解决方案

由于网络限制，本项目提供完整的离线依赖解决方案。

## 方案一：使用本地替换（推荐）

### 1. 修改 go.mod

在 `go.mod` 文件末尾添加：

```go
// 本地替换，解决网络依赖问题
// 实际项目中应使用网络依赖
replace (
	// 示例：本地路径替换
	// github.com/gorilla/websocket => ./local/websocket
	// github.com/mattn/go-sqlite3 => ./local/go-sqlite3
	// github.com/stretchr/testify => ./local/testify
)
```

### 2. 创建本地依赖目录结构

```
local/
├── websocket/     # 本地 websocket 库
├── go-sqlite3/    # 本地 sqlite3 库
└── testify/       # 本地 testify 库
```

### 3. 使用最小化实现

对于核心功能，可以使用 Go 标准库替代：

```go
// 替代方案：
// - websocket: 使用 net/http 的 Hijack 或标准库
// - sqlite3: 使用 map + json 文件存储
// - testify: 使用标准库 testing
```

## 方案二：使用标准库实现（极简版本）

创建一个 `internal/` 目录，提供最小化实现：

```
internal/
├── websocket/    # 最小 websocket 实现
├── storage/     # 内存/文件存储
└── testing/     # 简单测试工具
```

## 方案三：Docker 完整环境

使用 Docker 镜像包含所有依赖：

```dockerfile
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# 复制本地 vendor 或预下载的依赖
COPY vendor /go/src/vendor
COPY . /workspace
WORKDIR /workspace

# 使用 vendor 模式构建
RUN go build -mod=vendor -o o2ochat ./cmd/o2ochat
```

## 推荐工作流程

### 1. 初始设置（一次性）
```bash
# 方案 A: 使用本地替换
cp go.mod go.mod.backup
echo '
replace (
    github.com/gorilla/websocket => ./local/websocket
    github.com/mattn/go-sqlite3 => ./local/go-sqlite3
)' >> go.mod

# 方案 B: 使用标准库
cd internal
mkdir -p websocket storage testing
# 创建最小实现
```

### 2. 开发工作流程
```bash
# 构建（使用本地替换）
go build -mod=mod -o o2ochat ./cmd/o2ochat

# 测试（跳过网络依赖）
go test -short ./...

# 运行
./o2ochat
```

### 3. 生产部署
```bash
# Docker 构建（包含所有依赖）
docker build -t o2ochat:latest .
docker run -d -p 8080:8080 o2ochat:latest

# 或者使用预编译二进制文件
./o2ochat --config=production.yaml
```

## 常见问题

### Q: 如何解决 "cannot find module" 错误？
A: 使用 `replace` 指令指向本地路径，或创建最小实现。

### Q: 如何在离线环境构建？
A: 使用 `vendor/` 目录 + `go build -mod=vendor`。

### Q: 如何测试没有网络依赖？
A: 使用 `go test -short` 或 mock 实现。

## 参考文档

- `DEPENDENCY_INSTALL_GUIDE.md` - 依赖安装完整指南
- `OFFLINE_BUILD_GUIDE.md` - 离线构建完整指南
- `Makefile` - 自动化构建脚本

---

**注意**: 本文档假设您处于离线/内网环境。如果有网络访问，建议优先使用 `go mod` 标准流程。
