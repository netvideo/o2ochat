# 消息翻译功能

## 概述

O2OChat 支持为每条消息的发送和接收分别设置翻译开关，让用户可以灵活控制翻译行为。

## 功能特性

### 分开的翻译开关

- **接收翻译开关**: 控制是否翻译收到的消息
- **发送翻译开关**: 控制是否翻译发送的消息

### 独立控制

每个聊天可以独立设置翻译开关：
- 只为特定聊天启用翻译
- 接收和发送可以分别启用/禁用
- 支持全局默认设置

### 智能翻译

- 自动检测源语言
- 支持 17 种语言互译
- 显示原文和译文
- 翻译结果缓存

## 使用示例

### 基本使用

```go
package main

import (
	"fmt"
	"github.com/netvideo/o2ochat/pkg/translation"
)

func main() {
	// 创建翻译管理器
	manager := translation.NewTranslationManager()
	
	chatID := "chat-123"
	peerID := "peer-456"
	
	// 仅启用接收翻译
	manager.SetIncomingTranslation(chatID, true)
	
	// 仅启用发送翻译
	manager.SetOutgoingTranslation(chatID, true)
	
	// 检查状态
	incomingEnabled := manager.GetIncomingTranslationStatus(chatID)
	outgoingEnabled := manager.GetOutgoingTranslationStatus(chatID)
	
	fmt.Printf("接收翻译：%v\n", incomingEnabled)
	fmt.Printf("发送翻译：%v\n", outgoingEnabled)
}
```

### 完整配置

```go
settings := &translation.TranslationSettings{
	EnableIncomingTranslation: true,   // 启用接收翻译
	EnableOutgoingTranslation: true,   // 启用发送翻译
	SourceLanguage:           "zh-CN", // 源语言
	TargetLanguage:           "en",    // 目标语言
	AutoDetectSource:         true,    // 自动检测源语言
	Provider:                 ai.ProviderOllama, // 使用 Ollama
	ShowOriginal:            true,    // 显示原文
	CacheTranslations:       true,    // 缓存翻译
}

manager.SetSettings(chatID, peerID, settings)
```

### 全局设置

```go
// 设置全局默认配置
globalSettings := &translation.TranslationSettings{
	EnableIncomingTranslation: false,
	EnableOutgoingTranslation: false,
	TargetLanguage:           "en",
	Provider:                 ai.ProviderOllama,
}

manager.SetGlobalSettings(globalSettings)
```

## UI 界面设计

### 聊天界面

```
┌─────────────────────────────────┐
│ 聊天设置                        │
├─────────────────────────────────┤
│ 🔵 翻译设置                     │
│                                 │
│ [✓] 翻译收到的消息              │
│     从 [自动检测] → [英语]      │
│                                 │
│ [✓] 翻译发送的消息              │
│     从 [中文] → [英语]          │
│                                 │
│ [✓] 显示原文                    │
│ [✓] 缓存翻译结果                │
│                                 │
│ AI 提供商：[Ollama ▼]          │
│                                 │
└─────────────────────────────────┘
```

### 快捷开关

```
┌─────────────────────────────────┐
│ 消息气泡                         │
├─────────────────────────────────┤
│ 你好！                          │
│ Hello!                   [翻译] │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ 消息气泡 (已翻译)                │
├─────────────────────────────────┤
│ 你好！                          │
│ Hello!                   [原文] │
└─────────────────────────────────┘
```

## API 参考

### TranslationManager

#### 方法

| 方法 | 说明 |
|------|------|
| `SetIncomingTranslation(chatID, enabled)` | 设置接收翻译开关 |
| `SetOutgoingTranslation(chatID, enabled)` | 设置发送翻译开关 |
| `GetIncomingTranslationStatus(chatID)` | 获取接收翻译状态 |
| `GetOutgoingTranslationStatus(chatID)` | 获取发送翻译状态 |
| `SetSettings(chatID, peerID, settings)` | 设置完整配置 |
| `GetSettings(chatID)` | 获取聊天设置 |
| `SetGlobalSettings(settings)` | 设置全局配置 |
| `ShouldTranslate(chatID, direction)` | 检查是否需要翻译 |

### TranslationSettings

#### 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `EnableIncomingTranslation` | bool | 启用接收翻译 |
| `EnableOutgoingTranslation` | bool | 启用发送翻译 |
| `SourceLanguage` | string | 源语言代码 |
| `TargetLanguage` | string | 目标语言代码 |
| `AutoDetectSource` | bool | 自动检测源语言 |
| `Provider` | ProviderType | AI 提供商 |
| `ShowOriginal` | bool | 显示原文 |
| `CacheTranslations` | bool | 缓存翻译 |

## 使用场景

### 场景 1: 只翻译接收消息

用户 A 说英语，用户 B 说中文：

```go
// 用户 B 的设置
manager.SetIncomingTranslation(chatID, true)  // 翻译收到的英文
manager.SetOutgoingTranslation(chatID, false) // 发送中文，不翻译
```

### 场景 2: 只翻译发送消息

用户 A 说中文，用户 B 说英语：

```go
// 用户 A 的设置
manager.SetIncomingTranslation(chatID, false) // 接收英文，能看懂
manager.SetOutgoingTranslation(chatID, true)  // 发送时翻译成英文
```

### 场景 3: 双向翻译

两个用户说不同语言：

```go
// 双方都启用双向翻译
manager.SetIncomingTranslation(chatID, true)
manager.SetOutgoingTranslation(chatID, true)
```

### 场景 4: 临时翻译

偶尔需要翻译某条消息：

```go
// 默认关闭翻译
manager.SetIncomingTranslation(chatID, false)
manager.SetOutgoingTranslation(chatID, false)

// 需要时手动翻译单条消息
translator.Translate(ctx, &ai.TranslationRequest{
	Text: "Hello",
	SourceLang: "en",
	TargetLang: "zh-CN",
})
```

## 性能优化

### 缓存策略

```go
settings := &translation.TranslationSettings{
	CacheTranslations: true,      // 启用缓存
	// ... 其他设置
}
```

- 相同内容直接返回缓存
- 减少 AI 调用
- 降低成本
- 提高速度

### 批量翻译

```go
// 累积多条消息一起翻译
requests := []*ai.TranslationRequest{
	{Text: "Hello", SourceLang: "en", TargetLang: "zh-CN"},
	{Text: "World", SourceLang: "en", TargetLang: "zh-CN"},
}

responses, err := manager.TranslateBatch(ctx, requests)
```

## 隐私保护

### 本地翻译优先

```go
settings := &translation.TranslationSettings{
	Provider: ai.ProviderOllama, // 使用本地 Ollama
	// ... 其他设置
}
```

- 数据不出设备
- 保护隐私
- 离线可用

### 云端翻译

```go
settings := &translation.TranslationSettings{
	Provider: ai.ProviderOpenAI, // 使用 OpenAI
	// ... 其他设置
}
```

- 高质量翻译
- 需要网络
- 注意隐私政策

## 常见问题

### Q: 如何为不同聊天设置不同翻译？

```go
// 聊天 A：中英文翻译
manager.SetSettings("chat-a", "peer-a", &translation.TranslationSettings{
	SourceLanguage: "zh-CN",
	TargetLanguage: "en",
	EnableIncomingTranslation: true,
})

// 聊天 B：日中英文翻译
manager.SetSettings("chat-b", "peer-b", &translation.TranslationSettings{
	SourceLanguage: "ja",
	TargetLanguage: "zh-CN",
	EnableIncomingTranslation: true,
})
```

### Q: 如何临时关闭翻译？

```go
// 保存当前设置
settings, _ := manager.GetSettings(chatID)

// 关闭翻译
manager.SetIncomingTranslation(chatID, false)
manager.SetOutgoingTranslation(chatID, false)

// ... 使用完毕后恢复
manager.SetSettings(chatID, peerID, settings)
```

### Q: 支持哪些语言？

支持 17 种语言：
- 中文 (简体/繁体)
- 英文、日文、韩文
- 德文、法文、西班牙文
- 俄文、阿拉伯文、希伯来文
- 马来文、葡萄牙文、意大利文
- 藏文、蒙文、维吾尔文

## 相关文档

- [AI 翻译功能](AI_TRANSLATION.md)
- [消息模块](../pkg/message/)
- [UI 组件](../pkg/ui/)

---

**版本**: v1.0.0  
**更新时间**: 2026 年 2 月 28 日
