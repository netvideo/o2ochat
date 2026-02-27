package transport

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	managerErrConnectionFailed   = ErrConnectionFailed
	managerErrConnectionNotFound = ErrPeerNotReachable
)

type transportManager struct {
	config      *TransportConfig
	connections sync.Map
	listeners   sync.Map
	mu          sync.RWMutex
	connHandler func(Connection)
	ctx         context.Context
	cancel      context.CancelFunc
}

type TransportConfig struct {
	QUICConfig     *QUICConfig
	WebRTCConfig   *WebRTCConfig
	MaxConnections int
	KeepAlive      bool
}

func NewTransportManager(config *TransportConfig) TransportManager {
	if config == nil {
		config = &TransportConfig{
			QUICConfig:     DefaultQUICConfig(),
			WebRTCConfig:   DefaultWebRTCConfig(),
			MaxConnections: 100,
			KeepAlive:      true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &transportManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *transportManager) Connect(config *ConnectionConfig) (Connection, error) {
	if config == nil {
		config = DefaultConnectionConfig()
	}

	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(t.ctx, config.Timeout)
	defer cancel()

	var lastErr error
	for i := 0; i <= config.RetryCount; i++ {
		conn, err := t.tryConnect(ctx, config)
		if err == nil {
			return conn, nil
		}
		lastErr = err

		if i < config.RetryCount {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, lastErr)
}

func (t *transportManager) tryConnect(ctx context.Context, config *ConnectionConfig) (Connection, error) {
	for _, connType := range config.Priority {
		switch connType {
		case ConnectionTypeQUIC:
			if conn, err := t.connectQUIC(ctx, config); err == nil {
				return conn, nil
			}
		case ConnectionTypeWebRTC:
			if conn, err := t.connectWebRTC(ctx, config); err == nil {
				return conn, nil
			}
		}
	}

	return nil, ErrConnectionFailed
}

func (t *transportManager) connectQUIC(ctx context.Context, config *ConnectionConfig) (Connection, error) {
	var dialAddr string

	for _, addr := range config.IPv6Addresses {
		dialAddr = normalizeAddress(addr, "udp")
		conn, err := t.dialQUIC(ctx, dialAddr, config)
		if err == nil {
			return conn, nil
		}
	}

	for _, addr := range config.IPv4Addresses {
		dialAddr = normalizeAddress(addr, "udp")
		conn, err := t.dialQUIC(ctx, dialAddr, config)
		if err == nil {
			return conn, nil
		}
	}

	return nil, ErrConnectionFailed
}

func (t *transportManager) dialQUIC(ctx context.Context, addr string, config *ConnectionConfig) (Connection, error) {
	dialer := &net.Dialer{
		Timeout: config.Timeout,
	}

	netConn, err := dialer.DialContext(ctx, "udp", addr)
	if err != nil {
		return nil, err
	}

	conn := &quicConnection{
		id:            generateConnectionID(addr, config.PeerID),
		peerID:        config.PeerID,
		connType:      ConnectionTypeQUIC,
		netConn:       netConn,
		localAddr:     netConn.LocalAddr(),
		remoteAddr:    netConn.RemoteAddr(),
		state:         StateConnected,
		establishedAt: time.Now(),
		streams:       make(map[uint32]*quicStream),
		config:        t.config.QUICConfig,
	}

	t.connections.Store(conn.id, conn)

	if t.connHandler != nil {
		t.connHandler(conn)
	}

	return conn, nil
}

func (t *transportManager) connectWebRTC(ctx context.Context, config *ConnectionConfig) (Connection, error) {
	conn := &webrtcConnection{
		id:            generateConnectionID("webrtc", config.PeerID),
		peerID:        config.PeerID,
		connType:      ConnectionTypeWebRTC,
		state:         StateConnected,
		establishedAt: time.Now(),
		streams:       make(map[uint32]*webrtcStream),
		config:        t.config.WebRTCConfig,
	}

	t.connections.Store(conn.id, conn)

	if t.connHandler != nil {
		t.connHandler(conn)
	}

	return conn, nil
}

func (t *transportManager) Accept() (Connection, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var listener net.Listener
	t.listeners.Range(func(key, value interface{}) bool {
		if l, ok := value.(net.Listener); ok {
			listener = l
			return false
		}
		return true
	})

	if listener == nil {
		return nil, ErrConnectionClosed
	}

	conn, err := listener.Accept()
	if err != nil {
		return nil, err
	}

	quicConn := &quicConnection{
		id:            generateConnectionID(conn.RemoteAddr().String(), ""),
		connType:      ConnectionTypeQUIC,
		netConn:       conn,
		localAddr:     conn.LocalAddr(),
		remoteAddr:    conn.RemoteAddr(),
		state:         StateConnected,
		establishedAt: time.Now(),
		streams:       make(map[uint32]*quicStream),
		config:        t.config.QUICConfig,
	}

	t.connections.Store(quicConn.id, quicConn)

	if t.connHandler != nil {
		t.connHandler(quicConn)
	}

	return quicConn, nil
}

func (t *transportManager) Close() error {
	t.cancel()

	var closeErrors []error
	t.connections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(Connection); ok {
			if err := conn.Close(); err != nil {
				closeErrors = append(closeErrors, err)
			}
		}
		t.connections.Delete(key)
		return true
	})

	t.listeners.Range(func(key, value interface{}) bool {
		if listener, ok := value.(net.Listener); ok {
			listener.Close()
		}
		t.listeners.Delete(key)
		return true
	})

	if len(closeErrors) > 0 {
		return fmt.Errorf("close errors: %v", closeErrors)
	}

	return nil
}

func (t *transportManager) GetConnections() ([]ConnectionInfo, error) {
	var connections []ConnectionInfo
	t.connections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(Connection); ok {
			connections = append(connections, conn.GetInfo())
		}
		return true
	})
	return connections, nil
}

func (t *transportManager) FindConnection(peerID string) (Connection, error) {
	var foundConn Connection
	t.connections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(Connection); ok {
			if conn.GetInfo().PeerID == peerID {
				foundConn = conn
				return false
			}
		}
		return true
	})

	if foundConn == nil {
		return nil, ErrConnectionNotFound
	}

	return foundConn, nil
}

func (t *transportManager) Listen(addr string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	t.listeners.Store(addr, listener)

	go t.acceptLoop(listener)
	return nil
}

func (t *transportManager) acceptLoop(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		go t.handleIncomingConn(conn)
	}
}

func (t *transportManager) handleIncomingConn(conn net.Conn) {
	quicConn := &quicConnection{
		id:            generateConnectionID(conn.RemoteAddr().String(), ""),
		connType:      ConnectionTypeQUIC,
		netConn:       conn,
		localAddr:     conn.LocalAddr(),
		remoteAddr:    conn.RemoteAddr(),
		state:         StateConnected,
		establishedAt: time.Now(),
		streams:       make(map[uint32]*quicStream),
		config:        t.config.QUICConfig,
	}

	t.connections.Store(quicConn.id, quicConn)

	if t.connHandler != nil {
		t.connHandler(quicConn)
	}
}

func (t *transportManager) GetNetworkType() NetworkType {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return NetworkTypeUnknown
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			if ipNet.IP.To4() == nil && !ipNet.IP.IsLoopback() {
				return NetworkTypeIPv6
			}
		}
	}

	return NetworkTypeIPv4
}

func (t *transportManager) SetConnectionHandler(handler func(Connection)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connHandler = handler
}

func normalizeAddress(addr string, network string) string {
	if _, _, err := net.SplitHostPort(addr); err != nil {
		if network == "udp" && addr[0] == '[' {
			return addr + ":8080"
		}
		return addr + ":8080"
	}
	return addr
}

func generateConnectionID(addr, peerID string) string {
	data := []byte(fmt.Sprintf("%s-%s-%d", addr, peerID, time.Now().UnixNano()))
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:16])
}

func generateStreamID() uint32 {
	b := make([]byte, 4)
	rand.Read(b)
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
