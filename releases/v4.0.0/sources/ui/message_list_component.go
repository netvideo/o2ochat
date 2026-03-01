package ui

import (
	"sync"
)

type MessageListComponent struct {
	mu             sync.RWMutex
	messages       map[string][]*MessageItem
	peerID         string
	maxMessages    int
	onMessageClick func(messageID string)
	onNewMessage   func(message *MessageItem)
}

func NewMessageListComponent() *MessageListComponent {
	return &MessageListComponent{
		messages:   make(map[string][]*MessageItem),
		maxMessages: 1000,
	}
}

func (ml *MessageListComponent) SetPeerID(peerID string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.peerID = peerID
}

func (ml *MessageListComponent) AddMessage(message *MessageItem) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	peerID := message.To
	if message.IsOwn {
		peerID = message.From
	}

	if _, ok := ml.messages[peerID]; !ok {
		ml.messages[peerID] = make([]*MessageItem, 0)
	}

	ml.messages[peerID] = append(ml.messages[peerID], message)

	if len(ml.messages[peerID]) > ml.maxMessages {
		ml.messages[peerID] = ml.messages[peerID][len(ml.messages[peerID])-ml.maxMessages:]
	}

	if ml.onNewMessage != nil && peerID == ml.peerID {
		ml.onNewMessage(message)
	}
}

func (ml *MessageListComponent) UpdateMessage(messageID string, updateFunc func(*MessageItem)) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	for _, msgs := range ml.messages {
		for i, msg := range msgs {
			if msg.ID == messageID {
				updateFunc(msgs[i])
				return
			}
		}
	}
}

func (ml *MessageListComponent) ClearMessages(peerID string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	delete(ml.messages, peerID)
}

func (ml *MessageListComponent) GetMessages(peerID string) []*MessageItem {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	result := make([]*MessageItem, len(ml.messages[peerID]))
	copy(result, ml.messages[peerID])
	return result
}

func (ml *MessageListComponent) SetOnMessageClick(callback func(messageID string)) {
	ml.onMessageClick = callback
}

func (ml *MessageListComponent) SetOnNewMessage(callback func(message *MessageItem)) {
	ml.onNewMessage = callback
}

func (ml *MessageListComponent) GetMessageCount(peerID string) int {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return len(ml.messages[peerID])
}

func (ml *MessageListComponent) GetLatestMessage(peerID string) *MessageItem {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	msgs := ml.messages[peerID]
	if len(msgs) == 0 {
		return nil
	}
	return msgs[len(msgs)-1]
}
