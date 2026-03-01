package webrtc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

// CallType represents the type of call
type CallType string

const (
	CallTypeAudio CallType = "audio"
	CallTypeVideo CallType = "video"
)

// CallState represents the state of a call
type CallState string

const (
	CallStateNew        CallState = "new"
	CallStateConnecting CallState = "connecting"
	CallStateActive     CallState = "active"
	CallStateEnded      CallState = "ended"
)

// Call represents a WebRTC call
type Call struct {
	ID            string
	PeerID        string
	LocalPeerID   string
	Type          CallType
	State         CallState
	PeerConnection *webrtc.PeerConnection
	DataChannel   *webrtc.DataChannel
	StartedAt     time.Time
	EndedAt       time.Time
	mu            sync.RWMutex
}

// CallManager manages WebRTC calls
type CallManager struct {
	config         *webrtc.Configuration
	calls          map[string]*Call
	callbacks      *CallCallbacks
	mu             sync.RWMutex
	stats          CallStats
	localStream    *webrtc.MediaStream
	isInitialized  bool
}

// CallCallbacks represents call event callbacks
type CallCallbacks struct {
	OnCallStarted   func(call *Call)
	OnCallEnded     func(call *Call)
	OnRemoteStream  func(call *Call, stream *webrtc.MediaStream)
	OnError         func(call *Call, err error)
}

// CallStats represents call statistics
type CallStats struct {
	TotalCalls    int
	ActiveCalls   int
	TotalDuration time.Duration
}

// CallConfig represents call configuration
type CallConfig struct {
	ICEServers []webrtc.ICEServer
	STUNServer string
	TURNServer string
	TURNUser   string
	TURNPass   string
}

// DefaultCallConfig returns default call configuration
func DefaultCallConfig() *CallConfig {
	return &CallConfig{
		STUNServer: "stun:stun.l.google.com:19302",
	}
}

// NewCallManager creates a new call manager
func NewCallManager(config *CallConfig) (*CallManager, error) {
	if config == nil {
		config = DefaultCallConfig()
	}

	// Create WebRTC configuration
	webrtcConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Add TURN server if configured
	if config.TURNServer != "" {
		webrtcConfig.ICEServers = append(webrtcConfig.ICEServers, webrtc.ICEServer{
			URLs:       []string{config.TURNServer},
			Username:   config.TURNUser,
			Credential: config.TURNPass,
		})
	}

	manager := &CallManager{
		config:    &webrtcConfig,
		calls:     make(map[string]*Call),
		callbacks: &CallCallbacks{},
	}

	return manager, nil
}

// Initialize initializes the call manager
func (cm *CallManager) Initialize() error {
	if cm.isInitialized {
		return nil
	}

	// Initialize WebRTC API
	api := webrtc.NewAPI()

	// Create peer connection settings
	settings := webrtc.SettingEngine{}

	// Configure NAT traversal
	settings.SetNAT1To1IPs([]string{}, webrtc.ICECandidateTypeSrflx)

	// Add settings to API
	api = webrtc.NewAPI(webrtc.WithSettingEngine(settings))

	cm.isInitialized = true

	return nil
}

// StartCall starts a new call
func (cm *CallManager) StartCall(ctx context.Context, peerID string, callType CallType) (*Call, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Create peer connection
	peerConnection, err := webrtc.NewPeerConnection(*cm.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	// Create call
	call := &Call{
		ID:             generateCallID(),
		PeerID:         peerID,
		LocalPeerID:    "local-peer",
		Type:           callType,
		State:          CallStateNew,
		PeerConnection: peerConnection,
		StartedAt:      time.Now(),
	}

	// Store call
	cm.calls[call.ID] = call
	cm.stats.TotalCalls++
	cm.stats.ActiveCalls++

	// Setup data channel
	_, err = peerConnection.CreateDataChannel("call", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create data channel: %w", err)
	}

	// Create offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	// Set local description
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		return nil, fmt.Errorf("failed to set local description: %w", err)
	}

	call.State = CallStateConnecting

	// Notify callback
	if cm.callbacks.OnCallStarted != nil {
		go cm.callbacks.OnCallStarted(call)
	}

	return call, nil
}

// AcceptCall accepts an incoming call
func (cm *CallManager) AcceptCall(ctx context.Context, callID string, offer webrtc.SessionDescription) (*Call, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	call, exists := cm.calls[callID]
	if !exists {
		return nil, errors.New("call not found")
	}

	// Set remote description
	err := call.PeerConnection.SetRemoteDescription(offer)
	if err != nil {
		return nil, fmt.Errorf("failed to set remote description: %w", err)
	}

	// Create answer
	answer, err := call.PeerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create answer: %w", err)
	}

	// Set local description
	err = call.PeerConnection.SetLocalDescription(answer)
	if err != nil {
		return nil, fmt.Errorf("failed to set local description: %w", err)
	}

	call.State = CallStateActive

	return call, nil
}

// EndCall ends a call
func (cm *CallManager) EndCall(callID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	call, exists := cm.calls[callID]
	if !exists {
		return errors.New("call not found")
	}

	// Close peer connection
	if call.PeerConnection != nil {
		call.PeerConnection.Close()
	}

	// Update call state
	call.State = CallStateEnded
	call.EndedAt = time.Now()

	// Update stats
	cm.stats.ActiveCalls--
	duration := call.EndedAt.Sub(call.StartedAt)
	cm.stats.TotalDuration += duration

	// Notify callback
	if cm.callbacks.OnCallEnded != nil {
		go cm.callbacks.OnCallEnded(call)
	}

	return nil
}

// GetCall gets a call by ID
func (cm *CallManager) GetCall(callID string) *Call {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.calls[callID]
}

// GetActiveCalls gets all active calls
func (cm *CallManager) GetActiveCalls() []*Call {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	activeCalls := make([]*Call, 0)
	for _, call := range cm.calls {
		if call.State == CallStateActive {
			activeCalls = append(activeCalls, call)
		}
	}

	return activeCalls
}

// GetStats gets call statistics
func (cm *CallManager) GetStats() CallStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.stats
}

// SetCallbacks sets call callbacks
func (cm *CallManager) SetCallbacks(callbacks *CallCallbacks) {
	cm.callbacks = callbacks
}

// generateCallID generates a unique call ID
func generateCallID() string {
	return fmt.Sprintf("call-%d", time.Now().UnixNano())
}

// CallToJSON converts call to JSON
func CallToJSON(call *Call) ([]byte, error) {
	call.mu.RLock()
	defer call.mu.RUnlock()

	data := map[string]interface{}{
		"id":           call.ID,
		"peer_id":      call.PeerID,
		"local_peer_id": call.LocalPeerID,
		"type":         call.Type,
		"state":        call.State,
		"started_at":   call.StartedAt,
		"ended_at":     call.EndedAt,
	}

	return json.Marshal(data)
}

// CallFromJSON converts JSON to call
func CallFromJSON(data []byte) (*Call, error) {
	var callData map[string]interface{}
	err := json.Unmarshal(data, &callData)
	if err != nil {
		return nil, err
	}

	call := &Call{
		ID:          callData["id"].(string),
		PeerID:      callData["peer_id"].(string),
		LocalPeerID: callData["local_peer_id"].(string),
		Type:        CallType(callData["type"].(string)),
		State:       CallState(callData["state"].(string)),
	}

	if startedAt, ok := callData["started_at"].(string); ok {
		call.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
	}

	if endedAt, ok := callData["ended_at"].(string); ok {
		call.EndedAt, _ = time.Parse(time.RFC3339, endedAt)
	}

	return call, nil
}
