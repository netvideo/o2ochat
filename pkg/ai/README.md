# AI Translation Module (pkg/ai)

AI 翻译模块为 O2OChat 提供智能翻译功能，支持本地 Ollama 和云端 AI 服务。

## 功能特性

- 🔌 **多 AI 提供商支持** - Ollama, OpenAI GPT, Anthropic Claude, Google Gemini
- 🌍 **17 种语言翻译** - 中、英、日、韩、德、法、西、俄等
- 💾 **翻译缓存** - 提高性能，降低成本
- 🔄 **故障转移** - 自动切换到备用提供商
- 🏠 **本地优先** - 支持离线 Ollama 翻译
- 🔒 **隐私保护** - 本地翻译数据不出设备

## 核心组件

### translator.go
核心翻译接口和类型定义

### ollama.go
Ollama 本地集成实现

### openai.go
OpenAI API 集成实现

### manager.go
AI 管理器，统一接口、缓存、故障转移

## 使用示例

```go
config := &ai.AIManagerConfig{
    DefaultProvider: ai.ProviderOllama,
    EnableCache: true,
    Providers: []ai.ProviderConfig{
        {Name: ai.ProviderOllama, Enabled: true},
        {Name: ai.ProviderOpenAI, Enabled: true},
    },
}

manager, _ := ai.NewAIManager(config)
resp, _ := manager.Translate(ctx, &ai.TranslationRequest{
    Text: "Hello",
    SourceLang: "en",
    TargetLang: "zh-CN",
})
```

## 文档

- [完整使用指南](../../docs/AI_TRANSLATION.md)
- [消息翻译](../../docs/MESSAGE_TRANSLATION.md)

**版本**: v1.0.0
