package transport

import (
	"net"
	"time"
)

type TransportManager interface {
	Connect(config *ConnectionConfig) (Connection, error)
	Accept() (Connection, error)
	Close() error
	GetConnections() ([]ConnectionInfo, error)
	FindConnection(peerID string) (Connection, error)
	Listen(addr string) error
	GetNetworkType() NetworkType
	SetConnectionHandler(handler func(Connection))
}

type Connection interface {
	OpenStream(config *StreamConfig) (Stream, error)
	AcceptStream() (Stream, error)
	Close() error
	GetInfo() ConnectionInfo
	SendControlMessage(msg []byte) error
	ReceiveControlMessage() ([]byte, error)
	GetStats() *ConnectionStats
}

type Stream interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	GetStreamID() uint32
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	GetInfo() StreamInfo
}

type NATTraversal interface {
	GetPublicAddresses() ([]string, error)
	CreateHolePunching(localAddr, remoteAddr string) (net.Conn, error)
	CreateRelayConnection(relayServer string) (net.Conn, error)
}

type ConnectionHandler interface {
	OnConnection(conn Connection)
	OnDisconnection(conn Connection, err error)
	OnError(err error)
}

type StreamHandler interface {
	OnStreamOpen(stream Stream)
	OnStreamClose(stream Stream)
	OnStreamError(stream Stream, err error)
}
