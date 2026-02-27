package signaling

type SignalingClient interface {
	Connect(serverURL string) error
	Disconnect() error
	SendMessage(msg *SignalingMessage) error
	ReceiveMessage() (*SignalingMessage, error)
	Register(peerInfo *PeerInfo) error
	Unregister() error
	LookupPeer(peerID string) (*PeerInfo, error)
	GetState() ConnectionState
	GetConfig() *ClientConfig
	SetMessageHandler(handler func(*SignalingMessage))
	SetErrorHandler(handler func(error))
	Close() error
}

type SignalingServer interface {
	Start(addr string) error
	Stop() error
	Broadcast(msg *SignalingMessage) error
	GetOnlinePeers() ([]*PeerInfo, error)
	GetConfig() *ServerConfig
	IsRunning() bool
}

type DHTSignaling interface {
	Join(bootstrapNodes []string) error
	Publish(peerInfo *PeerInfo) error
	FindPeer(peerID string) (*PeerInfo, error)
	Leave() error
	GetLocalPeerID() string
}

type MessageHandler interface {
	OnMessage(msg *SignalingMessage)
	OnConnect(peerID string)
	OnDisconnect(peerID string)
	OnError(err error)
}
