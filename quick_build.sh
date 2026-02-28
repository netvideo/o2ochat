#!/bin/bash
# quick_build.sh - 无需网络连接的快速构建脚本

echo "🔨 O2OChat 快速构建脚本 (离线模式)"
echo "==================================="

# 方法：使用 go.mod 中的 replace 指令，指向本地最小实现
echo "📦 方法：使用本地最小实现..."

# 检查是否存在最小化 go.mod
if [ -f "go.mod.minimal" ]; then
    echo "✅ 找到 go.mod.minimal，使用离线配置..."
    cp go.mod go.mod.backup
    cp go.mod.minimal go.mod
    echo "📄 已切换到离线模式 go.mod"
else
    echo "⚠️  未找到 go.mod.minimal，使用当前 go.mod"
    echo "   尝试使用 vendor 模式..."
fi

# 尝试构建（不使用 vendor，直接使用本地模块）
echo ""
echo "🔨 开始构建..."
echo "   命令: go build -o o2ochat ./cmd/o2ochat"

if go build -o o2ochat ./cmd/o2ochat 2>&1; then
    echo ""
    echo "✅ 构建成功！"
    echo ""
    echo "📦 可执行文件信息:"
    ls -lh o2ochat
    echo ""
    echo "🚀 运行项目:"
    echo "  ./o2ochat --help"
    echo "  ./o2ochat"
    echo ""
    
    # 恢复原始 go.mod（如果使用了 minimal）
    if [ -f "go.mod.backup" ]; then
        mv go.mod.backup go.mod.minimal.active
        mv go.mod go.mod.offline
        mv go.mod.minimal.active go.mod
        echo "📄 已恢复原始 go.mod"
    fi
    
    exit 0
else
    echo ""
    echo "❌ 构建失败"
    echo ""
    echo "🔧 可能的原因和解决方案:"
    echo ""
    echo "1. 缺少依赖:"
    echo "   - 运行: ./fix_go_mod.sh"
    echo "   - 或者: make vendor"
    echo ""
    echo "2. 网络问题导致 go mod 下载失败:"
    echo "   - 使用: go build -mod=vendor"
    echo "   - 或者: go build -mod=mod -o o2ochat ./cmd/o2ochat"
    echo ""
    echo "3. 代码错误:"
    echo "   - 检查: go vet ./..."
    echo "   - 查看错误信息并修复代码"
    echo ""
    
    # 恢复原始 go.mod
    if [ -f "go.mod.backup" ]; then
        mv go.mod.backup go.mod
        echo "📄 已恢复原始 go.mod"
    fi
    
    exit 1
fi
