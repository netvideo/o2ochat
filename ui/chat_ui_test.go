package ui

import (
	"testing"
	"time"
)

func TestNewChatUI(t *testing.T) {
	chat := NewChatUI()
	if chat == nil {
		t.Error("expected non-nil ChatUI")
	}

	defaultChat, ok := chat.(*DefaultChatUI)
	if !ok {
		t.Error("expected DefaultChatUI type")
	}

	if defaultChat.openChats == nil {
		t.Error("expected openChats map to be initialized")
	}
	if defaultChat.messages == nil {
		t.Error("expected messages map to be initialized")
	}
}

func TestChatUIOpenChat(t *testing.T) {
	chat := NewChatUI()

	err := chat.OpenChat("QmPeer123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = chat.OpenChat("")
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestChatUICloseChat(t *testing.T) {
	chat := NewChatUI()

	err := chat.CloseChat("QmPeer123")
	if err != ErrChatNotFound {
		t.Errorf("expected ErrChatNotFound, got %v", err)
	}

	chat.OpenChat("QmPeer123")
	err = chat.CloseChat("QmPeer123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatUIAddMessage(t *testing.T) {
	chat := NewChatUI()

	msg := &MessageItem{
		ID:        "msg1",
		From:      "QmPeer456",
		To:        "QmPeer123",
		Content:   "Hello",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Status:    MessageStatusSent,
		IsOwn:     false,
	}

	err := chat.AddMessage(msg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = chat.AddMessage(nil)
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestChatUIUpdateMessageStatus(t *testing.T) {
	chat := NewChatUI()

	msg := &MessageItem{
		ID:        "msg1",
		From:      "QmPeer456",
		To:        "QmPeer123",
		Content:   "Hello",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Status:    MessageStatusSent,
		IsOwn:     false,
	}
	chat.AddMessage(msg)

	err := chat.UpdateMessageStatus("msg1", MessageStatusDelivered)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = chat.UpdateMessageStatus("nonexistent", MessageStatusDelivered)
	if err != ErrMessageNotFound {
		t.Errorf("expected ErrMessageNotFound, got %v", err)
	}
}

func TestChatUIClearChat(t *testing.T) {
	chat := NewChatUI()

	err := chat.ClearChat("QmPeer123")
	if err != ErrChatNotFound {
		t.Errorf("expected ErrChatNotFound, got %v", err)
	}

	msg := &MessageItem{
		ID:        "msg1",
		From:      "QmPeer456",
		To:        "QmPeer123",
		Content:   "Hello",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	chat.AddMessage(msg)

	err = chat.ClearChat("QmPeer123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatUISearchMessages(t *testing.T) {
	chat := NewChatUI()

	msg1 := &MessageItem{
		ID:        "msg1",
		From:      "QmPeer456",
		To:        "QmPeer123",
		Content:   "Hello world",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	msg2 := &MessageItem{
		ID:        "msg2",
		From:      "QmPeer456",
		To:        "QmPeer123",
		Content:   "Goodbye",
		Type:      MessageTypeText,
		Timestamp: time.Now(),
	}
	chat.AddMessage(msg1)
	chat.AddMessage(msg2)

	results, err := chat.SearchMessages("QmPeer123", "Hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	results, err = chat.SearchMessages("QmPeer123", "nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}

	_, err = chat.SearchMessages("QmPeer999", "Hello")
	if err != ErrChatNotFound {
		t.Errorf("expected ErrChatNotFound, got %v", err)
	}
}

func TestChatUIGetChatHistory(t *testing.T) {
	chat := NewChatUI()

	for i := 0; i < 5; i++ {
		msg := &MessageItem{
			ID:        string(rune('0' + i)),
			From:      "QmPeer456",
			To:        "QmPeer123",
			Content:   "Message",
			Type:      MessageTypeText,
			Timestamp: time.Now(),
		}
		chat.AddMessage(msg)
	}

	history, err := chat.GetChatHistory("QmPeer123", 3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(history) != 3 {
		t.Errorf("expected 3 messages, got %d", len(history))
	}

	history, err = chat.GetChatHistory("QmPeer123", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(history) != 5 {
		t.Errorf("expected 5 messages, got %d", len(history))
	}

	_, err = chat.GetChatHistory("QmPeer999", 10)
	if err != ErrChatNotFound {
		t.Errorf("expected ErrChatNotFound, got %v", err)
	}
}

func TestChatUISetInputCallback(t *testing.T) {
	chat := NewChatUI()

	callbackCalled := false
	callback := func(text string, attachments []string) {
		callbackCalled = true
	}

	err := chat.SetInputCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultChat := chat.(*DefaultChatUI)
	defaultChat.inputCallback("test", nil)
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}

func TestChatUISetReactionCallback(t *testing.T) {
	chat := NewChatUI()

	callbackCalled := false
	callback := func(messageID string, reaction string) {
		callbackCalled = true
	}

	err := chat.SetReactionCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultChat := chat.(*DefaultChatUI)
	defaultChat.reactionCallback("msg1", "👍")
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}
