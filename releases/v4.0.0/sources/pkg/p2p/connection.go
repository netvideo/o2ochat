// Package p2p provides WebRTC peer-to-peer connection management
package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
)

// ConnectionState represents the current state of a P2P connection
type ConnectionState int

const (
	StateNew ConnectionState = iota
	StateConnecting
	StateConnected
	StateDisconnected
	StateFailed
	StateClosed
)

func (s ConnectionState) String() string {
	switch s {
	case StateNew:
		return "new"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateDisconnected:
		return "disconnected"
	case StateFailed:
		return "failed"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// ICEConfig holds ICE server configuration
type ICEConfig struct {
	STUNServers  []string `json:"stun_servers"`
	TURNServers  []string `json:"turn_servers"`
	TURNUsername string   `json:"turn_username,omitempty"`
	TURNPassword string   `json:"turn_password,omitempty"`
}

// DefaultICEConfig returns default ICE configuration
func DefaultICEConfig() *ICEConfig {
	return &ICEConfig{
		STUNServers: []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
		},
		TURNServers: []string{},
	}
}

// ConnectionStats holds connection statistics
type ConnectionStats struct {
	BytesSent     uint64
	BytesReceived uint64
	PacketsSent   uint64
	PacketsLost   uint64
	RTT           time.Duration
	LocalAddr     string
	RemoteAddr    string
	ICEState      string
	UpdatedAt     time.Time
}

// PeerConnection defines the interface for P2P connections
type PeerConnection interface {
	// Connection management
	Connect(ctx context.Context, offer webrtc.SessionDescription) (*webrtc.SessionDescription, error)
	Accept(ctx context.Context, answer webrtc.SessionDescription) error
	Close() error
	State() ConnectionState

	// Data channel
	CreateDataChannel(label string) (*webrtc.DataChannel, error)
	OnDataChannel(func(*webrtc.DataChannel))

	// ICE handling
	AddICECandidate(candidate webrtc.ICECandidateInit) error
	OnICECandidate(func(*webrtc.ICECandidate))

	// Events
	OnConnectionStateChange(func(ConnectionState))
	OnDisconnect(func())

	// Stats
	GetStats() ConnectionStats
}

// connection implements PeerConnection
type connection struct {
	peerID    string
	pc        *webrtc.PeerConnection
	state     ConnectionState
	iceConfig *ICEConfig
	mu        sync.RWMutex

	// Callbacks
	onStateChange  func(ConnectionState)
	onDisconnect   func()
	onDataChannel  func(*webrtc.DataChannel)
	onICECandidate func(*webrtc.ICECandidate)

	// Stats
	stats   ConnectionStats
	statsMu sync.RWMutex
}

// Ensure connection implements PeerConnection
var _ PeerConnection = (*connection)(nil)

// ConnectionOptions holds configuration options for creating a connection
type ConnectionOptions struct {
	PeerID      string
	ICEConfig   *ICEConfig
	IsInitiator bool
}

// NewConnection creates a new P2P connection
func NewConnection(opts ConnectionOptions) (PeerConnection, error) {
	if opts.ICEConfig == nil {
		opts.ICEConfig = DefaultICEConfig()
	}

	// Build ICE servers
	iceServers := []webrtc.ICEServer{}
	for _, stun := range opts.ICEConfig.STUNServers {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs: []string{stun},
		})
	}
	for _, turn := range opts.ICEConfig.TURNServers {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs:       []string{turn},
			Username:   opts.ICEConfig.TURNUsername,
			Credential: opts.ICEConfig.TURNPassword,
		})
	}

	config := webrtc.Configuration{
		ICEServers:    iceServers,
		BundlePolicy:  webrtc.BundlePolicyMaxBundle,
		RTCPMuxPolicy: webrtc.RTCPMuxPolicyRequire,
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	conn := &connection{
		peerID:    opts.PeerID,
		pc:        pc,
		state:     StateNew,
		iceConfig: opts.ICEConfig,
	}

	// Set up event handlers
	pc.OnConnectionStateChange(conn.onConnectionStateChange)
	pc.OnDataChannel(conn.onDataChannelInternal)
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if conn.onICECandidate != nil && c != nil {
			conn.onICECandidate(c)
		}
	})

	return conn, nil
}

// Connect initiates a connection by creating an offer
func (c *connection) Connect(ctx context.Context, remoteOffer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	c.mu.Lock()
	c.state = StateConnecting
	c.mu.Unlock()

	// Set remote description (the offer)
	if err := c.pc.SetRemoteDescription(remoteOffer); err != nil {
		return nil, fmt.Errorf("failed to set remote description: %w", err)
	}

	// Create answer
	answer, err := c.pc.CreateAnswer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create answer: %w", err)
	}

	// Set local description
	if err := c.pc.SetLocalDescription(answer); err != nil {
		return nil, fmt.Errorf("failed to set local description: %w", err)
	}

	// Wait for ICE gathering to complete or timeout
	done := make(chan struct{})
	go func() {
		for {
			if c.pc.ICEGatheringState() == webrtc.ICEGatheringStateComplete {
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-done:
		return c.pc.LocalDescription(), nil
	case <-ctx.Done():
		return c.pc.LocalDescription(), ctx.Err()
	}
}

// Accept accepts an incoming connection answer
func (c *connection) Accept(ctx context.Context, answer webrtc.SessionDescription) error {
	c.mu.Lock()
	c.state = StateConnecting
	c.mu.Unlock()

	if err := c.pc.SetRemoteDescription(answer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	return nil
}

// Close closes the connection
func (c *connection) Close() error {
	c.mu.Lock()
	c.state = StateClosed
	c.mu.Unlock()

	if c.pc != nil {
		return c.pc.Close()
	}
	return nil
}

// State returns the current connection state
func (c *connection) State() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// CreateDataChannel creates a new data channel
func (c *connection) CreateDataChannel(label string) (*webrtc.DataChannel, error) {
	return c.pc.CreateDataChannel(label, nil)
}

// OnDataChannel sets the callback for incoming data channels
func (c *connection) OnDataChannel(handler func(*webrtc.DataChannel)) {
	c.onDataChannel = handler
}

func (c *connection) onDataChannelInternal(dc *webrtc.DataChannel) {
	if c.onDataChannel != nil {
		c.onDataChannel(dc)
	}
}

// AddICECandidate adds an ICE candidate
func (c *connection) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	return c.pc.AddICECandidate(candidate)
}

// OnICECandidate sets the callback for ICE candidates
func (c *connection) OnICECandidate(handler func(*webrtc.ICECandidate)) {
	c.onICECandidate = handler
}

// OnConnectionStateChange sets the callback for connection state changes
func (c *connection) OnConnectionStateChange(handler func(ConnectionState)) {
	c.onStateChange = handler
}

// OnDisconnect sets the callback for disconnection
func (c *connection) OnDisconnect(handler func()) {
	c.onDisconnect = handler
}

func (c *connection) onConnectionStateChange(state webrtc.PeerConnectionState) {
	c.mu.Lock()
	switch state {
	case webrtc.PeerConnectionStateConnecting:
		c.state = StateConnecting
	case webrtc.PeerConnectionStateConnected:
		c.state = StateConnected
	case webrtc.PeerConnectionStateDisconnected:
		c.state = StateDisconnected
	case webrtc.PeerConnectionStateFailed:
		c.state = StateFailed
	case webrtc.PeerConnectionStateClosed:
		c.state = StateClosed
	}
	c.mu.Unlock()

	if c.onStateChange != nil {
		c.onStateChange(c.state)
	}

	if c.state == StateDisconnected || c.state == StateFailed || c.state == StateClosed {
		if c.onDisconnect != nil {
			c.onDisconnect()
		}
	}
}

// GetStats returns connection statistics
func (c *connection) GetStats() ConnectionStats {
	c.statsMu.RLock()
	defer c.statsMu.RUnlock()
	return c.stats
}

// UpdateStats updates connection statistics
func (c *connection) updateStats() {
	// This would be called periodically to update stats from the peer connection
	// Implementation depends on webrtc package stats API
}

// MarshalJSON implements json.Marshaler for ConnectionState
func (s ConnectionState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements json.Unmarshaler for ConnectionState
func (s *ConnectionState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	switch str {
	case "new":
		*s = StateNew
	case "connecting":
		*s = StateConnecting
	case "connected":
		*s = StateConnected
	case "disconnected":
		*s = StateDisconnected
	case "failed":
		*s = StateFailed
	case "closed":
		*s = StateClosed
	default:
		return fmt.Errorf("unknown connection state: %s", str)
	}
	return nil
}

// Errors
type ConnectionError struct {
	Op  string
	Err error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection %s: %v", e.Op, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}
