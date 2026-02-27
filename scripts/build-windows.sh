#!/bin/bash
# Windows 构建脚本

set -e

echo "=== O2OChat Windows 构建脚本 ==="

# 设置环境变量
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1

# Windows 编译器设置
if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    export CC=x86_64-w64-mingw32-gcc
    export CXX=x86_64-w64-mingw32-g++
fi

# 获取版本
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# 构建标志
LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -H windowsgui"

echo ""
echo "构建信息:"
echo "  版本: $VERSION"
echo "  构建时间: $BUILD_TIME"
echo "  目标: Windows amd64"
echo ""

# 创建输出目录
mkdir -p dist

echo "开始构建..."
if go build -ldflags "$LDFLAGS" -o dist/o2ochat.exe ./cmd/o2ochat; then
    echo "✓ 构建成功: dist/o2ochat.exe"
    
    # 显示文件信息
    ls -lh dist/o2ochat.exe
    
    echo ""
    echo "Windows 构建完成！"
    echo "输出文件: dist/o2ochat.exe"
else
    echo "✗ 构建失败"
    exit 1
fi
