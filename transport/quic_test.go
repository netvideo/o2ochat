package transport

import (
	"testing"
	"time"
)

func TestQUICConnectionOpenStream(t *testing.T) {
	conn := &quicConnection{
		id:          "test-conn",
		peerID:      "test-peer",
		connType:    ConnectionTypeQUIC,
		state:       StateConnected,
		streams:     make(map[uint32]*quicStream),
		config:      DefaultQUICConfig(),
		controlChan: make(chan []byte, 10),
	}

	stream, err := conn.OpenStream(nil)
	if err != nil {
		t.Fatalf("OpenStream failed: %v", err)
	}

	if stream == nil {
		t.Fatal("Expected stream to be created")
	}

	if stream.GetStreamID() == 0 {
		t.Error("Expected non-zero stream ID")
	}
}

func TestQUICConnectionClose(t *testing.T) {
	conn := &quicConnection{
		id:          "test-conn",
		state:       StateConnected,
		streams:     make(map[uint32]*quicStream),
		controlChan: make(chan []byte, 10),
	}

	err := conn.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if conn.state != StateDisconnected {
		t.Error("Expected state to be disconnected")
	}

	err = conn.Close()
	if err != nil {
		t.Error("Expected second close to succeed")
	}
}

func TestQUICConnectionGetInfo(t *testing.T) {
	conn := &quicConnection{
		id:            "test-conn",
		peerID:        "test-peer",
		connType:      ConnectionTypeQUIC,
		state:         StateConnected,
		establishedAt: time.Now(),
		streams:       make(map[uint32]*quicStream),
		controlChan:   make(chan []byte, 10),
	}

	info := conn.GetInfo()

	if info.ID != conn.id {
		t.Errorf("Expected ID %s, got %s", conn.id, info.ID)
	}

	if info.PeerID != conn.peerID {
		t.Errorf("Expected PeerID %s, got %s", conn.peerID, info.PeerID)
	}

	if info.Type != ConnectionTypeQUIC {
		t.Errorf("Expected type QUIC, got %s", info.Type)
	}

	if info.State != StateConnected {
		t.Errorf("Expected state connected, got %s", info.State)
	}
}

func TestQUICConnectionSendControlMessage(t *testing.T) {
	conn := &quicConnection{
		id:          "test-conn",
		state:       StateConnected,
		streams:     make(map[uint32]*quicStream),
		controlChan: make(chan []byte, 10),
	}

	msg := []byte("control message")
	err := conn.SendControlMessage(msg)
	if err != nil {
		t.Errorf("SendControlMessage failed: %v", err)
	}
}

func TestQUICConnectionReceiveControlMessage(t *testing.T) {
	conn := &quicConnection{
		id:          "test-conn",
		state:       StateConnected,
		streams:     make(map[uint32]*quicStream),
		controlChan: make(chan []byte, 10),
	}

	go func() {
		conn.SendControlMessage([]byte("test"))
	}()

	msg, err := conn.ReceiveControlMessage()
	if err != nil {
		t.Fatalf("ReceiveControlMessage failed: %v", err)
	}

	if string(msg) != "test" {
		t.Errorf("Expected 'test', got %s", string(msg))
	}
}

func TestQUICConnectionGetStats(t *testing.T) {
	conn := &quicConnection{
		id:          "test-conn",
		state:       StateConnected,
		streams:     make(map[uint32]*quicStream),
		controlChan: make(chan []byte, 10),
	}

	stats := conn.GetStats()

	if stats.BytesSent != 0 {
		t.Errorf("Expected 0 bytes sent, got %d", stats.BytesSent)
	}

	if stats.BytesReceived != 0 {
		t.Errorf("Expected 0 bytes received, got %d", stats.BytesReceived)
	}
}

func TestQUICStreamRead(t *testing.T) {
	stream := &quicStream{
		id:        1,
		direction: "outbound",
	}

	_, err := stream.Read(make([]byte, 10))
	if err == nil {
		t.Error("Expected error for uninitialized stream")
	}
}

func TestQUICStreamWrite(t *testing.T) {
	stream := &quicStream{
		id:        1,
		direction: "outbound",
	}

	_, err := stream.Write([]byte("test"))
	if err == nil {
		t.Error("Expected error for uninitialized stream")
	}
}

func TestQUICStreamClose(t *testing.T) {
	conn := &quicConnection{
		streams:     make(map[uint32]*quicStream),
		controlChan: make(chan []byte, 10),
	}

	stream := &quicStream{
		id:        1,
		conn:      conn,
		direction: "outbound",
	}

	conn.streams[1] = stream

	err := stream.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if _, exists := conn.streams[1]; exists {
		t.Error("Expected stream to be removed")
	}
}

func TestQUICStreamGetInfo(t *testing.T) {
	stream := &quicStream{
		id:        1,
		direction: "outbound",
	}

	info := stream.GetInfo()

	if info.ID != 1 {
		t.Errorf("Expected ID 1, got %d", info.ID)
	}

	if info.Direction != "outbound" {
		t.Errorf("Expected direction outbound, got %s", info.Direction)
	}

	if info.State != "open" {
		t.Errorf("Expected state open, got %s", info.State)
	}
}

func TestQUICStreamDeadlines(t *testing.T) {
	stream := &quicStream{
		id:        1,
		direction: "outbound",
	}

	err := stream.SetDeadline(time.Now().Add(time.Second))
	if err == nil {
		t.Error("Expected error for uninitialized stream")
	}

	err = stream.SetReadDeadline(time.Now().Add(time.Second))
	if err == nil {
		t.Error("Expected error for uninitialized stream")
	}

	err = stream.SetWriteDeadline(time.Now().Add(time.Second))
	if err == nil {
		t.Error("Expected error for uninitialized stream")
	}
}

func TestQUICConnectionUpdateStats(t *testing.T) {
	conn := &quicConnection{
		id:          "test-conn",
		state:       StateConnected,
		streams:     make(map[uint32]*quicStream),
		controlChan: make(chan []byte, 10),
	}

	conn.updateStats(100, 200)

	stats := conn.GetStats()

	if stats.BytesSent != 100 {
		t.Errorf("Expected 100 bytes sent, got %d", stats.BytesSent)
	}

	if stats.BytesReceived != 200 {
		t.Errorf("Expected 200 bytes received, got %d", stats.BytesReceived)
	}
}
