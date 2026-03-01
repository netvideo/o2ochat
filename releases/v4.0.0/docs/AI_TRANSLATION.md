# O2OChat AI 翻译功能

## 概述

O2OChat 集成了 AI 翻译功能，支持本地 Ollama 和主要 AI 服务商的 API 调用。

## 支持的 AI 提供商

### 1. Ollama (本地)
- **优点**: 离线使用、隐私保护、免费
- **模型**: llama2, mistral, codellama 等
- **URL**: http://localhost:11434

### 2. OpenAI
- **优点**: 高质量翻译、上下文理解
- **模型**: gpt-3.5-turbo, gpt-4
- **API**: https://api.openai.com/v1

### 3. Anthropic (即将支持)
- **模型**: claude-3-opus, claude-3-sonnet
- **API**: https://api.anthropic.com

### 4. Google (即将支持)
- **模型**: gemini-pro, gemini-ultra
- **API**: https://generativelanguage.googleapis.com

### 5. DeepL (即将支持)
- **优点**: 专业翻译质量
- **API**: https://api.deepl.com

## 配置示例

### 基本配置

```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/netvideo/o2ochat/pkg/ai"
)

func main() {
	// 创建 AI 管理器配置
	config := &ai.AIManagerConfig{
		DefaultProvider: ai.ProviderOllama,
		EnableCache:     true,
		CacheSize:       1000,
		CacheTTL:        time.Hour,
		FallbackEnabled: true,
		Timeout:         30 * time.Second,
		Providers: []ai.ProviderConfig{
			{
				Name:    ai.ProviderOllama,
				Enabled: true,
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
				Timeout: 30 * time.Second,
			},
			{
				Name:    ai.ProviderOpenAI,
				Enabled: true,
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "your-api-key",
				Model:   "gpt-3.5-turbo",
				Timeout: 30 * time.Second,
			},
		},
	}
	
	// 创建 AI 管理器
	manager, err := ai.NewAIManager(config)
	if err != nil {
		panic(err)
	}
	
	// 翻译文本
	ctx := context.Background()
	req := &ai.TranslationRequest{
		Text:       "你好，世界！",
		SourceLang: "zh-CN",
		TargetLang: "en",
	}
	
	resp, err := manager.Translate(ctx, req)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("原文：%s\n", req.Text)
	fmt.Printf("译文：%s\n", resp.TranslatedText)
	fmt.Printf("提供商：%s\n", resp.Provider)
	fmt.Printf("耗时：%v\n", resp.Duration)
}
```

## Ollama 本地部署

### 1. 安装 Ollama

```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Windows
# 下载 https://ollama.ai/download
```

### 2. 拉取模型

```bash
# 拉取 llama2 模型
ollama pull llama2

# 拉取其他模型
ollama pull mistral
ollama pull codellama
```

### 3. 启动服务

```bash
ollama serve
```

### 4. 测试

```bash
curl http://localhost:11434/api/tags
```

## API 使用

### 单一翻译

```go
// 创建请求
req := &ai.TranslationRequest{
	Text:       "Hello, welcome to O2OChat!",
	SourceLang: "en",
	TargetLang: "zh-CN",
	Context:    "chat message",
	Formality:  "informal",
}

// 执行翻译
resp, err := manager.Translate(ctx, req)
```

### 批量翻译

```go
// 创建多个请求
reqs := []*ai.TranslationRequest{
	{Text: "Hello", SourceLang: "en", TargetLang: "zh-CN"},
	{Text: "World", SourceLang: "en", TargetLang: "zh-CN"},
	{Text: "Welcome", SourceLang: "en", TargetLang: "zh-CN"},
}

// 批量翻译
responses, err := manager.TranslateBatch(ctx, reqs)
```

### 切换提供商

```go
// 切换到 OpenAI
err := manager.SetActiveProvider(ai.ProviderOpenAI)

// 获取当前提供商
current := manager.GetActiveProvider()

// 查看所有可用提供商
providers := manager.ListProviders()
```

### 健康检查

```go
// 检查所有提供商健康状态
health := manager.CheckHealth(ctx)

for provider, healthy := range health {
	if healthy {
		fmt.Printf("%s: OK\n", provider)
	} else {
		fmt.Printf("%s: DOWN\n", provider)
	}
}
```

## 支持的语言

| 代码 | 语言 | 代码 | 语言 |
|------|------|------|------|
| zh-CN | 简体中文 | ja | 日本語 |
| zh-TW | 繁體中文 | ko | 한국어 |
| en | English | de | Deutsch |
| fr | Français | es | Español |
| ru | Русский | ar | العربية |
| he | עברית | ms | Bahasa Melayu |
| pt-BR | Português | it | Italiano |
| bo | བོད་ཡིག | mn | Монгол |
| ug | ئۇيغۇرچە | | |

## 高级功能

### 上下文感知翻译

```go
req := &ai.TranslationRequest{
	Text:       "Bank",
	SourceLang: "en",
	TargetLang: "zh-CN",
	Context:    "financial institution", // 上下文：金融机构
}
// 结果：银行 (而不是河岸)
```

### 正式/非正式语气

```go
req := &ai.TranslationRequest{
	Text:       "您好",
	SourceLang: "zh-CN",
	TargetLang: "en",
	Formality:  "formal", // 正式
}
// 结果：Hello (formal)

req.Formality = "informal" // 非正式
// 结果：Hi
```

### 缓存配置

```go
config := &ai.AIManagerConfig{
	EnableCache: true,
	CacheSize:   1000,      // 缓存 1000 条
	CacheTTL:    time.Hour, // 1 小时过期
}
```

### 故障转移

```go
config := &ai.AIManagerConfig{
	FallbackEnabled: true,
	Providers: []ai.ProviderConfig{
		{Name: ai.ProviderOllama, Enabled: true}, // 主
		{Name: ai.ProviderOpenAI, Enabled: true}, // 备用
	},
}
// Ollama 失败时自动切换到 OpenAI
```

## 性能优化

### 1. 使用缓存
- 重复内容直接从缓存获取
- 减少 API 调用
- 降低成本

### 2. 批量翻译
- 合并多个请求
- 减少网络开销
- 提高效率

### 3. 本地优先
- 优先使用 Ollama
- 离线可用
- 保护隐私

### 4. 选择合适的模型
- 简单翻译：llama2, mistral
- 高质量：gpt-4, claude-3
- 专业领域：codellama (代码)

## 成本估算

| 提供商 | 模型 | 价格 (每 1K tokens) | 适用场景 |
|--------|------|------------------|---------|
| Ollama | llama2 | 免费 (本地) | 日常翻译 |
| OpenAI | gpt-3.5-turbo | $0.001 | 一般质量 |
| OpenAI | gpt-4 | $0.03 | 高质量 |
| DeepL | Pro | €4.99/月 | 专业翻译 |

## 安全与隐私

### 本地 Ollama
- ✅ 数据不出本地
- ✅ 无需网络
- ✅ 完全控制

### 云服务
- ⚠️ 数据发送到第三方
- ⚠️ 需要 API Key
- ⚠️ 注意隐私政策

### 最佳实践
1. 敏感内容使用本地 Ollama
2. 使用 HTTPS 加密传输
3. 定期轮换 API Key
4. 设置使用限额

## 故障排除

### Ollama 连接失败

```bash
# 检查服务是否运行
ollama list

# 重启服务
ollama serve
```

### API Key 无效

```bash
# 检查 API Key 是否正确
# 查看服务商文档
```

### 翻译质量差

```bash
# 尝试更强大的模型
# 提供更多上下文
# 调整 temperature 参数
```

## 相关文档

- [AI 模块代码](../pkg/ai/)
- [示例代码](../examples/ai/)
- [API 文档](../docs/AI_API.md)

---

**版本**: v1.0.0  
**更新时间**: 2026 年 2 月 28 日
