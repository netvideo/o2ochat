package media

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrSessionNotStarted = errors.New("session not started")
	ErrSessionClosed     = errors.New("session closed")
	ErrInvalidConfig     = errors.New("invalid config")
	ErrNotFound          = errors.New("not found")
)

type CallSessionImpl struct {
	mu               sync.RWMutex
	sessionID        string
	config           *CallConfig
	peerInfo         *PeerInfo
	audioCodec       Codec
	videoCodec       Codec
	audioProcessor   AudioProcessor
	rtpAudio         RTPProcessor
	rtpVideo         RTPProcessor
	jitterAudio      JitterBuffer
	jitterVideo      JitterBuffer
	state            SessionState
	localAudioMuted  bool
	localVideoMuted  bool
	remoteAudioMuted bool
	remoteVideoMuted bool
	startedAt        time.Time
	stats            *CallStats
	sendFrameChan    chan *MediaFrame
	recvFrameChan    chan *MediaFrame
	closeChan        chan struct{}
	wg               sync.WaitGroup
}

type SessionState int

const (
	SessionStateIdle SessionState = iota
	SessionStateConnecting
	SessionStateConnected
	SessionStateDisconnected
	SessionStateClosed
)

func (s SessionState) String() string {
	switch s {
	case SessionStateIdle:
		return "idle"
	case SessionStateConnecting:
		return "connecting"
	case SessionStateConnected:
		return "connected"
	case SessionStateDisconnected:
		return "disconnected"
	case SessionStateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

func NewCallSessionImpl(sessionID string, config *CallConfig, peerInfo *PeerInfo) (*CallSessionImpl, error) {
	if sessionID == "" {
		return nil, ErrInvalidConfig
	}

	if config == nil {
		config = DefaultCallConfig()
	}

	session := &CallSessionImpl{
		sessionID:        sessionID,
		config:           config,
		peerInfo:         peerInfo,
		state:            SessionStateIdle,
		localAudioMuted:  false,
		localVideoMuted:  false,
		remoteAudioMuted: false,
		remoteVideoMuted: false,
		sendFrameChan:    make(chan *MediaFrame, 100),
		recvFrameChan:    make(chan *MediaFrame, 100),
		closeChan:        make(chan struct{}),
		stats:            &CallStats{},
	}

	if err := session.initCodecs(); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *CallSessionImpl) initCodecs() error {
	if s.config.AudioConfig != nil && s.config.AudioConfig.Enabled {
		audioCodec, err := NewCodec(MediaTypeAudio, s.config.AudioConfig)
		if err != nil {
			return err
		}
		s.audioCodec = audioCodec

		processor, err := NewAudioProcessor(s.config.AudioConfig.SampleRate, s.config.AudioConfig.Channels)
		if err != nil {
			return err
		}
		s.audioProcessor = processor

		ssrc := generateSSRC()
		s.rtpAudio, _ = NewRTPProcessor(ssrc, 96, s.config.AudioConfig.SampleRate)
		s.jitterAudio, _ = NewJitterBufferImpl(150 * time.Millisecond)
	}

	if s.config.VideoConfig != nil && s.config.VideoConfig.Enabled {
		videoCodec, err := NewCodec(MediaTypeVideo, s.config.VideoConfig)
		if err != nil {
			return err
		}
		s.videoCodec = videoCodec

		ssrc := generateSSRC()
		s.rtpVideo, _ = NewRTPProcessor(ssrc, 100, 90000)
		s.jitterVideo, _ = NewJitterBufferImpl(200 * time.Millisecond)
	}

	return nil
}

func (s *CallSessionImpl) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == SessionStateConnected || s.state == SessionStateConnecting {
		return nil
	}

	if s.state == SessionStateClosed {
		return ErrSessionClosed
	}

	s.state = SessionStateConnecting
	s.startedAt = time.Now()

	if s.config.AudioConfig != nil && s.config.AudioConfig.Enabled {
		if s.rtpAudio != nil {
			s.rtpAudio.SetMaxPacketSize(1200)
		}
	}

	if s.config.VideoConfig != nil && s.config.VideoConfig.Enabled {
		if s.rtpVideo != nil {
			s.rtpVideo.SetMaxPacketSize(1400)
		}
	}

	s.state = SessionStateConnected

	return nil
}

func (s *CallSessionImpl) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == SessionStateClosed {
		return ErrSessionClosed
	}

	s.state = SessionStateDisconnected

	close(s.closeChan)

	if s.audioCodec != nil {
		s.audioCodec.Close()
	}

	if s.videoCodec != nil {
		s.videoCodec.Close()
	}

	if s.audioProcessor != nil {
		s.audioProcessor.Close()
	}

	if s.rtpAudio != nil {
		s.rtpAudio.Close()
	}

	if s.rtpVideo != nil {
		s.rtpVideo.Close()
	}

	if s.jitterAudio != nil {
		s.jitterAudio.Close()
	}

	if s.jitterVideo != nil {
		s.jitterVideo.Close()
	}

	return nil
}

func (s *CallSessionImpl) Pause(mediaType MediaType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != SessionStateConnected {
		return ErrSessionNotStarted
	}

	switch mediaType {
	case MediaTypeAudio:
		s.localAudioMuted = true
	case MediaTypeVideo:
		s.localVideoMuted = true
	}

	return nil
}

func (s *CallSessionImpl) Resume(mediaType MediaType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != SessionStateConnected {
		return ErrSessionNotStarted
	}

	switch mediaType {
	case MediaTypeAudio:
		s.localAudioMuted = false
	case MediaTypeVideo:
		s.localVideoMuted = false
	}

	return nil
}

func (s *CallSessionImpl) SwitchDevice(mediaType MediaType, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != SessionStateConnected {
		return ErrSessionNotStarted
	}

	return nil
}

func (s *CallSessionImpl) AdjustBitrate(targetBitrate int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil {
		return ErrInvalidConfig
	}

	if targetBitrate < s.config.MinBitrate {
		targetBitrate = s.config.MinBitrate
	}

	if targetBitrate > s.config.MaxBitrate {
		targetBitrate = s.config.MaxBitrate
	}

	remainingBitrate := targetBitrate

	if s.config.AudioConfig != nil && s.config.AudioConfig.Enabled {
		audioBitrate := 64000
		if audioBitrate > remainingBitrate/2 {
			audioBitrate = remainingBitrate / 2
		}
		s.config.AudioConfig.Bitrate = audioBitrate
		remainingBitrate -= audioBitrate
	}

	if s.config.VideoConfig != nil && s.config.VideoConfig.Enabled && remainingBitrate > 0 {
		s.config.VideoConfig.Bitrate = remainingBitrate
	}

	return nil
}

func (s *CallSessionImpl) SendFrame(frame *MediaFrame) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.state != SessionStateConnected {
		return ErrSessionNotStarted
	}

	if frame == nil || len(frame.Payload) == 0 {
		return nil
	}

	if frame.Type == MediaTypeAudio && s.localAudioMuted {
		return nil
	}

	if frame.Type == MediaTypeVideo && s.localVideoMuted {
		return nil
	}

	select {
	case s.sendFrameChan <- frame:
		return nil
	default:
		return errors.New("send buffer full")
	}
}

func (s *CallSessionImpl) ReceiveFrame() (*MediaFrame, error) {
	select {
	case frame := <-s.recvFrameChan:
		return frame, nil
	case <-s.closeChan:
		return nil, ErrSessionClosed
	}
}

func (s *CallSessionImpl) GetSessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sessionID
}

func (s *CallSessionImpl) GetRemoteInfo() *PeerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.peerInfo
}

func (s *CallSessionImpl) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == SessionStateClosed {
		return nil
	}

	s.state = SessionStateClosed

	close(s.closeChan)
	close(s.sendFrameChan)
	close(s.recvFrameChan)

	s.wg.Wait()

	return nil
}

func (s *CallSessionImpl) GetState() SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.state
}

func (s *CallSessionImpl) IsAudioMuted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.localAudioMuted
}

func (s *CallSessionImpl) IsVideoMuted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.localVideoMuted
}

func (s *CallSessionImpl) GetStats() *CallStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &CallStats{
		AudioStats:   &StreamStats{},
		VideoStats:   &StreamStats{},
		NetworkStats: &NetworkStats{},
	}

	if s.rtpAudio != nil {
		rtpStats := s.rtpAudio.GetRTPStats()
		stats.AudioStats.Bitrate = int(rtpStats.PacketsSent * 160 * 8 / 1000)
		stats.AudioStats.FramesSent = int64(rtpStats.PacketsSent)
		stats.AudioStats.FramesReceived = int64(rtpStats.PacketsReceived)
		stats.NetworkStats.PacketLoss = float64(rtpStats.PacketsLost) / float64(rtpStats.PacketsSent+1)
		stats.NetworkStats.Jitter = time.Duration(rtpStats.Jitter) * time.Microsecond
	}

	if s.rtpVideo != nil {
		rtpStats := s.rtpVideo.GetRTPStats()
		stats.VideoStats.Bitrate = int(rtpStats.PacketsSent * 1400 * 8 / 1000)
		stats.VideoStats.FramesSent = int64(rtpStats.PacketsSent)
		stats.VideoStats.FramesReceived = int64(rtpStats.PacketsReceived)
	}

	if s.jitterAudio != nil {
		bufStatus := s.jitterAudio.GetBufferStatus()
		stats.AudioStats.Latency = time.Duration(bufStatus.Size) * time.Millisecond
	}

	return stats
}

func (s *CallSessionImpl) ProcessOutgoingFrame(frame *MediaFrame) error {
	if frame == nil || len(frame.Payload) == 0 {
		return nil
	}

	var codec Codec
	var rtp RTPProcessor
	var jitter JitterBuffer

	if frame.Type == MediaTypeAudio {
		codec = s.audioCodec
		rtp = s.rtpAudio
		jitter = s.jitterAudio
	} else {
		codec = s.videoCodec
		rtp = s.rtpVideo
		jitter = s.jitterVideo
	}

	_ = jitter

	if codec == nil || rtp == nil {
		return nil
	}

	encoded, err := codec.EncodeFrame(frame.Payload)
	if err != nil {
		return err
	}

	frame.Payload = encoded

	packets, err := rtp.Packetize(frame)
	if err != nil {
		return err
	}

	_ = packets

	return nil
}

func (s *CallSessionImpl) ProcessIncomingPacket(packet *RTPPacket) error {
	if packet == nil {
		return ErrInvalidPacket
	}

	var rtp RTPProcessor
	var jitter JitterBuffer

	if packet.PayloadType < 100 {
		rtp = s.rtpAudio
		jitter = s.jitterAudio
	} else {
		rtp = s.rtpVideo
		jitter = s.jitterVideo
	}

	if rtp == nil || jitter == nil {
		return nil
	}

	frame, err := rtp.Depacketize(packet)
	if err != nil {
		return err
	}

	if frame == nil {
		return nil
	}

	var codec Codec
	if frame.Type == MediaTypeAudio {
		codec = s.audioCodec
	} else {
		codec = s.videoCodec
	}

	if codec != nil {
		decoded, err := codec.DecodeFrame(frame.Payload)
		if err == nil {
			frame.Payload = decoded
		}
	}

	select {
	case s.recvFrameChan <- frame:
		return nil
	default:
		return errors.New("receive buffer full")
	}
}

func generateSSRC() uint32 {
	return uint32(time.Now().UnixNano() & 0xFFFFFFFF)
}

type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*CallSessionImpl
	config      *CallConfig
	audioConfig *MediaConfig
	videoConfig *MediaConfig
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*CallSessionImpl),
	}
}

func (m *SessionManager) CreateSession(sessionID string, config *CallConfig, peerInfo *PeerInfo) (*CallSessionImpl, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[sessionID]; exists {
		return nil, errors.New("session already exists")
	}

	session, err := NewCallSessionImpl(sessionID, config, peerInfo)
	if err != nil {
		return nil, err
	}

	m.sessions[sessionID] = session

	return session, nil
}

func (m *SessionManager) GetSession(sessionID string) (*CallSessionImpl, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, ErrNotFound
	}

	return session, nil
}

func (m *SessionManager) RemoveSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil
	}

	session.Close()

	delete(m.sessions, sessionID)

	return nil
}

func (m *SessionManager) ListSessions() []*CallSessionImpl {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*CallSessionImpl, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

func (m *SessionManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, session := range m.sessions {
		session.Close()
	}

	m.sessions = make(map[string]*CallSessionImpl)

	return nil
}
