// Package message provides message handling for P2P communications
package message

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MessageType represents the type of message
type MessageType int

const (
	TypeText MessageType = iota
	TypeFile
	TypeControl
	TypeAck
)

func (t MessageType) String() string {
	switch t {
	case TypeText:
		return "text"
	case TypeFile:
		return "file"
	case TypeControl:
		return "control"
	case TypeAck:
		return "ack"
	default:
		return "unknown"
	}
}

// ControlCommand represents control message commands
type ControlCommand string

const (
	CmdPing       ControlCommand = "ping"
	CmdPong       ControlCommand = "pong"
	CmdDisconnect ControlCommand = "disconnect"
	CmdTyping     ControlCommand = "typing"
	CmdSeen       ControlCommand = "seen"
)

// Message represents a message in the system
type Message struct {
	ID           string          `json:"id"`
	Type         MessageType     `json:"type"`
	SenderID     string          `json:"sender_id"`
	ReceiverID   string          `json:"receiver_id"`
	Timestamp    time.Time       `json:"timestamp"`
	Payload      json.RawMessage `json:"payload"`
	Acknowledged bool            `json:"acknowledged"`
	Retries      int             `json:"retries"`
}

// TextPayload represents text message payload
type TextPayload struct {
	Content string `json:"content"`
	Format  string `json:"format,omitempty"`
}

// FilePayload represents file message payload
type FilePayload struct {
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	FileType   string `json:"file_type"`
	Checksum   string `json:"checksum"`
	TransferID string `json:"transfer_id"`
}

// ControlPayload represents control message payload
type ControlPayload struct {
	Command ControlCommand  `json:"command"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// AckPayload represents acknowledgment payload
type AckPayload struct {
	MessageID string    `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
}

// Handler is the function signature for message handlers
type Handler func(msg *Message) error

// Manager defines the interface for message management
type Manager interface {
	// Send a message
	Send(ctx context.Context, msg *Message) error

	// Register a handler for a specific message type
	RegisterHandler(msgType MessageType, handler Handler)

	// Wait for acknowledgment
	WaitForAck(ctx context.Context, messageID string, timeout time.Duration) error

	// Message queue management
	Enqueue(msg *Message) error
	Dequeue() (*Message, error)

	// Close the manager
	Close() error
}

// manager implements Manager
type manager struct {
	handlers   map[MessageType]Handler
	handlersMu sync.RWMutex

	outbox   chan *Message
	outboxMu sync.Mutex

	ackChannels map[string]chan bool
	ackMu       sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	config Config
}

// Config holds message manager configuration
type Config struct {
	OutboxSize    int
	MaxRetries    int
	RetryInterval time.Duration
	AckTimeout    time.Duration
	WorkerCount   int
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		OutboxSize:    1000,
		MaxRetries:    3,
		RetryInterval: 5 * time.Second,
		AckTimeout:    30 * time.Second,
		WorkerCount:   4,
	}
}

// Ensure manager implements Manager
var _ Manager = (*manager)(nil)

// NewManager creates a new message manager
func NewManager(config Config) Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &manager{
		handlers:    make(map[MessageType]Handler),
		outbox:      make(chan *Message, config.OutboxSize),
		ackChannels: make(map[string]chan bool),
		ctx:         ctx,
		cancel:      cancel,
		config:      config,
	}

	// Start workers
	for i := 0; i < config.WorkerCount; i++ {
		m.wg.Add(1)
		go m.processOutbox()
	}

	return m
}

// Send sends a message
func (m *manager) Send(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Enqueue for processing
	select {
	case m.outbox <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RegisterHandler registers a handler for a message type
func (m *manager) RegisterHandler(msgType MessageType, handler Handler) {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.handlers[msgType] = handler
}

// WaitForAck waits for acknowledgment of a message
func (m *manager) WaitForAck(ctx context.Context, messageID string, timeout time.Duration) error {
	ackCh := make(chan bool, 1)

	m.ackMu.Lock()
	m.ackChannels[messageID] = ackCh
	m.ackMu.Unlock()

	defer func() {
		m.ackMu.Lock()
		delete(m.ackChannels, messageID)
		m.ackMu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case ack := <-ackCh:
		if ack {
			return nil
		}
		return fmt.Errorf("message was rejected")
	case <-ctx.Done():
		return fmt.Errorf("acknowledgment timeout: %w", ctx.Err())
	}
}

// Enqueue adds a message to the outbox
func (m *manager) Enqueue(msg *Message) error {
	select {
	case m.outbox <- msg:
		return nil
	default:
		return fmt.Errorf("outbox is full")
	}
}

// Dequeue retrieves a message from the outbox (internal use)
func (m *manager) Dequeue() (*Message, error) {
	select {
	case msg := <-m.outbox:
		return msg, nil
	default:
		return nil, fmt.Errorf("outbox is empty")
	}
}

// Close shuts down the message manager
func (m *manager) Close() error {
	m.cancel()
	m.wg.Wait()
	close(m.outbox)
	return nil
}

// processOutbox processes messages from the outbox
func (m *manager) processOutbox() {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case msg := <-m.outbox:
			if msg == nil {
				return
			}
			m.handleMessage(msg)
		}
	}
}

// handleMessage processes a single message
func (m *manager) handleMessage(msg *Message) {
	m.handlersMu.RLock()
	handler, exists := m.handlers[msg.Type]
	m.handlersMu.RUnlock()

	if !exists {
		return
	}

	// Retry logic
	for i := 0; i <= m.config.MaxRetries; i++ {
		err := handler(msg)
		if err == nil {
			return
		}

		if i < m.config.MaxRetries {
			time.Sleep(m.config.RetryInterval)
		}
	}
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixMilli())
}

// Encode encodes a message to JSON
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode decodes a message from JSON
func Decode(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewTextMessage creates a new text message
func NewTextMessage(senderID, receiverID, content string) *Message {
	payload, _ := json.Marshal(TextPayload{
		Content: content,
		Format:  "plain",
	})

	return &Message{
		ID:         generateMessageID(),
		Type:       TypeText,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Timestamp:  time.Now(),
		Payload:    payload,
	}
}

// NewFileMessage creates a new file message
func NewFileMessage(senderID, receiverID string, file FilePayload) *Message {
	payload, _ := json.Marshal(file)

	return &Message{
		ID:         generateMessageID(),
		Type:       TypeFile,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Timestamp:  time.Now(),
		Payload:    payload,
	}
}

// NewControlMessage creates a new control message
func NewControlMessage(senderID, receiverID string, command ControlCommand, data interface{}) *Message {
	var dataBytes json.RawMessage
	if data != nil {
		dataBytes, _ = json.Marshal(data)
	}

	payload, _ := json.Marshal(ControlPayload{
		Command: command,
		Data:    dataBytes,
	})

	return &Message{
		ID:         generateMessageID(),
		Type:       TypeControl,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Timestamp:  time.Now(),
		Payload:    payload,
	}
}

// NewAckMessage creates a new acknowledgment message
func NewAckMessage(messageID, senderID, receiverID string) *Message {
	payload, _ := json.Marshal(AckPayload{
		MessageID: messageID,
		Timestamp: time.Now(),
	})

	return &Message{
		ID:         generateMessageID(),
		Type:       TypeAck,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Timestamp:  time.Now(),
		Payload:    payload,
	}
}
