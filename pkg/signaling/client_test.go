// Package signaling provides WebSocket signaling for P2P connection establishment
package signaling

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestMessageTypeConstants 测试消息类型常量
func TestMessageTypeConstants(t *testing.T) {
	tests := []struct {
		msgType  MessageType
		expected string
	}{
		{TypeOffer, "offer"},
		{TypeAnswer, "answer"},
		{TypeICE, "ice"},
		{TypeJoin, "join"},
		{TypeLeave, "leave"},
		{TypePeerJoined, "peer_joined"},
		{TypePeerLeft, "peer_left"},
		{TypeError, "error"},
		{TypePing, "ping"},
		{TypePong, "pong"},
	}

	for _, tt := range tests {
		t.Run(string(tt.msgType), func(t *testing.T) {
			if string(tt.msgType) != tt.expected {
				t.Errorf("MessageType = %v, want %v", tt.msgType, tt.expected)
			}
		})
	}
}

// TestDefaultClientConfig 测试默认客户端配置
func TestDefaultClientConfig(t *testing.T) {
	config := DefaultClientConfig()

	if config.ReconnectInterval != 5*time.Second {
		t.Errorf("DefaultClientConfig.ReconnectInterval = %v, want 5s", config.ReconnectInterval)
	}

	if config.PingInterval != 30*time.Second {
		t.Errorf("DefaultClientConfig.PingInterval = %v, want 30s", config.PingInterval)
	}
}

// TestNewClient 测试创建客户端
func TestNewClient(t *testing.T) {
	config := ClientConfig{
		ServerURL: "ws://localhost:8080",
		PeerID:    "peer123",
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

// TestClientIsConnected 测试连接状态检查
func TestClientIsConnected(t *testing.T) {
	config := ClientConfig{
		ServerURL: "ws://localhost:8080",
		PeerID:    "peer123",
	}

	client := NewClient(config)

	// Should not be connected initially
	if client.IsConnected() {
		t.Error("IsConnected() should return false when not connected")
	}
}

// TestClientOnMessage 测试消息处理器设置
func TestClientOnMessage(t *testing.T) {
	config := ClientConfig{
		ServerURL: "ws://localhost:8080",
		PeerID:    "peer123",
	}

	client := NewClient(config)

	handlerCalled := false
	handler := func(msg Message) {
		handlerCalled = true
	}

	// Should not panic
	client.OnMessage(handler)

	// Verify handler is set (can't actually call it without a connection)
	_ = handlerCalled
}

// TestClientOnConnect 测试连接处理器设置
func TestClientOnConnect(t *testing.T) {
	config := ClientConfig{
		ServerURL: "ws://localhost:8080",
		PeerID:    "peer123",
	}

	client := NewClient(config)

	handlerCalled := false
	handler := func() {
		handlerCalled = true
	}

	// Should not panic
	client.OnConnect(handler)

	_ = handlerCalled
}

// TestClientOnDisconnect 测试断开连接处理器设置
func TestClientOnDisconnect(t *testing.T) {
	config := ClientConfig{
		ServerURL: "ws://localhost:8080",
		PeerID:    "peer123",
	}

	client := NewClient(config)

	handlerCalled := false
	handler := func() {
		handlerCalled = true
	}

	// Should not panic
	client.OnDisconnect(handler)

	_ = handlerCalled
}

// TestClientOnError 测试错误处理器设置
func TestClientOnError(t *testing.T) {
	config := ClientConfig{
		ServerURL: "ws://localhost:8080",
		PeerID:    "peer123",
	}

	client := NewClient(config)

	handlerCalled := false
	handler := func(err error) {
		handlerCalled = true
	}

	// Should not panic
	client.OnError(handler)

	_ = handlerCalled
}

// TestClientGetRoomInfo 测试获取房间信息
func TestClientGetRoomInfo(t *testing.T) {
	config := ClientConfig{
		ServerURL: "ws://localhost:8080",
		PeerID:    "peer123",
	}

	client := NewClient(config)

	roomInfo, err := client.GetRoomInfo("room123")
	// This might fail without a real connection, but should not panic
	_ = err
	_ = roomInfo
}

// TestMessageJSON 测试消息JSON序列化
func TestMessageJSON(t *testing.T) {
	msg := Message{
		Type:         TypeOffer,
		RoomID:       "room123",
		PeerID:       "peer123",
		TargetPeerID: "target456",
		Payload:      json.RawMessage(`{"sdp":"test"}`),
		Timestamp:    time.Now().Unix(),
	}

	// Marshal to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshaled JSON is empty")
	}

	// Unmarshal back
	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	// Verify fields
	if decoded.Type != msg.Type {
		t.Errorf("Decoded Type = %v, want %v", decoded.Type, msg.Type)
	}

	if decoded.RoomID != msg.RoomID {
		t.Errorf("Decoded RoomID = %v, want %v", decoded.RoomID, msg.RoomID)
	}

	if decoded.PeerID != msg.PeerID {
		t.Errorf("Decoded PeerID = %v, want %v", decoded.PeerID, msg.PeerID)
	}

	if decoded.TargetPeerID != msg.TargetPeerID {
		t.Errorf("Decoded TargetPeerID = %v, want %v", decoded.TargetPeerID, msg.TargetPeerID)
	}

	if string(decoded.Payload) != string(msg.Payload) {
		t.Errorf("Decoded Payload = %v, want %v", string(decoded.Payload), string(msg.Payload))
	}
}

// TestSDPMessageJSON 测试SDP消息JSON序列化
func TestSDPMessageJSON(t *testing.T) {
	sdp := SDPMessage{
		SDP:  "v=0\r\no=- 0 0 IN IP4 0.0.0.0\r\ns=-\r\nt=0 0\r\n",
		Type: "offer",
	}

	data, err := json.Marshal(sdp)
	if err != nil {
		t.Fatalf("Failed to marshal SDP: %v", err)
	}

	var decoded SDPMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SDP: %v", err)
	}

	if decoded.SDP != sdp.SDP {
		t.Errorf("Decoded SDP = %v, want %v", decoded.SDP, sdp.SDP)
	}

	if decoded.Type != sdp.Type {
		t.Errorf("Decoded Type = %v, want %v", decoded.Type, sdp.Type)
	}
}

// TestICEMessageJSON 测试ICE消息JSON序列化
func TestICEMessageJSON(t *testing.T) {
	sdpMLineIndex := uint16(0)
	sdpMid := "0"

	ice := ICEMessage{
		Candidate:        "candidate:1 1 UDP 2130706431 192.168.1.1 5000 typ host",
		SDPMLineIndex:    &sdpMLineIndex,
		SDPMid:           &sdpMid,
		UsernameFragment: "abc123",
	}

	data, err := json.Marshal(ice)
	if err != nil {
		t.Fatalf("Failed to marshal ICE: %v", err)
	}

	var decoded ICEMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ICE: %v", err)
	}

	if decoded.Candidate != ice.Candidate {
		t.Errorf("Decoded Candidate = %v, want %v", decoded.Candidate, ice.Candidate)
	}

	if decoded.SDPMLineIndex == nil || *decoded.SDPMLineIndex != *ice.SDPMLineIndex {
		t.Error("Decoded SDPMLineIndex mismatch")
	}

	if decoded.SDPMid == nil || *decoded.SDPMid != *ice.SDPMid {
		t.Error("Decoded SDPMid mismatch")
	}

	if decoded.UsernameFragment != ice.UsernameFragment {
		t.Errorf("Decoded UsernameFragment = %v, want %v", decoded.UsernameFragment, ice.UsernameFragment)
	}
}

// TestRoomInfoJSON 测试房间信息JSON序列化
func TestRoomInfoJSON(t *testing.T) {
	room := RoomInfo{
		RoomID:    "room123",
		Peers:     []string{"peer1", "peer2", "peer3"},
		CreatedAt: time.Now().Unix(),
	}

	data, err := json.Marshal(room)
	if err != nil {
		t.Fatalf("Failed to marshal RoomInfo: %v", err)
	}

	var decoded RoomInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal RoomInfo: %v", err)
	}

	if decoded.RoomID != room.RoomID {
		t.Errorf("Decoded RoomID = %v, want %v", decoded.RoomID, room.RoomID)
	}

	if len(decoded.Peers) != len(room.Peers) {
		t.Errorf("Decoded Peers length = %v, want %v", len(decoded.Peers), len(room.Peers))
	}

	for i, peer := range room.Peers {
		if decoded.Peers[i] != peer {
			t.Errorf("Decoded Peers[%d] = %v, want %v", i, decoded.Peers[i], peer)
		}
	}

	if decoded.CreatedAt != room.CreatedAt {
		t.Errorf("Decoded CreatedAt = %v, want %v", decoded.CreatedAt, room.CreatedAt)
	}
}

// BenchmarkBuildMerkleTree 基准测试构建Merkle树
func BenchmarkBuildMerkleTree(b *testing.B) {
	chunks := make([][]byte, 100)
	for i := range chunks {
		chunks[i] = make([]byte, 1024)
		for j := range chunks[i] {
			chunks[i][j] = byte(i + j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := BuildMerkleTree(chunks)
		if tree == nil {
			b.Fatal("BuildMerkleTree returned nil")
		}
	}
}

// BenchmarkMerkleTreeVerifyChunk 基准测试验证块
func BenchmarkMerkleTreeVerifyChunk(b *testing.B) {
	chunks := make([][]byte, 100)
	for i := range chunks {
		chunks[i] = make([]byte, 1024)
		for j := range chunks[i] {
			chunks[i][j] = byte(i + j)
		}
	}

	tree := BuildMerkleTree(chunks)
	if tree == nil {
		b.Fatal("BuildMerkleTree returned nil")
	}

	chunkIndex := 50
	chunk := chunks[chunkIndex]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.VerifyChunk(chunkIndex, chunk)
	}
}

// BenchmarkCalculateFileChecksum 基准测试计算文件校验和
func BenchmarkCalculateFileChecksum(b *testing.B) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "benchmark_checksum_*.bin")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write 1MB of data
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	if _, err := tempFile.Write(data); err != nil {
		b.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CalculateFileChecksum(tempFile.Name())
		if err != nil {
			b.Fatalf("CalculateFileChecksum() error: %v", err)
		}
	}
}
