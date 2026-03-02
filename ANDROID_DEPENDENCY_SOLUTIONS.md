# 🚀 Android Gradle 依赖下载解决方案

**更新时间**: 2026 年 3 月 1 日 17:20 CST

---

## ⏳ 当前问题

- ❌ Gradle 依赖下载失败（网络原因）
- ✅ 所有配置已正确
- ✅ Android SDK 已就绪

---

## 🚀 解决方案

### 方案 1: 使用国内镜像源 (推荐)

修改 `android/build.gradle`：

```gradle
buildscript {
    repositories {
        maven { url 'https://maven.aliyun.com/repository/google' }
        maven { url 'https://maven.aliyun.com/repository/public' }
        maven { url 'https://maven.aliyun.com/repository/gradle-plugin' }
        google()
        mavenCentral()
    }
}

allprojects {
    repositories {
        maven { url 'https://maven.aliyun.com/repository/google' }
        maven { url 'https://maven.aliyun.com/repository/public' }
        google()
        mavenCentral()
    }
}
```

### 方案 2: 使用 Gradle 国内镜像

修改 `android/gradle/wrapper/gradle-wrapper.properties`：

```properties
distributionUrl=https\://mirrors.cloud.tencent.com/gradle/gradle-8.2-bin.zip
```

### 方案 3: 离线模式（一次性下载所有依赖）

```bash
cd /mnt/f/o2ochat/android

# 使用国内镜像
export GRADLE_OPTS="-Dorg.gradle.jvmargs='-Xmx2048m -Dfile.encoding=UTF-8'"

# 下载依赖
/mnt/i/迅雷下载/gradle-8.2-bin/gradle-8.2/bin/gradle dependencies --refresh-dependencies
```

### 方案 4: 手动下载依赖（适合无网络环境）

1. 在有网络的机器上运行：
```bash
gradle dependencies --export-dependencies
```

2. 复制 `.gradle` 缓存目录到目标机器

### 方案 5: 使用缓存

```bash
cd /mnt/f/o2ochat/android

# 使用 Gradle 缓存
/mnt/i/迅雷下载/gradle-8.2-bin/gradle-8.2/bin/gradle assembleDebug --offline
```

---

## 📝 推荐步骤

### 步骤 1: 修改镜像源

```bash
cd /mnt/f/o2ochat/android

# 备份原文件
cp build.gradle build.gradle.bak

# 使用阿里云镜像
cat > build.gradle << 'GRADLE'
buildscript {
    repositories {
        maven { url 'https://maven.aliyun.com/repository/google' }
        maven { url 'https://maven.aliyun.com/repository/public' }
        maven { url 'https://maven.aliyun.com/repository/gradle-plugin' }
        google()
        mavenCentral()
    }
    dependencies {
        classpath 'com.android.tools.build:gradle:8.2.2'
        classpath "org.jetbrains.kotlin:kotlin-gradle-plugin:1.9.22"
    }
}

allprojects {
    repositories {
        maven { url 'https://maven.aliyun.com/repository/google' }
        maven { url 'https://maven.aliyun.com/repository/public' }
        google()
        mavenCentral()
    }
}

task clean(type: Delete) {
    delete rootProject.buildDir
}
GRADLE
```

### 步骤 2: 下载依赖

```bash
# 设置环境变量
export GRADLE=/mnt/i/迅雷下载/gradle-8.2-bin/gradle-8.2/bin/gradle
export ANDROID_HOME=$HOME/android-sdk
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64

# 下载依赖
$GRADLE dependencies --refresh-dependencies

# 编译 APK
$GRADLE assembleDebug --no-daemon
```

---

## 🎯 预期结果

使用阿里云镜像后，下载速度应提升至：
- 初始依赖下载：~5-10 分钟
- 后续编译：<2 分钟

---

## 📞 故障排除

### 问题 1: 依赖下载超时

```bash
# 增加超时时间
export GRADLE_OPTS="-Dorg.gradle.daemon=false -Dorg.gradle.http.timeout=300000"
```

### 问题 2: 镜像源不可用

尝试其他镜像：
- 腾讯云：https://mirrors.cloud.tencent.com/gradle/
- 阿里云：https://maven.aliyun.com/repository/
- 七牛云：https://kodo.qiniu.com/

### 问题 3: Gradle 缓存损坏

```bash
# 清理缓存
rm -rf ~/.gradle/caches/
rm -rf android/build/
rm -rf android/.gradle/

# 重新下载
$GRADLE clean build --refresh-dependencies
```

---

**推荐优先使用阿里云镜像方案！**

---

**更新时间**: 2026 年 3 月 1 日 17:20 CST  
**状态**: ⏳ 等待镜像源配置后重新编译
