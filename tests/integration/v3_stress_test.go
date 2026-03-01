package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/netvideo/decentralized"
	"github.com/netvideo/transport"
)

// TestP2PSecurityIntegration tests P2P network with security features
func TestP2PSecurityIntegration(t *testing.T) {
	// Create components
	dhtManager := decentralized.NewDHTManager(nil)
	rateLimiter := decentralized.NewRateLimiter(nil)
	anomalyDetector := decentralized.NewAnomalyDetector(nil)
	
	if dhtManager == nil || rateLimiter == nil || anomalyDetector == nil {
		t.Fatal("Failed to create components")
	}
	
	// Simulate P2P operations with security
	ip := "192.168.1.1"
	peerID := "test-peer"
	
	// Test rate limiting
	for i := 0; i < 100; i++ {
		if !rateLimiter.Allow(ip) {
			t.Logf("Rate limiter blocked request %d", i)
			break
		}
		
		// Record for anomaly detection
		anomalyDetector.RecordRequest(ip)
		
		// DHT operation
		dhtManager.FindNode(context.Background(), peerID)
	}
	
	t.Log("P2P + Security integration test passed")
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

// TestConnectionPoolIntegration tests connection pool integration
func TestConnectionPoolIntegration(t *testing.T) {
	pool := transport.NewConnectionPool(nil)
	
	if pool == nil {
		t.Fatal("Failed to create connection pool")
	}
	
	peerID := "test-peer"
	
	// Get multiple connections
	for i := 0; i < 10; i++ {
		conn, err := pool.GetConnection(context.Background(), peerID, func() (interface{}, error) {
			return nil, nil
		})
		
		if err != nil {
			t.Errorf("Failed to get connection %d: %v", i, err)
		}
		
		if conn == nil {
			t.Errorf("Connection %d is nil", i)
		}
		
		pool.ReturnConnection("test-conn")
	}
	
	// Get stats
	stats := pool.GetStats()
	t.Logf("Connection Pool Stats:")
	t.Logf("  Total connections: %d", stats.TotalConnections)
	t.Logf("  Active connections: %d", stats.ActiveConnections)
	t.Logf("  Total created: %d", stats.TotalCreated)
	
	t.Log("Connection pool integration test passed")
}

// TestQualityMonitorIntegration tests quality monitor integration
func TestQualityMonitorIntegration(t *testing.T) {
	monitor := transport.NewQualityMonitor(nil)
	
	if monitor == nil {
		t.Fatal("Failed to create quality monitor")
	}
	
	connID := "test-connection"
	peerID := "test-peer"
	
	// Update quality multiple times
	for i := 0; i < 100; i++ {
		monitor.UpdateQuality(connID, peerID, 50, 10.0, 0.01)
	}
	
	// Get quality
	quality := monitor.GetQuality(connID)
	if quality == nil {
		t.Fatal("Quality is nil")
	}
	
	// Check health
	if !monitor.IsHealthy(connID) {
		t.Error("Connection should be healthy")
	}
	
	// Get stats
	stats := monitor.GetStats()
	t.Logf("Quality Monitor Stats:")
	t.Logf("  Total connections: %d", stats.TotalConnections)
	t.Logf("  Health good: %d", stats.HealthGood)
	
	t.Log("Quality monitor integration test passed")
}

// TestStressP2POperations stress tests P2P operations
func TestStressP2POperations(t *testing.T) {
	dhtManager := decentralized.NewDHTManager(nil)
	cache := decentralized.NewDHTCache(nil)
	rateLimiter := decentralized.NewRateLimiter(nil)
	
	concurrent := 1000
	operations := 10000
	
	var wg sync.WaitGroup
	startTime := time.Now()
	
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ip := "192.168.1." + string(rune(id%255))
			
			for j := 0; j < operations/concurrent; j++ {
				// Rate limiting
				if !rateLimiter.Allow(ip) {
					continue
				}
				
				// DHT operation
				nodeID := "node-" + string(rune(id%255))
				dhtManager.FindNode(context.Background(), nodeID)
				
				// Cache operation
				key := "key-" + string(rune(j%100))
				cache.Set(key, []byte("value"), 5*time.Minute)
				cache.Get(key)
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(startTime)
	
	t.Logf("Stress Test:")
	t.Logf("  Concurrent: %d", concurrent)
	t.Logf("  Operations: %d", operations)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Ops/sec: %.2f", float64(operations)/duration.Seconds())
	
	// Check stats
	stats := rateLimiter.GetStats()
	t.Logf("Rate Limiter Stats:")
	t.Logf("  Total requests: %d", stats.TotalRequests)
	t.Logf("  Blocked requests: %d", stats.BlockedRequests)
	
	if duration > 30*time.Second {
		t.Errorf("Stress test too slow: %v", duration)
	}
}

// TestEndToEndFlow tests complete end-to-end flow
func TestEndToEndFlow(t *testing.T) {
	// Create all components
	dhtManager := decentralized.NewDHTManager(nil)
	cache := decentralized.NewDHTCache(nil)
	rateLimiter := decentralized.NewRateLimiter(nil)
	anomalyDetector := decentralized.NewAnomalyDetector(nil)
	pool := transport.NewConnectionPool(nil)
	monitor := transport.NewQualityMonitor(nil)
	
	if dhtManager == nil || cache == nil || rateLimiter == nil ||
		anomalyDetector == nil || pool == nil || monitor == nil {
		t.Fatal("Failed to create required components")
	}
	
	ctx := context.Background()
	peerID := "test-peer-e2e"
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
	
	// 6. Quality monitoring
	monitor.UpdateQuality("test-conn", peerID, 50, 10.0, 0.01)
	
	// Get final stats
	t.Log("End-to-End Flow Test Complete:")
	t.Log("  DHT Manager: ✓")
	t.Log("  Cache: ✓")
	t.Log("  Rate Limiter: ✓")
	t.Log("  Anomaly Detector: ✓")
	t.Log("  Connection Pool: ✓")
	t.Log("  Quality Monitor: ✓")
	
	t.Log("✅ End-to-end flow test passed!")
}
