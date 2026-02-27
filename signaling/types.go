package signaling

import (
	"time"
)

type MessageType string

const (
	MessageTypeOffer     MessageType = "offer"
	MessageTypeAnswer    MessageType = "answer"
	MessageTypeCandidate MessageType = "candidate"
	MessageTypeInvite    MessageType = "invite"
	MessageTypeBye       MessageType = "bye"
	MessageTypePing      MessageType = "ping"
	MessageTypePong      MessageType = "pong"
	MessageTypeRegister  MessageType = "register"
	MessageTypeLookup    MessageType = "lookup"
)

type SignalingMessage struct {
	Type      MessageType `json:"type"`
	From      string      `json:"from"`
	To        string      `json:"to"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Signature []byte      `json:"signature"`
	Nonce     string      `json:"nonce"`
}

type SDPInfo struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}

type PeerInfo struct {
	PeerID    string    `json:"peer_id"`
	IPv6Addrs []string  `json:"ipv6_addrs,omitempty"`
	IPv4Addrs []string  `json:"ipv4_addrs,omitempty"`
	PublicKey []byte    `json:"public_key"`
	LastSeen  time.Time `json:"last_seen"`
	Online    bool      `json:"online"`
}

type ServerConfig struct {
	MaxConnections    int           `json:"max_connections"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	MessageTimeout    time.Duration `json:"message_timeout"`
	EnableCompression bool          `json:"enable_compression"`
	TLSCertFile       string        `json:"tls_cert_file"`
	TLSKeyFile        string        `json:"tls_key_file"`
	Port              int           `json:"port"`
}

type ClientConfig struct {
	ServerURL            string        `json:"server_url"`
	ReconnectInterval    time.Duration `json:"reconnect_interval"`
	MaxReconnectAttempts int           `json:"max_reconnect_attempts"`
	HeartbeatInterval    time.Duration `json:"heartbeat_interval"`
	MessageTimeout       time.Duration `json:"message_timeout"`
	EnableCompression    bool          `json:"enable_compression"`
	TLSConfig            *TLSConfig    `json:"tls_config"`
}

type TLSConfig struct {
	CAFile             string `json:"ca_file"`
	CertFile           string `json:"cert_file"`
	KeyFile            string `json:"key_file"`
	ServerName         string `json:"server_name"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

type ConnectionState string

func (s ConnectionState) String() string {
	return string(s)
}

const (
	StateDisconnected ConnectionState = "disconnected"
	StateConnecting   ConnectionState = "connecting"
	StateConnected    ConnectionState = "connected"
	StateReconnecting ConnectionState = "reconnecting"
	StateFailed       ConnectionState = "failed"
)

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		MaxConnections:    1000,
		HeartbeatInterval: 30 * time.Second,
		MessageTimeout:    10 * time.Second,
		EnableCompression: true,
		Port:              8080,
	}
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		ServerURL:            "ws://localhost:8080",
		ReconnectInterval:    5 * time.Second,
		MaxReconnectAttempts: 10,
		HeartbeatInterval:    30 * time.Second,
		MessageTimeout:       10 * time.Second,
		EnableCompression:    true,
	}
}
