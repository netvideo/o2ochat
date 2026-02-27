package storage

import (
	"testing"
	"time"
)

func TestMessageStorage_BasicOperations(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	msgStorage := manager.GetMessageStorage()

	message := &ChatMessage{
		ID:        "msg-001",
		From:      "QmSender123",
		To:        "QmReceiver456",
		Content:   []byte("Hello, this is a test message!"),
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Delivered: false,
		Read:      false,
		Encrypted: true,
	}

	t.Run("StoreMessage", func(t *testing.T) {
		err := msgStorage.StoreMessage(message)
		if err != nil {
			t.Fatalf("Failed to store message: %v", err)
		}
	})

	t.Run("GetMessage", func(t *testing.T) {
		retrieved, err := msgStorage.GetMessage(message.ID)
		if err != nil {
			t.Fatalf("Failed to get message: %v", err)
		}

		if retrieved.ID != message.ID {
			t.Errorf("ID mismatch: expected %s, got %s", message.ID, retrieved.ID)
		}

		if retrieved.From != message.From {
			t.Errorf("From mismatch: expected %s, got %s", message.From, retrieved.From)
		}

		if retrieved.To != message.To {
			t.Errorf("To mismatch: expected %s, got %s", message.To, retrieved.To)
		}
	})

	t.Run("GetMessageNotFound", func(t *testing.T) {
		_, err := msgStorage.GetMessage("nonexistent")
		if err != ErrMessageNotFound {
			t.Errorf("Expected ErrMessageNotFound, got %v", err)
		}
	})
}

func TestMessageStorage_ConversationMessages(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	msgStorage := manager.GetMessageStorage()

	peerID := "QmPeer123"
	for i := 0; i < 5; i++ {
		message := &ChatMessage{
			ID:        string(rune('a' + i)),
			From:      peerID,
			To:        "QmOtherPeer",
			Content:   []byte("Message " + string(rune('0'+i))),
			Type:      MessageTypeText,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Delivered: true,
			Read:      i > 2,
			Encrypted: true,
		}
		msgStorage.StoreMessage(message)
	}

	t.Run("GetConversationMessages", func(t *testing.T) {
		messages, err := msgStorage.GetConversationMessages(peerID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get conversation messages: %v", err)
		}

		if len(messages) != 5 {
			t.Errorf("Expected 5 messages, got %d", len(messages))
		}
	})

	t.Run("GetMessageStats", func(t *testing.T) {
		stats, err := msgStorage.GetMessageStats(peerID)
		if err != nil {
			t.Fatalf("Failed to get message stats: %v", err)
		}

		if stats.TotalCount != 5 {
			t.Errorf("Expected total count 5, got %d", stats.TotalCount)
		}

		if stats.UnreadCount != 3 {
			t.Errorf("Expected unread count 3, got %d", stats.UnreadCount)
		}
	})
}

func TestMessageStorage_SearchMessages(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	msgStorage := manager.GetMessageStorage()

	peerID := "QmSearchPeer"
	messages := []string{"apple", "banana", "cherry", "date"}
	for i, content := range messages {
		msg := &ChatMessage{
			ID:        string(rune('a' + i)),
			From:      peerID,
			To:        "QmOther",
			Content:   []byte(content),
			Type:      MessageTypeText,
			Timestamp: time.Now(),
			Delivered: true,
			Read:      true,
			Encrypted: false,
		}
		msgStorage.StoreMessage(msg)
	}

	t.Run("SearchMessages", func(t *testing.T) {
		results, err := msgStorage.SearchMessages("a", peerID, 10)
		if err != nil {
			t.Fatalf("Failed to search messages: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results (apple, banana, date), got %d", len(results))
		}
	})
}

func TestMessageStorage_DeleteMessage(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	msgStorage := manager.GetMessageStorage()

	message := &ChatMessage{
		ID:        "msg-delete",
		From:      "QmFrom",
		To:        "QmTo",
		Content:   []byte("To be deleted"),
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Delivered: false,
		Read:      false,
		Encrypted: true,
	}

	msgStorage.StoreMessage(message)

	t.Run("DeleteMessage", func(t *testing.T) {
		err := msgStorage.DeleteMessage(message.ID)
		if err != nil {
			t.Fatalf("Failed to delete message: %v", err)
		}

		_, err = msgStorage.GetMessage(message.ID)
		if err != ErrMessageNotFound {
			t.Errorf("Expected ErrMessageNotFound after delete, got %v", err)
		}
	})

	t.Run("DeleteNonExistent", func(t *testing.T) {
		err := msgStorage.DeleteMessage("nonexistent")
		if err != ErrMessageNotFound {
			t.Errorf("Expected ErrMessageNotFound, got %v", err)
		}
	})
}

func TestMessageStorage_CleanupOldMessages(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	msgStorage := manager.GetMessageStorage()

	now := time.Now()
	oldTime := now.AddDate(0, -2, 0)

	oldMessage := &ChatMessage{
		ID:        "old-msg",
		From:      "QmFrom",
		To:        "QmTo",
		Content:   []byte("Old message"),
		Type:      MessageTypeText,
		Timestamp: oldTime,
		Delivered: true,
		Read:      true,
		Encrypted: true,
	}

	newMessage := &ChatMessage{
		ID:        "new-msg",
		From:      "QmFrom",
		To:        "QmTo",
		Content:   []byte("New message"),
		Type:      MessageTypeText,
		Timestamp: now,
		Delivered: true,
		Read:      true,
		Encrypted: true,
	}

	msgStorage.StoreMessage(oldMessage)
	msgStorage.StoreMessage(newMessage)

	t.Run("CleanupOldMessages", func(t *testing.T) {
		cutoffTime := now.AddDate(0, -1, 0)
		err := msgStorage.CleanupOldMessages(cutoffTime)
		if err != nil {
			t.Fatalf("Failed to cleanup old messages: %v", err)
		}

		_, err = msgStorage.GetMessage(oldMessage.ID)
		if err != ErrMessageNotFound {
			t.Error("Old message should have been cleaned up")
		}

		_, err = msgStorage.GetMessage(newMessage.ID)
		if err != nil {
			t.Error("New message should still exist")
		}
	})
}

func TestMessageStorage_UpdateMessage(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	msgStorage := manager.GetMessageStorage()

	message := &ChatMessage{
		ID:        "msg-update",
		From:      "QmFrom",
		To:        "QmTo",
		Content:   []byte("Original content"),
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Delivered: false,
		Read:      false,
		Encrypted: true,
	}

	msgStorage.StoreMessage(message)

	message.Content = []byte("Updated content")
	message.Delivered = true
	message.Read = true

	msgStorage.StoreMessage(message)

	retrieved, err := msgStorage.GetMessage(message.ID)
	if err != nil {
		t.Fatalf("Failed to get updated message: %v", err)
	}

	if string(retrieved.Content) != "Updated content" {
		t.Errorf("Expected updated content, got %s", string(retrieved.Content))
	}

	if !retrieved.Delivered {
		t.Error("Expected delivered to be true")
	}

	if !retrieved.Read {
		t.Error("Expected read to be true")
	}
}
