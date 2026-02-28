package ai

import (
	"strings"
	"sync"
)

// Term represents a glossary term
type Term struct {
	SourceText  string   `json:"source_text"`
	TargetText  string   `json:"target_text"`
	SourceLang  string   `json:"source_lang"`
	TargetLang  string   `json:"target_lang"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Context     []string `json:"context"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
	UsageCount  int      `json:"usage_count"`
	mu          sync.RWMutex
}

// Glossary manages translation glossary
type Glossary struct {
	terms      map[string]*Term    // key: sourceLang|targetLang|sourceText
	categories map[string][]string // category -> term keys
	mu         sync.RWMutex
	stats      GlossaryStats
	mu2        sync.RWMutex
}

// GlossaryStats represents glossary statistics
type GlossaryStats struct {
	TotalTerms      int `json:"total_terms"`
	TotalCategories int `json:"total_categories"`
	TotalMatches    int `json:"total_matches"`
}

// GlossaryConfig represents glossary configuration
type GlossaryConfig struct {
	AutoExtract bool `json:"auto_extract"`
}

// NewGlossary creates a new glossary
func NewGlossary() *Glossary {
	return &Glossary{
		terms:      make(map[string]*Term),
		categories: make(map[string][]string),
	}
}

// AddTerm adds a term to the glossary
func (g *Glossary) AddTerm(term *Term) error {
	key := g.generateKey(term.SourceLang, term.TargetLang, term.SourceText)

	g.mu.Lock()
	defer g.mu.Unlock()

	if existing, exists := g.terms[key]; exists {
		existing.mu.Lock()
		existing.TargetText = term.TargetText
		existing.Description = term.Description
		existing.UpdatedAt = term.UpdatedAt
		if term.Category != "" {
			existing.Category = term.Category
		}
		existing.mu.Unlock()
	} else {
		g.terms[key] = term
		if term.Category != "" {
			g.categories[term.Category] = append(g.categories[term.Category], key)
			g.mu2.Lock()
			g.stats.TotalCategories = len(g.categories)
			g.mu2.Unlock()
		}
	}

	g.mu2.Lock()
	g.stats.TotalTerms = len(g.terms)
	g.mu2.Unlock()

	return nil
}

// RemoveTerm removes a term from the glossary
func (g *Glossary) RemoveTerm(sourceLang, targetLang, sourceText string) bool {
	key := g.generateKey(sourceLang, targetLang, sourceText)

	g.mu.Lock()
	defer g.mu.Unlock()

	term, exists := g.terms[key]
	if !exists {
		return false
	}

	if term.Category != "" {
		categoryTerms := g.categories[term.Category]
		for i, k := range categoryTerms {
			if k == key {
				g.categories[term.Category] = append(categoryTerms[:i], categoryTerms[i+1:]...)
				break
			}
		}
		if len(g.categories[term.Category]) == 0 {
			delete(g.categories, term.Category)
		}
	}

	delete(g.terms, key)

	g.mu2.Lock()
	g.stats.TotalTerms = len(g.terms)
	g.stats.TotalCategories = len(g.categories)
	g.mu2.Unlock()

	return true
}

// GetTerm gets a term from the glossary
func (g *Glossary) GetTerm(sourceLang, targetLang, sourceText string) *Term {
	key := g.generateKey(sourceLang, targetLang, sourceText)

	g.mu.RLock()
	defer g.mu.RUnlock()

	term, exists := g.terms[key]
	if !exists {
		return nil
	}

	term.mu.Lock()
	term.UsageCount++
	term.mu.Unlock()

	g.mu2.Lock()
	g.stats.TotalMatches++
	g.mu2.Unlock()

	return term
}

// FindTerms finds terms in text
func (g *Glossary) FindTerms(text, sourceLang, targetLang string) []*Term {
	g.mu.RLock()
	defer g.mu.RUnlock()

	found := make([]*Term, 0)
	textLower := strings.ToLower(text)

	for _, term := range g.terms {
		if term.SourceLang != sourceLang || term.TargetLang != targetLang {
			continue
		}

		term.mu.RLock()
		sourceLower := strings.ToLower(term.SourceText)
		if strings.Contains(textLower, sourceLower) {
			found = append(found, term)
		}
		term.mu.RUnlock()
	}

	return found
}

// GetCategories returns all categories
func (g *Glossary) GetCategories() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	categories := make([]string, 0, len(g.categories))
	for cat := range g.categories {
		categories = append(categories, cat)
	}
	return categories
}

// GetTermsByCategory returns terms in a category
func (g *Glossary) GetTermsByCategory(category string) []*Term {
	g.mu.RLock()
	defer g.mu.RUnlock()

	keys, exists := g.categories[category]
	if !exists {
		return nil
	}

	terms := make([]*Term, 0, len(keys))
	for _, key := range keys {
		if term, exists := g.terms[key]; exists {
			terms = append(terms, term)
		}
	}

	return terms
}

// GetStats returns glossary statistics
func (g *Glossary) GetStats() GlossaryStats {
	g.mu2.RLock()
	defer g.mu2.RUnlock()
	return g.stats
}

// Export exports the glossary to a map
func (g *Glossary) Export() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	data := make(map[string]interface{})
	terms := make([]map[string]interface{}, 0, len(g.terms))

	for _, term := range g.terms {
		term.mu.RLock()
		termData := map[string]interface{}{
			"source_text": term.SourceText,
			"target_text": term.TargetText,
			"source_lang": term.SourceLang,
			"target_lang": term.TargetLang,
			"category":    term.Category,
			"description": term.Description,
			"context":     term.Context,
			"usage_count": term.UsageCount,
		}
		term.mu.RUnlock()
		terms = append(terms, termData)
	}

	data["terms"] = terms
	data["stats"] = g.stats

	return data
}

// Import imports glossary from a map
func (g *Glossary) Import(data map[string]interface{}) error {
	termsData, ok := data["terms"].([]interface{})
	if !ok {
		return nil
	}

	for _, termData := range termsData {
		termMap, ok := termData.(map[string]interface{})
		if !ok {
			continue
		}

		term := &Term{
			SourceText:  termMap["source_text"].(string),
			TargetText:  termMap["target_text"].(string),
			SourceLang:  termMap["source_lang"].(string),
			TargetLang:  termMap["target_lang"].(string),
			Category:    termMap["category"].(string),
			Description: termMap["description"].(string),
		}

		g.AddTerm(term)
	}

	return nil
}

// Clear clears the glossary
func (g *Glossary) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.terms = make(map[string]*Term)
	g.categories = make(map[string][]string)

	g.mu2.Lock()
	g.stats = GlossaryStats{}
	g.mu2.Unlock()
}

// generateKey generates a glossary key
func (g *Glossary) generateKey(sourceLang, targetLang, sourceText string) string {
	return sourceLang + "|" + targetLang + "|" + sourceText
}

// ApplyGlossary applies glossary terms to text
func ApplyGlossary(text string, terms []*Term) string {
	result := text
	for _, term := range terms {
		term.mu.RLock()
		result = strings.ReplaceAll(result, term.SourceText, term.TargetText)
		term.mu.RUnlock()
	}
	return result
}
