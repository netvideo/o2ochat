#!/bin/bash
# run_tests.sh - 运行项目测试

echo "🧪 O2OChat 测试运行脚本"
echo "=========================="

# 设置 Go 环境
export GOWORK=off
export GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
export GOSUMDB=off

echo ""
echo "📊 运行 P2P 包测试..."
go test ./pkg/p2p -v -run TestNewPeerConnection 2>&1 | head -50

echo ""
echo "✅ 测试运行完成！"
