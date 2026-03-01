package group

import (
	"sync"
	"time"
)

// GroupMessage represents a group message
type GroupMessage struct {
	ID        string
	GroupID   string
	SenderID  string
	Content   string
	Type      MessageType
	Timestamp time.Time
	Status    MessageStatus
	mu        sync.RWMutex
}

// MessageType represents message type
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeVideo    MessageType = "video"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeFile     MessageType = "file"
	MessageTypeSystem   MessageType = "system"
)

// MessageStatus represents message status
type MessageStatus string

const (
	MessageStatusSent     MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead     MessageStatus = "read"
	MessageStatusFailed   MessageStatus = "failed"
)

// GroupMessageManager manages group messages
type GroupMessageManager struct {
	messages   map[string][]*GroupMessage // groupID -> messages
	callbacks  *MessageCallbacks
	mu         sync.RWMutex
	stats      MessageStats
}

// MessageCallbacks represents message event callbacks
type MessageCallbacks struct {
	OnMessageSent      func(msg *GroupMessage)
	OnMessageDelivered func(msg *GroupMessage)
	OnMessageRead      func(msg *GroupMessage)
}

// MessageStats represents message statistics
type MessageStats struct {
	TotalMessages   int
	MessagesSent    int
	MessagesDelivered int
	MessagesRead    int
	MessagesFailed  int
}

// NewGroupMessageManager creates a new group message manager
func NewGroupMessageManager() *GroupMessageManager {
	return &GroupMessageManager{
		messages:  make(map[string][]*GroupMessage),
		callbacks: &MessageCallbacks{},
	}
}

// SendMessage sends a message to group
func (gmm *GroupMessageManager) SendMessage(groupID, senderID, content string, msgType MessageType) (*GroupMessage, error) {
	gmm.mu.Lock()
	defer gmm.mu.Unlock()

	// Create message
	msg := &GroupMessage{
		ID:        generateMessageID(),
		GroupID:   groupID,
		SenderID:  senderID,
		Content:   content,
		Type:      msgType,
		Timestamp: time.Now(),
		Status:    MessageStatusSent,
	}

	// Store message
	if _, exists := gmm.messages[groupID]; !exists {
		gmm.messages[groupID] = make([]*GroupMessage, 0)
	}
	gmm.messages[groupID] = append(gmm.messages[groupID], msg)

	gmm.stats.TotalMessages++
	gmm.stats.MessagesSent++

	// Notify callback
	if gmm.callbacks.OnMessageSent != nil {
		go gmm.callbacks.OnMessageSent(msg)
	}

	return msg, nil
}

// GetMessages gets messages from group
func (gmm *GroupMessageManager) GetMessages(groupID string, limit int, beforeTime time.Time) ([]*GroupMessage, error) {
	gmm.mu.RLock()
	defer gmm.mu.RUnlock()

	messages, exists := gmm.messages[groupID]
	if !exists {
		return []*GroupMessage{}, nil
	}

	// Filter and limit messages
	result := make([]*GroupMessage, 0)
	for i := len(messages) - 1; i >= 0; i-- {
		if beforeTime.IsZero() || messages[i].Timestamp.Before(beforeTime) {
			result = append(result, messages[i])
			if len(result) >= limit {
				break
			}
		}
	}

	// Reverse to get chronological order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// SetMessageDelivered marks message as delivered
func (gmm *GroupMessageManager) SetMessageDelivered(groupID, messageID string) error {
	gmm.mu.Lock()
	defer gmm.mu.Unlock()

	msg, err := gmm.getMessage(groupID, messageID)
	if err != nil {
		return err
	}

	msg.mu.Lock()
	if msg.Status == MessageStatusSent {
		msg.Status = MessageStatusDelivered
		gmm.stats.MessagesDelivered++
	}
	msg.mu.Unlock()

	// Notify callback
	if gmm.callbacks.OnMessageDelivered != nil {
		go gmm.callbacks.OnMessageDelivered(msg)
	}

	return nil
}

// SetMessageRead marks message as read
func (gmm *GroupMessageManager) SetMessageRead(groupID, messageID string) error {
	gmm.mu.Lock()
	defer gmm.mu.Unlock()

	msg, err := gmm.getMessage(groupID, messageID)
	if err != nil {
		return err
	}

	msg.mu.Lock()
	if msg.Status == MessageStatusSent || msg.Status == MessageStatusDelivered {
		msg.Status = MessageStatusRead
		gmm.stats.MessagesRead++
	}
	msg.mu.Unlock()

	// Notify callback
	if gmm.callbacks.OnMessageRead != nil {
		go gmm.callbacks.OnMessageRead(msg)
	}

	return nil
}

// SetMessageFailed marks message as failed
func (gmm *GroupMessageManager) SetMessageFailed(groupID, messageID string, reason string) error {
	gmm.mu.Lock()
	defer gmm.mu.Unlock()

	msg, err := gmm.getMessage(groupID, messageID)
	if err != nil {
		return err
	}

	msg.mu.Lock()
	msg.Status = MessageStatusFailed
	gmm.stats.MessagesFailed++
	msg.mu.Unlock()

	return nil
}

// DeleteMessage deletes a message
func (gmm *GroupMessageManager) DeleteMessage(groupID, messageID string) error {
	gmm.mu.Lock()
	defer gmm.mu.Unlock()

	messages, exists := gmm.messages[groupID]
	if !exists {
		return ErrGroupNotFound
	}

	for i, msg := range messages {
		if msg.ID == messageID {
			gmm.messages[groupID] = append(messages[:i], messages[i+1:]...)
			gmm.stats.TotalMessages--
			return nil
		}
	}

	return ErrMessageNotFound
}

// GetMessage gets a message by ID
func (gmm *GroupMessageManager) GetMessage(groupID, messageID string) (*GroupMessage, error) {
	gmm.mu.RLock()
	defer gmm.mu.RUnlock()
	return gmm.getMessage(groupID, messageID)
}

// getMessage gets a message by ID (internal, assumes lock held)
func (gmm *GroupMessageManager) getMessage(groupID, messageID string) (*GroupMessage, error) {
	messages, exists := gmm.messages[groupID]
	if !exists {
		return nil, ErrGroupNotFound
	}

	for _, msg := range messages {
		if msg.ID == messageID {
			return msg, nil
		}
	}

	return nil, ErrMessageNotFound
}

// SetCallbacks sets message callbacks
func (gmm *GroupMessageManager) SetCallbacks(callbacks *MessageCallbacks) {
	gmm.callbacks = callbacks
}

// GetStats gets message statistics
func (gmm *GroupMessageManager) GetStats() MessageStats {
	gmm.mu.RLock()
	defer gmm.mu.RUnlock()
	return gmm.stats
}

// ClearMessages clears all messages for a group
func (gmm *GroupMessageManager) ClearMessages(groupID string) error {
	gmm.mu.Lock()
	defer gmm.mu.Unlock()

	_, exists := gmm.messages[groupID]
	if !exists {
		return ErrGroupNotFound
	}

	count := len(gmm.messages[groupID])
	gmm.messages[groupID] = make([]*GroupMessage, 0)
	gmm.stats.TotalMessages -= count

	return nil
}

// SyncMessages syncs messages for a group
func (gmm *GroupMessageManager) SyncMessages(groupID string, lastSyncTime time.Time) ([]*GroupMessage, error) {
	gmm.mu.RLock()
	defer gmm.mu.RUnlock()

	messages, exists := gmm.messages[groupID]
	if !exists {
		return []*GroupMessage{}, nil
	}

	// Get messages since last sync
	result := make([]*GroupMessage, 0)
	for _, msg := range messages {
		if msg.Timestamp.After(lastSyncTime) {
			result = append(result, msg)
		}
	}

	return result, nil
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return "msg-" + time.Now().Format("20060102150405.000")
}

// Message errors
var (
	ErrMessageNotFound = "message not found"
)
