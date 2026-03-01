# O2OChat Android Build Guide

## Prerequisites

### Required Tools

1. **Go 1.22+** ✅ (Installed: go1.26.0)
2. **Android SDK** ⏳ (Need to install)
3. **Android NDK** ⏳ (Need to install)
4. **Gradle 7.0+** ⏳ (Need to install)
5. **gomobile** ⏳ (Need to install)

---

## Install Android SDK

### Method 1: Android Studio (Recommended)

```bash
# Download and install Android Studio
wget https://dl.google.com/dl/android/studio/ide-zips/2023.1.1.2/android-studio-2023.1.1.2-linux.tar.gz
sudo tar -xzf android-studio*.tar.gz -C /opt/

# Launch and install SDK via GUI
/opt/android-studio/bin/studio.sh
```

### Method 2: Command Line

```bash
# Download SDK tools
wget https://dl.google.com/android/repository/commandlinetools-linux-9477386_latest.zip
unzip commandlinetools*.zip
sudo mkdir -p /opt/android-sdk/cmdline-tools
sudo mv cmdline-tools /opt/android-sdk/cmdline-tools/latest

# Set environment
export ANDROID_HOME=/opt/android-sdk
export PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin

# Install SDK
yes | sdkmanager --licenses
sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0"
```

---

## Build Android APK

### Using Gradle (Kotlin Project)

```bash
cd /mnt/f/o2ochat/android
chmod +x gradlew
./gradlew assembleDebug
```

Output: `app/build/outputs/apk/debug/app-debug.apk`

### Using Go Mobile (Pure Go)

```bash
cd /mnt/f/o2ochat
gomobile build -target=android -o o2ochat-android.apk ./cmd/o2ochat
```

Output: `./o2ochat-android.apk`

---

## Current Status

### Completed
- ✅ Go 1.26.0 installed
- ✅ Android project structure created
- ✅ Build guide written

### Pending
- ⏳ Install Android SDK
- ⏳ Install Android NDK
- ⏳ Install Gradle
- ⏳ Install gomobile
- ⏳ Build APK

---

## Next Steps

1. Install Android SDK (see guide above)
2. Run: `cd android && ./gradlew assembleDebug`
3. Get APK: `app/build/outputs/apk/debug/app-debug.apk`

---

**Ready to compile after Android SDK installation!**
