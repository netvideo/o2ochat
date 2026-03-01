package ui

import (
	"errors"
	"sync"
)

type DefaultChatUI struct {
	mu                sync.RWMutex
	openChats         map[string]bool
	messages          map[string][]*MessageItem
	inputCallback     func(text string, attachments []string)
	reactionCallback  func(messageID string, reaction string)
}

func NewChatUI() ChatUI {
	return &DefaultChatUI{
		openChats: make(map[string]bool),
		messages:  make(map[string][]*MessageItem),
	}
}

func (c *DefaultChatUI) OpenChat(peerID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if peerID == "" {
		return ErrInvalidParameter
	}

	c.openChats[peerID] = true
	return nil
}

func (c *DefaultChatUI) CloseChat(peerID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.openChats[peerID]; !ok {
		return ErrChatNotFound
	}

	delete(c.openChats, peerID)
	return nil
}

func (c *DefaultChatUI) AddMessage(message *MessageItem) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if message == nil {
		return ErrInvalidParameter
	}

	peerID := message.To
	if message.IsOwn {
		peerID = message.From
	}

	c.messages[peerID] = append(c.messages[peerID], message)
	return nil
}

func (c *DefaultChatUI) UpdateMessageStatus(messageID string, status MessageStatus) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, msgs := range c.messages {
		for _, msg := range msgs {
			if msg.ID == messageID {
				msg.Status = status
				return nil
			}
		}
	}

	return ErrMessageNotFound
}

func (c *DefaultChatUI) ClearChat(peerID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.messages[peerID]; !ok {
		return ErrChatNotFound
	}

	c.messages[peerID] = nil
	return nil
}

func (c *DefaultChatUI) SearchMessages(peerID, query string) ([]*MessageItem, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	msgs, ok := c.messages[peerID]
	if !ok {
		return nil, ErrChatNotFound
	}

	var results []*MessageItem
	for _, msg := range msgs {
		if contains(msg.Content, query) {
			results = append(results, msg)
		}
	}

	return results, nil
}

func (c *DefaultChatUI) GetChatHistory(peerID string, limit int) ([]*MessageItem, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	msgs, ok := c.messages[peerID]
	if !ok {
		return nil, ErrChatNotFound
	}

	if limit <= 0 || limit > len(msgs) {
		limit = len(msgs)
	}

	start := len(msgs) - limit
	return msgs[start:], nil
}

func (c *DefaultChatUI) SetInputCallback(callback func(text string, attachments []string)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.inputCallback = callback
	return nil
}

func (c *DefaultChatUI) SetReactionCallback(callback func(messageID string, reaction string)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.reactionCallback = callback
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

var ErrMessageNotFound = errors.New("ui: message not found")
