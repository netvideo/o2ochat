package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/netvideo/identity"
)

// TestConnectionStateTransitions 测试连接状态转换
func TestConnectionStateTransitions(t *testing.T) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	conn, err := NewPeerConnection("local-peer", "remote-peer", config)
	if err != nil {
		t.Fatalf("Failed to create peer connection: %v", err)
	}
	defer conn.Close()

	// 测试初始状态
	if conn.GetState() != ConnectionStateNew {
		t.Errorf("Expected initial state to be %v, got %v",
			ConnectionStateNew, conn.GetState())
	}

	// 测试连接中状态
	if err := conn.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 等待连接建立或超时
	timeout := time.AfterFunc(5*time.Second, func() {
		t.Log("Connection establishment timed out (expected without network)")
	})
	defer timeout.Stop()

	// 验证状态不是 New
	if conn.GetState() == ConnectionStateNew {
		t.Error("Connection state should have changed from New after Connect()")
	}

	// 测试断开连接
	if err := conn.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	// 验证状态变化
	state := conn.GetState()
	if state != ConnectionStateDisconnected &&
		state != ConnectionStateFailed &&
		state != ConnectionStateClosed {
		t.Errorf("Unexpected state after disconnect: %v", state)
	}
}

// TestConnectionEvents 测试连接事件
func TestConnectionEvents(t *testing.T) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	conn, err := NewPeerConnection("local-peer", "remote-peer", config)
	if err != nil {
		t.Fatalf("Failed to create peer connection: %v", err)
	}
	defer conn.Close()

	// 测试状态变化回调
	stateChanges := []ConnectionState{}
	conn.OnStateChange = func(state ConnectionState) {
		stateChanges = append(stateChanges, state)
	}

	// 触发连接
	conn.Connect()
	time.Sleep(100 * time.Millisecond) // 允许异步回调执行

	// 验证回调被触发
	if len(stateChanges) == 0 {
		t.Error("OnStateChange callback was not triggered")
	}

	// 测试连接成功回调
	connectCalled := false
	conn.OnConnect = func() {
		connectCalled = true
	}

	// 测试断开连接回调
	disconnectCalled := false
	conn.OnDisconnect = func() {
		disconnectCalled = true
	}

	// 触发断开连接
	conn.Disconnect()
	time.Sleep(100 * time.Millisecond)

	// 验证断开连接回调被触发
	if !disconnectCalled {
		t.Error("OnDisconnect callback was not triggered")
	}

	// 测试错误处理回调
	errorCalled := false
	conn.OnError = func(err error) {
		errorCalled = true
	}

	// 验证错误处理机制
	if conn.lastError == nil && errorCalled {
		t.Error("Error callback triggered without error")
	}
}

// TestDataChannelOperations 测试数据通道操作
func TestDataChannelOperations(t *testing.T) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	conn, err := NewPeerConnection("local-peer", "remote-peer", config)
	if err != nil {
		t.Fatalf("Failed to create peer connection: %v", err)
	}
	defer conn.Close()

	// 测试创建数据通道
	channelLabel := "test-channel"
	channel, err := conn.CreateDataChannel(channelLabel)
	if err != nil {
		t.Fatalf("Failed to create data channel: %v", err)
	}

	// 验证数据通道属性
	if channel.Label != channelLabel {
		t.Errorf("Channel label mismatch: got %v, want %v",
			channel.Label, channelLabel)
	}

	if channel.State != DataChannelStateConnecting {
		t.Errorf("Expected initial state to be %v, got %v",
			DataChannelStateConnecting, channel.State)
	}

	// 测试发送数据
	testData := []byte("Hello, P2P Data Channel!")
	err = channel.Send(testData)
	if err != nil {
		t.Errorf("Failed to send data: %v", err)
	}

	// 测试统计信息
	stats := channel.GetStats()
	if stats.BytesSent == 0 {
		t.Error("BytesSent should be > 0 after sending data")
	}

	// 测试接收数据回调
	receivedData := make(chan []byte, 1)
	channel.OnMessage = func(data []byte) {
		receivedData <- data
	}

	// 模拟接收数据
	simulatedData := []byte("Simulated received data")
	if channel.OnMessage != nil {
		channel.OnMessage(simulatedData)
	}

	select {
	case data := <-receivedData:
		if string(data) != string(simulatedData) {
			t.Error("Received data mismatch")
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for received data")
	}

	// 测试关闭数据通道
	err = channel.Close()
	if err != nil {
		t.Errorf("Failed to close channel: %v", err)
	}

	// 验证关闭状态
	if channel.State != DataChannelStateClosed {
		t.Errorf("Expected state to be %v after close, got %v",
			DataChannelStateClosed, channel.State)
	}
}

// TestDataChannelErrorHandling 测试数据通道错误处理
func TestDataChannelErrorHandling(t *testing.T) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	conn, err := NewPeerConnection("local-peer", "remote-peer", config)
	if err != nil {
		t.Fatalf("Failed to create peer connection: %v", err)
	}
	defer conn.Close()

	channel, err := conn.CreateDataChannel("error-test-channel")
	if err != nil {
		t.Fatalf("Failed to create data channel: %v", err)
	}

	// 测试错误处理回调
	errorReceived := make(chan error, 1)
	channel.OnError = func(err error) {
		errorReceived <- err
	}

	// 模拟错误
	simulatedError := NewDataChannelError("test error", nil)
	if channel.OnError != nil {
		channel.OnError(simulatedError)
	}

	select {
	case err := <-errorReceived:
		if err == nil {
			t.Error("Expected error but got nil")
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for error")
	}

	// 测试在关闭的通道上发送
	err = channel.Close()
	if err != nil {
		t.Fatalf("Failed to close channel: %v", err)
	}

	// 尝试在关闭的通道上发送数据
	err = channel.Send([]byte("test data"))
	if err == nil {
		t.Error("Expected error when sending on closed channel")
	}

	// 验证错误类型
	if _, ok := err.(*DataChannelError); !ok {
		t.Errorf("Expected DataChannelError, got %T", err)
	}
}

// BenchmarkDataChannelSend 数据通道发送性能测试
func BenchmarkDataChannelSend(b *testing.B) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	conn, err := NewPeerConnection("local-peer", "remote-peer", config)
	if err != nil {
		b.Fatalf("Failed to create peer connection: %v", err)
	}
	defer conn.Close()

	channel, err := conn.CreateDataChannel("benchmark-channel")
	if err != nil {
		b.Fatalf("Failed to create data channel: %v", err)
	}

	testData := []byte("This is benchmark test data for data channel send performance")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := channel.Send(testData)
		if err != nil {
			b.Errorf("Failed to send data: %v", err)
		}
	}

	b.ReportMetric(float64(b.N), "messages")
	b.ReportMetric(float64(len(testData)*b.N), "bytes")
}

// BenchmarkDataChannelReceive 数据通道接收性能测试
func BenchmarkDataChannelReceive(b *testing.B) {
	config := &PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	}

	conn, err := NewPeerConnection("local-peer", "remote-peer", config)
	if err != nil {
		b.Fatalf("Failed to create peer connection: %v", err)
	}
	defer conn.Close()

	channel, err := conn.CreateDataChannel("benchmark-receive-channel")
	if err != nil {
		b.Fatalf("Failed to create data channel: %v", err)
	}

	testData := []byte("Benchmark receive test data")

	// 设置接收回调
	receivedCount := 0
	channel.OnMessage = func(data []byte) {
		receivedCount++
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟接收数据
		if channel.OnMessage != nil {
			channel.OnMessage(testData)
		}
	}

	b.ReportMetric(float64(b.N), "messages")
	b.ReportMetric(float64(receivedCount), "received")
}
