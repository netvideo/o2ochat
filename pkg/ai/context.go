package ai

import (
	"context"
	"strings"
	"sync"
)

// TranslationContext manages translation context
type TranslationContext struct {
	Text        string            `json:"text"`
	Domain      string            `json:"domain"`       // e.g., "medical", "legal", "technical"
	Tone        string            `json:"tone"`         // e.g., "formal", "informal"
	Audience    string            `json:"audience"`     // e.g., "general", "expert"
	Purpose     string            `json:"purpose"`      // e.g., "translation", "localization"
	CustomTerms map[string]string `json:"custom_terms"` // Custom terminology mappings
	mu          sync.RWMutex
}

// ContextExtractor extracts context from text
type ContextExtractor struct {
	domainKeywords map[string][]string
	toneIndicators map[string][]string
}

// NewContextExtractor creates a new context extractor
func NewContextExtractor() *ContextExtractor {
	return &ContextExtractor{
		domainKeywords: map[string][]string{
			"medical":   {"doctor", "patient", "treatment", "diagnosis", "medicine", "医院", "医生", "治疗"},
			"legal":     {"law", "court", "judge", "contract", "legal", "法律", "法院", "合同"},
			"technical": {"software", "hardware", "code", "API", "system", "软件", "硬件", "系统"},
			"business":  {"company", "market", "finance", "investment", "business", "公司", "市场", "投资"},
			"general":   {},
		},
		toneIndicators: map[string][]string{
			"formal":   {"please", "thank you", "respectfully", "您", "请", "谢谢"},
			"informal": {"hey", "hi", "thanks", "你", "嗨", "谢"},
			"neutral":  {},
		},
	}
}

// ExtractContext extracts context from text
func (ce *ContextExtractor) ExtractContext(text string, sourceLang, targetLang string) *TranslationContext {
	context := &TranslationContext{
		Text:        text,
		Domain:      "general",
		Tone:        "neutral",
		Audience:    "general",
		Purpose:     "translation",
		CustomTerms: make(map[string]string),
	}

	// Detect domain
	textLower := strings.ToLower(text)
	maxDomainCount := 0
	for domain, keywords := range ce.domainKeywords {
		count := 0
		for _, keyword := range keywords {
			if strings.Contains(textLower, strings.ToLower(keyword)) {
				count++
			}
		}
		if count > maxDomainCount {
			maxDomainCount = count
			context.Domain = domain
		}
	}

	// Detect tone
	maxToneCount := 0
	for tone, indicators := range ce.toneIndicators {
		count := 0
		for _, indicator := range indicators {
			if strings.Contains(textLower, strings.ToLower(indicator)) {
				count++
			}
		}
		if count > maxToneCount {
			maxToneCount = count
			context.Tone = tone
		}
	}

	return context
}

// AddCustomTerm adds a custom term to context
func (tc *TranslationContext) AddCustomTerm(source, target string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.CustomTerms[source] = target
}

// GetCustomTerm gets a custom term from context
func (tc *TranslationContext) GetCustomTerm(source string) (string, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	term, exists := tc.CustomTerms[source]
	return term, exists
}

// GetContextString returns context as a string for translation requests
func (tc *TranslationContext) GetContextString() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	var parts []string
	if tc.Domain != "general" {
		parts = append(parts, "Domain: "+tc.Domain)
	}
	if tc.Tone != "neutral" {
		parts = append(parts, "Tone: "+tc.Tone)
	}
	if tc.Audience != "general" {
		parts = append(parts, "Audience: "+tc.Audience)
	}
	if tc.Purpose != "translation" {
		parts = append(parts, "Purpose: "+tc.Purpose)
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, ", ")
}

// ApplyContext applies context to translation request
func ApplyContext(req *TranslationRequest, context *TranslationContext) {
	if context == nil {
		return
	}

	// Add context to request
	if req.Context == "" {
		req.Context = context.GetContextString()
	} else {
		req.Context = req.Context + "; " + context.GetContextString()
	}
}

// ContextAwareTranslator wraps translator with context awareness
type ContextAwareTranslator struct {
	translator Translator
	extractor  *ContextExtractor
}

// NewContextAwareTranslator creates a new context-aware translator
func NewContextAwareTranslator(translator Translator) *ContextAwareTranslator {
	return &ContextAwareTranslator{
		translator: translator,
		extractor:  NewContextExtractor(),
	}
}

// Translate translates with context awareness
func (cat *ContextAwareTranslator) Translate(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error) {
	// Extract context if not provided
	if req.Context == "" {
		context := cat.extractor.ExtractContext(req.Text, req.SourceLang, req.TargetLang)
		ApplyContext(req, context)
	}

	// Apply custom terms
	if req.Context != "" {
		// In a real implementation, custom terms would be applied here
		// For now, just pass through
	}

	// Translate with context
	return cat.translator.Translate(ctx, req)
}

// SetCustomTerms sets custom terms for context
func (cat *ContextAwareTranslator) SetCustomTerms(terms map[string]string) {
	// In a real implementation, this would update the context
	_ = terms
}

// GetDomainKeywords returns domain keywords
func (ce *ContextExtractor) GetDomainKeywords() map[string][]string {
	return ce.domainKeywords
}

// GetToneIndicators returns tone indicators
func (ce *ContextExtractor) GetToneIndicators() map[string][]string {
	return ce.toneIndicators
}

// UpdateDomainKeywords updates domain keywords
func (ce *ContextExtractor) UpdateDomainKeywords(domain string, keywords []string) {
	ce.domainKeywords[domain] = keywords
}

// UpdateToneIndicators updates tone indicators
func (ce *ContextExtractor) UpdateToneIndicators(tone string, indicators []string) {
	ce.toneIndicators[tone] = indicators
}
