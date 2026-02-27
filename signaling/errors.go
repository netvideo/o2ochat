package signaling

import "errors"

var (
	ErrNotConnected         = errors.New("not connected")
	ErrAlreadyConnected     = errors.New("already connected")
	ErrConnectionFailed     = errors.New("connection failed")
	ErrMessageSendFailed    = errors.New("message send failed")
	ErrMessageReceiveFailed = errors.New("message receive failed")
	ErrInvalidMessage       = errors.New("invalid message")
	ErrInvalidPeerID        = errors.New("invalid peer ID")
	ErrPeerNotFound         = errors.New("peer not found")
	ErrPeerOffline          = errors.New("peer is offline")
	ErrServerNotRunning     = errors.New("server not running")
	ErrServerAlreadyRunning = errors.New("server already running")
	ErrRegistrationFailed   = errors.New("registration failed")
	ErrUnregistrationFailed = errors.New("unregistration failed")
	ErrLookupFailed         = errors.New("lookup failed")
	ErrSignatureInvalid     = errors.New("invalid signature")
	ErrNonceInvalid         = errors.New("invalid nonce")
	ErrMessageTimeout       = errors.New("message timeout")
	ErrHeartbeatTimeout     = errors.New("heartbeat timeout")
	ErrMaxConnections       = errors.New("max connections reached")
	ErrDHTJoinFailed        = errors.New("DHT join failed")
	ErrDHTPublishFailed     = errors.New("DHT publish failed")
	ErrDHTLookupFailed      = errors.New("DHT lookup failed")
	ErrCompressionFailed    = errors.New("compression failed")
	ErrDecompressionFailed  = errors.New("decompression failed")
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrConnectionClosed     = errors.New("connection closed")
	ErrReconnectFailed      = errors.New("reconnect failed")
)

type SignalingError struct {
	Code    string
	Message string
	Err     error
}

func (e *SignalingError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *SignalingError) Unwrap() error {
	return e.Err
}

func NewSignalingError(code, message string, err error) *SignalingError {
	return &SignalingError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
