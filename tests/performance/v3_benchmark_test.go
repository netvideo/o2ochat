package performance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/netvideo/decentralized"
	"github.com/netvideo/transport"
)

// BenchmarkDHTRouting benchmarks DHT routing performance
func BenchmarkDHTRouting(b *testing.B) {
	manager := decentralized.NewDHTManager(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		manager.FindNode(context.Background(), nodeID)
	}
}

// BenchmarkDHTCache benchmarks DHT cache performance
func BenchmarkDHTCache(b *testing.B) {
	cache := decentralized.NewDHTCache(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%1000)
		cache.Set(key, []byte("value"), 5*time.Minute)
		cache.Get(key)
	}
}

// BenchmarkRateLimiter benchmarks rate limiter performance
func BenchmarkRateLimiter(b *testing.B) {
	limiter := decentralized.NewRateLimiter(nil)
	ip := "192.168.1.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ip)
	}
}

// BenchmarkAnomalyDetector benchmarks anomaly detector performance
func BenchmarkAnomalyDetector(b *testing.B) {
	detector := decentralized.NewAnomalyDetector(nil)
	ip := "192.168.1.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.RecordRequest(ip)
	}
}

// BenchmarkConnectionPool benchmarks connection pool performance
func BenchmarkConnectionPool(b *testing.B) {
	pool := transport.NewConnectionPool(nil)
	peerID := "test-peer"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.GetConnection(context.Background(), peerID, func() (interface{}, error) {
			return nil, nil
		})
	}
}

// BenchmarkQualityMonitor benchmarks quality monitor performance
func BenchmarkQualityMonitor(b *testing.B) {
	monitor := transport.NewQualityMonitor(nil)
	connID := "test-connection"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.UpdateQuality(connID, "test-peer", 50, 10.0, 0.01)
	}
}

// BenchmarkConcurrentDHT benchmarks concurrent DHT operations
func BenchmarkConcurrentDHT(b *testing.B) {
	manager := decentralized.NewDHTManager(nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			nodeID := fmt.Sprintf("node-%d", i)
			manager.FindNode(context.Background(), nodeID)
			i++
		}
	})
}

// BenchmarkConcurrentCache benchmarks concurrent cache operations
func BenchmarkConcurrentCache(b *testing.B) {
	cache := decentralized.NewDHTCache(nil)

	b.ResetTimer()
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id%1000)
			cache.Set(key, []byte("value"), 5*time.Minute)
			cache.Get(key)
		}(i)
	}
	wg.Wait()
}

// BenchmarkConcurrentRateLimiter benchmarks concurrent rate limiter operations
func BenchmarkConcurrentRateLimiter(b *testing.B) {
	limiter := decentralized.NewRateLimiter(nil)

	b.ResetTimer()
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ip := fmt.Sprintf("192.168.1.%d", id%255)
			limiter.Allow(ip)
		}(i)
	}
	wg.Wait()
}

// StressTestDHT stress tests DHT operations
func StressTestDHT(t *testing.T) {
	manager := decentralized.NewDHTManager(nil)

	concurrent := 1000
	operations := 10000

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations/concurrent; j++ {
				nodeID := fmt.Sprintf("node-%d-%d", id, j)
				manager.FindNode(context.Background(), nodeID)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("DHT Stress Test:")
	t.Logf("  Concurrent: %d", concurrent)
	t.Logf("  Operations: %d", operations)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Ops/sec: %.2f", float64(operations)/duration.Seconds())
}

// StressTestCache stress tests cache operations
func StressTestCache(t *testing.T) {
	cache := decentralized.NewDHTCache(nil)

	concurrent := 1000
	operations := 10000

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations/concurrent; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				cache.Set(key, []byte("value"), 5*time.Minute)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Cache Stress Test:")
	t.Logf("  Concurrent: %d", concurrent)
	t.Logf("  Operations: %d", operations)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Ops/sec: %.2f", float64(operations)/duration.Seconds())
}

// StressTestRateLimiter stress tests rate limiter operations
func StressTestRateLimiter(t *testing.T) {
	limiter := decentralized.NewRateLimiter(nil)

	concurrent := 1000
	operations := 10000

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations/concurrent; j++ {
				ip := fmt.Sprintf("192.168.1.%d", id%255)
				limiter.Allow(ip)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Rate Limiter Stress Test:")
	t.Logf("  Concurrent: %d", concurrent)
	t.Logf("  Operations: %d", operations)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Ops/sec: %.2f", float64(operations)/duration.Seconds())
}
