# macOS / iOS 版编译指南

**更新时间**: 2026 年 3 月 2 日  
**重要提示**: iOS/macOS 编译需要 macOS 系统 + Xcode

---

## 📊 系统要求

### macOS 系统要求

| 项目 | 要求 | 推荐 |
|------|------|------|
| **操作系统** | macOS 12.0+ | macOS 14.0+ |
| **Xcode** | 14.0+ | Xcode 15.0+ |
| **内存** | 8GB | 16GB+ |
| **磁盘** | 20GB 可用 | 50GB+ SSD |

---

## 🚀 macOS 编译步骤

### 步骤 1: 安装 Xcode

1. **打开 App Store**
2. **搜索 "Xcode"**
3. **点击"获取"并安装**
4. **等待下载完成**（约 10-15GB）

或从官网下载：
```
https://developer.apple.com/xcode/
```

### 步骤 2: 配置 Xcode

```bash
# 打开 Xcode
open -a Xcode

# 同意许可协议
sudo xcodebuild -license accept

# 安装命令行工具
xcode-select --install
```

### 步骤 3: 下载项目

```bash
# 方法 1: 使用 Git
git clone https://github.com/netvideo/o2ochat.git
cd o2ochat/ios/O2OChat

# 方法 2: 使用 XcodeGen
cd o2ochat/ios
gem install xcodegen
xcodegen generate
```

### 步骤 4: 打开项目

```bash
# 使用 Xcode 打开
open O2OChat.xcodeproj

# 或使用 XcodeGen 生成的 workspace
open O2OChat.xcworkspace
```

### 步骤 5: 配置签名

1. **在 Xcode 中**
   ```
   项目名称 -> Signing & Capabilities
   ```

2. **选择 Team**
   - 使用个人 Team（免费）
   - 或使用开发者账号（付费）

3. **修改 Bundle Identifier**
   ```
   Bundle Identifier: com.yourname.o2ochat
   ```

### 步骤 6: 编译 iOS 版本

1. **选择目标设备**
   ```
   选择：Any iOS Device (arm64)
   或选择具体设备
   ```

2. **编译**
   ```
   Product -> Build
   或 Command + B
   ```

3. **输出位置**
   ```
   Products/O2OChat.app
   ```

### 步骤 7: 编译 macOS 版本

1. **切换 Scheme**
   ```
   Scheme 选择器 -> O2OChat-macOS
   ```

2. **编译**
   ```
   Product -> Build
   或 Command + B
   ```

3. **输出位置**
   ```
   Products/O2OChat.app
   ```

---

## 📱 安装到 iOS 设备

### 方法 1: 直接运行

1. **连接 iPhone/iPad**
2. **在 Xcode 中选择设备**
3. **点击运行按钮**
4. **应用会自动安装到设备**

### 方法 2: 导出 IPA

1. **Archive 项目**
   ```
   Product -> Archive
   ```

2. **导出 IPA**
   ```
   Distribute App -> Ad Hoc / Development
   选择签名证书
   导出 IPA 文件
   ```

3. **使用 Finder 安装**
   ```
   连接设备 -> Finder
   拖拽 IPA 到设备
   ```

### 方法 3: 使用 TestFlight

1. **上传到 App Store Connect**
   ```
   Product -> Archive
   Distribute App -> App Store Connect
   ```

2. **在 TestFlight 中添加测试人员**
3. **测试人员安装 TestFlight 应用**
4. **从 TestFlight 安装 O2OChat**

---

## 💻 macOS 应用安装

### Debug 版本

```bash
# 编译后在 Products 目录
cd ~/Library/Developer/Xcode/DerivedData/
find . -name "O2OChat.app" -type d

# 复制到 Applications
cp -R O2OChat.app /Applications/
```

### Release 版本

1. **切换 Build Configuration**
   ```
   Product -> Scheme -> Edit Scheme
   Run -> Info -> Build Configuration: Release
   ```

2. **Archive**
   ```
   Product -> Archive
   ```

3. **导出应用**
   ```
   Distribute App -> 选择导出选项
   导出 .app 文件
   ```

---

## 🔧 常见问题解决

### 问题 1: 签名错误

**错误信息**:
```
No signing certificate "iOS Development" found
```

**解决方案**:
```
1. Xcode -> Preferences -> Accounts
2. 添加 Apple ID
3. 选择 Team
4. 重新编译
```

### 问题 2: 依赖缺失

**错误信息**:
```
No such module 'XXX'
```

**解决方案**:
```bash
# 使用 XcodeGen 重新生成项目
cd /mnt/f/o2ochat/ios
gem install xcodegen
xcodegen generate
```

### 问题 3: 部署目标过低

**错误信息**:
```
The deployment target is too low
```

**解决方案**:
```
1. 项目设置 -> General
2. 修改 Deployment Target
3. iOS: 15.0+
4. macOS: 12.0+
```

### 问题 4: Swift 版本不匹配

**解决方案**:
```
1. 项目设置 -> Build Settings
2. 搜索 Swift Version
3. 选择 Swift 5.x
4. 清理构建缓存
```

---

## 📝 XcodeGen 配置说明

### project.yml 说明

```yaml
name: O2OChat
options:
  bundleIdPrefix: com.o2ochat
  deploymentTarget:
    iOS: 15.0
    macOS: 12.0

targets:
  O2OChat-iOS:
    type: application
    platform: iOS
    sources:
      - path: O2OChat/Sources
    settings:
      base:
        PRODUCT_BUNDLE_IDENTIFIER: com.o2ochat.ios
        
  O2OChat-macOS:
    type: application
    platform: macOS
    sources:
      - path: O2OChat/Sources
    settings:
      base:
        PRODUCT_BUNDLE_IDENTIFIER: com.o2ochat.macos
```

---

## 📞 相关资源

### 官方文档

- **Xcode**: https://developer.apple.com/xcode/
- **Swift**: https://swift.org/
- **iOS 开发**: https://developer.apple.com/ios/
- **macOS 开发**: https://developer.apple.com/macos/

### 项目位置

- **macOS 项目**: `/mnt/f/o2ochat/ios/O2OChat`
- **project.yml**: `/mnt/f/o2ochat/ios/project.yml`

---

## ✅ 编译检查清单

### 编译前检查

- [ ] macOS 12.0+ 系统
- [ ] Xcode 14.0+ 已安装
- [ ] 已接受 Xcode 许可
- [ ] 命令行工具已安装
- [ ] 项目已下载
- [ ] 已配置签名

### 编译后检查

- [ ] .app 文件已生成
- [ ] 可以正常运行
- [ ] 无编译错误
- [ ] 无运行时错误

### 发布前检查

- [ ] Release 模式编译
- [ ] 已签名
- [ ] 在真机上测试
- [ ] 性能测试通过
- [ ] 已测试主要功能

---

**更新时间**: 2026 年 3 月 2 日  
**状态**: ⏳ 需要 macOS + Xcode  
**平台**: macOS 12.0+, iOS 15.0+  

**🚀 请在 macOS 系统上使用 Xcode 编译 iOS/macOS 版！**
