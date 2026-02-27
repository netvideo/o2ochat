package performance

import (
	"testing"
	"time"

	"github.com/netvideo/transport"
)

func BenchmarkQUICConnection(b *testing.B) {
	manager := transport.NewTransportManager()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟 QUIC 连接建立
		config := &transport.ConnectionConfig{
			PeerID:        "test-peer",
			IPv6Addresses: []string{"[::1]:24242"},
			Priority:      []transport.ConnectionType{transport.ConnectionTypeQUIC},
			Timeout:       5 * time.Second,
			RetryCount:    1,
		}

		// 注意：这是基准测试，实际连接会失败
		// 真实测试需要启动实际的 QUIC 服务器
		_ = config
	}
}

func BenchmarkConnectionSetup(b *testing.B) {
	manager := transport.NewTransportManager()
	defer manager.Close()

	// 启动监听
	err := manager.Listen("[::1]:0")
	if err != nil {
		b.Skipf("监听失败：%v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		config := &transport.ConnectionConfig{
			PeerID:        "test-peer",
			IPv6Addresses: []string{"[::1]:0"},
			Priority:      []transport.ConnectionType{transport.ConnectionTypeQUIC},
			Timeout:       1 * time.Second,
			RetryCount:    0,
		}

		// 尝试连接（会失败，但测试连接建立逻辑）
		_, _ = manager.Connect(config)
	}
}

func BenchmarkStreamCreation(b *testing.B) {
	// 模拟流创建性能测试
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		config := &transport.StreamConfig{
			Reliable:   true,
			Ordered:    true,
			BufferSize: 1024,
		}

		// 注意：这是基准测试框架
		// 实际测试需要真实的连接
		_ = config
	}
}

func BenchmarkDataTransfer(b *testing.B) {
	// 模拟数据传输性能测试
	dataSize := 1024 * 1024 // 1MB
	data := make([]byte, dataSize)

	b.ResetTimer()
	b.SetBytes(int64(dataSize))

	for i := 0; i < b.N; i++ {
		// 模拟数据传输
		_ = data
	}
}
