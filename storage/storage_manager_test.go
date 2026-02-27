package storage

import (
	"os"
	"testing"
	"time"
)

func setupTestStorage(t *testing.T) (*SQLiteStorageManager, func()) {
	tmpDir, err := os.MkdirTemp("", "o2ochat-storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	config := &StorageConfig{
		Type:        StorageTypeSQLite,
		Path:        tmpDir,
		MaxSize:     1024 * 1024 * 100,
		Compression: false,
		Encryption:  false,
		CacheSize:   64,
	}

	manager := NewSQLiteStorageManager()
	if err := manager.Initialize(config); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	cleanup := func() {
		manager.Close()
		os.RemoveAll(tmpDir)
	}

	return manager, cleanup
}

func TestStorageManager_Initialize(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	if !manager.initialized {
		t.Error("Storage manager should be initialized")
	}

	if manager.db == nil {
		t.Error("Database should not be nil")
	}

	if manager.messageStorage == nil {
		t.Error("MessageStorage should not be nil")
	}

	if manager.chunkStorage == nil {
		t.Error("ChunkStorage should not be nil")
	}

	if manager.configStorage == nil {
		t.Error("ConfigStorage should not be nil")
	}

	if manager.cacheManager == nil {
		t.Error("CacheManager should not be nil")
	}
}

func TestStorageManager_PutGet(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	key := "test:section:key1"
	value := []byte("test value 1")
	options := &StorageOptions{
		TTL:         time.Hour,
		Compression: false,
		Encryption:  false,
	}

	err := manager.Put(key, value, options)
	if err != nil {
		t.Fatalf("Failed to put value: %v", err)
	}

	retrieved, err := manager.Get(key)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}
}

func TestStorageManager_Delete(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	key := "test:section:key2"
	value := []byte("test value 2")

	err := manager.Put(key, value, nil)
	if err != nil {
		t.Fatalf("Failed to put value: %v", err)
	}

	exists, err := manager.Exists(key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Key should exist")
	}

	err = manager.Delete(key)
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	exists, err = manager.Exists(key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Key should not exist after deletion")
	}
}

func TestStorageManager_BatchOperations(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	entries := map[string][]byte{
		"batch:key1": []byte("value1"),
		"batch:key2": []byte("value2"),
		"batch:key3": []byte("value3"),
	}

	err := manager.BatchPut(entries, nil)
	if err != nil {
		t.Fatalf("Failed to batch put: %v", err)
	}

	keys := []string{"batch:key1", "batch:key2", "batch:key3"}
	retrieved, err := manager.BatchGet(keys)
	if err != nil {
		t.Fatalf("Failed to batch get: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(retrieved))
	}

	for i, key := range keys {
		expectedKey := key
		expectedValue := entries[key]
		if value, exists := retrieved[expectedKey]; !exists {
			t.Errorf("Key %s not found in batch get", expectedKey)
		} else if string(value) != string(expectedValue) {
			t.Errorf("Key %s: expected %s, got %s", expectedKey, string(expectedValue), string(value))
		}
		_ = i
	}

	err = manager.BatchDelete(keys)
	if err != nil {
		t.Fatalf("Failed to batch delete: %v", err)
	}

	for _, key := range keys {
		exists, _ := manager.Exists(key)
		if exists {
			t.Errorf("Key %s should not exist after batch delete", key)
		}
	}
}

func TestStorageManager_Exists(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	key := "test:exists:key"
	value := []byte("exists test")

	exists, err := manager.Exists(key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Key should not exist before put")
	}

	err = manager.Put(key, value, nil)
	if err != nil {
		t.Fatalf("Failed to put value: %v", err)
	}

	exists, err = manager.Exists(key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Key should exist after put")
	}
}

func TestStorageManager_List(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	entries := map[string][]byte{
		"list:section:key1": []byte("value1"),
		"list:section:key2": []byte("value2"),
		"list:section:key3": []byte("value3"),
		"list:other:key1":   []byte("value4"),
	}

	err := manager.BatchPut(entries, nil)
	if err != nil {
		t.Fatalf("Failed to batch put: %v", err)
	}

	keys, err := manager.List("list:section")
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d: %v", len(keys), keys)
	}
}

func TestStorageManager_GetStats(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		key := "stats:key" + string(rune(i))
		value := []byte("value" + string(rune(i)))
		manager.Put(key, value, nil)
	}

	stats, err := manager.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if stats.TotalSize == 0 {
		t.Error("TotalSize should not be zero")
	}
}

func TestStorageManager_CleanupExpired(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	key := "ttl:expired:key"
	value := []byte("ttl test value")
	options := &StorageOptions{
		TTL: 100 * time.Millisecond,
	}

	err := manager.Put(key, value, options)
	if err != nil {
		t.Fatalf("Failed to put value with TTL: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	err = manager.CleanupExpired()
	if err != nil {
		t.Fatalf("Failed to cleanup expired: %v", err)
	}
}

func TestStorageManager_BackupRestore(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	key := "backup:key"
	value := []byte("backup test value")

	err := manager.Put(key, value, nil)
	if err != nil {
		t.Fatalf("Failed to put value: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "o2ochat-backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backupPath := tmpDir + "/backup.db"
	err = manager.Backup(backupPath)
	if err != nil {
		t.Fatalf("Failed to backup: %v", err)
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("Backup file should exist")
	}
}

func TestStorageManager_Close(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	err := manager.Close()
	if err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	if !manager.closed {
		t.Error("Manager should be closed")
	}

	err = manager.Put("test:key", []byte("test"), nil)
	if err != ErrStorageNotInitialized {
		t.Error("Should return ErrStorageNotInitialized after close")
	}
}

func TestStorageManager_MessageStorage(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	msgStorage := manager.GetMessageStorage()
	if msgStorage == nil {
		t.Fatal("MessageStorage should not be nil")
	}

	message := &ChatMessage{
		ID:        "msg-test-1",
		From:      "QmPeer123",
		To:        "QmPeer456",
		Content:   []byte("Hello, World!"),
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Delivered: false,
		Read:      false,
		Encrypted: true,
	}

	err := msgStorage.StoreMessage(message)
	if err != nil {
		t.Fatalf("Failed to store message: %v", err)
	}

	retrieved, err := msgStorage.GetMessage(message.ID)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	if retrieved.ID != message.ID {
		t.Errorf("Expected ID %s, got %s", message.ID, retrieved.ID)
	}

	if string(retrieved.Content) != string(message.Content) {
		t.Errorf("Expected content %s, got %s", string(message.Content), string(retrieved.Content))
	}

	if retrieved.From != message.From {
		t.Errorf("Expected From %s, got %s", message.From, retrieved.From)
	}

	if retrieved.To != message.To {
		t.Errorf("Expected To %s, got %s", message.To, retrieved.To)
	}
}

func TestStorageManager_ConfigStorage(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	configStorage := manager.GetConfigStorage()
	if configStorage == nil {
		t.Fatal("ConfigStorage should not be nil")
	}

	err := configStorage.StoreConfig("app", "theme", "dark")
	if err != nil {
		t.Fatalf("Failed to store config: %v", err)
	}

	value, err := configStorage.GetConfig("app", "theme", "light")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if value != "dark" {
		t.Errorf("Expected theme 'dark', got '%v'", value)
	}

	allConfig, err := configStorage.GetAllConfig("app")
	if err != nil {
		t.Fatalf("Failed to get all config: %v", err)
	}

	if _, exists := allConfig["theme"]; !exists {
		t.Error("theme should exist in config")
	}
}

func TestStorageManager_CacheManager(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	cache := manager.GetCacheManager()
	if cache == nil {
		t.Fatal("CacheManager should not be nil")
	}

	key := "cache:test:key"
	value := []byte("cache test value")
	ttl := time.Hour

	err := cache.Set(key, value, ttl)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	retrieved, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}

	stats, err := cache.GetCacheStats()
	if err != nil {
		t.Fatalf("Failed to get cache stats: %v", err)
	}

	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
}
