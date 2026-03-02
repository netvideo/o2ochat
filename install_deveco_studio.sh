#!/bin/bash
# DevEco Studio 自动安装脚本

set -e

echo "=========================================="
echo "DevEco Studio 自动安装脚本"
echo "=========================================="
echo ""

# 检查系统
echo "📋 检查系统环境..."
if [ ! -f /etc/os-release ]; then
    echo "❌ 不支持的系统"
    exit 1
fi

source /etc/os-release
if [[ "$ID" != "ubuntu" && "$ID" != "debian" && "$ID" != "fedora" ]]; then
    echo "⚠️  请手动安装，参考 INSTALL_DEVECO_STUDIO.md"
    exit 1
fi

echo "✅ 系统：$NAME $VERSION"
echo ""

# 检查依赖
echo "📋 安装必要依赖..."
sudo apt update
sudo apt install -y wget unzip libgtk-3-0 libxtst6 libnss3 libasound2
echo "✅ 依赖已安装"
echo ""

# 下载
echo "📥 下载 DevEco Studio..."
cd /tmp
if [ -f DevEco-Studio-5.0.0.100-linux.zip ]; then
    echo "✅ 已下载"
else
    echo "⏳ 开始下载（约 3-5GB，可能需要 10-30 分钟）..."
    echo "💡 或手动从 https://developer.harmonyos.com/cn/develop/deveco-studio 下载"
    echo ""
    # 由于文件较大，提供手动下载选项
    read -p "是否继续自动下载？(y/n): " answer
    if [[ $answer == "y" || $answer == "Y" ]]; then
        wget https://contentcenter-prod-1300007963.file.myqcloud.com/download/DevEco-Studio-5.0.0.100-linux.zip
    else
        echo "❌ 取消安装"
        echo "📝 请参考 INSTALL_DEVECO_STUDIO.md 手动安装"
        exit 1
    fi
fi

# 解压
echo "📦 解压 DevEco Studio..."
sudo mkdir -p /opt
sudo unzip -q DevEco-Studio-5.0.0.100-linux.zip -d /opt/
sudo chmod -R 755 /opt/deveco-studio
echo "✅ 已解压到 /opt/deveco-studio"
echo ""

# 创建快捷方式
echo "🖥️  创建桌面快捷方式..."
mkdir -p ~/.local/share/applications
cat > ~/.local/share/applications/deveco-studio.desktop << 'DESKTOP'
[Desktop Entry]
Version=1.0
Type=Application
Name=DevEco Studio
Exec=/opt/deveco-studio/bin/devestudio.sh
Icon=/opt/deveco-studio/bin/devestudio.png
Comment=HarmonyOS Development IDE
Categories=Development;IDE;
Terminal=false
DESKTOP
chmod +x ~/.local/share/applications/deveco-studio.desktop
echo "✅ 快捷方式已创建"
echo ""

# 完成
echo "=========================================="
echo "✅ DevEco Studio 安装完成！"
echo "=========================================="
echo ""
echo "📝 下一步:"
echo "1. 启动 DevEco Studio:"
echo "   /opt/deveco-studio/bin/devestudio.sh"
echo ""
echo "2. 首次启动配置:"
echo "   - 接受许可协议"
echo "   - 配置 SDK 路径（默认 /opt/deveco-studio/sdk）"
echo "   - 安装 HarmonyOS SDK API 12"
echo ""
echo "3. 打开项目:"
echo "   File -> Open -> /mnt/f/o2ochat/harmony/O2OChat"
echo ""
echo "4. 编译项目:"
echo "   Build -> Build Hap(s) / APP(s)"
echo ""
echo "📚 详细指南：INSTALL_DEVECO_STUDIO.md"
echo ""
