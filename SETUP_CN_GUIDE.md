# O2OChat v4.0.0 国内镜像环境搭建指南

使用国内镜像，加速下载和编译！

---

## 快速开始

### 1. 配置国内镜像

```bash
# 设置 GOPROXY (阿里云镜像)
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# 设置 GOSUMDB (可选)
go env -w GOSUMDB=off

# 验证配置
go env GOPROXY
```

### 2. 安装 Go (国内)

#### 方法 1: 使用 Go 中国镜像

```bash
# 下载 Go 1.22 (使用国内镜像)
cd /tmp
wget https://golang.google.cn/dl/go1.22.4.linux-amd64.tar.gz

# 解压安装
sudo tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz

# 添加到 PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### 方法 2: 使用包管理器

```bash
# Ubuntu/Debian (使用国内源)
sudo apt update
sudo apt install golang-go

# CentOS/RHEL (使用国内源)
sudo yum install golang
```

### 3. 下载依赖 (使用国内镜像)

```bash
cd /mnt/f/o2ochat

# 清理缓存 (可选)
go clean -modcache

# 下载依赖 (使用阿里云镜像)
go mod download

# 整理依赖
go mod tidy

# 验证依赖
go mod verify
```

### 4. 编译 (使用国内镜像)

```bash
# 设置国内镜像
export GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
export GOSUMDB=off

# 编译 Linux 版本
go build -v -o o2ochat-linux ./cmd/o2ochat

# 编译 Windows 版本
GOOS=windows GOARCH=amd64 go build -v -o o2ochat-windows.exe ./cmd/o2ochat

# 编译 macOS 版本
GOOS=darwin GOARCH=amd64 go build -v -o o2ochat-macos ./cmd/o2ochat
```

---

## 自动化安装脚本

保存为 setup_cn.sh:

```bash
#!/bin/bash
set -e

echo "开始安装 O2OChat 环境..."

# 1. 配置 GOPROXY
echo "配置 GOPROXY..."
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
go env -w GOSUMDB=off
echo "GOPROXY 配置完成"

# 2. 检查 Go 安装
echo "检查 Go 安装..."
if ! command -v go &> /dev/null; then
    echo "Go 未安装，请先安装 Go 1.22+"
    exit 1
fi
go version
echo "Go 已安装"

# 3. 下载依赖
echo "下载依赖..."
go mod download
go mod tidy
echo "依赖下载完成"

# 4. 编译
echo "开始编译..."
go build -v -o o2ochat ./cmd/o2ochat
echo "编译完成"

# 5. 验证
echo "验证安装..."
./o2ochat --version
echo "验证完成"

echo ""
echo "O2OChat 环境搭建完成！"
```

运行：

```bash
chmod +x setup_cn.sh
./setup_cn.sh
```

---

## 国内镜像源

### GOPROXY 镜像

| 镜像源 | URL |
|--------|-----|
| 阿里云 | https://mirrors.aliyun.com/goproxy/ |
| 七牛云 | https://goproxy.cn |
| 腾讯云 | https://mirrors.cloud.tencent.com/go/ |

### Docker 镜像

| 镜像源 | URL |
|--------|-----|
| 阿里云 | registry.cn-hangzhou.aliyuncs.com |
| 腾讯云 | mirror.ccs.tencentyun.com |
| 网易 | hub-mirror.c.163.com |

---

## 验证安装

### 1. 检查 Go 版本

```bash
go version
```

### 2. 检查 GOPROXY

```bash
go env GOPROXY
```

### 3. 测试编译

```bash
cd /mnt/f/o2ochat
go build -v -o o2ochat-test ./cmd/o2ochat
./o2ochat-test --version
```

---

## 常见问题

### Q: 下载依赖速度慢？

```bash
# 使用阿里云镜像
export GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# 清理缓存重新下载
go clean -modcache
go mod download
```

### Q: 校验和验证失败？

```bash
# 临时禁用校验和验证
export GOSUMDB=off

# 删除校验和文件
rm go.sum
go mod tidy
```

### Q: 编译时找不到包？

```bash
# 更新依赖
go get -u ./...

# 整理依赖
go mod tidy
```

---

使用国内镜像，编译速度提升 10 倍！

准备就绪，开始编译！
