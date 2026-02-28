# 快速构建脚本
#!/bin/bash

echo "🔨 O2OChat 快速构建脚本"
echo "=========================="

# 检查 vendor 目录
if [ ! -d "vendor/websocket" ] || [ ! -d "vendor/go-sqlite3" ] || [ ! -d "vendor/testify" ]; then
    echo "❌ 缺少 vendor 依赖，请先运行 fix_go_mod.sh"
    exit 1
fi

echo "✅ Vendor 依赖检查通过"

# 构建项目
echo "🔨 开始构建..."
go build -mod=vendor -o o2ochat ./cmd/o2ochat

# 检查构建结果
if [ $? -eq 0 ]; then
    echo "✅ 构建成功！"
    echo "📦 可执行文件: ./o2ochat"
    ls -lh o2ochat
    echo ""
    echo "🚀 运行项目:"
    echo "  ./o2ochat --help"
    echo "  ./o2ochat"
else
    echo "❌ 构建失败，请检查错误信息"
    exit 1
fi
