package media

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrAlreadyExists = errors.New("already exists")
)

type MediaManagerImpl struct {
	mu             sync.RWMutex
	initialized    bool
	deviceList     map[MediaType][]*DeviceInfo
	sessionManager *SessionManager
	config         *MediaManagerConfig
	state          ManagerState
	startTime      time.Time
}

type ManagerState int

const (
	ManagerStateUninitialized ManagerState = iota
	ManagerStateInitialized
	ManagerStateRunning
	ManagerStateStopped
)

func (s ManagerState) String() string {
	switch s {
	case ManagerStateUninitialized:
		return "uninitialized"
	case ManagerStateInitialized:
		return "initialized"
	case ManagerStateRunning:
		return "running"
	case ManagerStateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

type MediaManagerConfig struct {
	EnableAudio           bool
	EnableVideo           bool
	AudioDeviceID         string
	VideoDeviceID         string
	DefaultAudioBitrate   int
	DefaultVideoBitrate   int
	MaxConcurrentSessions int
}

func DefaultMediaManagerConfig() *MediaManagerConfig {
	return &MediaManagerConfig{
		EnableAudio:           true,
		EnableVideo:           true,
		DefaultAudioBitrate:   64000,
		DefaultVideoBitrate:   500000,
		MaxConcurrentSessions: 10,
	}
}

func NewMediaManager() (MediaManager, error) {
	return NewMediaManagerWithConfig(DefaultMediaManagerConfig())
}

func (m *MediaManagerImpl) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	m.deviceList = m.enumerateDevices()

	m.sessionManager = NewSessionManager()

	m.initialized = true
	m.state = ManagerStateInitialized
	m.startTime = time.Now()

	return nil
}

func (m *MediaManagerImpl) enumerateDevices() map[MediaType][]*DeviceInfo {
	devices := make(map[MediaType][]*DeviceInfo)

	audioDevices := []*DeviceInfo{
		{
			ID:      "default-audio-input",
			Name:    "Default Microphone",
			Type:    MediaTypeAudio,
			Default: true,
		},
		{
			ID:      "default-audio-output",
			Name:    "Default Speaker",
			Type:    MediaTypeAudio,
			Default: true,
		},
	}
	devices[MediaTypeAudio] = audioDevices

	videoDevices := []*DeviceInfo{
		{
			ID:      "default-video",
			Name:    "Default Camera",
			Type:    MediaTypeVideo,
			Default: true,
		},
	}
	devices[MediaTypeVideo] = videoDevices

	return devices
}

func (m *MediaManagerImpl) GetDevices(mediaType MediaType) ([]*DeviceInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, ErrNotInitialized
	}

	devices, exists := m.deviceList[mediaType]
	if !exists {
		return []*DeviceInfo{}, nil
	}

	return devices, nil
}

func (m *MediaManagerImpl) CreateCallSession(config *CallConfig) (CallSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, ErrNotInitialized
	}

	if config == nil {
		config = DefaultCallConfig()
	}

	sessionID := generateSessionID()

	peerInfo := &PeerInfo{
		PeerID:     sessionID,
		PublicKey:  nil,
		AudioMuted: false,
		VideoMuted: false,
	}

	session, err := m.sessionManager.CreateSession(sessionID, config, peerInfo)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (m *MediaManagerImpl) JoinCall(sessionID string) (CallSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, ErrNotInitialized
	}

	session, err := m.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (m *MediaManagerImpl) LeaveCall(sessionID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	return m.sessionManager.RemoveSession(sessionID)
}

func (m *MediaManagerImpl) GetCallStats(sessionID string) (*CallStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, ErrNotInitialized
	}

	session, err := m.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session.GetStats(), nil
}

func (m *MediaManagerImpl) Destroy() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil
	}

	m.sessionManager.CloseAll()

	m.initialized = false
	m.state = ManagerStateStopped

	return nil
}

func (m *MediaManagerImpl) GetState() ManagerState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.state
}

func (m *MediaManagerImpl) GetUptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.startTime.IsZero() {
		return 0
	}

	return time.Since(m.startTime)
}

func (m *MediaManagerImpl) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.sessionManager.ListSessions())
}

func (m *MediaManagerImpl) SetConfig(config *MediaManagerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		return errors.New("invalid config")
	}

	m.config = config
	return nil
}

func (m *MediaManagerImpl) GetConfig() *MediaManagerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config
}

func (m *MediaManagerImpl) RefreshDevices() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	m.deviceList = m.enumerateDevices()

	return nil
}

func generateSessionID() string {
	return "session-" + time.Now().Format("20060102150405")
}

func NewMediaManagerWithConfig(config *MediaManagerConfig) (MediaManager, error) {
	manager := &MediaManagerImpl{
		initialized:    false,
		deviceList:     make(map[MediaType][]*DeviceInfo),
		sessionManager: NewSessionManager(),
		config:         config,
		state:          ManagerStateUninitialized,
	}

	if err := manager.Initialize(); err != nil {
		return nil, err
	}

	return manager, nil
}

type MediaManagerFactory struct{}

func NewMediaManagerFactory() *MediaManagerFactory {
	return &MediaManagerFactory{}
}

func (f *MediaManagerFactory) CreateDefault() (MediaManager, error) {
	return NewMediaManager()
}

func (f *MediaManagerFactory) CreateWithConfig(config *MediaManagerConfig) (MediaManager, error) {
	return NewMediaManagerWithConfig(config)
}
