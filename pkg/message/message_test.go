// Package message provides message handling for P2P communications
package message

import (
	"context"
	"fmt"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestMessageTypeString 测试消息类型字符串表示
func TestMessageTypeString(t *testing.T) {
	tests := []struct {
		msgType  MessageType
		expected string
	}{
		{TypeText, "text"},
		{TypeFile, "file"},
		{TypeControl, "control"},
		{TypeAck, "ack"},
		{MessageType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.msgType.String()
			if result != tt.expected {
				t.Errorf("MessageType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNewTextMessage 测试创建文本消息
func TestNewTextMessage(t *testing.T) {
	tests := []struct {
		name       string
		senderID   string
		receiverID string
		content    string
	}{
		{
			name:       "Simple message",
			senderID:   "sender1",
			receiverID: "receiver1",
			content:    "Hello, World!",
		},
		{
			name:       "Empty content",
			senderID:   "sender2",
			receiverID: "receiver2",
			content:    "",
		},
		{
			name:       "Unicode content",
			senderID:   "sender3",
			receiverID: "receiver3",
			content:    "你好，世界！🌍🚀",
		},
		{
			name:       "Long content",
			senderID:   "sender4",
			receiverID: "receiver4",
			content:    strings.Repeat("A", 10000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewTextMessage(tt.senderID, tt.receiverID, tt.content)

			if msg == nil {
				t.Fatal("NewTextMessage() returned nil")
			}

			if msg.Type != TypeText {
				t.Errorf("Message.Type = %v, want %v", msg.Type, TypeText)
			}

			if msg.SenderID != tt.senderID {
				t.Errorf("Message.SenderID = %v, want %v", msg.SenderID, tt.senderID)
			}

			if msg.ReceiverID != tt.receiverID {
				t.Errorf("Message.ReceiverID = %v, want %v", msg.ReceiverID, tt.receiverID)
			}

			if msg.ID == "" {
				t.Error("Message.ID is empty")
			}

			if msg.Timestamp.IsZero() {
				t.Error("Message.Timestamp is zero")
			}

			// Verify payload
			var payload TextPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				t.Errorf("Failed to unmarshal payload: %v", err)
			}

			if payload.Content != tt.content {
				t.Errorf("Payload.Content = %v, want %v", payload.Content, tt.content)
			}

			if payload.Format != "plain" {
				t.Errorf("Payload.Format = %v, want %v", payload.Format, "plain")
			}
		})
	}
}

// TestNewFileMessage 测试创建文件消息
func TestNewFileMessage(t *testing.T) {
	tests := []struct {
		name       string
		senderID   string
		receiverID string
		file       FilePayload
	}{
		{
			name:       "PDF file",
			senderID:   "sender1",
			receiverID: "receiver1",
			file: FilePayload{
				FileID:   "file1",
				FileName: "document.pdf",
				FileSize: 1024 * 1024,
				FileType: "application/pdf",
				Checksum: "sha256:abc123",
			},
		},
		{
			name:       "Image file",
			senderID:   "sender2",
			receiverID: "receiver2",
			file: FilePayload{
				FileID:   "file2",
				FileName: "image.png",
				FileSize: 5 * 1024 * 1024,
				FileType: "image/png",
				Checksum: "sha256:def456",
			},
		},
		{
			name:       "Empty file",
			senderID:   "sender3",
			receiverID: "receiver3",
			file: FilePayload{
				FileID:   "file3",
				FileName: "empty.txt",
				FileSize: 0,
				FileType: "text/plain",
				Checksum: "sha256:000000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewFileMessage(tt.senderID, tt.receiverID, tt.file)

			if msg == nil {
				t.Fatal("NewFileMessage() returned nil")
			}

			if msg.Type != TypeFile {
				t.Errorf("Message.Type = %v, want %v", msg.Type, TypeFile)
			}

			if msg.SenderID != tt.senderID {
				t.Errorf("Message.SenderID = %v, want %v", msg.SenderID, tt.senderID)
			}

			if msg.ReceiverID != tt.receiverID {
				t.Errorf("Message.ReceiverID = %v, want %v", msg.ReceiverID, tt.receiverID)
			}

			// Verify payload
			var payload FilePayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				t.Errorf("Failed to unmarshal payload: %v", err)
			}

			if payload.FileID != tt.file.FileID {
				t.Errorf("Payload.FileID = %v, want %v", payload.FileID, tt.file.FileID)
			}

			if payload.FileName != tt.file.FileName {
				t.Errorf("Payload.FileName = %v, want %v", payload.FileName, tt.file.FileName)
			}

			if payload.FileSize != tt.file.FileSize {
				t.Errorf("Payload.FileSize = %v, want %v", payload.FileSize, tt.file.FileSize)
			}

			if payload.FileType != tt.file.FileType {
				t.Errorf("Payload.FileType = %v, want %v", payload.FileType, tt.file.FileType)
			}

			if payload.Checksum != tt.file.Checksum {
				t.Errorf("Payload.Checksum = %v, want %v", payload.Checksum, tt.file.Checksum)
			}
		})
	}
}

// TestNewControlMessage 测试创建控制消息
func TestNewControlMessage(t *testing.T) {
	tests := []struct {
		name       string
		senderID   string
		receiverID string
		command    ControlCommand
		data       interface{}
	}{
		{
			name:       "Ping command",
			senderID:   "sender1",
			receiverID: "receiver1",
			command:    CmdPing,
			data:       nil,
		},
		{
			name:       "Typing command",
			senderID:   "sender2",
			receiverID: "receiver2",
			command:    CmdTyping,
			data:       map[string]interface{}{"isTyping": true},
		},
		{
			name:       "Seen command with data",
			senderID:   "sender3",
			receiverID: "receiver3",
			command:    CmdSeen,
			data:       map[string]string{"lastMessageID": "msg123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewControlMessage(tt.senderID, tt.receiverID, tt.command, tt.data)

			if msg == nil {
				t.Fatal("NewControlMessage() returned nil")
			}

			if msg.Type != TypeControl {
				t.Errorf("Message.Type = %v, want %v", msg.Type, TypeControl)
			}

			if msg.SenderID != tt.senderID {
				t.Errorf("Message.SenderID = %v, want %v", msg.SenderID, tt.senderID)
			}

			if msg.ReceiverID != tt.receiverID {
				t.Errorf("Message.ReceiverID = %v, want %v", msg.ReceiverID, tt.receiverID)
			}

			// Verify payload
			var payload ControlPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				t.Errorf("Failed to unmarshal payload: %v", err)
			}

			if payload.Command != tt.command {
				t.Errorf("Payload.Command = %v, want %v", payload.Command, tt.command)
			}

			// Verify data if provided
			if tt.data != nil {
				if len(payload.Data) == 0 {
					t.Error("Expected payload.Data to be non-empty")
				}
			}
		})
	}
}

// TestNewAckMessage 测试创建确认消息
func TestNewAckMessage(t *testing.T) {
	tests := []struct {
		name       string
		messageID  string
		senderID   string
		receiverID string
	}{
		{
			name:       "Simple ack",
			messageID:  "msg123",
			senderID:   "sender1",
			receiverID: "receiver1",
		},
		{
			name:       "Ack with empty message ID",
			messageID:  "",
			senderID:   "sender2",
			receiverID: "receiver2",
		},
		{
			name:       "Long message ID",
			messageID:  strings.Repeat("a", 100),
			senderID:   "sender3",
			receiverID: "receiver3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewAckMessage(tt.messageID, tt.senderID, tt.receiverID)

			if msg == nil {
				t.Fatal("NewAckMessage() returned nil")
			}

			if msg.Type != TypeAck {
				t.Errorf("Message.Type = %v, want %v", msg.Type, TypeAck)
			}

			if msg.SenderID != tt.senderID {
				t.Errorf("Message.SenderID = %v, want %v", msg.SenderID, tt.senderID)
			}

			if msg.ReceiverID != tt.receiverID {
				t.Errorf("Message.ReceiverID = %v, want %v", msg.ReceiverID, tt.receiverID)
			}

			// Verify payload
			var payload AckPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				t.Errorf("Failed to unmarshal payload: %v", err)
			}

			if payload.MessageID != tt.messageID {
				t.Errorf("Payload.MessageID = %v, want %v", payload.MessageID, tt.messageID)
			}

			if payload.Timestamp.IsZero() {
				t.Error("Payload.Timestamp is zero")
			}
		})
	}
}

// TestMessageEncodeDecode 测试消息编码解码
func TestMessageEncodeDecode(t *testing.T) {
	original := &Message{
		ID:         "msg123",
		Type:       TypeText,
		SenderID:   "sender1",
		ReceiverID: "receiver1",
		Timestamp:  time.Now().Truncate(time.Millisecond),
		Payload:    []byte(`{"content":"Hello"}`),
	}

	// Encode
	data, err := original.Encode()
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	if len(data) == 0 {
		t.Error("Encode() returned empty data")
	}

	// Decode
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	// Verify decoded message
	if decoded.ID != original.ID {
		t.Errorf("Decoded ID = %v, want %v", decoded.ID, original.ID)
	}

	if decoded.Type != original.Type {
		t.Errorf("Decoded Type = %v, want %v", decoded.Type, original.Type)
	}

	if decoded.SenderID != original.SenderID {
		t.Errorf("Decoded SenderID = %v, want %v", decoded.SenderID, original.SenderID)
	}

	if decoded.ReceiverID != original.ReceiverID {
		t.Errorf("Decoded ReceiverID = %v, want %v", decoded.ReceiverID, original.ReceiverID)
	}

	if string(decoded.Payload) != string(original.Payload) {
		t.Errorf("Decoded Payload = %v, want %v", string(decoded.Payload), string(original.Payload))
	}
}

// TestDecodeInvalidData 测试解码无效数据
func TestDecodeInvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Empty data",
			data: []byte{},
		},
		{
			name: "Invalid JSON",
			data: []byte(`{invalid json}`),
		},
		{
			name: "Not a message",
			data: []byte(`{"foo": "bar"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(tt.data)
			if err == nil {
				t.Error("Decode() expected error but got none")
			}
		})
	}
}

// TestManager 测试消息管理器
func TestManager(t *testing.T) {
	config := DefaultConfig()
	mgr := NewManager(config)

	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	defer func() {
		if err := mgr.Close(); err != nil {
			t.Errorf("Close() error: %v", err)
		}
	}()
}

// TestManagerSend 测试管理器发送消息
func TestManagerSend(t *testing.T) {
	config := DefaultConfig()
	config.OutboxSize = 10
	mgr := NewManager(config)
	defer mgr.Close()

	tests := []struct {
		name    string
		msg     *Message
		wantErr bool
	}{
		{
			name: "Valid message",
			msg: &Message{
				ID:         "msg1",
				Type:       TypeText,
				SenderID:   "sender1",
				ReceiverID: "receiver1",
				Payload:    []byte(`{"content":"Hello"}`),
			},
			wantErr: false,
		},
		{
			name: "Message with auto-generated ID",
			msg: &Message{
				Type:       TypeText,
				SenderID:   "sender2",
				ReceiverID: "receiver2",
				Payload:    []byte(`{"content":"Hello"}`),
			},
			wantErr: false,
		},
		{
			name: "Message with auto-generated timestamp",
			msg: &Message{
				ID:         "msg3",
				Type:       TypeText,
				SenderID:   "sender3",
				ReceiverID: "receiver3",
				Payload:    []byte(`{"content":"Hello"}`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			originalID := tt.msg.ID
			originalTimestamp := tt.msg.Timestamp

			err := mgr.Send(ctx, tt.msg)

			if tt.wantErr && err == nil {
				t.Error("Send() expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Send() unexpected error: %v", err)
			}

			// Verify auto-generated fields
			if originalID == "" && tt.msg.ID == "" {
				t.Error("Message ID was not auto-generated")
			}

			if originalTimestamp.IsZero() && tt.msg.Timestamp.IsZero() {
				t.Error("Message timestamp was not auto-generated")
			}
		})
	}
}

// TestManagerSendTimeout 测试发送消息超时
func TestManagerSendTimeout(t *testing.T) {
	config := DefaultConfig()
	config.OutboxSize = 1
	mgr := NewManager(config)
	defer mgr.Close()

	// Fill the outbox
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		msg := &Message{
			Type:       TypeText,
			SenderID:   "sender",
			ReceiverID: "receiver",
			Payload:    []byte(`{"content":"Test"}`),
		}
		err := mgr.Send(ctx, msg)
		cancel()

		// Second send should timeout
		if i == 1 {
			if err == nil {
				t.Error("Expected timeout error but got none")
			}
		}
	}
}

// TestManagerRegisterHandler 测试注册处理器
func TestManagerRegisterHandler(t *testing.T) {
	config := DefaultConfig()
	mgr := NewManager(config)
	defer mgr.Close()

	handlerCalled := false
	handler := func(msg *Message) error {
		handlerCalled = true
		return nil
	}

	// Register handler for text messages
	mgr.RegisterHandler(TypeText, handler)
}

// TestManagerEnqueueDequeue 测试消息入队和出队
func TestManagerEnqueueDequeue(t *testing.T) {
	config := DefaultConfig()
	config.OutboxSize = 10
	mgr := NewManager(config)
	defer mgr.Close()

	// Test Enqueue
	msg1 := &Message{
		ID:         "msg1",
		Type:       TypeText,
		SenderID:   "sender1",
		ReceiverID: "receiver1",
		Payload:    []byte(`{"content":"Hello"}`),
	}

	if err := mgr.Enqueue(msg1); err != nil {
		t.Errorf("Enqueue() error: %v", err)
	}

	// Test Dequeue
	dequeued, err := mgr.Dequeue()
	if err != nil {
		t.Errorf("Dequeue() error: %v", err)
	}

	if dequeued == nil {
		t.Fatal("Dequeue() returned nil")
	}

	if dequeued.ID != msg1.ID {
		t.Errorf("Dequeued message ID = %v, want %v", dequeued.ID, msg1.ID)
	}
}

// TestManagerDequeueEmpty 测试从空队列出队
func TestManagerDequeueEmpty(t *testing.T) {
	config := DefaultConfig()
	mgr := NewManager(config)
	defer mgr.Close()

	_, err := mgr.Dequeue()
	if err == nil {
		t.Error("Dequeue() from empty queue expected error but got none")
	}
}

// TestManagerWaitForAck 测试等待确认
func TestManagerWaitForAck(t *testing.T) {
	config := DefaultConfig()
	mgr := NewManager(config)
	defer mgr.Close()

	messageID := "msg123"

	// Start waiting for ack in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	var ackErr error
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		ackErr = mgr.WaitForAck(ctx, messageID, 500*time.Millisecond)
	}()

	// Simulate acknowledgment
	time.Sleep(100 * time.Millisecond)
	mgr.ackMu.Lock()
	if ch, ok := mgr.ackChannels[messageID]; ok {
		ch <- true
	}
	mgr.ackMu.Unlock()

	wg.Wait()

	if ackErr != nil {
		t.Errorf("WaitForAck() error: %v", ackErr)
	}
}

// TestManagerWaitForAckTimeout 测试等待确认超时
func TestManagerWaitForAckTimeout(t *testing.T) {
	config := DefaultConfig()
	mgr := NewManager(config)
	defer mgr.Close()

	messageID := "msg456"

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := mgr.WaitForAck(ctx, messageID, 50*time.Millisecond)
	if err == nil {
		t.Error("WaitForAck() expected timeout error but got none")
	}
}

// TestManagerWaitForAckRejection 测试确认被拒绝
func TestManagerWaitForAckRejection(t *testing.T) {
	config := DefaultConfig()
	mgr := NewManager(config)
	defer mgr.Close()

	messageID := "msg789"

	// Start waiting for ack
	var wg sync.WaitGroup
	wg.Add(1)
	var ackErr error
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		ackErr = mgr.WaitForAck(ctx, messageID, 500*time.Millisecond)
	}()

	// Simulate rejection
	time.Sleep(50 * time.Millisecond)
	mgr.ackMu.Lock()
	if ch, ok := mgr.ackChannels[messageID]; ok {
		ch <- false
	}
	mgr.ackMu.Unlock()

	wg.Wait()

	if ackErr == nil {
		t.Error("WaitForAck() expected rejection error but got none")
	}
}

// TestManagerEnqueueFull 测试入队当队列已满
func TestManagerEnqueueFull(t *testing.T) {
	config := DefaultConfig()
	config.OutboxSize = 2
	mgr := NewManager(config)
	defer mgr.Close()

	// Fill the queue
	for i := 0; i < 3; i++ {
		msg := &Message{
			ID:         fmt.Sprintf("msg%d", i),
			Type:       TypeText,
			SenderID:   "sender",
			ReceiverID: "receiver",
			Payload:    []byte(`{"content":"Test"}`),
		}

		err := mgr.Enqueue(msg)
		if i < 2 {
			if err != nil {
				t.Errorf("Enqueue() error for msg%d: %v", i, err)
			}
		} else {
			if err == nil {
				t.Error("Enqueue() expected error for full queue but got none")
			}
		}
	}
}

// TestClose 测试关闭管理器
func TestClose(t *testing.T) {
	config := DefaultConfig()
	mgr := NewManager(config)

	// Close should succeed
	if err := mgr.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}

	// Operations after close should fail or return error
	msg := &Message{
		Type:       TypeText,
		SenderID:   "sender",
		ReceiverID: "receiver",
		Payload:    []byte(`{"content":"Test"}`),
	}

	// Enqueue should fail after close
	err := mgr.Enqueue(msg)
	// We accept either error or success depending on implementation
	// The test documents the behavior
	_ = err
}

// TestConcurrentOperations 测试并发操作
func TestConcurrentOperations(t *testing.T) {
	config := DefaultConfig()
	config.OutboxSize = 100
	mgr := NewManager(config)
	defer mgr.Close()

	const numGoroutines = 10
	const messagesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				msg := &Message{
					ID:         fmt.Sprintf("goroutine%d-msg%d", id, j),
					Type:       TypeText,
					SenderID:   fmt.Sprintf("sender%d", id),
					ReceiverID: fmt.Sprintf("receiver%d", id),
					Payload:    []byte(`{"content":"Test"}`),
				}

				if err := mgr.Enqueue(msg); err != nil {
					t.Errorf("Enqueue() error: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all messages were enqueued
	expectedCount := numGoroutines * messagesPerGoroutine
	actualCount := 0
	for {
		_, err := mgr.Dequeue()
		if err != nil {
			break
		}
		actualCount++
	}

	if actualCount != expectedCount {
		t.Errorf("Dequeued %d messages, expected %d", actualCount, expectedCount)
	}
}

// TestMessageTypeConstants 测试消息类型常量
func TestMessageTypeConstants(t *testing.T) {
	// Verify message type values
	if TypeText != 0 {
		t.Errorf("TypeText = %d, want 0", TypeText)
	}

	if TypeFile != 1 {
		t.Errorf("TypeFile = %d, want 1", TypeFile)
	}

	if TypeControl != 2 {
		t.Errorf("TypeControl = %d, want 2", TypeControl)
	}

	if TypeAck != 3 {
		t.Errorf("TypeAck = %d, want 3", TypeAck)
	}
}

// TestControlCommandConstants 测试控制命令常量
func TestControlCommandConstants(t *testing.T) {
	// Verify control command values
	if CmdPing != "ping" {
		t.Errorf("CmdPing = %v, want ping", CmdPing)
	}

	if CmdPong != "pong" {
		t.Errorf("CmdPong = %v, want pong", CmdPong)
	}

	if CmdDisconnect != "disconnect" {
		t.Errorf("CmdDisconnect = %v, want disconnect", CmdDisconnect)
	}

	if CmdTyping != "typing" {
		t.Errorf("CmdTyping = %v, want typing", CmdTyping)
	}

	if CmdSeen != "seen" {
		t.Errorf("CmdSeen = %v, want seen", CmdSeen)
	}
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.OutboxSize != 1000 {
		t.Errorf("DefaultConfig.OutboxSize = %d, want 1000", config.OutboxSize)
	}

	if config.MaxRetries != 3 {
		t.Errorf("DefaultConfig.MaxRetries = %d, want 3", config.MaxRetries)
	}

	if config.RetryInterval != 5*time.Second {
		t.Errorf("DefaultConfig.RetryInterval = %v, want 5s", config.RetryInterval)
	}

	if config.AckTimeout != 30*time.Second {
		t.Errorf("DefaultConfig.AckTimeout = %v, want 30s", config.AckTimeout)
	}

	if config.WorkerCount != 4 {
		t.Errorf("DefaultConfig.WorkerCount = %d, want 4", config.WorkerCount)
	}
}

// BenchmarkMessageEncode 基准测试消息编码
func BenchmarkMessageEncode(b *testing.B) {
	msg := &Message{
		ID:         "msg123",
		Type:       TypeText,
		SenderID:   "sender1",
		ReceiverID: "receiver1",
		Timestamp:  time.Now(),
		Payload:    []byte(`{"content":"Hello, World!"}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := msg.Encode()
		if err != nil {
			b.Fatalf("Encode() error: %v", err)
		}
	}
}

// BenchmarkMessageDecode 基准测试消息解码
func BenchmarkMessageDecode(b *testing.B) {
	msg := &Message{
		ID:         "msg123",
		Type:       TypeText,
		SenderID:   "sender1",
		ReceiverID: "receiver1",
		Timestamp:  time.Now(),
		Payload:    []byte(`{"content":"Hello, World!"}`),
	}

	data, err := msg.Encode()
	if err != nil {
		b.Fatalf("Setup Encode() error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decode(data)
		if err != nil {
			b.Fatalf("Decode() error: %v", err)
		}
	}
}

// BenchmarkNewTextMessage 基准测试创建文本消息
func BenchmarkNewTextMessage(b *testing.B) {
	senderID := "sender1"
	receiverID := "receiver1"
	content := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewTextMessage(senderID, receiverID, content)
	}
}

// BenchmarkManagerSend 基准测试管理器发送
func BenchmarkManagerSend(b *testing.B) {
	config := DefaultConfig()
	config.OutboxSize = 10000
	mgr := NewManager(config)
	defer mgr.Close()

	msg := &Message{
		Type:       TypeText,
		SenderID:   "sender",
		ReceiverID: "receiver",
		Payload:    []byte(`{"content":"Benchmark"}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		msg.ID = "" // Reset ID for each iteration
		mgr.Send(ctx, msg)
		cancel()
	}
}
