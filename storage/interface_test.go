package storage

import (
	"testing"
	"time"
)

func TestStorageTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected StorageType
	}{
		{"SQLite", StorageTypeSQLite},
		{"BoltDB", StorageTypeBoltDB},
		{"LevelDB", StorageTypeLevelDB},
		{"Badger", StorageTypeBadger},
		{"Memory", StorageTypeMemory},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("storage type should not be empty")
			}
		})
	}
}

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected MessageType
	}{
		{"Text", MessageTypeText},
		{"Image", MessageTypeImage},
		{"File", MessageTypeFile},
		{"Voice", MessageTypeVoice},
		{"Video", MessageTypeVideo},
		{"System", MessageTypeSystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("message type should not be empty")
			}
		})
	}
}

func TestStorageConfig(t *testing.T) {
	config := DefaultStorageConfig()

	if config.Type != StorageTypeSQLite {
		t.Errorf("expected SQLite, got %v", config.Type)
	}
	if config.Path == "" {
		t.Error("path should not be empty")
	}
	if config.MaxSize <= 0 {
		t.Error("max size should be positive")
	}
	if config.CacheSize <= 0 {
		t.Error("cache size should be positive")
	}
}

func TestStorageOptions(t *testing.T) {
	options := DefaultStorageOptions()

	if options.TTL <= 0 {
		t.Error("TTL should be positive")
	}
}

func TestStorageEntry(t *testing.T) {
	entry := &StorageEntry{
		Key:       "test-key",
		Value:     []byte("test-value"),
		Metadata:  map[string]string{"type": "test"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Size:      10,
	}

	if entry.Key == "" {
		t.Error("key should not be empty")
	}
	if entry.Value == nil {
		t.Error("value should not be nil")
	}
	if entry.Size <= 0 {
		t.Error("size should be positive")
	}
}

func TestChatMessage(t *testing.T) {
	message := &ChatMessage{
		ID:        "msg-123",
		From:      "peer-1",
		To:        "peer-2",
		Content:   []byte("Hello"),
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Delivered: false,
		Read:      false,
		Encrypted: true,
	}

	if message.ID == "" {
		t.Error("ID should not be empty")
	}
	if message.From == "" {
		t.Error("from should not be empty")
	}
	if message.To == "" {
		t.Error("to should not be empty")
	}
}

func TestStorageStats(t *testing.T) {
	stats := &StorageStats{
		TotalSize:    1000,
		UsedSize:     500,
		FreeSize:     500,
		FileCount:    100,
		ChunkCount:   50,
		MessageCount: 200,
	}

	if stats.TotalSize <= 0 {
		t.Error("total size should be positive")
	}
	if stats.UsedSize < 0 {
		t.Error("used size should not be negative")
	}
}

func TestMessageStats(t *testing.T) {
	stats := &MessageStats{
		TotalCount:  100,
		UnreadCount: 10,
		TotalSize:   50000,
	}

	if stats.TotalCount < 0 {
		t.Error("total count should not be negative")
	}
	if stats.UnreadCount < 0 {
		t.Error("unread count should not be negative")
	}
}

func TestChunkStats(t *testing.T) {
	stats := &ChunkStats{
		TotalCount: 100,
		TotalSize:  104857600,
		ChunkCount: 100,
	}

	if stats.TotalCount < 0 {
		t.Error("total count should not be negative")
	}
	if stats.TotalSize < 0 {
		t.Error("total size should not be negative")
	}
}

func TestCacheStats(t *testing.T) {
	stats := &CacheStats{
		Size:      1024,
		Items:     100,
		Hits:      500,
		Misses:    50,
		Evictions: 10,
	}

	if stats.Size < 0 {
		t.Error("size should not be negative")
	}
	if stats.Items < 0 {
		t.Error("items should not be negative")
	}
}

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		err  error
		name string
	}{
		{ErrStorageNotInitialized, "ErrStorageNotInitialized"},
		{ErrKeyNotFound, "ErrKeyNotFound"},
		{ErrKeyExists, "ErrKeyExists"},
		{ErrInvalidKey, "ErrInvalidKey"},
		{ErrInvalidValue, "ErrInvalidValue"},
		{ErrStorageFull, "ErrStorageFull"},
		{ErrCompressionFailed, "ErrCompressionFailed"},
		{ErrDecompressionFailed, "ErrDecompressionFailed"},
		{ErrEncryptionFailed, "ErrEncryptionFailed"},
		{ErrDecryptionFailed, "ErrDecryptionFailed"},
		{ErrBackupFailed, "ErrBackupFailed"},
		{ErrRestoreFailed, "ErrRestoreFailed"},
		{ErrInvalidConfig, "ErrInvalidConfig"},
		{ErrInvalidPath, "ErrInvalidPath"},
		{ErrMessageNotFound, "ErrMessageNotFound"},
		{ErrChunkNotFound, "ErrChunkNotFound"},
		{ErrConfigNotFound, "ErrConfigNotFound"},
		{ErrCacheNotFound, "ErrCacheNotFound"},
		{ErrInvalidTTL, "ErrInvalidTTL"},
		{ErrDatabaseCorrupted, "ErrDatabaseCorrupted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("error message should not be empty")
			}
		})
	}
}

func TestStorageError(t *testing.T) {
	innerErr := ErrStorageNotInitialized
	storageErr := NewStorageError("STORAGE_INIT", "storage not initialized", innerErr)

	if storageErr.Code != "STORAGE_INIT" {
		t.Errorf("expected code STORAGE_INIT, got %s", storageErr.Code)
	}
	if storageErr.Message != "storage not initialized" {
		t.Errorf("expected message 'storage not initialized', got %s", storageErr.Message)
	}
	if storageErr.Unwrap() != innerErr {
		t.Error("unwrap should return inner error")
	}
	if storageErr.Error() == "" {
		t.Error("error should not be empty")
	}
}

func TestInterfaceCompatibility(t *testing.T) {
	var _ StorageManager = nil
	var _ MessageStorage = nil
	var _ ChunkStorage = nil
	var _ ConfigStorage = nil
	var _ CacheManager = nil
}
