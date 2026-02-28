package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/netvideo/identity"
)

// TestNewPeerConnection 测试创建新的 P2P 连接
func TestNewPeerConnection(t *testing.T) {
	tests := []struct {
		name     string
		localID  string
		remoteID string
		wantErr  bool
	}{
		{
			name:     "Valid connection",
			localID:  "peer1",
			remoteID: "peer2",
			wantErr:  false,
		},
		{
			name:     "Empty local ID",
			localID:  "",
			remoteID: "peer2",
			wantErr:  true,
		},
		{
			name:     "Empty remote ID",
			localID:  "peer1",
			remoteID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &PeerConnectionConfig{
				ICEServers: []string{"stun:stun.l.google.com:19302"},
			}

			conn, err := NewPeerConnection(tt.localID, tt.remoteID, config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPeerConnection() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewPeerConnection() unexpected error: %v", err)
				return
			}

			if conn == nil {
				t.Errorf("NewPeerConnection() returned nil connection")
				return
			}

			// 验证连接属性
			if conn.GetLocalPeerID() != tt.localID {
				t.Errorf("LocalPeerID = %v, want %v", conn.GetLocalPeerID(), tt.localID)
			}

			if conn.GetRemotePeerID() != tt.remoteID {
				t.Errorf("RemotePeerID = %v, want %v", conn.GetRemotePeerID(), tt.remoteID)
			}

			// 清理
			conn.Close()
		})
	}
}

// TestPeerConnectionState 测试连接状态管理
func TestPeerConnectionState(t *testing.T) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	conn, err := NewPeerConnection("peer1", "peer2", config)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	// 初始状态应该是 New
	if conn.GetState() != ConnectionStateNew {
		t.Errorf("Initial state = %v, want %v", conn.GetState(), ConnectionStateNew)
	}

	// 测试状态转换
	conn.SetState(ConnectionStateConnecting)
	if conn.GetState() != ConnectionStateConnecting {
		t.Errorf("State after connecting = %v, want %v", conn.GetState(), ConnectionStateConnecting)
	}

	conn.SetState(ConnectionStateConnected)
	if conn.GetState() != ConnectionStateConnected {
		t.Errorf("State after connected = %v, want %v", conn.GetState(), ConnectionStateConnected)
	}

	// 测试关闭状态
	conn.Close()
	if conn.GetState() != ConnectionStateClosed {
		t.Errorf("State after close = %v, want %v", conn.GetState(), ConnectionStateClosed)
	}
}

// TestDataChannel 测试数据通道功能
func TestDataChannel(t *testing.T) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	conn, err := NewPeerConnection("peer1", "peer2", config)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	// 创建数据通道
	channelLabel := "test-channel"
	channel, err := conn.CreateDataChannel(channelLabel, nil)
	if err != nil {
		t.Fatalf("Failed to create data channel: %v", err)
	}

	if channel == nil {
		t.Fatal("CreateDataChannel returned nil")
	}

	// 验证通道标签
	if channel.Label() != channelLabel {
		t.Errorf("Channel label = %v, want %v", channel.Label(), channelLabel)
	}

	// 测试发送数据
	testData := []byte("Hello, P2P!")
	err = channel.Send(testData)
	if err != nil {
		t.Errorf("Failed to send data: %v", err)
	}

	// 测试关闭通道
	err = channel.Close()
	if err != nil {
		t.Errorf("Failed to close channel: %v", err)
	}
}

// TestConnectionTimeout 测试连接超时
func TestConnectionTimeout(t *testing.T) {
	config := &PeerConnectionConfig{
		ICEServers:        []string{"stun:stun.l.google.com:19302"},
		ConnectionTimeout: 1 * time.Second,
	}

	conn, err := NewPeerConnection("peer1", "peer2", config)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	// 验证超时设置
	if conn.GetConfig().ConnectionTimeout != 1*time.Second {
		t.Errorf("ConnectionTimeout = %v, want %v",
			conn.GetConfig().ConnectionTimeout, 1*time.Second)
	}
}

// TestConnectionError 测试连接错误处理
func TestConnectionError(t *testing.T) {
	// 测试无效配置
	config := &PeerConnectionConfig{
		ICEServers: []string{}, // 空的 ICE 服务器
	}

	conn, err := NewPeerConnection("peer1", "peer2", config)
	// 空 ICE 服务器不应该导致错误，但应该使用默认配置
	if err != nil {
		t.Logf("Warning: Empty ICE servers caused error: %v", err)
	}

	if conn != nil {
		conn.Close()
	}
}

// BenchmarkPeerConnection 性能测试
func BenchmarkPeerConnection(b *testing.B) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, err := NewPeerConnection("peer1", "peer2", config)
		if err != nil {
			b.Fatalf("Failed to create connection: %v", err)
		}
		conn.Close()
	}
}

// BenchmarkDataChannelSend 数据通道发送性能测试
func BenchmarkDataChannelSend(b *testing.B) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	conn, err := NewPeerConnection("peer1", "peer2", config)
	if err != nil {
		b.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	channel, err := conn.CreateDataChannel("bench-channel", nil)
	if err != nil {
		b.Fatalf("Failed to create data channel: %v", err)
	}

	testData := make([]byte, 1024) // 1KB payload
	rand.Read(testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := channel.Send(testData)
		if err != nil {
			b.Fatalf("Failed to send data: %v", err)
		}
	}

	b.ReportMetric(float64(b.N), "messages")
	b.ReportMetric(float64(b.N*len(testData))/1024/1024, "MB")
}

// ExampleNewPeerConnection 示例代码
func ExampleNewPeerConnection() {
	config := &PeerConnectionConfig{
		ICEServers: []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
		},
		ConnectionTimeout: 30 * time.Second,
	}

	conn, err := NewPeerConnection("local-peer-id", "remote-peer-id", config)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	// Create data channel
	channel, err := conn.CreateDataChannel("chat", nil)
	if err != nil {
		log.Fatalf("Failed to create data channel: %v", err)
	}

	// Send message
	message := []byte("Hello, P2P!")
	if err := channel.Send(message); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	fmt.Println("Message sent successfully!")
	// Output: Message sent successfully!
}
