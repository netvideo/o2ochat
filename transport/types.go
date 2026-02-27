package transport

import (
	"net"
	"time"
)

type ConnectionType string

const (
	ConnectionTypeQUIC   ConnectionType = "quic"
	ConnectionTypeWebRTC ConnectionType = "webrtc"
)

type ConnectionState string

const (
	StateDisconnected ConnectionState = "disconnected"
	StateConnecting   ConnectionState = "connecting"
	StateConnected    ConnectionState = "connected"
	StateFailed       ConnectionState = "failed"
	StateClosing      ConnectionState = "closing"
)

type ConnectionConfig struct {
	PeerID        string           `json:"peer_id"`
	IPv6Addresses []string         `json:"ipv6_addresses"`
	IPv4Addresses []string         `json:"ipv4_addresses"`
	Priority      []ConnectionType `json:"priority"`
	Timeout       time.Duration    `json:"timeout"`
	RetryCount    int              `json:"retry_count"`
}

type ConnectionInfo struct {
	ID            string          `json:"id"`
	PeerID        string          `json:"peer_id"`
	Type          ConnectionType  `json:"type"`
	LocalAddr     net.Addr        `json:"local_addr"`
	RemoteAddr    net.Addr        `json:"remote_addr"`
	State         ConnectionState `json:"state"`
	EstablishedAt time.Time       `json:"established_at"`
	Stats         ConnectionStats `json:"stats"`
}

type ConnectionStats struct {
	BytesSent       uint64        `json:"bytes_sent"`
	BytesReceived   uint64        `json:"bytes_received"`
	PacketsSent     uint64        `json:"packets_sent"`
	PacketsReceived uint64        `json:"packets_received"`
	Retransmits     uint64        `json:"retransmits"`
	Latency         time.Duration `json:"latency"`
	Bandwidth       int64         `json:"bandwidth"`
	PacketLoss      float64       `json:"packet_loss"`
}

type StreamConfig struct {
	Reliable   bool `json:"reliable"`
	Ordered    bool `json:"ordered"`
	MaxRetries int  `json:"max_retries"`
	BufferSize int  `json:"buffer_size"`
}

type StreamInfo struct {
	ID        uint32 `json:"id"`
	Direction string `json:"direction"`
	State     string `json:"state"`
}

type NetworkType string

const (
	NetworkTypeIPv6    NetworkType = "ipv6"
	NetworkTypeIPv4    NetworkType = "ipv4"
	NetworkTypeUnknown NetworkType = "unknown"
)

type QUICConfig struct {
	MaxIncomingStreams    int           `json:"max_incoming_streams"`
	MaxIncomingUniStreams int           `json:"max_incoming_uni_streams"`
	KeepAlive             bool          `json:"keep_alive"`
	HandshakeIdleTimeout  time.Duration `json:"handshake_idle_timeout"`
	MaxIdleTimeout        time.Duration `json:"max_idle_timeout"`
	Enable0RTT            bool          `json:"enable_0rtt"`
}

type WebRTCConfig struct {
	ICEServers []ICEServer `json:"ice_servers"`
	PortRange  PortRange   `json:"port_range"`
}

type ICEServer struct {
	URLs           []string `json:"urls"`
	Username       string   `json:"username"`
	Credential     string   `json:"credential"`
	CredentialType string   `json:"credential_type"`
}

type PortRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

func DefaultConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		Priority:   []ConnectionType{ConnectionTypeQUIC, ConnectionTypeWebRTC},
		Timeout:    10 * time.Second,
		RetryCount: 3,
	}
}

func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		Reliable:   true,
		Ordered:    true,
		MaxRetries: 3,
		BufferSize: 64 * 1024,
	}
}

func DefaultQUICConfig() *QUICConfig {
	return &QUICConfig{
		MaxIncomingStreams:    100,
		MaxIncomingUniStreams: 100,
		KeepAlive:             true,
		HandshakeIdleTimeout:  30 * time.Second,
		MaxIdleTimeout:        5 * time.Minute,
		Enable0RTT:            true,
	}
}

func DefaultWebRTCConfig() *WebRTCConfig {
	return &WebRTCConfig{
		ICEServers: []ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		PortRange: PortRange{Min: 10000, Max: 60000},
	}
}
