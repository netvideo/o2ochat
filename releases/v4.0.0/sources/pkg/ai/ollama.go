package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaTranslator implements Translator for Ollama
type OllamaTranslator struct {
	config     *ProviderConfig
	httpClient HTTPClient
	model      string
}

// OllamaRequest represents Ollama API request
type OllamaRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	Stream  bool          `json:"stream"`
	Options OllamaOptions `json:"options,omitempty"`
}

// OllamaOptions represents Ollama model options
type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
}

// OllamaResponse represents Ollama API response
type OllamaResponse struct {
	Model         string `json:"model"`
	Response      string `json:"response"`
	Done          bool   `json:"done"`
	Context       []int  `json:"context,omitempty"`
	TotalDuration int64  `json:"total_duration,omitempty"`
}

// NewOllamaTranslator creates a new Ollama translator
func NewOllamaTranslator(config *ProviderConfig) *OllamaTranslator {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model := config.Model
	if model == "" {
		model = "llama2" // Default model
	}

	return &OllamaTranslator{
		config:     config,
		httpClient: createHTTPClient(config.Timeout),
		model:      model,
	}
}

// Translate translates text using Ollama
func (o *OllamaTranslator) Translate(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error) {
	startTime := time.Now()

	// Normalize languages
	sourceLang := NormalizeLanguage(req.SourceLang)
	targetLang := NormalizeLanguage(req.TargetLang)

	// Build translation prompt
	prompt := o.buildTranslationPrompt(req.Text, sourceLang, targetLang, req.Context)

	// Create Ollama request
	ollamaReq := &OllamaRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false,
		Options: OllamaOptions{
			Temperature: 0.3, // Lower temperature for more accurate translations
			NumPredict:  2048,
		},
	}

	// Send request
	response, err := o.sendRequest(ctx, ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}

	duration := time.Since(startTime)

	// Parse response
	translatedText := response.Response

	return &TranslationResponse{
		TranslatedText: translatedText,
		SourceLang:     sourceLang,
		TargetLang:     targetLang,
		Model:          o.model,
		Provider:       ProviderOllama,
		Duration:       duration,
		Metadata: map[string]interface{}{
			"total_duration": response.TotalDuration,
		},
	}, nil
}

// TranslateBatch translates multiple texts
func (o *OllamaTranslator) TranslateBatch(ctx context.Context, reqs []*TranslationRequest) ([]*TranslationResponse, error) {
	responses := make([]*TranslationResponse, len(reqs))

	for i, req := range reqs {
		resp, err := o.Translate(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("batch translation failed at index %d: %w", i, err)
		}
		responses[i] = resp
	}

	return responses, nil
}

// GetSupportedLanguages returns supported languages
func (o *OllamaTranslator) GetSupportedLanguages() []string {
	return SupportedLanguages()
}

// GetProviderInfo returns provider information
func (o *OllamaTranslator) GetProviderInfo() ProviderInfo {
	return ProviderInfo{
		Name:           ProviderOllama,
		Model:          o.model,
		BaseURL:        o.config.BaseURL,
		SupportedLangs: SupportedLanguages(),
		MaxTextLength:  4096, // Depends on model context
		Features:       []string{"local", "offline", "stream"},
	}
}

// buildTranslationPrompt creates the translation prompt
func (o *OllamaTranslator) buildTranslationPrompt(text, sourceLang, targetLang, context string) string {
	sourceName := LanguageNames()[sourceLang]
	targetName := LanguageNames()[targetLang]

	if sourceName == "" {
		sourceName = sourceLang
	}
	if targetName == "" {
		targetName = targetLang
	}

	prompt := fmt.Sprintf("Translate the following text from %s to %s:\n\n", sourceName, targetName)

	if context != "" {
		prompt += fmt.Sprintf("Context: %s\n\n", context)
	}

	prompt += fmt.Sprintf("Original: %s\n\n", text)
	prompt += "Translation:"

	return prompt
}

// sendRequest sends HTTP request to Ollama API
func (o *OllamaTranslator) sendRequest(ctx context.Context, req *OllamaRequest) (*OllamaResponse, error) {
	url := fmt.Sprintf("%s/api/generate", o.config.BaseURL)

	jsonData, err := marshalJSON(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if err := checkHTTPStatus(resp); err != nil {
		return nil, err
	}

	body, err := readResponseBody(resp)
	if err != nil {
		return nil, err
	}

	var ollamaResp OllamaResponse
	if err := unmarshalJSON(body, &ollamaResp); err != nil {
		return nil, err
	}

	return &ollamaResp, nil
}

// ListModels lists available Ollama models
func (o *OllamaTranslator) ListModels(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/api/tags", o.config.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if err := checkHTTPStatus(resp); err != nil {
		return nil, err
	}

	body, err := readResponseBody(resp)
	if err != nil {
		return nil, err
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := unmarshalJSON(body, &result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Models))
	for i, model := range result.Models {
		models[i] = model.Name
	}

	return models, nil
}

// IsAvailable checks if Ollama service is available
func (o *OllamaTranslator) IsAvailable(ctx context.Context) bool {
	url := fmt.Sprintf("%s/api/tags", o.config.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return false
	}

	return checkHTTPStatus(resp) == nil
}
