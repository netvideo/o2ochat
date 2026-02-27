package ui

import (
	"sync"
	"time"
)

type ChatSession struct {
	PeerID           string
	PeerName         string
	LastMessage      string
	LastMessageTime  time.Time
	UnreadCount      int
	IsTyping         bool
	IsPinned         bool
	IsMuted          bool
	IsArchived       bool
}

type ChatManager struct {
	mu          sync.RWMutex
	sessions    map[string]*ChatSession
	activeChat  string
	chatUI      ChatUI
	onNewMessage func(peerID string, message *MessageItem)
	onTyping    func(peerID string, isTyping bool)
	onSessionUpdate func(session *ChatSession)
}

func NewChatManager() *ChatManager {
	return &ChatManager{
		sessions: make(map[string]*ChatSession),
	}
}

func (cm *ChatManager) SetChatUI(chatUI ChatUI) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.chatUI = chatUI
}

func (cm *ChatManager) CreateSession(peerID, peerName string) *ChatSession {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if session, ok := cm.sessions[peerID]; ok {
		return session
	}

	session := &ChatSession{
		PeerID:          peerID,
		PeerName:        peerName,
		LastMessageTime: time.Now(),
	}

	cm.sessions[peerID] = session

	if cm.onSessionUpdate != nil {
		cm.onSessionUpdate(session)
	}

	return session
}

func (cm *ChatManager) GetSession(peerID string) (*ChatSession, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	session, ok := cm.sessions[peerID]
	return session, ok
}

func (cm *ChatManager) GetAllSessions() []*ChatSession {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	sessions := make([]*ChatSession, 0, len(cm.sessions))
	for _, s := range cm.sessions {
		sessions = append(sessions, s)
	}

	return sessions
}

func (cm *ChatManager) GetActiveSession() (*ChatSession, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.activeChat == "" {
		return nil, false
	}

	session, ok := cm.sessions[cm.activeChat]
	return session, ok
}

func (cm *ChatManager) SetActiveChat(peerID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.activeChat = peerID

	if session, ok := cm.sessions[peerID]; ok {
		session.UnreadCount = 0
	}

	if cm.chatUI != nil {
		cm.chatUI.OpenChat(peerID)
	}

	if cm.onSessionUpdate != nil {
		if session, ok := cm.sessions[peerID]; ok {
			cm.onSessionUpdate(session)
		}
	}
}

func (cm *ChatManager) CloseChat(peerID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.chatUI != nil {
		cm.chatUI.CloseChat(peerID)
	}

	if cm.activeChat == peerID {
		cm.activeChat = ""
	}
}

func (cm *ChatManager) AddMessage(peerID string, message *MessageItem) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	session, ok := cm.sessions[peerID]
	if !ok {
		session = cm.createSessionUnlocked(peerID, message.From)
	}

	session.LastMessage = message.Content
	session.LastMessageTime = message.Timestamp

	if !message.IsOwn && cm.activeChat != peerID {
		session.UnreadCount++
	}

	if cm.chatUI != nil {
		cm.chatUI.AddMessage(message)
	}

	if cm.onNewMessage != nil {
		cm.onNewMessage(peerID, message)
	}

	if cm.onSessionUpdate != nil {
		cm.onSessionUpdate(session)
	}
}

func (cm *ChatManager) UpdateMessageStatus(peerID, messageID string, status MessageStatus) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.chatUI != nil {
		cm.chatUI.UpdateMessageStatus(messageID, status)
	}
}

func (cm *ChatManager) SetTyping(peerID string, isTyping bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	session, ok := cm.sessions[peerID]
	if !ok {
		return
	}

	session.IsTyping = isTyping

	if cm.onTyping != nil {
		cm.onTyping(peerID, isTyping)
	}
}

func (cm *ChatManager) DeleteSession(peerID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.sessions, peerID)

	if cm.activeChat == peerID {
		cm.activeChat = ""
	}

	if cm.chatUI != nil {
		cm.chatUI.ClearChat(peerID)
	}
}

func (cm *ChatManager) PinSession(peerID string, pinned bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if session, ok := cm.sessions[peerID]; ok {
		session.IsPinned = pinned
		if cm.onSessionUpdate != nil {
			cm.onSessionUpdate(session)
		}
	}
}

func (cm *ChatManager) MuteSession(peerID string, muted bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if session, ok := cm.sessions[peerID]; ok {
		session.IsMuted = muted
		if cm.onSessionUpdate != nil {
			cm.onSessionUpdate(session)
		}
	}
}

func (cm *ChatManager) ArchiveSession(peerID string, archived bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if session, ok := cm.sessions[peerID]; ok {
		session.IsArchived = archived
		if cm.onSessionUpdate != nil {
			cm.onSessionUpdate(session)
		}
	}
}

func (cm *ChatManager) SearchMessages(peerID, query string) []*MessageItem {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.chatUI != nil {
		results, _ := cm.chatUI.SearchMessages(peerID, query)
		return results
	}

	return nil
}

func (cm *ChatManager) GetChatHistory(peerID string, limit int) []*MessageItem {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.chatUI != nil {
		history, _ := cm.chatUI.GetChatHistory(peerID, limit)
		return history
	}

	return nil
}

func (cm *ChatManager) SetOnNewMessage(callback func(peerID string, message *MessageItem)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onNewMessage = callback
}

func (cm *ChatManager) SetOnTyping(callback func(peerID string, isTyping bool)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onTyping = callback
}

func (cm *ChatManager) SetOnSessionUpdate(callback func(session *ChatSession)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onSessionUpdate = callback
}

func (cm *ChatManager) createSessionUnlocked(peerID, peerName string) *ChatSession {
	session := &ChatSession{
		PeerID:          peerID,
		PeerName:        peerName,
		LastMessageTime: time.Now(),
	}

	cm.sessions[peerID] = session
	return session
}

type ChatSessionList struct {
	mu       sync.RWMutex
	sessions []*ChatSession
}

func NewChatSessionList() *ChatSessionList {
	return &ChatSessionList{
		sessions: make([]*ChatSession, 0),
	}
}

func (csl *ChatSessionList) Update(sessions []*ChatSession) {
	csl.mu.Lock()
	defer csl.mu.Unlock()

	csl.sessions = sessions
}

func (csl *ChatSessionList) GetSessions() []*ChatSession {
	csl.mu.RLock()
	defer csl.mu.RUnlock()

	result := make([]*ChatSession, len(csl.sessions))
	copy(result, csl.sessions)
	return result
}

func (csl *ChatSessionList) GetPinnedSessions() []*ChatSession {
	csl.mu.RLock()
	defer csl.mu.RUnlock()

	var pinned []*ChatSession
	for _, s := range csl.sessions {
		if s.IsPinned {
			pinned = append(pinned, s)
		}
	}
	return pinned
}

func (csl *ChatSessionList) GetArchivedSessions() []*ChatSession {
	csl.mu.RLock()
	defer csl.mu.RUnlock()

	var archived []*ChatSession
	for _, s := range csl.sessions {
		if s.IsArchived {
			archived = append(archived, s)
		}
	}
	return archived
}
