# DevEco Studio 手动安装指南

由于需要 sudo 权限，请手动执行以下步骤。

---

## 📋 安装步骤

### 步骤 1: 安装依赖

```bash
sudo apt update
sudo apt install -y wget unzip libgtk-3-0 libxtst6 libnss3
```

### 步骤 2: 下载 DevEco Studio

**方法 A: 使用浏览器下载**

1. 访问：https://developer.harmonyos.com/cn/develop/deveco-studio
2. 点击"下载"
3. 选择 Linux 版本
4. 下载完成后，文件位于 `~/Downloads/`

**方法 B: 使用终端下载**

```bash
cd ~/Downloads
wget https://contentcenter-prod-1300007963.file.myqcloud.com/download/DevEco-Studio-5.0.0.100-linux.zip
```

### 步骤 3: 解压安装

```bash
sudo mkdir -p /opt
sudo unzip ~/Downloads/DevEco-Studio-5.0.0.100-linux.zip -d /opt/
sudo chmod -R 755 /opt/deveco-studio
```

### 步骤 4: 创建快捷方式

```bash
mkdir -p ~/.local/share/applications
cat > ~/.local/share/applications/deveco-studio.desktop << 'DESKTOP'
[Desktop Entry]
Version=1.0
Type=Application
Name=DevEco Studio
Exec=/opt/deveco-studio/bin/devestudio.sh
Comment=HarmonyOS Development IDE
Terminal=false
DESKTOP
chmod +x ~/.local/share/applications/deveco-studio.desktop
```

### 步骤 5: 启动 DevEco Studio

```bash
/opt/deveco-studio/bin/devestudio.sh
```

---

## ⚙️ 首次启动配置

### 1. 接受许可协议

- 阅读并接受华为开发者协议

### 2. 配置 SDK

- SDK 默认路径：`/opt/deveco-studio/sdk`
- 点击"Next"

### 3. 安装 HarmonyOS SDK

- 选择 API Version: **12 (5.0.0)**
- 点击"Install"
- 等待下载完成（约 5-10GB）

### 4. 完成配置

- 点击"Finish"

---

## 🚀 打开项目并编译

### 1. 打开项目

```bash
# 从 DevEco Studio 启动
File -> Open -> /mnt/f/o2ochat/harmony/O2OChat
```

### 2. 等待项目同步

- DevEco Studio 会自动识别项目
- 等待 Gradle/Hvigor 同步完成

### 3. 编译项目

```bash
# 菜单操作
Build -> Build Hap(s) / APP(s)

# 或命令行编译
cd /mnt/f/o2ochat/harmony/O2OChat
# 需要 DevEco Studio 的命令行工具
```

### 4. 输出位置

```
entry/build/outputs/entry-default-signed.hap
```

---

## 🔍 验证安装

### 检查安装目录

```bash
ls -la /opt/deveco-studio/
ls -la /opt/deveco-studio/sdk/
```

### 启动应用

```bash
/opt/deveco-studio/bin/devestudio.sh
```

---

## ⚠️ 常见问题

### 问题 1: 无法启动

```bash
# 检查依赖
ldd /opt/deveco-studio/bin/devestudio.sh

# 安装缺失依赖
sudo apt install -y libatk1.0-0 libatk-bridge2.0-0 libcups2 libdrm2
```

### 问题 2: 权限问题

```bash
sudo chown -R $USER:$USER /opt/deveco-studio
```

### 问题 3: SDK 下载失败

- 检查网络连接
- 使用国内镜像
- 手动下载 SDK 并导入

---

## 📞 相关资源

- **官方网站**: https://developer.harmonyos.com/cn/develop/deveco-studio
- **开发文档**: https://developer.harmonyos.com/cn/docs
- **项目位置**: /mnt/f/o2ochat/harmony/O2OChat

---

**更新时间**: 2026 年 3 月 2 日  
**状态**: ⏳ 待手动安装
