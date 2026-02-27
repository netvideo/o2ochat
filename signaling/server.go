package signaling

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsServer struct {
	config      *ServerConfig
	httpServer  *http.Server
	listener    net.Listener
	peers       sync.Map
	running     bool
	mu          sync.RWMutex
	broadcast   chan *SignalingMessage
	connections map[string]*websocket.Conn
	connMu      sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewWebSocketServer(config *ServerConfig) SignalingServer {
	if config == nil {
		config = DefaultServerConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &wsServer{
		config:      config,
		running:     false,
		broadcast:   make(chan *SignalingMessage, 100),
		connections: make(map[string]*websocket.Conn),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *wsServer) Start(addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrServerAlreadyRunning
	}

	if addr == "" {
		addr = fmt.Sprintf(":%d", s.config.Port)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  s.config.MessageTimeout,
		WriteTimeout: s.config.MessageTimeout,
	}

	var err error
	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		s.httpServer.TLSConfig = &tls.Config{}
		s.listener, err = tls.Listen("tcp", addr, s.httpServer.TLSConfig)
		if err != nil {
			return fmt.Errorf("failed to start TLS listener: %w", err)
		}
		go s.httpServer.Serve(s.listener)
	} else {
		s.listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to start listener: %w", err)
		}
		go s.httpServer.Serve(s.listener)
	}

	s.running = true
	s.wg.Add(1)
	go s.broadcastLoop()

	return nil
}

func (s *wsServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return ErrServerNotRunning
	}

	s.cancel()
	s.wg.Wait()

	s.connMu.RLock()
	for _, conn := range s.connections {
		conn.Close()
	}
	s.connMu.RUnlock()

	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
	}

	s.running = false
	return nil
}

func (s *wsServer) Broadcast(msg *SignalingMessage) error {
	if !s.running {
		return ErrServerNotRunning
	}

	select {
	case s.broadcast <- msg:
		return nil
	case <-time.After(s.config.MessageTimeout):
		return ErrMessageTimeout
	}
}

func (s *wsServer) GetOnlinePeers() ([]*PeerInfo, error) {
	if !s.running {
		return nil, ErrServerNotRunning
	}

	var peers []*PeerInfo
	s.peers.Range(func(key, value interface{}) bool {
		if peerInfo, ok := value.(*PeerInfo); ok {
			peers = append(peers, peerInfo)
		}
		return true
	})

	return peers, nil
}

func (s *wsServer) GetConfig() *ServerConfig {
	return s.config
}

func (s *wsServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *wsServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !s.IsRunning() {
		http.Error(w, "server not running", http.StatusServiceUnavailable)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	peerID := r.URL.Query().Get("peer_id")
	if peerID == "" {
		peerID = generatePeerID()
	}

	s.connMu.Lock()
	s.connections[peerID] = conn
	s.connMu.Unlock()

	s.wg.Add(1)
	go s.handleConnection(peerID, conn)
}

func (s *wsServer) handleConnection(peerID string, conn *websocket.Conn) {
	defer func() {
		s.connMu.Lock()
		delete(s.connections, peerID)
		s.connMu.Unlock()
		s.peers.Delete(peerID)
		conn.Close()
		s.wg.Done()
	}()

	go s.sendHeartbeat(peerID, conn)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			conn.SetReadDeadline(time.Now().Add(s.config.MessageTimeout))
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				}
				return
			}

			var msg SignalingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				s.sendError(peerID, conn, ErrInvalidMessage)
				continue
			}

			s.processMessage(peerID, &msg, conn)
		}
	}
}

func (s *wsServer) processMessage(peerID string, msg *SignalingMessage, conn *websocket.Conn) {
	switch msg.Type {
	case MessageTypeRegister:
		var peerInfo PeerInfo
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if peerIDStr, ok := data["peer_id"].(string); ok {
				peerInfo.PeerID = peerIDStr
			}
			if ipv6, ok := data["ipv6_addrs"].([]interface{}); ok {
				for _, v := range ipv6 {
					if addr, ok := v.(string); ok {
						peerInfo.IPv6Addrs = append(peerInfo.IPv6Addrs, addr)
					}
				}
			}
			if ipv4, ok := data["ipv4_addrs"].([]interface{}); ok {
				for _, v := range ipv4 {
					if addr, ok := v.(string); ok {
						peerInfo.IPv4Addrs = append(peerInfo.IPv4Addrs, addr)
					}
				}
			}
		}
		peerInfo.Online = true
		peerInfo.LastSeen = time.Now()
		s.peers.Store(peerID, &peerInfo)

	case MessageTypePing:
		if peerInfo, ok := s.peers.Load(peerID); ok {
			p := peerInfo.(*PeerInfo)
			p.LastSeen = time.Now()
			s.peers.Store(peerID, p)
		}

		pongMsg := &SignalingMessage{
			Type:      MessageTypePong,
			Timestamp: time.Now(),
		}
		s.sendMessage(peerID, conn, pongMsg)

	case MessageTypeOffer, MessageTypeAnswer, MessageTypeCandidate:
		s.routeMessage(msg, conn)

	case MessageTypeLookup:
		targetPeerID, ok := msg.Data.(string)
		if !ok {
			s.sendError(peerID, conn, ErrInvalidPeerID)
			return
		}

		if targetPeer, ok := s.peers.Load(targetPeerID); ok {
			response := &SignalingMessage{
				Type:      MessageTypeLookup,
				To:        peerID,
				Data:      targetPeer.(*PeerInfo),
				Timestamp: time.Now(),
			}
			s.sendMessage(peerID, conn, response)
		} else {
			s.sendError(peerID, conn, ErrPeerNotFound)
		}
	}
}

func (s *wsServer) routeMessage(msg *SignalingMessage, conn *websocket.Conn) {
	targetConn := s.getConnection(msg.To)
	if targetConn == nil {
		s.sendError(msg.From, conn, ErrPeerNotFound)
		return
	}

	if err := s.sendMessage(msg.To, targetConn, msg); err != nil {
		s.sendError(msg.From, conn, ErrMessageSendFailed)
	}
}

func (s *wsServer) getConnection(peerID string) *websocket.Conn {
	s.connMu.RLock()
	defer s.connMu.RUnlock()
	return s.connections[peerID]
}

func (s *wsServer) sendMessage(peerID string, conn *websocket.Conn, msg *SignalingMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	conn.SetWriteDeadline(time.Now().Add(s.config.MessageTimeout))
	return conn.WriteMessage(websocket.TextMessage, data)
}

func (s *wsServer) sendError(peerID string, conn *websocket.Conn, err error) {
	errorMsg := &SignalingMessage{
		Type:      "error",
		Data:      err.Error(),
		Timestamp: time.Now(),
	}
	s.sendMessage(peerID, conn, errorMsg)
}

func (s *wsServer) sendHeartbeat(peerID string, conn *websocket.Conn) {
	ticker := time.NewTicker(s.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			pingMsg := &SignalingMessage{
				Type:      MessageTypePing,
				Timestamp: time.Now(),
			}
			if err := s.sendMessage(peerID, conn, pingMsg); err != nil {
				return
			}
		}
	}
}

func (s *wsServer) broadcastLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.broadcast:
			s.connMu.RLock()
			for peerID, conn := range s.connections {
				if msg.To == "" || msg.To == peerID {
					go s.sendMessage(peerID, conn, msg)
				}
			}
			s.connMu.RUnlock()
		}
	}
}

func (s *wsServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func generatePeerID() string {
	return fmt.Sprintf("peer_%d", time.Now().UnixNano())
}
