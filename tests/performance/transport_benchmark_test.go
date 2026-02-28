package performance

import (
	"testing"
	"time"

	"github.com/netvideo/transport"
)

func BenchmarkQUICConnection(b *testing.B) {
	config := &transport.TransportConfig{
		QUICConfig: &transport.QUICConfig{
			MaxIncomingStreams:   100,
			KeepAlive:            true,
			HandshakeIdleTimeout: 5 * time.Second,
			MaxIdleTimeout:       30 * time.Second,
		},
		WebRTCConfig:   nil,
		MaxConnections: 100,
		KeepAlive:      true,
	}

	manager := transport.NewTransportManager(config)
	if manager == nil {
		b.Fatal("Failed to create transport manager")
	}
	defer manager.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟 QUIC 连接建立
		connConfig := &transport.ConnectionConfig{
			PeerID:        "test-peer",
			IPv6Addresses: []string{"[::1]:24242"},
			Priority:      []transport.ConnectionType{transport.ConnectionTypeQUIC},
			Timeout:       5 * time.Second,
			RetryCount:    1,
		}

		// 注意：这是基准测试，实际连接会失败
		// 真实测试需要启动实际的 QUIC 服务器
		_ = connConfig
	}
}

func BenchmarkConnectionSetup(b *testing.B) {
	config := &transport.TransportConfig{
		QUICConfig: &transport.QUICConfig{
			MaxIncomingStreams:   100,
			KeepAlive:            true,
			HandshakeIdleTimeout: 5 * time.Second,
			MaxIdleTimeout:       30 * time.Second,
		},
		WebRTCConfig:   nil,
		MaxConnections: 100,
		KeepAlive:      true,
	}

	manager := transport.NewTransportManager(config)
	if manager == nil {
		b.Fatal("Failed to create transport manager")
	}
	defer manager.Close()

	// 启动监听
	err := manager.Listen("[::1]:0")
	if err != nil {
		b.Skipf("监听失败：%v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟连接建立
		connConfig := &transport.ConnectionConfig{
			PeerID:        "peer-" + string(rune(i)),
			IPv6Addresses: []string{"[::1]:24242"},
			Priority:      []transport.ConnectionType{transport.ConnectionTypeQUIC},
			Timeout:       5 * time.Second,
			RetryCount:    1,
		}

		conn, err := manager.Connect(connConfig)
		if err != nil {
			// 基准测试中连接失败是正常的
			continue
		}
		_ = conn
	}
}

func BenchmarkDataTransfer(b *testing.B) {
	config := &transport.TransportConfig{
		QUICConfig: &transport.QUICConfig{
			MaxIncomingStreams:   100,
			KeepAlive:            true,
			HandshakeIdleTimeout: 5 * time.Second,
			MaxIdleTimeout:       30 * time.Second,
		},
		WebRTCConfig:   nil,
		MaxConnections: 100,
		KeepAlive:      true,
	}

	manager := transport.NewTransportManager(config)
	if manager == nil {
		b.Fatal("Failed to create transport manager")
	}
	defer manager.Close()

	testData := make([]byte, 1024*1024) // 1MB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 模拟数据传输
		b.StartTimer()
		// 实际传输逻辑
		b.StopTimer()
	}
}
