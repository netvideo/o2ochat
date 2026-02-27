# O2OChat Windows 程序构建指南

本指南介绍如何为 Windows 平台构建 O2OChat 应用程序。

## 目录

- [快速开始](#快速开始)
- [环境准备](#环境准备)
- [构建方法](#构建方法)
- [Windows 特性](#windows-特性)
- [故障排除](#故障排除)
- [发布检查清单](#发布检查清单)

## 快速开始

### 一键构建

```bash
# 克隆项目
git clone https://github.com/netvideo/o2ochat.git
cd o2ochat

# 运行 Windows 构建脚本
chmod +x scripts/build-windows.sh
./scripts/build-windows.sh
```

**输出**: `dist/o2ochat.exe`

## 环境准备

### 必需组件

#### 1. Go 编译器 (1.22+)

```bash
# 检查 Go 版本
go version

# 如果版本低于 1.22，请升级
# 参考: https://golang.org/doc/install
```

#### 2. MinGW-w64 交叉编译工具链

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install mingw-w64
```

**macOS:**
```bash
brew install mingw-w64
```

**Windows (MSYS2):**
```bash
pacman -S mingw-w64-x86_64-toolchain
```

#### 3. 验证安装

```bash
# 检查 MinGW 安装
x86_64-w64-mingw32-gcc --version

# 应该输出类似:
# x86_64-w64-mingw32-gcc (GCC) 10.x.x
```

### 可选组件

#### UPX 压缩工具（推荐）

```bash
# Ubuntu/Debian
sudo apt-get install upx

# macOS
brew install upx

# Windows
choco install upx
```

使用 UPX 可以显著减小可执行文件大小：
```bash
upx --best o2ochat.exe
```

## 构建方法

### 方法 1: 使用构建脚本（推荐）

```bash
# 设置环境变量
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=off

# 运行构建脚本
./scripts/build-windows.sh
```

**脚本功能:**
- 自动检测环境
- 设置正确的编译器
- 优化构建标志
- 可选的 UPX 压缩

### 方法 2: 手动构建

#### 步骤 1: 设置环境变量

```bash
# 目标平台
export GOOS=windows
export GOARCH=amd64

# 启用 CGO（某些依赖需要）
export CGO_ENABLED=1

# Windows 编译器
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++

# Go 代理（国内加速）
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=off
```

#### 步骤 2: 下载依赖

```bash
go mod tidy
```

#### 步骤 3: 构建可执行文件

```bash
# 基础构建
go build -o o2ochat.exe ./cmd/o2ochat

# 优化构建（推荐）
go build -ldflags "-s -w" -o o2ochat.exe ./cmd/o2ochat

# Windows GUI 应用（无控制台窗口）
go build -ldflags "-s -w -H windowsgui" -o o2ochat.exe ./cmd/o2ochat
```

#### 步骤 4: 验证构建

```bash
# 检查文件
ls -lh o2ochat.exe

# 测试运行（在 Windows 上）
./o2ochat.exe --version
```

### 方法 3: Windows 本地构建

在 Windows 系统上直接构建（无需交叉编译）：

#### 使用 PowerShell

```powershell
# 设置环境变量
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

# 构建
go build -ldflags "-s -w" -o o2ochat.exe ./cmd/o2ochat

# 验证
.\o2ochat.exe --version
```

#### 使用命令提示符 (CMD)

```cmd
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

go build -ldflags "-s -w" -o o2ochat.exe ./cmd/o2ochat

o2ochat.exe --version
```

## Windows 特性

### 系统托盘集成

Windows 版本支持系统托盘图标和菜单：

```go
// 创建托盘图标
tray.Run()
```

### Windows 服务支持

可以作为 Windows 服务运行：

```powershell
# 安装服务
sc create O2OChat binPath= "C:\Program Files\O2OChat\o2ochat.exe --service"

# 启动服务
sc start O2OChat

# 停止服务
sc stop O2OChat

# 删除服务
sc delete O2OChat
```

### 自动更新

支持 Windows 自动更新机制：

```go
// 检查更新
checker := updater.NewChecker("https://update.o2ochat.io")
update, err := checker.Check()
if err == nil && update.Available {
    // 下载并安装更新
    updater.Install(update)
}
```

### Windows 防火墙集成

自动添加防火墙规则：

```powershell
# 添加防火墙规则
netsh advfirewall firewall add rule name="O2OChat" dir=in action=allow program="o2ochat.exe"
```

## 故障排除

### 常见问题

#### 1. 交叉编译失败

**错误**: `gcc: error: unrecognized command line option '-mthreads'`

**解决**:
```bash
# 安装正确的 MinGW
sudo apt-get install mingw-w64

# 验证安装
x86_64-w64-mingw32-gcc --version
```

#### 2. 运行时缺少 DLL

**错误**: `The program can't start because libgcc_s_seh-1.dll is missing`

**解决**:
```bash
# 静态链接 GCC 运行时
go build -ldflags "-s -w -linkmode external -extldflags '-static'" -o o2ochat.exe ./cmd/o2ochat
```

#### 3. Windows 病毒误报

**问题**: Windows Defender 或杀毒软件误报

**解决**:
- 添加数字签名
- 提交到杀毒软件厂商白名单
- 使用更知名的证书颁发机构

#### 4. 图标不显示

**问题**: Windows 任务栏或资源管理器不显示自定义图标

**解决**:
```bash
# 使用 rsrc 工具嵌入图标
go get github.com/akavel/rsrc
rsrc -ico icon.ico -o rsrc.syso
go build -o o2ochat.exe ./cmd/o2ochat
```

### 调试 Windows 版本

```bash
# 构建调试版本
go build -tags debug -gcflags="-N -l" -o o2ochat-debug.exe ./cmd/o2ochat

# Windows 上调试
dlv exec o2ochat-debug.exe
```

## 发布检查清单

- [ ] 版本号已更新
- [ ] 二进制文件已签名
- [ ] 测试通过
- [ ] 文档已更新
- [ ] 安装程序已创建
- [ ] 病毒扫描通过
- [ ] GitHub Release 已创建

## 桌面 GUI 应用

O2OChat 提供 Windows 桌面 GUI 应用，使用 Fyne 框架开发。

### 构建桌面应用

```bash
cd windows/O2OChat
go mod tidy
go build -o o2ochat-gui.exe .
```

### 运行桌面应用

```bash
./o2ochat-gui.exe
```

### 依赖要求

- Go 1.22+
- GCC 编译器 (MinGW)
- GTK3 运行时库

```powershell
# Windows 上安装 MinGW
choco install mingw

# 或使用 MSYS2
pacman -S mingw-w64-x86_64-gcc
```

### 交叉编译 (Linux/macOS → Windows)

```bash
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc

cd windows/O2OChat
go build -o o2ochat-gui.exe .
```

## 相关链接

- [Go 交叉编译](https://golang.org/doc/install/source#environment)
- [MinGW-w64 文档](http://mingw-w64.org/)
- [Windows API Go 绑定](https://github.com/akavel/rsrc)
- [UPX 压缩工具](https://upx.github.io/)

---

**最后更新**: 2024年
**作者**: O2OChat Team
