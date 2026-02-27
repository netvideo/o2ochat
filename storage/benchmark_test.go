package storage

import (
	"os"
	"testing"
	"time"
)

func BenchmarkStorageManager_Put(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	key := "bench:put:key"
	value := make([]byte, 1024)
	for i := range value {
		value[i] = byte(i % 256)
	}
	options := &StorageOptions{
		TTL:         time.Hour,
		Compression: false,
		Encryption:  false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Put(key, value, options)
	}
}

func BenchmarkStorageManager_Get(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	key := "bench:get:key"
	value := make([]byte, 1024)
	for i := range value {
		value[i] = byte(i % 256)
	}
	manager.Put(key, value, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Get(key)
	}
}

func BenchmarkStorageManager_BatchPut(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	entries := make(map[string][]byte)
	for i := 0; i < 100; i++ {
		key := "bench:batch:" + string(rune(i))
		value := make([]byte, 256)
		for j := range value {
			value[j] = byte(j % 256)
		}
		entries[key] = value
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.BatchPut(entries, nil)
	}
}

func BenchmarkStorageManager_BatchGet(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	keys := make([]string, 100)
	entries := make(map[string][]byte)
	for i := 0; i < 100; i++ {
		key := "bench:batch:get:" + string(rune(i))
		value := make([]byte, 256)
		for j := range value {
			value[j] = byte(j % 256)
		}
		entries[key] = value
		keys[i] = key
	}
	manager.BatchPut(entries, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.BatchGet(keys)
	}
}

func BenchmarkCacheManager_Set(b *testing.B) {
	cache := NewLRUCacheManager(1024 * 1024 * 100)
	defer cache.Close()

	key := "bench:cache:set"
	value := make([]byte, 1024)
	for i := range value {
		value[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(key, value, time.Hour)
	}
}

func BenchmarkCacheManager_Get(b *testing.B) {
	cache := NewLRUCacheManager(1024 * 1024 * 100)
	defer cache.Close()

	key := "bench:cache:get"
	value := make([]byte, 1024)
	for i := range value {
		value[i] = byte(i % 256)
	}
	cache.Set(key, value, time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(key)
	}
}

func BenchmarkMessageStorage_StoreMessage(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	msgStorage := manager.GetMessageStorage()
	message := &ChatMessage{
		ID:        "",
		From:      "QmPeerBenchmark",
		To:        "QmPeerOther",
		Content:   make([]byte, 512),
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Delivered: false,
		Read:      false,
		Encrypted: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message.ID = "bench:msg:" + string(rune(i))
		msgStorage.StoreMessage(message)
	}
}

func BenchmarkMessageStorage_GetMessage(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	msgStorage := manager.GetMessageStorage()
	message := &ChatMessage{
		ID:        "bench:get:msg",
		From:      "QmPeerBenchmark",
		To:        "QmPeerOther",
		Content:   make([]byte, 512),
		Type:      MessageTypeText,
		Timestamp: time.Now(),
		Delivered: true,
		Read:      true,
		Encrypted: true,
	}
	msgStorage.StoreMessage(message)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgStorage.GetMessage("bench:get:msg")
	}
}

func BenchmarkConfigStorage_StoreConfig(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	configStorage := manager.GetConfigStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key" + string(rune(i%256))
		configStorage.StoreConfig("benchmark", key, "value")
	}
}

func BenchmarkConfigStorage_GetConfig(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	configStorage := manager.GetConfigStorage()
	for i := 0; i < 100; i++ {
		key := "bench:config:get:" + string(rune(i))
		configStorage.StoreConfig("benchmark", key, "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configStorage.GetConfig("benchmark", "bench:config:get:50", "default")
	}
}

func BenchmarkChunkStorage_StoreChunk(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	chunkStorage := manager.GetChunkStorage()
	chunkData := make([]byte, 1024*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunkStorage.StoreChunk("bench:file", i, chunkData)
	}
}

func BenchmarkChunkStorage_GetChunk(b *testing.B) {
	manager, cleanup := setupBenchmarkStorage(b)
	defer cleanup()

	chunkStorage := manager.GetChunkStorage()
	chunkData := make([]byte, 1024*1024)
	chunkStorage.StoreChunk("bench:file:get", 0, chunkData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunkStorage.GetChunk("bench:file:get", 0)
	}
}

func setupBenchmarkStorage(b *testing.B) (*SQLiteStorageManager, func()) {
	tmpDir, err := os.MkdirTemp("", "o2ochat-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}

	config := &StorageConfig{
		Type:        StorageTypeSQLite,
		Path:        tmpDir,
		MaxSize:     1024 * 1024 * 1024,
		Compression: false,
		Encryption:  false,
		CacheSize:   256,
	}

	manager := NewSQLiteStorageManager()
	if err := manager.Initialize(config); err != nil {
		os.RemoveAll(tmpDir)
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	cleanup := func() {
		manager.Close()
		os.RemoveAll(tmpDir)
	}

	return manager, cleanup
}
