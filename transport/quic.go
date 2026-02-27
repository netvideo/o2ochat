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

type quicConnection struct {
	id            string
	peerID        string
	connType      ConnectionType
	netConn       net.Conn
	localAddr     net.Addr
	remoteAddr    net.Addr
	state         ConnectionState
	establishedAt time.Time
	closed        int32
	streams       map[uint32]*quicStream
	streamsMu     sync.RWMutex
	config        *QUICConfig
	stats         ConnectionStats
	statsMu       sync.RWMutex
	controlChan   chan []byte
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

type quicStream struct {
	id        uint32
	conn      *quicConnection
	stream    net.Conn
	direction string
	closed    int32
	config    *StreamConfig
	ctx       context.Context
	cancel    context.CancelFunc
}

func (q *quicConnection) OpenStream(config *StreamConfig) (Stream, error) {
	if atomic.LoadInt32(&q.closed) == 1 {
		return nil, ErrConnectionClosed
	}

	if config == nil {
		config = DefaultStreamConfig()
	}

	streamID := generateStreamID()

	stream := &quicStream{
		id:        streamID,
		conn:      q,
		direction: "outbound",
		config:    config,
	}

	q.streamsMu.Lock()
	q.streams[streamID] = stream
	q.streamsMu.Unlock()

	return stream, nil
}

func (q *quicConnection) AcceptStream() (Stream, error) {
	if atomic.LoadInt32(&q.closed) == 1 {
		return nil, ErrConnectionClosed
	}

	return nil, errors.New("AcceptStream not implemented for QUIC")
}

func (q *quicConnection) Close() error {
	if !atomic.CompareAndSwapInt32(&q.closed, 0, 1) {
		return nil
	}

	q.state = StateClosing

	if q.cancel != nil {
		q.cancel()
	}

	q.streamsMu.RLock()
	for _, stream := range q.streams {
		stream.Close()
	}
	q.streamsMu.RUnlock()

	var closeErr error
	if q.netConn != nil {
		closeErr = q.netConn.Close()
	}

	q.state = StateDisconnected
	return closeErr
}

func (q *quicConnection) GetInfo() ConnectionInfo {
	stats := q.GetStats()
	return ConnectionInfo{
		ID:            q.id,
		PeerID:        q.peerID,
		Type:          q.connType,
		LocalAddr:     q.localAddr,
		RemoteAddr:    q.remoteAddr,
		State:         q.state,
		EstablishedAt: q.establishedAt,
		Stats:         *stats,
	}
}

func (q *quicConnection) SendControlMessage(msg []byte) error {
	if atomic.LoadInt32(&q.closed) == 1 {
		return ErrConnectionClosed
	}

	select {
	case q.controlChan <- msg:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("control message send timeout")
	}
}

func (q *quicConnection) ReceiveControlMessage() ([]byte, error) {
	if atomic.LoadInt32(&q.closed) == 1 {
		return nil, ErrConnectionClosed
	}

	select {
	case msg := <-q.controlChan:
		return msg, nil
	case <-time.After(5 * time.Second):
		return nil, errors.New("control message receive timeout")
	}
}

func (q *quicConnection) GetStats() *ConnectionStats {
	q.statsMu.RLock()
	defer q.statsMu.RUnlock()
	return &q.stats
}

func (q *quicConnection) updateStats(bytesSent, bytesReceived uint64) {
	q.statsMu.Lock()
	defer q.statsMu.Unlock()
	q.stats.BytesSent += bytesSent
	q.stats.BytesReceived += bytesReceived
}

func (q *quicConnection) startKeepAlive() {
	if q.config == nil || !q.config.KeepAlive {
		return
	}

	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-q.ctx.Done():
				return
			case <-ticker.C:
				if atomic.LoadInt32(&q.closed) == 1 {
					return
				}
			}
		}
	}()
}

func (q *quicConnection) removeStream(streamID uint32) {
	q.streamsMu.Lock()
	defer q.streamsMu.Unlock()
	delete(q.streams, streamID)
}

func (qs *quicStream) Read(p []byte) (n int, err error) {
	if atomic.LoadInt32(&qs.closed) == 1 {
		return 0, io.EOF
	}

	if qs.stream == nil {
		return 0, errors.New("stream not initialized")
	}

	n, err = qs.stream.Read(p)
	if n > 0 {
		qs.conn.updateStats(0, uint64(n))
	}
	return n, err
}

func (qs *quicStream) Write(p []byte) (n int, err error) {
	if atomic.LoadInt32(&qs.closed) == 1 {
		return 0, io.EOF
	}

	if qs.stream == nil {
		return 0, errors.New("stream not initialized")
	}

	n, err = qs.stream.Write(p)
	if n > 0 {
		qs.conn.updateStats(uint64(n), 0)
	}
	return n, err
}

func (qs *quicStream) Close() error {
	if !atomic.CompareAndSwapInt32(&qs.closed, 0, 1) {
		return nil
	}

	if qs.cancel != nil {
		qs.cancel()
	}

	if qs.stream != nil {
		qs.stream.Close()
	}

	qs.conn.removeStream(qs.id)
	return nil
}

func (qs *quicStream) GetStreamID() uint32 {
	return qs.id
}

func (qs *quicStream) SetDeadline(t time.Time) error {
	if qs.stream == nil {
		return errors.New("stream not initialized")
	}
	return qs.stream.SetDeadline(t)
}

func (qs *quicStream) SetReadDeadline(t time.Time) error {
	if qs.stream == nil {
		return errors.New("stream not initialized")
	}
	return qs.stream.SetReadDeadline(t)
}

func (qs *quicStream) SetWriteDeadline(t time.Time) error {
	if qs.stream == nil {
		return errors.New("stream not initialized")
	}
	return qs.stream.SetWriteDeadline(t)
}

func (qs *quicStream) GetInfo() StreamInfo {
	state := "open"
	if atomic.LoadInt32(&qs.closed) == 1 {
		state = "closed"
	}

	return StreamInfo{
		ID:        qs.id,
		Direction: qs.direction,
		State:     state,
	}
}
