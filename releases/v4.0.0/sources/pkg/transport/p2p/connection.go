package p2p

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/ice"
)

// PeerConnection 封装 WebRTC 对等连接
type PeerConnection struct {
	id       string
	peerID   string
	conn     *webrtc.PeerConnection
	signaler Signaler

	// 通道
	onDataChannel     func(*webrtc.DataChannel)
	onConnectionState func(webrtc.PeerConnectionState)
	onICECandidate    func(*webrtc.ICECandidate)

	// 内部状态
	mu            sync.RWMutex
	dataChannels  map[string]*webrtc.DataChannel
	iceCandidates []*webrtc.ICECandidate
	isInitiator   bool
}

// Signaler 信令接口
type Signaler interface {
	SendSignal(peerID string, signal *SignalMessage) error
	OnSignal(handler func(peerID string, signal *SignalMessage))
}

// SignalMessage 信令消息
type SignalMessage struct {
	Type      string            `json:"type"` // offer, answer, candidate, join, leave
	SDP       string            `json:"sdp,omitempty"`
	Candidate *ICECandidateJSON `json:"candidate,omitempty"`
	From      string            `json:"from"`
	To        string            `json:"to"`
	Timestamp int64             `json:"timestamp"`
}

// ICECandidateJSON ICE 候选者 JSON 格式
type ICECandidateJSON struct {
	Candidate        string `json:"candidate"`
	SDPMid           string `json:"sdpMid"`
	SDPMLineIndex    uint16 `json:"sdpMLineIndex"`
	UsernameFragment string `json:"usernameFragment,omitempty"`
}

// Config P2P 配置
type Config struct {
	ICE         []string // ICE 服务器列表
	STUNServers []string
	TURNServers []TURNServer
	IsInitiator bool
}

// TURNServer TURN 服务器配置
type TURNServer struct {
	URL      string
	Username string
	Password string
}

// NewPeerConnection 创建新的 P2P 连接
func NewPeerConnection(id, peerID string, config *Config, signaler Signaler) (*PeerConnection, error) {
	pc := &PeerConnection{
		id:           id,
		peerID:       peerID,
		signaler:     signaler,
		dataChannels: make(map[string]*webrtc.DataChannel),
		isInitiator:  config.IsInitiator,
	}

	// 创建 WebRTC 配置
	webrtcConfig := &webrtc.Configuration{
		ICEServers:         []webrtc.ICEServer{},
		ICETransportPolicy: webrtc.ICETransportPolicyAll,
		BundlePolicy:       webrtc.BundlePolicyBalanced,
		RTCPMuxPolicy:      webrtc.RTCPMuxPolicyRequire,
	}

	// 添加 STUN 服务器
	for _, stun := range config.STUNServers {
		webrtcConfig.ICEServers = append(webrtcConfig.ICEServers, webrtc.ICEServer{
			URLs: []string{stun},
		})
	}

	// 添加 TURN 服务器
	for _, turn := range config.TURNServers {
		webrtcConfig.ICEServers = append(webrtcConfig.ICEServers, webrtc.ICEServer{
			URLs:       []string{turn.URL},
			Username:   turn.Username,
			Credential: turn.Password,
		})
	}

	// 创建 PeerConnection
	conn, err := webrtc.NewPeerConnection(*webrtcConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 PeerConnection 失败: %w", err)
	}

	pc.conn = conn

	// 设置事件处理器
	pc.setupEventHandlers()

	return pc, nil
}

// setupEventHandlers 设置事件处理器
func (pc *PeerConnection) setupEventHandlers() {
	// ICE 候选者事件
	pc.conn.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		pc.mu.Lock()
		pc.iceCandidates = append(pc.iceCandidates, c)
		pc.mu.Unlock()

		// 发送 ICE 候选者给对方
		candidate := &ICECandidateJSON{
			Candidate:        c.ToJSON().Candidate,
			SDPMid:           c.ToJSON().SDPMid,
			SDPMLineIndex:    c.ToJSON().SDPMLineIndex,
			UsernameFragment: c.ToJSON().UsernameFragment,
		}

		if pc.onICECandidate != nil {
			pc.onICECandidate(c)
		}

		// 通过信令发送
		if pc.signaler != nil {
			signal := &SignalMessage{
				Type:      "candidate",
				Candidate: candidate,
				From:      pc.id,
				To:        pc.peerID,
				Timestamp: time.Now().Unix(),
			}

			if err := pc.signaler.SendSignal(pc.peerID, signal); err != nil {
				log.Printf("发送 ICE 候选者失败: %v", err)
			}
		}
	})

	// 连接状态变化事件
	pc.conn.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("连接状态变化: %s", s.String())

		if pc.onConnectionState != nil {
			pc.onConnectionState(s)
		}

		switch s {
		case webrtc.PeerConnectionStateConnected:
			log.Println("P2P 连接已建立")
		case webrtc.PeerConnectionStateFailed:
			log.Println("P2P 连接失败")
		case webrtc.PeerConnectionStateDisconnected:
			log.Println("P2P 连接断开")
		case webrtc.PeerConnectionStateClosed:
			log.Println("P2P 连接已关闭")
		}
	})

	// DataChannel 事件
	pc.conn.OnDataChannel(func(dc *webrtc.DataChannel) {
		log.Printf("新的 DataChannel: %s", dc.Label())

		pc.mu.Lock()
		pc.dataChannels[dc.Label()] = dc
		pc.mu.Unlock()

		// 设置 DataChannel 事件处理器
		pc.setupDataChannelHandlers(dc)

		if pc.onDataChannel != nil {
			pc.onDataChannel(dc)
		}
	})
}

// setupDataChannelHandlers 设置 DataChannel 事件处理器
func (pc *PeerConnection) setupDataChannelHandlers(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		log.Printf("DataChannel %s 已打开", dc.Label())
	})

	dc.OnClose(func() {
		log.Printf("DataChannel %s 已关闭", dc.Label())
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.Printf("DataChannel %s 收到消息: %d bytes", dc.Label(), len(msg.Data))
		// 这里处理收到的消息
	})

	dc.OnError(func(err error) {
		log.Printf("DataChannel %s 错误: %v", dc.Label(), err)
	})
}

// CreateOffer 创建 SDP Offer
func (pc *PeerConnection) CreateOffer() (*webrtc.SessionDescription, error) {
	if !pc.isInitiator {
		return nil, fmt.Errorf("只有发起者才能创建 Offer")
	}

	offer, err := pc.conn.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Offer 失败: %w", err)
	}

	// 设置本地描述
	if err := pc.conn.SetLocalDescription(offer); err != nil {
		return nil, fmt.Errorf("设置本地描述失败: %w", err)
	}

	return &offer, nil
}

// CreateAnswer 创建 SDP Answer
func (pc *PeerConnection) CreateAnswer(offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	if pc.isInitiator {
		return nil, fmt.Errorf("发起者不能创建 Answer")
	}

	// 设置远程描述
	if err := pc.conn.SetRemoteDescription(*offer); err != nil {
		return nil, fmt.Errorf("设置远程描述失败: %w", err)
	}

	// 创建 Answer
	answer, err := pc.conn.CreateAnswer(nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Answer 失败: %w", err)
	}

	// 设置本地描述
	if err := pc.conn.SetLocalDescription(answer); err != nil {
		return nil, fmt.Errorf("设置本地描述失败: %w", err)
	}

	return &answer, nil
}

// SetRemoteDescription 设置远程 SDP
func (pc *PeerConnection) SetRemoteDescription(desc *webrtc.SessionDescription) error {
	return pc.conn.SetRemoteDescription(*desc)
}

// AddICECandidate 添加 ICE 候选者
func (pc *PeerConnection) AddICECandidate(candidate *ICECandidateJSON) error {
	if candidate == nil {
		return fmt.Errorf("ICE 候选者为空")
	}

	c := webrtc.ICECandidateInit{
		Candidate:        candidate.Candidate,
		SDPMid:           &candidate.SDPMid,
		SDPMLineIndex:    &candidate.SDPMLineIndex,
		UsernameFragment: &candidate.UsernameFragment,
	}

	return pc.conn.AddICECandidate(c)
}

// CreateDataChannel 创建 DataChannel
func (pc *PeerConnection) CreateDataChannel(label string, config *webrtc.DataChannelInit) (*webrtc.DataChannel, error) {
	return pc.conn.CreateDataChannel(label, config)
}

// GetDataChannel 获取 DataChannel
func (pc *PeerConnection) GetDataChannel(label string) (*webrtc.DataChannel, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dc, ok := pc.dataChannels[label]
	return dc, ok
}

// Close 关闭连接
func (pc *PeerConnection) Close() error {
	return pc.conn.Close()
}

// GetConnectionState 获取连接状态
func (pc *PeerConnection) GetConnectionState() webrtc.PeerConnectionState {
	return pc.conn.ConnectionState()
}

// GetICEConnectionState 获取 ICE 连接状态
func (pc *PeerConnection) GetICEConnectionState() webrtc.ICEConnectionState {
	return pc.conn.ICEConnectionState()
}

// GetSignalingState 获取信令状态
func (pc *PeerConnection) GetSignalingState() webrtc.SignalingState {
	return pc.conn.SignalingState()
}

// IsInitiator 是否是发起者
func (pc *PeerConnection) IsInitiator() bool {
	return pc.isInitiator
}

// GetID 获取连接 ID
func (pc *PeerConnection) GetID() string {
	return pc.id
}

// GetPeerID 获取对端 ID
func (pc *PeerConnection) GetPeerID() string {
	return pc.peerID
}

// SetOnDataChannel 设置 DataChannel 回调
func (pc *PeerConnection) SetOnDataChannel(handler func(*webrtc.DataChannel)) {
	pc.onDataChannel = handler
}

// SetOnConnectionStateChange 设置连接状态变化回调
func (pc *PeerConnection) SetOnConnectionStateChange(handler func(webrtc.PeerConnectionState)) {
	pc.onConnectionState = handler
}

// SetOnICECandidate 设置 ICE 候选者回调
func (pc *PeerConnection) SetOnICECandidate(handler func(*webrtc.ICECandidate)) {
	pc.onICECandidate = handler
}

// GetStats 获取连接统计信息
func (pc *PeerConnection) GetStats() webrtc.StatsReport {
	return pc.conn.GetStats()
}

// GetSenders 获取发送器列表
func (pc *PeerConnection) GetSenders() []*webrtc.RTPSender {
	return pc.conn.GetSenders()
}

// GetReceivers 获取接收器列表
func (pc *PeerConnection) GetReceivers() []*webrtc.RTPReceiver {
	return pc.conn.GetReceivers()
}

// GetTransceivers 获取收发器列表
func (pc *PeerConnection) GetTransceivers() []*webrtc.RTPTransceiver {
	return pc.conn.GetTransceivers()
}

// AddTrack 添加媒体轨道
func (pc *PeerConnection) AddTrack(track webrtc.TrackLocal) (*webrtc.RTPSender, error) {
	return pc.conn.AddTrack(track)
}

// RemoveTrack 移除媒体轨道
func (pc *PeerConnection) RemoveTrack(sender *webrtc.RTPSender) error {
	return pc.conn.RemoveTrack(sender)
}

// AddTransceiverFromTrack 从轨道添加收发器
func (pc *PeerConnection) AddTransceiverFromTrack(track webrtc.TrackLocal, init ...webrtc.RTPTransceiverInit) (*webrtc.RTPTransceiver, error) {
	return pc.conn.AddTransceiverFromTrack(track, init...)
}

// AddTransceiverFromKind 从类型添加收发器
func (pc *PeerConnection) AddTransceiverFromKind(kind webrtc.RTPCodecType, init ...webrtc.RTPTransceiverInit) (*webrtc.RTPTransceiver, error) {
	return pc.conn.AddTransceiverFromKind(kind, init...)
}

// RestartIce 重启 ICE
func (pc *PeerConnection) RestartIce() error {
	return pc.conn.RestartIce()
}
