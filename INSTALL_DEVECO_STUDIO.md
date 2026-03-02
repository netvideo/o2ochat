# DevEco Studio 安装指南

**更新时间**: 2026 年 3 月 2 日  
**系统**: Ubuntu 24.04.3 LTS x86_64

---

## 📊 系统要求

### ✅ 已满足

- **操作系统**: Ubuntu 24.04.3 LTS ✅
- **架构**: x86_64 ✅
- **内存**: 15GB (要求 8GB+) ✅
- **磁盘**: 118GB 可用 (要求 10GB+) ✅

---

## 🚀 安装步骤

### 方法 1: 官方下载（推荐）

#### 1. 下载 DevEco Studio

```bash
# 访问官方网站下载
# https://developer.harmonyos.com/cn/develop/deveco-studio

# 或使用wget下载（版本号可能会更新）
cd /tmp
wget https://contentcenter-prod-1300007963.file.myqcloud.com/download/DevEco-Studio-5.0.0.100-linux.zip

# 验证下载
ls -lh DevEco-Studio-5.0.0.100-linux.zip
```

#### 2. 解压安装

```bash
# 解压到/opt目录
sudo mkdir -p /opt
sudo unzip DevEco-Studio-5.0.0.100-linux.zip -d /opt/

# 设置权限
sudo chmod -R 755 /opt/deveco-studio

# 创建桌面快捷方式
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
```

#### 3. 首次启动

```bash
# 启动 DevEco Studio
/opt/deveco-studio/bin/devestudio.sh
```

---

### 方法 2: 使用脚本安装

```bash
cd /mnt/f/o2ochat
chmod +x install_deveco_studio.sh
./install_deveco_studio.sh
```

---

## ⚙️ 配置 SDK

### 1. 打开设置

启动 DevEco Studio 后：
- File -> Settings (或 Ctrl+Alt+S)
- 找到 HarmonyOS SDK

### 2. 安装 SDK

- 选择 API Version: 12 (5.0.0)
- 点击 Apply 安装
- 等待下载完成（约 5-10GB）

### 3. 配置环境变量

```bash
# 添加到 ~/.bashrc
export HARMONYOS_SDK_ROOT=/opt/deveco-studio/sdk
export PATH=$PATH:$HARMONYOS_SDK_ROOT/toolchains

# 应用配置
source ~/.bashrc
```

---

## 🔍 验证安装

### 1. 检查安装

```bash
# 检查安装目录
ls -la /opt/deveco-studio/

# 检查 SDK
ls -la /opt/deveco-studio/sdk/
```

### 2. 测试编译

```bash
# 打开项目
cd /mnt/f/o2ochat/harmony/O2OChat

# 在 DevEco Studio 中打开
/opt/deveco-studio/bin/devestudio.sh .

# 或命令行编译
cd /mnt/f/o2ochat/harmony/O2OChat
npm install
npm run build:debug
```

---

## ⚠️ 常见问题

### 问题 1: 权限问题

```bash
# 解决权限问题
sudo chown -R $USER:$USER /opt/deveco-studio
```

### 问题 2: 依赖缺失

```bash
# 安装必要依赖
sudo apt update
sudo apt install -y libgtk-3-0 libxtst6 libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libdrm2 libxkbcommon0 libxcomposite1 libxdamage1 libxfixes3 libxrandr2 libgbm1 libasound2
```

### 问题 3: 编译失败

```bash
# 清理缓存
cd /mnt/f/o2ochat/harmony/O2OChat
rm -rf .hvigor/
rm -rf build/
rm -rf entry/build/

# 重新编译
npm install
npm run build:debug
```

---

## 📞 相关资源

- **官方网站**: https://developer.harmonyos.com/cn/develop/deveco-studio
- **开发文档**: https://developer.harmonyos.com/cn/docs
- **ArkTS 语言**: https://developer.harmonyos.com/cn/docs/documentation/doc-guides/start-overview-0000000000029432
- **HarmonyOS SDK**: https://developer.harmonyos.com/cn/develop/sdkresources

---

**更新时间**: 2026 年 3 月 2 日  
**状态**: ⏳ 待安装
