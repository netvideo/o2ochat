package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/netvideo/decentralized"
	"github.com/netvideo/transport"
)

// TestDHTRoutingPerformance tests DHT routing performance
func TestDHTRoutingPerformance(t *testing.T) {
	manager := decentralized.NewDHTManager(nil)
	
	iterations := 1000
	startTime := time.Now()
	
	for i := 0; i < iterations; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		manager.FindNode(context.Background(), nodeID)
	}
	
	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)
	
	t.Logf("DHT Routing Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average time: %v", avgTime)
	
	if avgTime > 50*time.Millisecond {
		t.Errorf("DHT routing too slow: %v (target: <50ms)", avgTime)
	}
}

// TestDHTCachePerformance tests DHT cache performance
func TestDHTCachePerformance(t *testing.T) {
	cache := decentralized.NewDHTCache(nil)
	
	iterations := 10000
	startTime := time.Now()
	
	for i := 0; i < iterations; i++ {
		key := fmt.Sprintf("key-%d", i%1000)
		cache.Set(key, []byte("value"), 5*time.Minute)
		cache.Get(key)
	}
	
	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)
	
	t.Logf("DHT Cache Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average time: %v", avgTime)
	
	if avgTime > 1*time.Millisecond {
		t.Errorf("Cache operations too slow: %v (target: <1ms)", avgTime)
	}
}

// TestRateLimiterPerformance tests rate limiter performance
func TestRateLimiterPerformance(t *testing.T) {
	limiter := decentralized.NewRateLimiter(nil)
	ip := "192.168.1.1"
	
	iterations := 100000
	startTime := time.Now()
	
	for i := 0; i < iterations; i++ {
		limiter.Allow(ip)
	}
	
	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)
	
	t.Logf("Rate Limiter Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average time: %v", avgTime)
	
	if avgTime > 100*time.Microsecond {
		t.Errorf("Rate limiting too slow: %v (target: <100μs)", avgTime)
	}
}

// TestConnectionPoolPerformance tests connection pool performance
func TestConnectionPoolPerformance(t *testing.T) {
	pool := transport.NewConnectionPool(nil)
	peerID := "test-peer"
	
	iterations := 1000
	startTime := time.Now()
	
	for i := 0; i < iterations; i++ {
		pool.GetConnection(context.Background(), peerID, func() (interface{}, error) {
			return nil, nil
		})
	}
	
	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)
	
	t.Logf("Connection Pool Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average time: %v", avgTime)
	
	if avgTime > 10*time.Millisecond {
		t.Errorf("Connection pool too slow: %v (target: <10ms)", avgTime)
	}
}

// TestConcurrentDHTOperations tests concurrent DHT operations
func TestConcurrentDHTOperations(t *testing.T) {
	manager := decentralized.NewDHTManager(nil)
	
	concurrent := 100
	operations := 1000
	
	startTime := time.Now()
	
	done := make(chan bool, concurrent)
	
	for i := 0; i < concurrent; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < operations/concurrent; j++ {
				nodeID := fmt.Sprintf("node-%d-%d", id, j)
				manager.FindNode(context.Background(), nodeID)
			}
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < concurrent; i++ {
		<-done
	}
	
	duration := time.Since(startTime)
	opsPerSec := float64(operations) / duration.Seconds()
	
	t.Logf("Concurrent DHT Operations:")
	t.Logf("  Concurrent: %d", concurrent)
	t.Logf("  Operations: %d", operations)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Ops/sec: %.2f", opsPerSec)
	
	if opsPerSec < 1000 {
		t.Errorf("DHT throughput too low: %.2f ops/sec (target: >1000)", opsPerSec)
	}
}

// TestAnomalyDetectorPerformance tests anomaly detector performance
func TestAnomalyDetectorPerformance(t *testing.T) {
	detector := decentralized.NewAnomalyDetector(nil)
	ip := "192.168.1.1"
	
	iterations := 10000
	startTime := time.Now()
	
	for i := 0; i < iterations; i++ {
		detector.RecordRequest(ip)
	}
	
	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)
	
	t.Logf("Anomaly Detector Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average time: %v", avgTime)
	
	if avgTime > 50*time.Microsecond {
		t.Errorf("Anomaly detection too slow: %v (target: <50μs)", avgTime)
	}
}

// TestQualityMonitorPerformance tests quality monitor performance
func TestQualityMonitorPerformance(t *testing.T) {
	monitor := transport.NewQualityMonitor(nil)
	connID := "test-connection"
	
	iterations := 10000
	startTime := time.Now()
	
	for i := 0; i < iterations; i++ {
		monitor.UpdateQuality(connID, "test-peer", 50, 10.0, 0.01)
	}
	
	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)
	
	t.Logf("Quality Monitor Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average time: %v", avgTime)
	
	if avgTime > 10*time.Microsecond {
		t.Errorf("Quality monitoring too slow: %v (target: <10μs)", avgTime)
	}
}

// BenchmarkEndToEndPerformance benchmarks end-to-end performance
func BenchmarkEndToEndPerformance(b *testing.B) {
	dhtManager := decentralized.NewDHTManager(nil)
	cache := decentralized.NewDHTCache(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		peerID := "bench-peer"
		ip := "192.168.1.1"
		
		// Complete flow
		dhtManager.FindNode(ctx, peerID)
		cache.Set("key", []byte("value"), 5*time.Minute)
		cache.Get("key")
		_ = ip
	}
}
