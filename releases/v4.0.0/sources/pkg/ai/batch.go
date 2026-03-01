package ai

import (
	"context"
	"sync"
	"time"
)

// BatchTranslationRequest represents a batch translation request
type BatchTranslationRequest struct {
	Texts      []string `json:"texts"`
	SourceLang string   `json:"source_lang"`
	TargetLang string   `json:"target_lang"`
	Context    string   `json:"context,omitempty"`
}

// BatchTranslationResponse represents a batch translation response
type BatchTranslationResponse struct {
	Translations []string      `json:"translations"`
	SourceLang   string        `json:"source_lang"`
	TargetLang   string        `json:"target_lang"`
	Duration     time.Duration `json:"duration"`
}

// BatchTranslator handles batch translation
type BatchTranslator struct {
	translator  *Translator
	batchSize   int
	maxWaitTime time.Duration
	queue       chan *BatchItem
	results     chan *BatchResult
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	stats       BatchStats
	mu2         sync.RWMutex
}

// BatchItem represents an item in the batch queue
type BatchItem struct {
	Request    *TranslationRequest
	ResultChan chan *TranslationResult
	CreatedAt  time.Time
}

// BatchResult represents a batch result
type BatchResult struct {
	Request  *BatchTranslationRequest
	Response *BatchTranslationResponse
	Error    error
}

// BatchStats represents batch translation statistics
type BatchStats struct {
	TotalBatches      int           `json:"total_batches"`
	TotalTranslations int           `json:"total_translations"`
	AverageBatchSize  float64       `json:"average_batch_size"`
	AverageDuration   time.Duration `json:"average_duration"`
}

// BatchTranslationConfig represents batch translation configuration
type BatchTranslationConfig struct {
	BatchSize   int           `json:"batch_size"`
	MaxWaitTime time.Duration `json:"max_wait_time"`
	QueueSize   int           `json:"queue_size"`
}

// DefaultBatchTranslationConfig returns default batch translation configuration
func DefaultBatchTranslationConfig() *BatchTranslationConfig {
	return &BatchTranslationConfig{
		BatchSize:   10,
		MaxWaitTime: 100 * time.Millisecond,
		QueueSize:   100,
	}
}

// NewBatchTranslator creates a new batch translator
func NewBatchTranslator(translator *Translator, config *BatchTranslationConfig) *BatchTranslator {
	if config == nil {
		config = DefaultBatchTranslationConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	bt := &BatchTranslator{
		translator:  translator,
		batchSize:   config.BatchSize,
		maxWaitTime: config.MaxWaitTime,
		queue:       make(chan *BatchItem, config.QueueSize),
		results:     make(chan *BatchResult, config.QueueSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start batch processor
	go bt.processBatch()

	return bt
}

// Translate translates a single text with batching
func (bt *BatchTranslator) Translate(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error) {
	resultChan := make(chan *TranslationResult, 1)
	item := &BatchItem{
		Request:    req,
		ResultChan: resultChan,
		CreatedAt:  time.Now(),
	}

	bt.queue <- item

	select {
	case result := <-resultChan:
		return result.Response, result.Error
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// TranslateBatch translates multiple texts in batch
func (bt *BatchTranslator) TranslateBatch(ctx context.Context, req *BatchTranslationRequest) (*BatchTranslationResponse, error) {
	if len(req.Texts) == 0 {
		return &BatchTranslationResponse{
			Translations: make([]string, 0),
			SourceLang:   req.SourceLang,
			TargetLang:   req.TargetLang,
		}, nil
	}

	// Split into batches if too large
	if len(req.Texts) <= bt.batchSize {
		return bt.processSingleBatch(ctx, req)
	}

	// Process in multiple batches
	allTranslations := make([]string, 0, len(req.Texts))

	for i := 0; i < len(req.Texts); i += bt.batchSize {
		end := i + bt.batchSize
		if end > len(req.Texts) {
			end = len(req.Texts)
		}

		batchReq := &BatchTranslationRequest{
			Texts:      req.Texts[i:end],
			SourceLang: req.SourceLang,
			TargetLang: req.TargetLang,
			Context:    req.Context,
		}

		resp, err := bt.processSingleBatch(ctx, batchReq)
		if err != nil {
			return nil, err
		}

		allTranslations = append(allTranslations, resp.Translations...)
	}

	return &BatchTranslationResponse{
		Translations: allTranslations,
		SourceLang:   req.SourceLang,
		TargetLang:   req.TargetLang,
	}, nil
}

// GetStats returns batch translation statistics
func (bt *BatchTranslator) GetStats() BatchStats {
	bt.mu2.RLock()
	defer bt.mu2.RUnlock()
	return bt.stats
}

// Close closes the batch translator
func (bt *BatchTranslator) Close() {
	bt.cancel()
	close(bt.queue)
	close(bt.results)
	bt.wg.Wait()
}

// processBatch processes items from the queue in batches
func (bt *BatchTranslator) processBatch() {
	bt.wg.Add(1)
	defer bt.wg.Done()

	items := make([]*BatchItem, 0, bt.batchSize)
	timer := time.NewTimer(bt.maxWaitTime)
	defer timer.Stop()

	for {
		select {
		case <-bt.ctx.Done():
			// Process remaining items before exit
			if len(items) > 0 {
				bt.processItems(items)
			}
			return

		case item, ok := <-bt.queue:
			if !ok {
				return
			}

			items = append(items, item)

			// Process batch if full
			if len(items) >= bt.batchSize {
				bt.processItems(items)
				items = make([]*BatchItem, 0, bt.batchSize)
				timer.Reset(bt.maxWaitTime)
			}

		case <-timer.C:
			// Process accumulated items
			if len(items) > 0 {
				bt.processItems(items)
				items = make([]*BatchItem, 0, bt.batchSize)
			}
			timer.Reset(bt.maxWaitTime)
		}
	}
}

// processItems processes a batch of items
func (bt *BatchTranslator) processItems(items []*BatchItem) {
	if len(items) == 0 {
		return
	}

	// Group by language pair and context
	groups := make(map[string][]*BatchItem)
	for _, item := range items {
		key := item.Request.SourceLang + "|" + item.Request.TargetLang + "|" + item.Request.Context
		groups[key] = append(groups[key], item)
	}

	// Process each group
	for _, groupItems := range groups {
		if len(groupItems) == 0 {
			continue
		}

		// Create batch request
		texts := make([]string, 0, len(groupItems))
		for _, item := range groupItems {
			texts = append(texts, item.Request.Text)
		}

		batchReq := &BatchTranslationRequest{
			Texts:      texts,
			SourceLang: groupItems[0].Request.SourceLang,
			TargetLang: groupItems[0].Request.TargetLang,
			Context:    groupItems[0].Request.Context,
		}

		// Process batch (simplified - in real implementation, use batch API)
		startTime := time.Now()
		translations := make([]string, 0, len(texts))

		for _, text := range texts {
			resp, err := bt.translator.Translate(bt.ctx, &TranslationRequest{
				Text:       text,
				SourceLang: batchReq.SourceLang,
				TargetLang: batchReq.TargetLang,
				Context:    batchReq.Context,
			})

			if err != nil {
				translations = append(translations, "")
			} else {
				translations = append(translations, resp.TranslatedText)
			}
		}

		batchResp := &BatchTranslationResponse{
			Translations: translations,
			SourceLang:   batchReq.SourceLang,
			TargetLang:   batchReq.TargetLang,
			Duration:     time.Since(startTime),
		}

		// Send results
		for i, item := range groupItems {
			result := &TranslationResult{
				Response: &TranslationResponse{
					TranslatedText: batchResp.Translations[i],
					SourceLang:     batchResp.SourceLang,
					TargetLang:     batchResp.TargetLang,
					Model:          "batch",
					Provider:       ProviderOllama,
					Duration:       batchResp.Duration,
				},
			}
			item.ResultChan <- result
		}

		// Update stats
		bt.mu2.Lock()
		bt.stats.TotalBatches++
		bt.stats.TotalTranslations += len(groupItems)
		bt.stats.AverageBatchSize = float64(bt.stats.TotalTranslations) / float64(bt.stats.TotalBatches)
		totalDuration := time.Duration(bt.stats.TotalBatches-1)*bt.stats.AverageDuration + batchResp.Duration
		bt.stats.AverageDuration = totalDuration / time.Duration(bt.stats.TotalBatches)
		bt.mu2.Unlock()
	}
}

// processSingleBatch processes a single batch request
func (bt *BatchTranslator) processSingleBatch(ctx context.Context, req *BatchTranslationRequest) (*BatchTranslationResponse, error) {
	startTime := time.Now()
	translations := make([]string, 0, len(req.Texts))

	for _, text := range req.Texts {
		resp, err := bt.translator.Translate(ctx, &TranslationRequest{
			Text:       text,
			SourceLang: req.SourceLang,
			TargetLang: req.TargetLang,
			Context:    req.Context,
		})

		if err != nil {
			translations = append(translations, "")
		} else {
			translations = append(translations, resp.TranslatedText)
		}
	}

	return &BatchTranslationResponse{
		Translations: translations,
		SourceLang:   req.SourceLang,
		TargetLang:   req.TargetLang,
		Duration:     time.Since(startTime),
	}, nil
}
