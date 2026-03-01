package ui

import (
	"sort"
	"sync"
	"time"
)

type ChatControllerComponent struct {
	mu             sync.RWMutex
	sessions       map[string]*ChatSession
	activeChat     string
	messageList    *MessageListComponent
	contactList    *ContactListComponent
	
	onSendMessage     func(peerID, text string, attachments []string)
	onNewMessage      func(peerID string, message *MessageItem)
	onTyping          func(peerID string, isTyping bool)
	onSessionUpdate   func(session *ChatSession)
	onCallStart       func(peerID string, video bool)
	onFileTransfer    func(peerID string)
}

func NewChatControllerComponent() *ChatControllerComponent {
	cc := &ChatControllerComponent{
		sessions:    make(map[string]*ChatSession),
		messageList: NewMessageListComponent(),
		contactList: NewContactListComponent(),
	}

	cc.setupCallbacks()
	return cc
}

func (cc *ChatControllerComponent) setupCallbacks() {
	cc.messageList.SetOnNewMessage(func(msg *MessageItem) {
		if cc.onNewMessage != nil {
			peerID := msg.To
			if msg.IsOwn {
				peerID = msg.From
			}
			cc.onNewMessage(peerID, msg)
		}
	})
}

func (cc *ChatControllerComponent) CreateSession(peerID, peerName string) *ChatSession {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if session, ok := cc.sessions[peerID]; ok {
		return session
	}

	session := &ChatSession{
		PeerID:          peerID,
		PeerName:        peerName,
		LastMessageTime: time.Now(),
	}

	cc.sessions[peerID] = session

	if cc.onSessionUpdate != nil {
		cc.onSessionUpdate(session)
	}

	return session
}

func (cc *ChatControllerComponent) GetSession(peerID string) (*ChatSession, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	session, ok := cc.sessions[peerID]
	return session, ok
}

func (cc *ChatControllerComponent) GetAllSessions() []*ChatSession {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	sessions := make([]*ChatSession, 0, len(cc.sessions))
	for _, s := range cc.sessions {
		sessions = append(sessions, s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].IsPinned != sessions[j].IsPinned {
			return sessions[i].IsPinned
		}
		return sessions[i].LastMessageTime.After(sessions[j].LastMessageTime)
	})

	return sessions
}

func (cc *ChatControllerComponent) OpenChat(peerID string) {
	cc.mu.Lock()
	cc.activeChat = peerID
	cc.mu.Unlock()

	if session, ok := cc.sessions[peerID]; ok {
		session.UnreadCount = 0
		if cc.onSessionUpdate != nil {
			cc.onSessionUpdate(session)
		}
	}

	cc.messageList.SetPeerID(peerID)
}

func (cc *ChatControllerComponent) CloseChat() {
	cc.mu.Lock()
	cc.activeChat = ""
	cc.mu.Unlock()
}

func (cc *ChatControllerComponent) SendMessage(text string) {
	cc.mu.RLock()
	peerID := cc.activeChat
	cc.mu.RUnlock()

	if peerID == "" || text == "" {
		return
	}

	message := &MessageItem{
		ID:        generateMessageIDComponent(),
		From:      "self",
		To:        peerID,
		Content:   text,
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Status:    MessageStatusSending,
		IsOwn:     true,
	}

	cc.ReceiveMessage(peerID, message)

	if cc.onSendMessage != nil {
		cc.onSendMessage(peerID, text, nil)
	}
}

func (cc *ChatControllerComponent) ReceiveMessage(peerID string, message *MessageItem) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	session, ok := cc.sessions[peerID]
	if !ok {
		session = &ChatSession{
			PeerID:          peerID,
			PeerName:        peerID,
			LastMessageTime: time.Now(),
		}
		cc.sessions[peerID] = session
	}

	session.LastMessage = message.Content
	session.LastMessageTime = message.Timestamp

	if !message.IsOwn && cc.activeChat != peerID {
		session.UnreadCount++
	}

	cc.messageList.AddMessage(message)

	if cc.onNewMessage != nil {
		cc.onNewMessage(peerID, message)
	}

	if cc.onSessionUpdate != nil {
		cc.onSessionUpdate(session)
	}
}

func (cc *ChatControllerComponent) GetActiveChat() string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.activeChat
}

func (cc *ChatControllerComponent) GetMessageList() *MessageListComponent {
	return cc.messageList
}

func (cc *ChatControllerComponent) GetContactList() *ContactListComponent {
	return cc.contactList
}

func (cc *ChatControllerComponent) DeleteSession(peerID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	delete(cc.sessions, peerID)
	cc.messageList.ClearMessages(peerID)

	if cc.activeChat == peerID {
		cc.activeChat = ""
	}
}

func (cc *ChatControllerComponent) PinSession(peerID string, pinned bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if session, ok := cc.sessions[peerID]; ok {
		session.IsPinned = pinned
		if cc.onSessionUpdate != nil {
			cc.onSessionUpdate(session)
		}
	}
}

func (cc *ChatControllerComponent) MuteSession(peerID string, muted bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if session, ok := cc.sessions[peerID]; ok {
		session.IsMuted = muted
		if cc.onSessionUpdate != nil {
			cc.onSessionUpdate(session)
		}
	}
}

func (cc *ChatControllerComponent) SetTyping(peerID string, isTyping bool) {
	if cc.onTyping != nil {
		cc.onTyping(peerID, isTyping)
	}
}

func (cc *ChatControllerComponent) SetOnSendMessage(callback func(peerID, text string, attachments []string)) {
	cc.onSendMessage = callback
}

func (cc *ChatControllerComponent) SetOnNewMessage(callback func(peerID string, message *MessageItem)) {
	cc.onNewMessage = callback
}

func (cc *ChatControllerComponent) SetOnTyping(callback func(peerID string, isTyping bool)) {
	cc.onTyping = callback
}

func (cc *ChatControllerComponent) SetOnSessionUpdate(callback func(session *ChatSession)) {
	cc.onSessionUpdate = callback
}

func (cc *ChatControllerComponent) SetOnCallStart(callback func(peerID string, video bool)) {
	cc.onCallStart = callback
}

func (cc *ChatControllerComponent) SetOnFileTransfer(callback func(peerID string)) {
	cc.onFileTransfer = callback
}

func (cc *ChatControllerComponent) StartCall(video bool) {
	cc.mu.RLock()
	peerID := cc.activeChat
	cc.mu.RUnlock()

	if peerID != "" && cc.onCallStart != nil {
		cc.onCallStart(peerID, video)
	}
}

func (cc *ChatControllerComponent) StartFileTransfer() {
	cc.mu.RLock()
	peerID := cc.activeChat
	cc.mu.RUnlock()

	if peerID != "" && cc.onFileTransfer != nil {
		cc.onFileTransfer(peerID)
	}
}

func generateMessageIDComponent() string {
	return "msg_" + time.Now().Format("20060102150405")
}
