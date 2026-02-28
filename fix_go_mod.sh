#!/bin/bash
# fix_go_mod.sh - 修复 go mod 依赖问题

set -e

echo "🔧 O2OChat Go Mod 依赖修复脚本"
echo "================================"

# 1. 清理旧的 vendor 目录
echo "📁 步骤 1/5: 清理旧的 vendor 目录..."
rm -rf vendor
mkdir -p vendor
cd vendor

# 2. 克隆核心依赖
echo "📦 步骤 2/5: 克隆核心依赖..."

# 网络依赖
echo "   📡 克隆 gorilla/websocket..."
git clone --depth 1 https://github.com/gorilla/websocket.git 2>/dev/null || echo "   ⚠️  网络不可用，跳过 websocket"

echo "   🗄️  克隆 mattn/go-sqlite3..."
git clone --depth 1 https://github.com/mattn/go-sqlite3.git 2>/dev/null || echo "   ⚠️  网络不可用，跳过 go-sqlite3"

echo "   🧪 克隆 stretchr/testify..."
git clone --depth 1 https://github.com/stretchr/testify.git 2>/dev/null || echo "   ⚠️  网络不可用，跳过 testify"

# 可选依赖
echo "   📜 克隆 golang/protobuf..."
git clone --depth 1 https://github.com/golang/protobuf.git 2>/dev/null || echo "   ⚠️  网络不可用，跳过 protobuf"

echo "   🆔 克隆 google/uuid..."
git clone --depth 1 https://github.com/google/uuid.git 2>/dev/null || echo "   ⚠️  网络不可用，跳过 uuid"

echo "   🐍 克隆 spf13/cobra..."
git clone --depth 1 https://github.com/spf13/cobra.git 2>/dev/null || echo "   ⚠️  网络不可用，跳过 cobra"

cd ..

# 3. 检查克隆结果
echo "📊 步骤 3/5: 检查克隆结果..."
if [ -d "vendor/websocket" ]; then
    echo "   ✅ websocket 克隆成功"
else
    echo "   ⚠️  websocket 克隆失败，将使用 go mod 下载"
fi

if [ -d "vendor/go-sqlite3" ]; then
    echo "   ✅ go-sqlite3 克隆成功"
else
    echo "   ⚠️  go-sqlite3 克隆失败，将使用 go mod 下载"
fi

if [ -d "vendor/testify" ]; then
    echo "   ✅ testify 克隆成功"
else
    echo "   ⚠️  testify 克隆失败，将使用 go mod 下载"
fi

# 4. 更新 go.mod
echo "📝 步骤 4/5: 更新 go.mod..."

# 检查是否存在 vendor 目录中的依赖
if [ -d "vendor/websocket" ] && [ -d "vendor/go-sqlite3" ] && [ -d "vendor/testify" ]; then
    # 添加 replace 指令到 go.mod
    cat >> go.mod << 'EOF'

// Local vendor replacements
replace (
	github.com/gorilla/websocket => ./vendor/websocket
	github.com/mattn/go-sqlite3 => ./vendor/go-sqlite3
	github.com/stretchr/testify => ./vendor/testify
)
EOF
    echo "   ✅ go.mod 已更新，使用本地 vendor"
else
    echo "   ℹ️  部分依赖未克隆，将使用 go mod 下载"
fi

# 5. 验证设置
echo "✅ 步骤 5/5: 验证设置..."

# 检查 go.mod
echo "   📄 go.mod 状态:"
grep -A 5 "^replace" go.mod 2>/dev/null || echo "   ℹ️  无 replace 指令"

# 检查 vendor 目录
echo "   📁 vendor 目录状态:"
ls -d vendor/* 2>/dev/null | wc -l | xargs echo "      包含目录数:"

echo ""
echo "🎉 Go Mod 依赖修复完成！"
echo ""
echo "📚 下一步:"
echo "   1. 构建项目: go build -o o2ochat ./cmd/o2ochat"
echo "   2. 运行测试: go test ./..."
echo "   3. 如果使用了 vendor: go build -mod=vendor -o o2ochat ./cmd/o2ochat"
echo ""
echo "📖 参考文档:"
echo "   - DEPENDENCY_INSTALL_GUIDE.md - 依赖安装完整指南"
echo "   - OFFLINE_BUILD_GUIDE.md - 离线构建完整指南"
