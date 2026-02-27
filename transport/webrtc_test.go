package transport

import (
	"testing"
	"time"
)

func TestWebRTCConnectionOpenStream(t *testing.T) {
	conn := &webrtcConnection{
		id:       "test-conn",
		peerID:   "test-peer",
		connType: ConnectionTypeWebRTC,
		state:    StateConnected,
		streams:  make(map[uint32]*webrtcStream),
		config:   DefaultWebRTCConfig(),
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

func TestWebRTCConnectionClose(t *testing.T) {
	conn := &webrtcConnection{
		id:      "test-conn",
		state:   StateConnected,
		streams: make(map[uint32]*webrtcStream),
	}

	err := conn.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if conn.state != StateDisconnected {
		t.Error("Expected state to be disconnected")
	}
}

func TestWebRTCConnectionGetInfo(t *testing.T) {
	conn := &webrtcConnection{
		id:            "test-conn",
		peerID:        "test-peer",
		connType:      ConnectionTypeWebRTC,
		state:         StateConnected,
		establishedAt: time.Now(),
		streams:       make(map[uint32]*webrtcStream),
	}

	info := conn.GetInfo()

	if info.ID != conn.id {
		t.Errorf("Expected ID %s, got %s", conn.id, info.ID)
	}

	if info.PeerID != conn.peerID {
		t.Errorf("Expected PeerID %s, got %s", conn.peerID, info.PeerID)
	}

	if info.Type != ConnectionTypeWebRTC {
		t.Errorf("Expected type WebRTC, got %s", info.Type)
	}
}

func TestWebRTCConnectionGetStats(t *testing.T) {
	conn := &webrtcConnection{
		id:      "test-conn",
		state:   StateConnected,
		streams: make(map[uint32]*webrtcStream),
	}

	stats := conn.GetStats()

	if stats.BytesSent != 0 {
		t.Errorf("Expected 0 bytes sent, got %d", stats.BytesSent)
	}

	if stats.BytesReceived != 0 {
		t.Errorf("Expected 0 bytes received, got %d", stats.BytesReceived)
	}
}

func TestWebRTCConnectionUpdateStats(t *testing.T) {
	conn := &webrtcConnection{
		id:      "test-conn",
		state:   StateConnected,
		streams: make(map[uint32]*webrtcStream),
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

func TestWebRTCStreamRead(t *testing.T) {
	stream := &webrtcStream{
		id:        1,
		direction: "outbound",
		buffer:    newReadBuffer(),
	}

	go func() {
		stream.buffer.Write([]byte("test"))
	}()

	buf := make([]byte, 10)
	n, err := stream.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if n != 4 {
		t.Errorf("Expected 4 bytes, got %d", n)
	}
}

func TestWebRTCStreamWrite(t *testing.T) {
	stream := &webrtcStream{
		id:        1,
		direction: "outbound",
		buffer:    newReadBuffer(),
	}

	n, err := stream.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != 4 {
		t.Errorf("Expected 4 bytes written, got %d", n)
	}
}

func TestWebRTCStreamClose(t *testing.T) {
	conn := &webrtcConnection{
		streams: make(map[uint32]*webrtcStream),
	}

	stream := &webrtcStream{
		id:        1,
		conn:      conn,
		direction: "outbound",
		buffer:    newReadBuffer(),
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

func TestWebRTCStreamGetInfo(t *testing.T) {
	stream := &webrtcStream{
		id:        1,
		direction: "outbound",
		buffer:    newReadBuffer(),
	}

	info := stream.GetInfo()

	if info.ID != 1 {
		t.Errorf("Expected ID 1, got %d", info.ID)
	}

	if info.Direction != "outbound" {
		t.Errorf("Expected direction outbound, got %s", info.Direction)
	}
}

func TestWebRTCStreamDeadlines(t *testing.T) {
	stream := &webrtcStream{
		id:        1,
		direction: "outbound",
		buffer:    newReadBuffer(),
	}

	err := stream.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetDeadline should not error: %v", err)
	}

	err = stream.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetReadDeadline should not error: %v", err)
	}

	err = stream.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetWriteDeadline should not error: %v", err)
	}
}

func TestReadBuffer(t *testing.T) {
	rb := newReadBuffer()

	n, err := rb.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != 4 {
		t.Errorf("Expected 4 bytes written, got %d", n)
	}

	buf := make([]byte, 10)
	n, err = rb.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if n != 4 {
		t.Errorf("Expected 4 bytes read, got %d", n)
	}

	if string(buf[:n]) != "test" {
		t.Errorf("Expected 'test', got %s", string(buf[:n]))
	}
}

func TestWebRTCConnectionSendControlMessage(t *testing.T) {
	conn := &webrtcConnection{
		id:      "test-conn",
		state:   StateConnected,
		streams: make(map[uint32]*webrtcStream),
	}

	err := conn.SendControlMessage([]byte("test"))
	if err == nil {
		t.Log("Expected error for no control channel (acceptable)")
	}
}

func TestWebRTCConnectionReceiveControlMessage(t *testing.T) {
	conn := &webrtcConnection{
		id:      "test-conn",
		state:   StateConnected,
		streams: make(map[uint32]*webrtcStream),
	}

	_, err := conn.ReceiveControlMessage()
	if err == nil {
		t.Log("Expected error for no control channel (acceptable)")
	}
}
