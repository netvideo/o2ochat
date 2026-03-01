# Message Translation Settings (pkg/translation)

消息翻译设置模块为每条消息的发送和接收提供独立的翻译开关控制。

## 功能特性

- 🔀 **分开的翻译开关** - 接收和发送消息独立控制
- 🎯 **按聊天设置** - 每个聊天独立配置
- 🌍 **多语言支持** - 17 种语言互译
- ⚡ **快速切换** - 一键启用/禁用翻译
- 💾 **全局默认** - 可设置全局默认配置

## 核心组件

### settings.go
翻译设置管理器，包括：
- TranslationSettings - 翻译配置
- ChatTranslationSettings - 按聊天配置
- TranslationManager - 管理器

## 使用示例

```go
manager := translation.NewTranslationManager()

// 仅启用接收翻译
manager.SetIncomingTranslation(chatID, true)

// 仅启用发送翻译
manager.SetOutgoingTranslation(chatID, true)

// 检查状态
incoming := manager.GetIncomingTranslationStatus(chatID)
outgoing := manager.GetOutgoingTranslationStatus(chatID)
```

## 文档

- [消息翻译指南](../../docs/MESSAGE_TRANSLATION.md)
- [AI 翻译](../../docs/AI_TRANSLATION.md)

**版本**: v1.0.0
