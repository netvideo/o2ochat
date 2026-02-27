package signaling

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	config            *ClientConfig
	conn              *websocket.Conn
	state             ConnectionState
	mu                sync.RWMutex
	messageHandler    func(*SignalingMessage)
	errorHandler      func(error)
	messageChan       chan *SignalingMessage
	reconnectTimer    *time.Timer
	running           bool
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	peerID            string
	publicKey         []byte
	reconnectAttempts int
	lastHeartbeat     time.Time
	identityManager   interface {
		SignMessage(message []byte) ([]byte, error)
		VerifySignature(peerID string, message, signature []byte) bool
	}
}

func NewWebSocketClient(config *ClientConfig) SignalingClient {
	if config == nil {
		config = DefaultClientConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &wsClient{
		config:      config,
		state:       StateDisconnected,
		messageChan: make(chan *SignalingMessage, 100),
		running:     false,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (c *wsClient) Connect(serverURL string) error {
	if serverURL == "" {
		serverURL = c.config.ServerURL
	}

	c.mu.Lock()
	if c.state == StateConnected || c.state == StateConnecting {
		c.mu.Unlock()
		return ErrAlreadyConnected
	}
	c.state = StateConnecting
	c.mu.Unlock()

	var dialer websocket.Dialer
	if c.config.TLSConfig != nil {
		dialer = websocket.Dialer{
			TLSClientConfig: &tls.Config{
				ServerName:         c.config.TLSConfig.ServerName,
				InsecureSkipVerify: c.config.TLSConfig.InsecureSkipVerify,
			},
		}
	}

	conn, _, err := dialer.Dial(serverURL, nil)
	if err != nil {
		c.mu.Lock()
		c.state = StateFailed
		c.mu.Unlock()
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	c.mu.Lock()
	c.conn = conn
	c.state = StateConnected
	c.running = true
	c.mu.Unlock()

	c.wg.Add(1)
	go c.readLoop()

	c.wg.Add(1)
	go c.heartbeatLoop()

	c.scheduleReconnect()

	return nil
}

func (c *wsClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateDisconnected {
		return nil
	}

	if c.state != StateConnected && c.state != StateReconnecting {
		return ErrNotConnected
	}

	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}

	c.running = false
	c.cancel()
	c.wg.Wait()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.state = StateDisconnected
	return nil
}

func (c *wsClient) SendMessage(msg *SignalingMessage) error {
	c.mu.RLock()
	state := c.state
	conn := c.conn
	c.mu.RUnlock()

	if state != StateConnected || conn == nil {
		return ErrNotConnected
	}

	msg.Timestamp = time.Now()
	if msg.Nonce == "" {
		msg.Nonce = generateNonce()
	}

	if c.identityManager != nil && len(c.publicKey) > 0 {
		msgData, err := json.Marshal(struct {
			Type      MessageType `json:"type"`
			From      string      `json:"from"`
			To        string      `json:"to"`
			Data      interface{} `json:"data"`
			Timestamp time.Time   `json:"timestamp"`
			Nonce     string      `json:"nonce"`
		}{
			Type:      msg.Type,
			From:      msg.From,
			To:        msg.To,
			Data:      msg.Data,
			Timestamp: msg.Timestamp,
			Nonce:     msg.Nonce,
		})
		if err != nil {
			return err
		}

		signature, err := c.identityManager.SignMessage(msgData)
		if err != nil {
			return err
		}
		msg.Signature = signature
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	conn.SetWriteDeadline(time.Now().Add(c.config.MessageTimeout))
	return conn.WriteMessage(websocket.TextMessage, data)
}

func (c *wsClient) ReceiveMessage() (*SignalingMessage, error) {
	select {
	case msg := <-c.messageChan:
		return msg, nil
	case <-c.ctx.Done():
		return nil, ErrNotConnected
	case <-time.After(c.config.MessageTimeout):
		return nil, ErrMessageTimeout
	}
}

func (c *wsClient) Register(peerInfo *PeerInfo) error {
	msg := &SignalingMessage{
		Type: MessageTypeRegister,
		From: peerInfo.PeerID,
		Data: map[string]interface{}{
			"peer_id":    peerInfo.PeerID,
			"ipv6_addrs": peerInfo.IPv6Addrs,
			"ipv4_addrs": peerInfo.IPv4Addrs,
			"public_key": peerInfo.PublicKey,
		},
	}

	c.peerID = peerInfo.PeerID
	c.publicKey = peerInfo.PublicKey

	return c.SendMessage(msg)
}

func (c *wsClient) Unregister() error {
	c.mu.RLock()
	peerID := c.peerID
	c.mu.RUnlock()

	msg := &SignalingMessage{
		Type: MessageTypeBye,
		From: peerID,
	}

	return c.SendMessage(msg)
}

func (c *wsClient) LookupPeer(peerID string) (*PeerInfo, error) {
	msg := &SignalingMessage{
		Type: MessageTypeLookup,
		From: c.peerID,
		To:   "",
		Data: peerID,
	}

	if err := c.SendMessage(msg); err != nil {
		return nil, err
	}

	select {
	case response := <-c.messageChan:
		if response.Type == MessageTypeLookup {
			if peerInfo, ok := response.Data.(*PeerInfo); ok {
				return peerInfo, nil
			}
		}
		return nil, ErrLookupFailed
	case <-time.After(c.config.MessageTimeout):
		return nil, ErrMessageTimeout
	case <-c.ctx.Done():
		return nil, ErrNotConnected
	}
}

func (c *wsClient) GetState() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *wsClient) GetConfig() *ClientConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

func (c *wsClient) SetMessageHandler(handler func(*SignalingMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageHandler = handler
}

func (c *wsClient) SetErrorHandler(handler func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorHandler = handler
}

func (c *wsClient) Close() error {
	return c.Disconnect()
}

func (c *wsClient) readLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			conn.SetReadDeadline(time.Now().Add(c.config.MessageTimeout))
			_, message, err := conn.ReadMessage()
			if err != nil {
				c.handleDisconnection(err)
				return
			}

			var msg SignalingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				c.handleError(ErrInvalidMessage)
				continue
			}

			if err := c.verifyMessage(&msg); err != nil {
				c.handleError(err)
				continue
			}

			c.lastHeartbeat = time.Now()

			select {
			case c.messageChan <- &msg:
			default:
			}

			c.mu.RLock()
			handler := c.messageHandler
			c.mu.RUnlock()
			if handler != nil {
				handler(&msg)
			}
		}
	}
}

func (c *wsClient) handleDisconnection(err error) {
	c.mu.Lock()
	wasConnected := c.state == StateConnected
	c.state = StateReconnecting
	c.mu.Unlock()

	if wasConnected {
		c.handleError(ErrConnectionFailed)

		if c.reconnectAttempts < c.config.MaxReconnectAttempts {
			c.scheduleReconnect()
		} else {
			c.mu.Lock()
			c.state = StateFailed
			c.mu.Unlock()
			c.handleError(ErrConnectionFailed)
		}
	}
}

func (c *wsClient) scheduleReconnect() {
	c.mu.Lock()
	c.reconnectAttempts++
	interval := c.config.ReconnectInterval
	c.mu.Unlock()

	c.reconnectTimer = time.AfterFunc(interval, func() {
		if !c.running {
			return
		}

		err := c.Connect("")
		if err != nil {
			c.mu.Lock()
			if c.reconnectAttempts < c.config.MaxReconnectAttempts {
				c.mu.Unlock()
				c.scheduleReconnect()
			} else {
				c.mu.Unlock()
				c.handleError(ErrConnectionFailed)
			}
		} else {
			c.mu.Lock()
			c.reconnectAttempts = 0
			c.mu.Unlock()
		}
	})
}

func (c *wsClient) heartbeatLoop() {
	ticker := time.NewTicker(c.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			state := c.state
			c.mu.RUnlock()

			if state != StateConnected {
				continue
			}

			pingMsg := &SignalingMessage{
				Type: MessageTypePing,
				From: c.peerID,
			}

			if err := c.SendMessage(pingMsg); err != nil {
				c.handleDisconnection(err)
				return
			}

			if time.Since(c.lastHeartbeat) > c.config.HeartbeatInterval*2 {
				c.handleDisconnection(ErrHeartbeatTimeout)
				return
			}
		}
	}
}

func (c *wsClient) verifyMessage(msg *SignalingMessage) error {
	if c.identityManager == nil || len(msg.Signature) == 0 {
		return nil
	}

	msgData, err := json.Marshal(struct {
		Type      MessageType `json:"type"`
		From      string      `json:"from"`
		To        string      `json:"to"`
		Data      interface{} `json:"data"`
		Timestamp time.Time   `json:"timestamp"`
		Nonce     string      `json:"nonce"`
	}{
		Type:      msg.Type,
		From:      msg.From,
		To:        msg.To,
		Data:      msg.Data,
		Timestamp: msg.Timestamp,
		Nonce:     msg.Nonce,
	})
	if err != nil {
		return err
	}

	if !c.identityManager.VerifySignature(msg.From, msgData, msg.Signature) {
		return ErrSignatureInvalid
	}

	return nil
}

func (c *wsClient) handleError(err error) {
	c.mu.RLock()
	handler := c.errorHandler
	c.mu.RUnlock()

	if handler != nil {
		handler(err)
	}
}

func generateNonce() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}
