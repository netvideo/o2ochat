package ui

import (
	"sync"
	"time"
)

type CallState string

const (
	CallStateIdle       CallState = "idle"
	CallStateRinging    CallState = "ringing"
	CallStateConnecting CallState = "connecting"
	CallStateConnected  CallState = "connected"
	CallStateEnded      CallState = "ended"
	CallStateFailed     CallState = "failed"
)

type CallComponent struct {
	mu              sync.RWMutex
	sessionID       string
	peerID          string
	peerName        string
	peerAvatar      string
	state           CallState
	hasVideo        bool
	isIncoming      bool
	startTime       time.Time
	duration        time.Duration
	
	isMuted         bool
	isVideoOff      bool
	isScreenSharing bool
	isHold          bool
	
	quality         float64
	networkStats    *NetworkStats
	
	onAccept        func()
	onReject       func()
	onEnd          func()
	onMuteToggle   func()
	onVideoToggle  func()
	onScreenShare  func()
	onHold         func()
}

func NewCallComponent() *CallComponent {
	return &CallComponent{
		state:       CallStateIdle,
		networkStats: &NetworkStats{},
	}
}

func (c *CallComponent) StartIncomingCall(peerID, peerName, peerAvatar string, hasVideo bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != CallStateIdle {
		return ErrCallNotFound
	}

	c.sessionID = generateCallID()
	c.peerID = peerID
	c.peerName = peerName
	c.peerAvatar = peerAvatar
	c.hasVideo = hasVideo
	c.isIncoming = true
	c.state = CallStateRinging

	return nil
}

func (c *CallComponent) StartOutgoingCall(peerID, peerName, peerAvatar string, hasVideo bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != CallStateIdle {
		return ErrCallNotFound
	}

	c.sessionID = generateCallID()
	c.peerID = peerID
	c.peerName = peerName
	c.peerAvatar = peerAvatar
	c.hasVideo = hasVideo
	c.isIncoming = false
	c.state = CallStateConnecting

	return nil
}

func (c *CallComponent) Accept() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != CallStateRinging && c.state != CallStateConnecting {
		return
	}

	c.state = CallStateConnected
	c.startTime = time.Now()

	if c.onAccept != nil {
		c.onAccept()
	}
}

func (c *CallComponent) Reject() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != CallStateRinging {
		return
	}

	c.state = CallStateEnded

	if c.onReject != nil {
		c.onReject()
	}
}

func (c *CallComponent) End() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state = CallStateEnded
	if c.startTime.IsZero() {
		c.duration = time.Since(c.startTime)
	}

	if c.onEnd != nil {
		c.onEnd()
	}
}

func (c *CallComponent) ToggleMute() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isMuted = !c.isMuted

	if c.onMuteToggle != nil {
		c.onMuteToggle()
	}
}

func (c *CallComponent) ToggleVideo() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isVideoOff = !c.isVideoOff

	if c.onVideoToggle != nil {
		c.onVideoToggle()
	}
}

func (c *CallComponent) ToggleScreenShare() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isScreenSharing = !c.isScreenSharing

	if c.onScreenShare != nil {
		c.onScreenShare()
	}
}

func (c *CallComponent) ToggleHold() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isHold = !c.isHold

	if c.onHold != nil {
		c.onHold()
	}
}

func (c *CallComponent) UpdateNetworkStats(stats *NetworkStats) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.networkStats = stats
}

func (c *CallComponent) UpdateQuality(quality float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.quality = quality
}

func (c *CallComponent) GetState() CallState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *CallComponent) GetSessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}

func (c *CallComponent) GetPeerID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.peerID
}

func (c *CallComponent) GetPeerName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.peerName
}

func (c *CallComponent) GetDuration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.startTime.IsZero() {
		return 0
	}
	return time.Since(c.startTime)
}

func (c *CallComponent) IsMuted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isMuted
}

func (c *CallComponent) IsVideoOff() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isVideoOff
}

func (c *CallComponent) IsScreenSharing() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isScreenSharing
}

func (c *CallComponent) IsOnHold() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isHold
}

func (c *CallComponent) HasVideo() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hasVideo
}

func (c *CallComponent) IsIncoming() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isIncoming
}

func (c *CallComponent) GetNetworkStats() *NetworkStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.networkStats
}

func (c *CallComponent) GetQuality() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.quality
}

func (c *CallComponent) SetOnAccept(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onAccept = callback
}

func (c *CallComponent) SetOnReject(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onReject = callback
}

func (c *CallComponent) SetOnEnd(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEnd = callback
}

func (c *CallComponent) SetOnMuteToggle(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onMuteToggle = callback
}

func (c *CallComponent) SetOnVideoToggle(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onVideoToggle = callback
}

func (c *CallComponent) SetOnScreenShare(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onScreenShare = callback
}

func (c *CallComponent) SetOnHold(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onHold = callback
}

func (c *CallComponent) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == CallStateRinging || c.state == CallStateConnecting || c.state == CallStateConnected
}

func generateCallID() string {
	return "call_" + time.Now().Format("20060102150405")
}

type CallManagerComponent struct {
	mu           sync.RWMutex
	activeCall   *CallComponent
	history      []*CallHistoryEntry
	onCallStart  func(peerID string, video bool)
	onCallEnd    func(peerID string, duration time.Duration)
}

type CallHistoryEntry struct {
	SessionID  string
	PeerID     string
	PeerName   string
	HasVideo   bool
	Direction  string
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Status     string
}

func NewCallManagerComponent() *CallManagerComponent {
	return &CallManagerComponent{
		history: make([]*CallHistoryEntry, 0),
	}
}

func (cm *CallManagerComponent) StartCall(peerID, peerName string, hasVideo bool) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.activeCall != nil && cm.activeCall.IsActive() {
		return ErrOperationFailed
	}

	cm.activeCall = NewCallComponent()
	return cm.activeCall.StartOutgoingCall(peerID, peerName, "", hasVideo)
}

func (cm *CallManagerComponent) AcceptCall(peerID, peerName string, hasVideo bool) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.activeCall == nil {
		return ErrCallNotFound
	}

	cm.activeCall.StartIncomingCall(peerID, peerName, "", hasVideo)
	cm.activeCall.Accept()
	return nil
}

func (cm *CallManagerComponent) EndCall() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.activeCall != nil {
		cm.activeCall.End()
		cm.addToHistory(cm.activeCall)
		cm.activeCall = nil
	}
}

func (cm *CallManagerComponent) GetActiveCall() *CallComponent {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.activeCall
}

func (cm *CallManagerComponent) GetCallHistory() []*CallHistoryEntry {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*CallHistoryEntry, len(cm.history))
	copy(result, cm.history)
	return result
}

func (cm *CallManagerComponent) ClearHistory() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.history = make([]*CallHistoryEntry, 0)
}

func (cm *CallManagerComponent) addToHistory(call *CallComponent) {
	entry := &CallHistoryEntry{
		SessionID: call.GetSessionID(),
		PeerID:    call.GetPeerID(),
		PeerName:  call.GetPeerName(),
		HasVideo:  call.HasVideo(),
		Direction: "outgoing",
		StartTime: call.startTime,
		EndTime:   time.Now(),
		Duration:  call.GetDuration(),
		Status:    "completed",
	}

	if call.IsIncoming() {
		entry.Direction = "incoming"
	}

	cm.history = append(cm.history, entry)
}

func (cm *CallManagerComponent) SetOnCallStart(callback func(peerID string, video bool)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onCallStart = callback
}

func (cm *CallManagerComponent) SetOnCallEnd(callback func(peerID string, duration time.Duration)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onCallEnd = callback
}
