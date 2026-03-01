package p2p

import (
	"sync"
	"testing"
	"time"
)

// TestConnectionWithEmptyConfig 测试空配置
func TestConnectionWithEmptyConfig(t *testing.T) {
	conn, err := NewPeerConnection("local", "remote", nil)
	if err != nil {
		t.Fatalf("Failed to create connection with nil config: %v", err)
	}
	defer conn.Close()

	// 验证使用默认配置
	if conn.config == nil {
		t.Error("Config should not be nil after initialization")
	}
}

// TestConnectionWithSameIDs 测试相同的本地和远程ID
func TestConnectionWithSameIDs(t *testing.T) {
	_, err := NewPeerConnection("same-id", "same-id", nil)
	if err == nil {
		t.Error("Should fail when local and remote IDs are the same")
	}
}

// TestConnectionWithLongIDs 测试长ID
func TestConnectionWithLongIDs(t *testing.T) {
	longID := make([]byte, 1024) // 1KB ID
	for i := range longID {
		longID[i] = byte('a' + i%26)
	}

	conn, err := NewPeerConnection(string(longID), "remote", nil)
	if err != nil {
		t.Fatalf("Failed with long ID: %v", err)
	}
	defer conn.Close()

	if conn.localID != string(longID) {
		t.Error("Long ID should be preserved correctly")
	}
}

// TestConnectionWithSpecialChars 测试特殊字符ID
func TestConnectionWithSpecialChars(t *testing.T) {
	specialID := "peer-123_test.peer@domain:8080/path?q=1"

	conn, err := NewPeerConnection(specialID, "remote", nil)
	if err != nil {
		t.Fatalf("Failed with special char ID: %v", err)
	}
	defer conn.Close()

	if conn.localID != specialID {
		t.Error("Special char ID should be preserved correctly")
	}
}

// TestConcurrentConnections 测试并发连接
func TestConcurrentConnections(t *testing.T) {
	const numConnections = 100
	var wg sync.WaitGroup
	errors := make(chan error, numConnections)

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			localID := "local-" + string(rune('a'+index%26))
			remoteID := "remote-" + string(rune('a'+index%26))

			conn, err := NewPeerConnection(localID, remoteID, nil)
			if err != nil {
				errors <- err
				return
			}
			defer conn.Close()

			// 模拟一些操作
			time.Sleep(time.Millisecond * 10)
		}(i)
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
			t.Logf("Connection error: %v", err)
		}
	}

	if errorCount > 0 {
		t.Errorf("Had %d connection errors out of %d attempts",
			errorCount, numConnections)
	}

	t.Logf("Successfully created %d concurrent connections", numConnections)
}

// TestConnectionRaceConditions 测试竞态条件
func TestConnectionRaceConditions(t *testing.T) {
	conn, err := NewPeerConnection("local", "remote", nil)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	// 并发读写状态
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 读取状态
			_ = conn.GetState()
			// 获取统计
			_ = conn.GetStats()
		}()
	}

	wg.Wait()
}

// BenchmarkConnectionCreation 基准测试：创建连接
func BenchmarkConnectionCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		conn, err := NewPeerConnection("local", "remote", nil)
		if err != nil {
			b.Fatal(err)
		}
		conn.Close()
	}
}

// BenchmarkConcurrentConnections 基准测试：并发连接
func BenchmarkConcurrentConnections(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := NewPeerConnection("local", "remote", nil)
			if err != nil {
				b.Fatal(err)
			}
			conn.Close()
		}
	})
}

// BenchmarkConnectionStateAccess 基准测试：状态访问
func BenchmarkConnectionStateAccess(b *testing.B) {
	conn, err := NewPeerConnection("local", "remote", nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conn.GetState()
	}
}

// BenchmarkStatsAccess 基准测试：统计访问
func BenchmarkStatsAccess(b *testing.B) {
	conn, err := NewPeerConnection("local", "remote", nil)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conn.GetStats()
	}
}
