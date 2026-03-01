package ai

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

// OpenAITranslator implements Translator for OpenAI
type OpenAITranslator struct {
	config     *ProviderConfig
	httpClient HTTPClient
	model      string
}

// OpenAIRequest represents OpenAI Chat Completion API request
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// OpenAIMessage represents a chat message
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents OpenAI API response
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      OpenAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAITranslator creates a new OpenAI translator
func NewOpenAITranslator(config *ProviderConfig) *OpenAITranslator {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := config.Model
	if model == "" {
		model = "gpt-3.5-turbo" // Default model
	}

	return &OpenAITranslator{
		config:     config,
		httpClient: createHTTPClient(config.Timeout),
		model:      model,
	}
}

// Translate translates text using OpenAI
func (o *OpenAITranslator) Translate(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error) {
	startTime := time.Now()

	// Normalize languages
	sourceLang := NormalizeLanguage(req.SourceLang)
	targetLang := NormalizeLanguage(req.TargetLang)

	// Build messages for Chat Completion API
	messages := o.buildTranslationMessages(req.Text, sourceLang, targetLang, req.Context)

	// Create OpenAI request
	openaiReq := &OpenAIRequest{
		Model:       o.model,
		Messages:    messages,
		Temperature: 0.3,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	// Send request
	response, err := o.sendRequest(ctx, openaiReq)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}

	duration := time.Since(startTime)

	// Parse response
	var translatedText string
	if len(response.Choices) > 0 {
		translatedText = response.Choices[0].Message.Content
	}

	return &TranslationResponse{
		TranslatedText: translatedText,
		SourceLang:     sourceLang,
		TargetLang:     targetLang,
		Model:          o.model,
		Provider:       ProviderOpenAI,
		Usage: &UsageInfo{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
		Duration: duration,
		Metadata: map[string]interface{}{
			"model": response.Model,
			"id":    response.ID,
		},
	}, nil
}

// TranslateBatch translates multiple texts
func (o *OpenAITranslator) TranslateBatch(ctx context.Context, reqs []*TranslationRequest) ([]*TranslationResponse, error) {
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
func (o *OpenAITranslator) GetSupportedLanguages() []string {
	return SupportedLanguages()
}

// GetProviderInfo returns provider information
func (o *OpenAITranslator) GetProviderInfo() ProviderInfo {
	return ProviderInfo{
		Name:           ProviderOpenAI,
		Model:          o.model,
		BaseURL:        o.config.BaseURL,
		SupportedLangs: SupportedLanguages(),
		MaxTextLength:  4096,
		Features:       []string{"high_quality", "context_aware", "batch"},
	}
}

// buildTranslationMessages creates messages for Chat Completion API
func (o *OpenAITranslator) buildTranslationMessages(text, sourceLang, targetLang, context string) []OpenAIMessage {
	sourceName := LanguageNames()[sourceLang]
	targetName := LanguageNames()[targetLang]

	if sourceName == "" {
		sourceName = sourceLang
	}
	if targetName == "" {
		targetName = targetLang
	}

	systemContent := fmt.Sprintf("You are a professional translator. Translate from %s to %s.", sourceName, targetName)

	if context != "" {
		systemContent += " Context: " + context
	}

	systemContent += " Provide only the translation, nothing else."

	messages := []OpenAIMessage{
		{
			Role:    "system",
			Content: systemContent,
		},
		{
			Role:    "user",
			Content: text,
		},
	}

	return messages
}

// sendRequest sends HTTP request to OpenAI API
func (o *OpenAITranslator) sendRequest(ctx context.Context, req *OpenAIRequest) (*OpenAIResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", o.config.BaseURL)

	jsonData, err := marshalJSON(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.config.APIKey)

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

	var openaiResp OpenAIResponse
	if err := unmarshalJSON(body, &openaiResp); err != nil {
		return nil, err
	}

	return &openaiResp, nil
}

// IsAvailable checks if OpenAI API is available
func (o *OpenAITranslator) IsAvailable(ctx context.Context) bool {
	url := fmt.Sprintf("%s/models", o.config.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	httpReq.Header.Set("Authorization", "Bearer "+o.config.APIKey)

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return false
	}

	return checkHTTPStatus(resp) == nil
}
