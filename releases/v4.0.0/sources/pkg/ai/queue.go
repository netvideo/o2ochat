package ai

import (
	"context"
	"sync"
	"time"
)

// TranslationResult represents a translation result wrapper
type TranslationResult struct {
	Response *TranslationResponse
	Error    error
}

// StreamingTranslator handles streaming translation
type StreamingTranslator struct {
	translator Translator
}

// NewStreamingTranslator creates a new streaming translator
func NewStreamingTranslator(translator Translator) *StreamingTranslator {
	return &StreamingTranslator{
		translator: translator,
	}
}

// TranslateStream translates with streaming response
func (st *StreamingTranslator) TranslateStream(ctx context.Context, req *TranslationRequest) (<-chan *TranslationResult, error) {
	streamChan := make(chan *TranslationResult, 10)

	// Set stream flag
	req.Stream = true

	go func() {
		defer close(streamChan)

		// For now, use regular translation (streaming requires provider-specific implementation)
		resp, err := st.translator.Translate(ctx, req)
		if err != nil {
			streamChan <- &TranslationResult{Error: err}
			return
		}

		// Send response in chunks (simplified)
		chunkSize := 10
		text := resp.TranslatedText
		for i := 0; i < len(text); i += chunkSize {
			end := i + chunkSize
			if end > len(text) {
				end = len(text)
			}

			chunkResp := &TranslationResponse{
				TranslatedText: text[i:end],
				SourceLang:     resp.SourceLang,
				TargetLang:     resp.TargetLang,
				Model:          resp.Model,
				Provider:       resp.Provider,
				Duration:       resp.Duration / time.Duration((len(text)+chunkSize-1)/chunkSize),
			}

			streamChan <- &TranslationResult{Response: chunkResp}

			// Small delay to simulate streaming
			select {
			case <-time.After(10 * time.Millisecond):
			case <-ctx.Done():
				return
			}
		}
	}()

	return streamChan, nil
}

// TranslationQueue manages translation queue
type TranslationQueue struct {
	queue      chan *TranslationRequest
	results    chan *TranslationResult
	translator Translator
	maxSize    int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	stats      QueueStats
	mu2        sync.RWMutex
}

// QueueStats represents queue statistics
type QueueStats struct {
	QueueSize       int           `json:"queue_size"`
	TotalProcessed  int           `json:"total_processed"`
	AverageWaitTime time.Duration `json:"average_wait_time"`
	AverageDuration time.Duration `json:"average_duration"`
}

// TranslationQueueConfig represents queue configuration
type TranslationQueueConfig struct {
	MaxSize    int `json:"max_size"`
	NumWorkers int `json:"num_workers"`
}

// DefaultTranslationQueueConfig returns default queue configuration
func DefaultTranslationQueueConfig() *TranslationQueueConfig {
	return &TranslationQueueConfig{
		MaxSize:    1000,
		NumWorkers: 10,
	}
}

// NewTranslationQueue creates a new translation queue
func NewTranslationQueue(translator Translator, config *TranslationQueueConfig) *TranslationQueue {
	if config == nil {
		config = DefaultTranslationQueueConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	tq := &TranslationQueue{
		queue:      make(chan *TranslationRequest, config.MaxSize),
		results:    make(chan *TranslationResult, config.MaxSize),
		translator: translator,
		maxSize:    config.MaxSize,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start workers
	for i := 0; i < config.NumWorkers; i++ {
		tq.wg.Add(1)
		go tq.worker(i)
	}

	return tq
}

// Submit submits a translation request to the queue
func (tq *TranslationQueue) Submit(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error) {
	select {
	case tq.queue <- req:
		// Wait for result
		select {
		case result := <-tq.results:
			tq.mu2.Lock()
			tq.stats.TotalProcessed++
			tq.mu2.Unlock()
			return result.Response, result.Error
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetStats returns queue statistics
func (tq *TranslationQueue) GetStats() QueueStats {
	tq.mu2.RLock()
	defer tq.mu2.RUnlock()

	stats := tq.stats
	stats.QueueSize = len(tq.queue)
	return stats
}

// Close closes the translation queue
func (tq *TranslationQueue) Close() {
	tq.cancel()
	tq.wg.Wait()
	close(tq.queue)
	close(tq.results)
}

// worker processes translation requests
func (tq *TranslationQueue) worker(id int) {
	defer tq.wg.Done()

	for {
		select {
		case <-tq.ctx.Done():
			return
		case req, ok := <-tq.queue:
			if !ok {
				return
			}

			startTime := time.Now()
			resp, err := tq.translator.Translate(tq.ctx, req)
			duration := time.Since(startTime)

			result := &TranslationResult{
				Response: resp,
				Error:    err,
			}

			// Update stats
			tq.mu2.Lock()
			tq.stats.AverageWaitTime = (tq.stats.AverageWaitTime*time.Duration(tq.stats.TotalProcessed) + duration) / time.Duration(tq.stats.TotalProcessed+1)
			tq.stats.TotalProcessed++
			tq.mu2.Unlock()

			// Send result
			select {
			case tq.results <- result:
			case <-tq.ctx.Done():
				return
			}
		}
	}
}

// TranslateWithQueue translates using the queue
func TranslateWithQueue(ctx context.Context, queue *TranslationQueue, text, sourceLang, targetLang, context string) (*TranslationResponse, error) {
	req := &TranslationRequest{
		Text:       text,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Context:    context,
	}

	return queue.Submit(ctx, req)
}
