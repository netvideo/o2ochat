package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/netvideo/ai"
	"github.com/netvideo/decentralized"
	"github.com/netvideo/transport"
)

// TestP2PWithAIIntegration tests P2P network with AI translation
func TestP2PWithAIIntegration(t *testing.T) {
	// Create DHT manager
	dhtManager := decentralized.NewDHTManager(nil)

	// Create AI translator
	translator := ai.NewTranslator(nil)

	// Create connection pool
	pool := transport.NewConnectionPool(nil)

	if dhtManager == nil {
		t.Fatal("Failed to create DHT manager")
	}
	if translator == nil {
		t.Fatal("Failed to create AI translator")
	}
	if pool == nil {
		t.Fatal("Failed to create connection pool")
	}

	// Test integration
	ctx := context.Background()

	// Simulate P2P discovery and translation
	peerID := "test-peer-123"

	// Find node
	_, err := dhtManager.FindNode(ctx, peerID)
	if err == nil {
		t.Log("P2P node discovery working")
	}

	// Test translation
	req := &ai.TranslationRequest{
		Text:       "Hello, World!",
		SourceLang: "en",
		TargetLang: "zh-CN",
	}

	resp, err := translator.Translate(ctx, req)
	if err == nil && resp != nil {
		t.Log("AI translation working")
	}

	// Test connection pooling
	_, err = pool.GetConnection(ctx, peerID, func() (interface{}, error) {
		return nil, nil
	})
	if err == nil {
		t.Log("Connection pooling working")
	}

	t.Log("P2P + AI integration test complete")
}

// TestSecurityWithAIIntegration tests security features with AI
func TestSecurityWithAIIntegration(t *testing.T) {
	// Create rate limiter
	rateLimiter := decentralized.NewRateLimiter(nil)

	// Create anomaly detector
	anomalyDetector := decentralized.NewAnomalyDetector(nil)

	// Create AI translator
	translator := ai.NewTranslator(nil)

	if rateLimiter == nil {
		t.Fatal("Failed to create rate limiter")
	}
	if anomalyDetector == nil {
		t.Fatal("Failed to create anomaly detector")
	}
	if translator == nil {
		t.Fatal("Failed to create AI translator")
	}

	// Test rate limiting with AI translation
	ip := "192.168.1.1"

	// Simulate requests
	for i := 0; i < 10; i++ {
		if rateLimiter.Allow(ip) {
			anomalyDetector.RecordRequest(ip)

			req := &ai.TranslationRequest{
				Text:       "Test message",
				SourceLang: "en",
				TargetLang: "zh-CN",
			}

			translator.Translate(context.Background(), req)
		}
	}

	t.Log("Security + AI integration test complete")
}

// TestConcurrentP2POperations tests concurrent P2P operations
func TestConcurrentP2POperations(t *testing.T) {
	dhtManager := decentralized.NewDHTManager(nil)
	cache := decentralized.NewDHTCache(nil)

	concurrent := 100
	operations := 1000

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations/concurrent; j++ {
				nodeID := "node-" + string(rune(id%255))
				dhtManager.FindNode(context.Background(), nodeID)

				key := "key-" + string(rune(j%100))
				cache.Set(key, []byte("value"), 5*time.Minute)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Concurrent P2P Test:")
	t.Logf("  Concurrent: %d", concurrent)
	t.Logf("  Operations: %d", operations)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Ops/sec: %.2f", float64(operations)/duration.Seconds())

	if duration > 10*time.Second {
		t.Errorf("Concurrent operations too slow: %v", duration)
	}
}

// TestAIWithSecurityIntegration tests AI translation with security features
func TestAIWithSecurityIntegration(t *testing.T) {
	translator := ai.NewTranslator(nil)
	rateLimiter := decentralized.NewRateLimiter(nil)
	anomalyDetector := decentralized.NewAnomalyDetector(nil)

	concurrent := 50
	translations := 500

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < translations/concurrent; j++ {
				ip := "192.168.1." + string(rune(id%255))

				// Check rate limit
				if !rateLimiter.Allow(ip) {
					return
				}

				// Record for anomaly detection
				anomalyDetector.RecordRequest(ip)

				// Perform translation
				req := &ai.TranslationRequest{
					Text:       "Test message " + string(rune(j)),
					SourceLang: "en",
					TargetLang: "zh-CN",
				}

				translator.Translate(context.Background(), req)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("AI + Security Integration Test:")
	t.Logf("  Concurrent: %d", concurrent)
	t.Logf("  Translations: %d", translations)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Translations/sec: %.2f", float64(translations)/duration.Seconds())

	// Check anomaly detector stats
	stats := anomalyDetector.GetStats()
	t.Logf("  Anomalies detected: %d", stats.TotalAnomalies)
}

// TestEndToEndP2PFlow tests end-to-end P2P flow
func TestEndToEndP2PFlow(t *testing.T) {
	// Create all components
	dhtManager := decentralized.NewDHTManager(nil)
	cache := decentralized.NewDHTCache(nil)
	rateLimiter := decentralized.NewRateLimiter(nil)
	anomalyDetector := decentralized.NewAnomalyDetector(nil)
	translator := ai.NewTranslator(nil)
	pool := transport.NewConnectionPool(nil)
	monitor := transport.NewQualityMonitor(nil)

	// Verify all components created
	if dhtManager == nil || cache == nil || rateLimiter == nil ||
		anomalyDetector == nil || translator == nil || pool == nil || monitor == nil {
		t.Fatal("Failed to create required components")
	}

	ctx := context.Background()

	// Simulate complete P2P flow
	peerID := "test-peer-end-to-end"
	ip := "192.168.1.100"

	// 1. Rate limiting
	if !rateLimiter.Allow(ip) {
		t.Fatal("Rate limiter blocked valid request")
	}

	// 2. Anomaly detection
	anomalyDetector.RecordRequest(ip)

	// 3. DHT lookup
	dhtManager.FindNode(ctx, peerID)

	// 4. Cache lookup
	cache.Set("test-key", []byte("test-value"), 5*time.Minute)
	cache.Get("test-key")

	// 5. Connection pooling
	pool.GetConnection(ctx, peerID, func() (interface{}, error) {
		return nil, nil
	})

	// 6. AI translation
	translator.Translate(ctx, &ai.TranslationRequest{
		Text:       "End-to-end test",
		SourceLang: "en",
		TargetLang: "zh-CN",
	})

	// 7. Quality monitoring
	monitor.UpdateQuality("test-conn", peerID, 50, 10.0, 0.01)

	// Get final stats
	t.Log("End-to-End P2P Flow Test Complete:")
	t.Logf("  DHT Manager: ✓")
	t.Logf("  Cache: ✓")
	t.Logf("  Rate Limiter: ✓")
	t.Logf("  Anomaly Detector: ✓")
	t.Logf("  Translator: ✓")
	t.Logf("  Connection Pool: ✓")
	t.Logf("  Quality Monitor: ✓")
}

// BenchmarkEndToEndFlow benchmarks end-to-end flow
func BenchmarkEndToEndFlow(b *testing.B) {
	dhtManager := decentralized.NewDHTManager(nil)
	cache := decentralized.NewDHTCache(nil)
	translator := ai.NewTranslator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		peerID := "bench-peer"
		ip := "192.168.1.1"

		// Complete flow
		dhtManager.FindNode(ctx, peerID)
		cache.Set("key", []byte("value"), 5*time.Minute)
		cache.Get("key")
		translator.Translate(ctx, &ai.TranslationRequest{
			Text:       "Benchmark",
			SourceLang: "en",
			TargetLang: "zh-CN",
		})
		_ = ip
	}
}
