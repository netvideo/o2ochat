// Package ai provides AI translation capabilities for O2OChat
// Integrating local Ollama and major AI service providers
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ProviderType represents the AI service provider type
type ProviderType string

const (
	// ProviderOllama - Local Ollama instance
	ProviderOllama ProviderType = "ollama"
	// ProviderOpenAI - OpenAI GPT models
	ProviderOpenAI ProviderType = "openai"
	// ProviderAnthropic - Anthropic Claude models
	ProviderAnthropic ProviderType = "anthropic"
	// ProviderGoogle - Google Gemini models
	ProviderGoogle ProviderType = "google"
	// ProviderDeepL - DeepL translation
	ProviderDeepL ProviderType = "deepl"
)

// TranslationRequest represents a translation request
type TranslationRequest struct {
	Text        string  `json:"text"`
	SourceLang  string  `json:"source_lang"`
	TargetLang  string  `json:"target_lang"`
	Context     string  `json:"context,omitempty"`
	Formality   string  `json:"formality,omitempty"` // formal, informal
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
}

// TranslationResponse represents a translation response
type TranslationResponse struct {
	TranslatedText string                 `json:"translated_text"`
	SourceLang     string                 `json:"source_lang"`
	TargetLang     string                 `json:"target_lang"`
	Model          string                 `json:"model"`
	Provider       ProviderType           `json:"provider"`
	Usage          *UsageInfo             `json:"usage,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Duration       time.Duration          `json:"duration"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProviderConfig represents configuration for an AI provider
type ProviderConfig struct {
	Name       ProviderType  `json:"name"`
	Enabled    bool          `json:"enabled"`
	BaseURL    string        `json:"base_url,omitempty"`
	APIKey     string        `json:"api_key,omitempty"`
	Model      string        `json:"model,omitempty"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
}

// Translator provides unified translation interface
type Translator interface {
	Translate(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error)
	TranslateBatch(ctx context.Context, reqs []*TranslationRequest) ([]*TranslationResponse, error)
	GetSupportedLanguages() []string
	GetProviderInfo() ProviderInfo
}

// ProviderInfo contains information about a provider
type ProviderInfo struct {
	Name           ProviderType `json:"name"`
	Model          string       `json:"model"`
	BaseURL        string       `json:"base_url"`
	SupportedLangs []string     `json:"supported_langs"`
	MaxTextLength  int          `json:"max_text_length"`
	Features       []string     `json:"features"` // stream, batch, etc.
}

// SupportedLanguages returns a list of supported language codes
func SupportedLanguages() []string {
	return []string{
		"zh-CN", "zh-TW", "en", "ja", "ko",
		"de", "fr", "es", "ru", "ar",
		"he", "ms", "pt-BR", "it",
		"bo", "mn", "ug",
	}
}

// LanguageNames returns human-readable language names
func LanguageNames() map[string]string {
	return map[string]string{
		"zh-CN": "简体中文", "zh-TW": "繁體中文",
		"en": "English", "ja": "日本語", "ko": "한국어",
		"de": "Deutsch", "fr": "Français", "es": "Español",
		"ru": "Русский", "ar": "العربية", "he": "עברית",
		"ms": "Bahasa Melayu", "pt-BR": "Português", "it": "Italiano",
		"bo": "བོད་ཡིག", "mn": "Монгол", "ug": "ئۇيغۇرچە",
	}
}

// ValidateLanguage checks if a language code is supported
func ValidateLanguage(lang string) bool {
	supported := SupportedLanguages()
	for _, l := range supported {
		if strings.EqualFold(l, lang) {
			return true
		}
	}
	return false
}

// NormalizeLanguage normalizes language code
func NormalizeLanguage(lang string) string {
	lower := strings.ToLower(lang)
	switch lower {
	case "zh", "zh-cn", "zh_cn", "chinese":
		return "zh-CN"
	case "zh-tw", "zh_tw", "traditional chinese":
		return "zh-TW"
	case "en", "english":
		return "en"
	case "ja", "japanese":
		return "ja"
	case "ko", "korean":
		return "ko"
	default:
		for _, supported := range SupportedLanguages() {
			if strings.EqualFold(supported, lang) {
				return supported
			}
		}
		return lang
	}
}

// splitText splits long text into chunks for translation
func splitText(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	current := ""

	for _, word := range strings.Fields(text) {
		if len(current)+len(word)+1 > maxLen {
			chunks = append(chunks, strings.TrimSpace(current))
			current = ""
		}
		current += word + " "
	}

	if strings.TrimSpace(current) != "" {
		chunks = append(chunks, strings.TrimSpace(current))
	}

	return chunks
}

// mergeTranslations merges translated chunks
func mergeTranslations(translations []string) string {
	return strings.Join(translations, " ")
}

// isChinese detects if text contains Chinese characters
func isChinese(text string) bool {
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

// retry executes a function with retry logic
func retry(fn func() error, maxRetries int, delay time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(delay * time.Duration(i+1))
	}
	return err
}

// HTTPClient represents an HTTP client interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient is the default HTTP client implementation
type DefaultHTTPClient struct {
	client *http.Client
}

// NewDefaultHTTPClient creates a new HTTP client
func NewDefaultHTTPClient(timeout time.Duration) *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Do executes an HTTP request
func (c *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// createHTTPClient creates an HTTP client with timeout
func createHTTPClient(timeout time.Duration) HTTPClient {
	return NewDefaultHTTPClient(timeout)
}

// readResponseBody reads and returns the response body
func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// checkHTTPStatus checks if HTTP status is successful
func checkHTTPStatus(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}
	return nil
}

// marshalJSON marshals data to JSON
func marshalJSON(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// unmarshalJSON unmarshals JSON to data
func unmarshalJSON(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}
