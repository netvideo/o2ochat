// Package signaling provides WebSocket signaling for P2P connection establishment
package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType represents the type of signaling message
type MessageType string

const (
	TypeOffer      MessageType = "offer"
	TypeAnswer     MessageType = "answer"
	TypeICE        MessageType = "ice"
	TypeJoin       MessageType = "join"
	TypeLeave      MessageType = "leave"
	TypePeerJoined MessageType = "peer_joined"
	TypePeerLeft   MessageType = "peer_left"
	TypeError      MessageType = "error"
	TypePing       MessageType = "ping"
	TypePong       MessageType = "pong"
)

// Message represents a signaling message
type Message struct {
	Type         MessageType     `json:"type"`
	RoomID       string          `json:"room_id,omitempty"`
	PeerID       string          `json:"peer_id,omitempty"`
	TargetPeerID string          `json:"target_peer_id,omitempty"`
	Payload      json.RawMessage `json:"payload,omitempty"`
	Timestamp    int64           `json:"timestamp"`
}

// SDPMessage represents SDP offer/answer payload
type SDPMessage struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"` // "offer" or "answer"
}

// ICEMessage represents ICE candidate payload
type ICEMessage struct {
	Candidate        string  `json:"candidate"`
	SDPMLineIndex    *uint16 `json:"sdpMLineIndex,omitempty"`
	SDPMid           *string `json:"sdpMid,omitempty"`
	UsernameFragment string  `json:"usernameFragment,omitempty"`
}

// RoomInfo represents information about a room
type RoomInfo struct {
	RoomID    string   `json:"room_id"`
	Peers     []string `json:"peers"`
	CreatedAt int64    `json:"created_at"`
}

// ClientConfig holds configuration for the signaling client
type ClientConfig struct {
	ServerURL         string
	PeerID            string
	ReconnectInterval time.Duration
	PingInterval      time.Duration
	MessageHandler    func(Message)
}

// DefaultClientConfig returns default client configuration
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		ReconnectInterval: 5 * time.Second,
		PingInterval:      30 * time.Second,
	}
}

// Client defines the interface for signaling clients
type Client interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool

	// Room management
	JoinRoom(roomID string) error
	LeaveRoom(roomID string) error
	GetRoomInfo(roomID string) (*RoomInfo, error)

	// Signaling
	SendOffer(targetPeerID string, offer interface{}) error
	SendAnswer(targetPeerID string, answer interface{}) error
	SendICE(targetPeerID string, candidate interface{}) error

	// Event handlers
	OnMessage(handler func(Message))
	OnConnect(handler func())
	OnDisconnect(handler func())
	OnError(handler func(error))
}

// client implements Client
type client struct {
	config    ClientConfig
	conn      *websocket.Conn
	connected bool
	mu        sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	// Handlers
	onMessage    func(Message)
	onConnect    func()
	onDisconnect func()
	onError      func(error)

	// Reconnection
	reconnectMu sync.Mutex
}

// Ensure client implements Client
var _ Client = (*client)(nil)

// NewClient creates a new signaling client
func NewClient(config ClientConfig) Client {
	return &client{
		config: config,
	}
}

// Connect establishes a connection to the signaling server
func (c *client) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.connected {
		c.mu.Unlock()
		return fmt.Errorf("already connected")
	}
	c.mu.Unlock()

	// Parse and validate URL
	u, err := url.Parse(c.config.ServerURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	if c.config.PeerID != "" {
		q.Set("peer_id", c.config.PeerID)
	}
	u.RawQuery = q.Encode()

	// Connect
	wsURL := u.String()
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to signaling server: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.mu.Unlock()

	// Start message reader
	go c.readMessages()

	// Start ping handler
	go c.pingHandler()

	if c.onConnect != nil {
		c.onConnect()
	}

	return nil
}

// Disconnect closes the connection to the signaling server
func (c *client) Disconnect() error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return nil
	}

	conn := c.conn
	cancel := c.cancel
	c.connected = false
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	if conn != nil {
		// Send close message
		msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "disconnecting")
		conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))

		// Close connection
		conn.Close()
	}

	if c.onDisconnect != nil {
		c.onDisconnect()
	}

	return nil
}

// IsConnected returns whether the client is connected
func (c *client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// JoinRoom joins a signaling room
func (c *client) JoinRoom(roomID string) error {
	msg := Message{
		Type:      TypeJoin,
		RoomID:    roomID,
		PeerID:    c.config.PeerID,
		Timestamp: time.Now().Unix(),
	}

	return c.sendMessage(msg)
}

// LeaveRoom leaves a signaling room
func (c *client) LeaveRoom(roomID string) error {
	msg := Message{
		Type:      TypeLeave,
		RoomID:    roomID,
		PeerID:    c.config.PeerID,
		Timestamp: time.Now().Unix(),
	}

	return c.sendMessage(msg)
}

// GetRoomInfo gets information about a room
func (c *client) GetRoomInfo(roomID string) (*RoomInfo, error) {
	// This would typically request info from the server
	// For now, return a placeholder
	return &RoomInfo{
		RoomID:    roomID,
		Peers:     []string{},
		CreatedAt: time.Now().Unix(),
	}, nil
}

// SendOffer sends an SDP offer to a peer
func (c *client) SendOffer(targetPeerID string, offer interface{}) error {
	payload, err := json.Marshal(offer)
	if err != nil {
		return fmt.Errorf("failed to marshal offer: %w", err)
	}

	msg := Message{
		Type:         TypeOffer,
		PeerID:       c.config.PeerID,
		TargetPeerID: targetPeerID,
		Payload:      payload,
		Timestamp:    time.Now().Unix(),
	}

	return c.sendMessage(msg)
}

// SendAnswer sends an SDP answer to a peer
func (c *client) SendAnswer(targetPeerID string, answer interface{}) error {
	payload, err := json.Marshal(answer)
	if err != nil {
		return fmt.Errorf("failed to marshal answer: %w", err)
	}

	msg := Message{
		Type:         TypeAnswer,
		PeerID:       c.config.PeerID,
		TargetPeerID: targetPeerID,
		Payload:      payload,
		Timestamp:    time.Now().Unix(),
	}

	return c.sendMessage(msg)
}

// SendICE sends an ICE candidate to a peer
func (c *client) SendICE(targetPeerID string, candidate interface{}) error {
	payload, err := json.Marshal(candidate)
	if err != nil {
		return fmt.Errorf("failed to marshal ICE candidate: %w", err)
	}

	msg := Message{
		Type:         TypeICE,
		PeerID:       c.config.PeerID,
		TargetPeerID: targetPeerID,
		Payload:      payload,
		Timestamp:    time.Now().Unix(),
	}

	return c.sendMessage(msg)
}

// OnMessage sets the message handler
func (c *client) OnMessage(handler func(Message)) {
	c.onMessage = handler
}

// OnConnect sets the connect handler
func (c *client) OnConnect(handler func()) {
	c.onConnect = handler
}

// OnDisconnect sets the disconnect handler
func (c *client) OnDisconnect(handler func()) {
	c.onDisconnect = handler
}

// OnError sets the error handler
func (c *client) OnError(handler func(error)) {
	c.onError = handler
}

// sendMessage sends a message to the server
func (c *client) sendMessage(msg Message) error {
	c.mu.RLock()
	conn := c.conn
	connected := c.connected
	c.mu.RUnlock()

	if !connected || conn == nil {
		return fmt.Errorf("not connected to signaling server")
	}

	if err := conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// readMessages reads messages from the connection
func (c *client) readMessages() {
	defer func() {
		c.Disconnect()
	}()

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				if c.onError != nil {
					c.onError(fmt.Errorf("websocket error: %w", err))
				}
			}
			return
		}

		// Handle ping/pong
		switch msg.Type {
		case TypePing:
			pong := Message{
				Type:      TypePong,
				Timestamp: time.Now().Unix(),
			}
			c.sendMessage(pong)
			continue
		case TypePong:
			continue
		}

		// Call message handler
		if c.onMessage != nil {
			c.onMessage(msg)
		}
	}
}

// pingHandler handles periodic ping messages
func (c *client) pingHandler() {
	ticker := time.NewTicker(c.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if !c.IsConnected() {
				continue
			}

			ping := Message{
				Type:      TypePing,
				PeerID:    c.config.PeerID,
				Timestamp: time.Now().Unix(),
			}

			if err := c.sendMessage(ping); err != nil {
				if c.onError != nil {
					c.onError(fmt.Errorf("failed to send ping: %w", err))
				}
			}
		}
	}
}
