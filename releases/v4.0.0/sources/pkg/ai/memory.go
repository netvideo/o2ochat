package ai

import (
	"strings"
	"sync"
)

// TranslationMemorySegment represents a translation memory segment
type TranslationMemorySegment struct {
	ID           string  `json:"id"`
	SourceText   string  `json:"source_text"`
	TargetText   string  `json:"target_text"`
	SourceLang   string  `json:"source_lang"`
	TargetLang   string  `json:"target_lang"`
	Context      string  `json:"context"`
	QualityScore float64 `json:"quality_score"` // 0-100
	UsageCount   int     `json:"usage_count"`
	CreatedAt    int64   `json:"created_at"`
	LastUsedAt   int64   `json:"last_used_at"`
	mu           sync.RWMutex
}

// TranslationMemory manages translation memory
type TranslationMemory struct {
	segments   map[string]*TranslationMemorySegment // key: sourceLang|targetLang|sourceText
	fuzzyIndex map[string][]string                  // fuzzy search index
	mu         sync.RWMutex
	stats      TranslationMemoryStats
	mu2        sync.RWMutex
}

// TranslationMemoryStats represents translation memory statistics
type TranslationMemoryStats struct {
	TotalSegments int     `json:"total_segments"`
	TotalMatches  int     `json:"total_matches"`
	AverageScore  float64 `json:"average_score"`
	ExactMatches  int     `json:"exact_matches"`
	FuzzyMatches  int     `json:"fuzzy_matches"`
}

// TranslationMemoryMatch represents a TM match result
type TranslationMemoryMatch struct {
	Segment    *TranslationMemorySegment `json:"segment"`
	Similarity float64                   `json:"similarity"` // 0-100
	MatchType  string                    `json:"match_type"` // "exact", "fuzzy"
}

// TranslationMemoryConfig represents TM configuration
type TranslationMemoryConfig struct {
	MinFuzzyScore float64 `json:"min_fuzzy_score"` // Minimum fuzzy match score (default: 70)
	MaxSegments   int     `json:"max_segments"`    // Maximum segments to store
}

// DefaultTranslationMemoryConfig returns default TM configuration
func DefaultTranslationMemoryConfig() *TranslationMemoryConfig {
	return &TranslationMemoryConfig{
		MinFuzzyScore: 70.0,
		MaxSegments:   100000,
	}
}

// NewTranslationMemory creates a new translation memory
func NewTranslationMemory() *TranslationMemory {
	return &TranslationMemory{
		segments:   make(map[string]*TranslationMemorySegment),
		fuzzyIndex: make(map[string][]string),
	}
}

// AddSegment adds a segment to translation memory
func (tm *TranslationMemory) AddSegment(segment *TranslationMemorySegment) error {
	key := tm.generateKey(segment.SourceLang, segment.TargetLang, segment.SourceText)

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if exists
	if existing, exists := tm.segments[key]; exists {
		existing.mu.Lock()
		existing.TargetText = segment.TargetText
		existing.QualityScore = segment.QualityScore
		existing.LastUsedAt = segment.LastUsedAt
		existing.mu.Unlock()
	} else {
		// Evict if necessary
		if len(tm.segments) >= 100000 {
			tm.evictOldest()
		}
		tm.segments[key] = segment

		// Add to fuzzy index
		tm.addToFuzzyIndex(key, segment.SourceText)
	}

	tm.mu2.Lock()
	tm.stats.TotalSegments = len(tm.segments)
	tm.mu2.Unlock()

	return nil
}

// FindMatches finds matching segments
func (tm *TranslationMemory) FindMatches(sourceText, sourceLang, targetLang string) []*TranslationMemoryMatch {
	key := tm.generateKey(sourceLang, targetLang, sourceText)

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	matches := make([]*TranslationMemoryMatch, 0)

	// Exact match
	if segment, exists := tm.segments[key]; exists {
		segment.mu.RLock()
		matches = append(matches, &TranslationMemoryMatch{
			Segment:    segment,
			Similarity: 100.0,
			MatchType:  "exact",
		})
		segment.mu.RUnlock()

		tm.mu2.Lock()
		tm.stats.ExactMatches++
		tm.stats.TotalMatches++
		tm.mu2.Unlock()

		return matches
	}

	// Fuzzy match
	fuzzyMatches := tm.findFuzzyMatches(sourceText, sourceLang, targetLang)
	matches = append(matches, fuzzyMatches...)

	if len(fuzzyMatches) > 0 {
		tm.mu2.Lock()
		tm.stats.FuzzyMatches++
		tm.stats.TotalMatches++
		tm.mu2.Unlock()
	}

	return matches
}

// GetBestMatch gets the best matching segment
func (tm *TranslationMemory) GetBestMatch(sourceText, sourceLang, targetLang, minScore float64) *TranslationMemoryMatch {
	matches := tm.FindMatches(sourceText, sourceLang, targetLang)
	if len(matches) == 0 {
		return nil
	}

	best := matches[0]
	for _, match := range matches {
		if match.Similarity > best.Similarity && match.Similarity >= minScore {
			best = match
		}
	}

	if best.Similarity >= minScore {
		return best
	}

	return nil
}

// GetStats returns TM statistics
func (tm *TranslationMemory) GetStats() TranslationMemoryStats {
	tm.mu2.RLock()
	defer tm.mu2.RUnlock()
	return tm.stats
}

// Clear clears the translation memory
func (tm *TranslationMemory) Clear() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.segments = make(map[string]*TranslationMemorySegment)
	tm.fuzzyIndex = make(map[string][]string)

	tm.mu2.Lock()
	tm.stats = TranslationMemoryStats{}
	tm.mu2.Unlock()
}

// Export exports TM to a map
func (tm *TranslationMemory) Export() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	data := make(map[string]interface{})
	segments := make([]map[string]interface{}, 0, len(tm.segments))

	for _, segment := range tm.segments {
		segment.mu.RLock()
		segmentData := map[string]interface{}{
			"id":            segment.ID,
			"source_text":   segment.SourceText,
			"target_text":   segment.TargetText,
			"source_lang":   segment.SourceLang,
			"target_lang":   segment.TargetLang,
			"context":       segment.Context,
			"quality_score": segment.QualityScore,
			"usage_count":   segment.UsageCount,
		}
		segment.mu.RUnlock()
		segments = append(segments, segmentData)
	}

	data["segments"] = segments
	data["stats"] = tm.stats

	return data
}

// Import imports TM from a map
func (tm *TranslationMemory) Import(data map[string]interface{}) error {
	segmentsData, ok := data["segments"].([]interface{})
	if !ok {
		return nil
	}

	for _, segmentData := range segmentsData {
		segmentMap, ok := segmentData.(map[string]interface{})
		if !ok {
			continue
		}

		segment := &TranslationMemorySegment{
			ID:           segmentMap["id"].(string),
			SourceText:   segmentMap["source_text"].(string),
			TargetText:   segmentMap["target_text"].(string),
			SourceLang:   segmentMap["source_lang"].(string),
			TargetLang:   segmentMap["target_lang"].(string),
			Context:      segmentMap["context"].(string),
			QualityScore: segmentMap["quality_score"].(float64),
			UsageCount:   int(segmentMap["usage_count"].(float64)),
		}

		tm.AddSegment(segment)
	}

	return nil
}

// generateKey generates a TM key
func (tm *TranslationMemory) generateKey(sourceLang, targetLang, sourceText string) string {
	return sourceLang + "|" + targetLang + "|" + sourceText
}

// addToFuzzyIndex adds a segment to fuzzy index
func (tm *TranslationMemory) addToFuzzyIndex(key, sourceText string) {
	// Simple word-based indexing
	words := strings.Fields(strings.ToLower(sourceText))
	for _, word := range words {
		if len(word) < 3 {
			continue
		}
		tm.fuzzyIndex[word] = append(tm.fuzzyIndex[word], key)
	}
}

// findFuzzyMatches finds fuzzy matching segments
func (tm *TranslationMemory) findFuzzyMatches(sourceText, sourceLang, targetLang string) []*TranslationMemoryMatch {
	matches := make([]*TranslationMemoryMatch, 0)

	// Get candidate keys from index
	candidateKeys := make(map[string]bool)
	words := strings.Fields(strings.ToLower(sourceText))
	for _, word := range words {
		if len(word) < 3 {
			continue
		}
		if keys, exists := tm.fuzzyIndex[word]; exists {
			for _, key := range keys {
				if strings.HasPrefix(key, sourceLang+"|"+targetLang+"|") {
					candidateKeys[key] = true
				}
			}
		}
	}

	// Calculate similarity for candidates
	for key := range candidateKeys {
		segment, exists := tm.segments[key]
		if !exists {
			continue
		}

		segment.mu.RLock()
		similarity := tm.calculateSimilarity(sourceText, segment.SourceText)
		segment.mu.RUnlock()

		if similarity >= 70.0 { // Minimum fuzzy threshold
			segment.mu.RLock()
			matches = append(matches, &TranslationMemoryMatch{
				Segment:    segment,
				Similarity: similarity,
				MatchType:  "fuzzy",
			})
			segment.mu.RUnlock()

			// Update usage
			segment.mu.Lock()
			segment.UsageCount++
			segment.LastUsedAt = time.Now().Unix()
			segment.mu.Unlock()
		}
	}

	// Sort by similarity (simple bubble sort for small lists)
	for i := 0; i < len(matches)-1; i++ {
		for j := 0; j < len(matches)-i-1; j++ {
			if matches[j].Similarity < matches[j+1].Similarity {
				matches[j], matches[j+1] = matches[j+1], matches[j]
			}
		}
	}

	return matches
}

// calculateSimilarity calculates text similarity (simple Levenshtein-based)
func (tm *TranslationMemory) calculateSimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if s1 == s2 {
		return 100.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Simple word overlap similarity
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	matchCount := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 {
				matchCount++
				break
			}
		}
	}

	total := len(words1) + len(words2)
	if total == 0 {
		return 0.0
	}

	return float64(2*matchCount) / float64(total) * 100.0
}

// evictOldest evicts the oldest unused segment
func (tm *TranslationMemory) evictOldest() {
	var oldestKey string
	var oldestTime int64 = time.Now().Unix()

	for key, segment := range tm.segments {
		segment.mu.RLock()
		if segment.LastUsedAt < oldestTime {
			oldestTime = segment.LastUsedAt
			oldestKey = key
		}
		segment.mu.RUnlock()
	}

	if oldestKey != "" {
		segment, exists := tm.segments[oldestKey]
		if !exists {
			return
		}

		// Remove from fuzzy index
		words := strings.Fields(strings.ToLower(segment.SourceText))
		for _, word := range words {
			if len(word) < 3 {
				continue
			}
			if keys, exists := tm.fuzzyIndex[word]; exists {
				for i, k := range keys {
					if k == oldestKey {
						tm.fuzzyIndex[word] = append(keys[:i], keys[i+1:]...)
						break
					}
				}
			}
		}

		delete(tm.segments, oldestKey)
	}
}

// ApplyTM applies translation memory to translation
func ApplyTM(translator Translator, sourceText, sourceLang, targetLang string, tm *TranslationMemory, minScore float64) (string, error) {
	// Try to find match in TM
	match := tm.GetBestMatch(sourceText, sourceLang, targetLang, minScore)
	if match != nil {
		// Use TM match
		match.Segment.mu.RLock()
		targetText := match.Segment.TargetText
		match.Segment.mu.RUnlock()
		return targetText, nil
	}

	// Translate normally
	resp, err := translator.Translate(context.Background(), &TranslationRequest{
		Text:       sourceText,
		SourceLang: sourceLang,
		TargetLang: targetLang,
	})

	if err != nil {
		return "", err
	}

	// Add to TM
	segment := &TranslationMemorySegment{
		ID:           generateUUID(),
		SourceText:   sourceText,
		TargetText:   resp.TranslatedText,
		SourceLang:   sourceLang,
		TargetLang:   targetLang,
		QualityScore: 100.0,
		UsageCount:   1,
		CreatedAt:    time.Now().Unix(),
		LastUsedAt:   time.Now().Unix(),
	}

	tm.AddSegment(segment)

	return resp.TranslatedText, nil
}
