package transport

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type webrtcConnection struct {
	id            string
	peerID        string
	connType      ConnectionType
	state         ConnectionState
	establishedAt time.Time
	closed        int32
	streams       map[uint32]*webrtcStream
	streamsMu     sync.RWMutex
	config        *WebRTCConfig
	stats         ConnectionStats
	statsMu       sync.RWMutex
	dataChannels  map[string]*webrtcDataChannel
	channelsMu    sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

type webrtcStream struct {
	id        uint32
	conn      *webrtcConnection
	direction string
	closed    int32
	config    *StreamConfig
	buffer    *readBuffer
	ctx       context.Context
	cancel    context.CancelFunc
}

type webrtcDataChannel struct {
	label  string
	stream *webrtcStream
}

type readBuffer struct {
	data []byte
	mu   sync.Mutex
	cond *sync.Cond
}

func newReadBuffer() *readBuffer {
	rb := &readBuffer{
		data: make([]byte, 0),
	}
	rb.cond = sync.NewCond(&rb.mu)
	return rb
}

func (rb *readBuffer) Write(p []byte) (n int, err error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.data = append(rb.data, p...)
	rb.cond.Signal()
	return len(p), nil
}

func (rb *readBuffer) Read(p []byte) (n int, err error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	for len(rb.data) == 0 {
		rb.cond.Wait()
	}

	n = copy(p, rb.data)
	rb.data = rb.data[n:]
	return n, nil
}

func (w *webrtcConnection) OpenStream(config *StreamConfig) (Stream, error) {
	if atomic.LoadInt32(&w.closed) == 1 {
		return nil, ErrConnectionClosed
	}

	if config == nil {
		config = DefaultStreamConfig()
	}

	streamID := generateStreamID()

	stream := &webrtcStream{
		id:        streamID,
		conn:      w,
		direction: "outbound",
		config:    config,
		buffer:    newReadBuffer(),
	}

	w.streamsMu.Lock()
	w.streams[streamID] = stream
	w.streamsMu.Unlock()

	return stream, nil
}

func (w *webrtcConnection) AcceptStream() (Stream, error) {
	if atomic.LoadInt32(&w.closed) == 1 {
		return nil, ErrConnectionClosed
	}

	return nil, errors.New("AcceptStream not implemented for WebRTC")
}

func (w *webrtcConnection) Close() error {
	if !atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		return nil
	}

	w.state = StateClosing

	if w.cancel != nil {
		w.cancel()
	}

	w.streamsMu.RLock()
	for _, stream := range w.streams {
		stream.Close()
	}
	w.streamsMu.RUnlock()

	w.channelsMu.RLock()
	for _, channel := range w.dataChannels {
		if channel.stream != nil {
			channel.stream.Close()
		}
	}
	w.channelsMu.RUnlock()

	w.state = StateDisconnected
	return nil
}

func (w *webrtcConnection) GetInfo() ConnectionInfo {
	stats := w.GetStats()
	return ConnectionInfo{
		ID:            w.id,
		PeerID:        w.peerID,
		Type:          w.connType,
		LocalAddr:     &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 0},
		RemoteAddr:    &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 0},
		State:         w.state,
		EstablishedAt: w.establishedAt,
		Stats:         *stats,
	}
}

func (w *webrtcConnection) SendControlMessage(msg []byte) error {
	if atomic.LoadInt32(&w.closed) == 1 {
		return ErrConnectionClosed
	}

	channel := w.getControlChannel()
	if channel == nil {
		return errors.New("no control channel available")
	}

	_, err := channel.stream.buffer.Write(msg)
	if err != nil {
		return err
	}

	w.updateStats(uint64(len(msg)), 0)
	return nil
}

func (w *webrtcConnection) ReceiveControlMessage() ([]byte, error) {
	if atomic.LoadInt32(&w.closed) == 1 {
		return nil, ErrConnectionClosed
	}

	channel := w.getControlChannel()
	if channel == nil {
		return nil, errors.New("no control channel available")
	}

	buffer := make([]byte, 1024)
	n, err := channel.stream.buffer.Read(buffer)
	if err != nil {
		return nil, err
	}

	w.updateStats(0, uint64(n))
	return buffer[:n], nil
}

func (w *webrtcConnection) GetStats() *ConnectionStats {
	w.statsMu.RLock()
	defer w.statsMu.RUnlock()
	return &w.stats
}

func (w *webrtcConnection) updateStats(bytesSent, bytesReceived uint64) {
	w.statsMu.Lock()
	defer w.statsMu.Unlock()
	w.stats.BytesSent += bytesSent
	w.stats.BytesReceived += bytesReceived
}

func (w *webrtcConnection) getControlChannel() *webrtcDataChannel {
	w.channelsMu.RLock()
	defer w.channelsMu.RUnlock()

	if channel, ok := w.dataChannels["control"]; ok {
		return channel
	}
	return nil
}

func (w *webrtcConnection) removeStream(streamID uint32) {
	w.streamsMu.Lock()
	defer w.streamsMu.Unlock()
	delete(w.streams, streamID)
}

func (ws *webrtcStream) Read(p []byte) (n int, err error) {
	if atomic.LoadInt32(&ws.closed) == 1 {
		return 0, io.EOF
	}

	if ws.buffer == nil {
		return 0, errors.New("stream not initialized")
	}

	n, err = ws.buffer.Read(p)
	if n > 0 {
		ws.conn.updateStats(0, uint64(n))
	}
	return n, err
}

func (ws *webrtcStream) Write(p []byte) (n int, err error) {
	if atomic.LoadInt32(&ws.closed) == 1 {
		return 0, io.EOF
	}

	if ws.buffer == nil {
		return 0, errors.New("stream not initialized")
	}

	n, err = ws.buffer.Write(p)
	if n > 0 {
		ws.conn.updateStats(uint64(n), 0)
	}
	return n, err
}

func (ws *webrtcStream) Close() error {
	if !atomic.CompareAndSwapInt32(&ws.closed, 0, 1) {
		return nil
	}

	if ws.cancel != nil {
		ws.cancel()
	}

	ws.conn.removeStream(ws.id)
	return nil
}

func (ws *webrtcStream) GetStreamID() uint32 {
	return ws.id
}

func (ws *webrtcStream) SetDeadline(t time.Time) error {
	return nil
}

func (ws *webrtcStream) SetReadDeadline(t time.Time) error {
	return nil
}

func (ws *webrtcStream) SetWriteDeadline(t time.Time) error {
	return nil
}

func (ws *webrtcStream) GetInfo() StreamInfo {
	state := "open"
	if atomic.LoadInt32(&ws.closed) == 1 {
		state = "closed"
	}

	return StreamInfo{
		ID:        ws.id,
		Direction: ws.direction,
		State:     state,
	}
}
