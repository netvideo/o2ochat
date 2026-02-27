package transport

import "errors"

var (
	ErrNotConnected        = errors.New("not connected")
	ErrAlreadyConnected    = errors.New("already connected")
	ErrConnectionFailed    = errors.New("connection failed")
	ErrConnectionClosed    = errors.New("connection closed")
	ErrNoAvailableAddress  = errors.New("no available address")
	ErrHandshakeTimeout    = errors.New("handshake timeout")
	ErrStreamOpenFailed    = errors.New("stream open failed")
	ErrStreamClosed        = errors.New("stream closed")
	ErrWriteTimeout        = errors.New("write timeout")
	ErrReadTimeout         = errors.New("read timeout")
	ErrInvalidConfig       = errors.New("invalid config")
	ErrInvalidAddress      = errors.New("invalid address")
	ErrUnsupportedProtocol = errors.New("unsupported protocol")
	ErrNATTraversalFailed  = errors.New("NAT traversal failed")
	ErrHolePunchingFailed  = errors.New("hole punching failed")
	ErrRelayFailed         = errors.New("relay connection failed")
	ErrListenerClosed      = errors.New("listener closed")
	ErrMaxStreamsReached   = errors.New("max streams reached")
	ErrBandwidthExceeded   = errors.New("bandwidth exceeded")
	ErrNetworkChanged      = errors.New("network changed")
	ErrEncryptionFailed    = errors.New("encryption failed")
	ErrPeerNotReachable    = errors.New("peer not reachable")
	ErrConnectionNotFound  = errors.New("connection not found")
)

type TransportError struct {
	Code    string
	Message string
	Err     error
}

func (e *TransportError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *TransportError) Unwrap() error {
	return e.Err
}

func NewTransportError(code, message string, err error) *TransportError {
	return &TransportError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
