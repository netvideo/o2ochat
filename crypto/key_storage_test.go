package crypto

import (
	"bytes"
	"testing"
)

func TestNewMemoryKeyStorage(t *testing.T) {
	masterKey := make([]byte, 32)
	storage := NewMemoryKeyStorage(masterKey)
	if storage == nil {
		t.Fatal("Expected storage to be created")
	}
}

func TestMemoryKeyStorageStoreGet(t *testing.T) {
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	storage := NewMemoryKeyStorage(masterKey)

	key := []byte("secret key")
	metadata := map[string]string{"purpose": "test"}

	err := storage.StoreKey("test-key", key, metadata)
	if err != nil {
		t.Fatalf("StoreKey failed: %v", err)
	}

	retrievedKey, retrievedMeta, err := storage.GetKey("test-key")
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}

	if !bytes.Equal(key, retrievedKey) {
		t.Error("Retrieved key does not match")
	}

	if retrievedMeta["purpose"] != "test" {
		t.Error("Metadata does not match")
	}
}

func TestMemoryKeyStorageDuplicate(t *testing.T) {
	storage := NewMemoryKeyStorage(nil)

	err := storage.StoreKey("key1", []byte("key"), nil)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.StoreKey("key1", []byte("key2"), nil)
	if err != ErrKeyExists {
		t.Errorf("Expected ErrKeyExists, got: %v", err)
	}
}

func TestMemoryKeyStorageNotFound(t *testing.T) {
	storage := NewMemoryKeyStorage(nil)

	_, _, err := storage.GetKey("nonexistent")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got: %v", err)
	}
}

func TestMemoryKeyStorageDelete(t *testing.T) {
	storage := NewMemoryKeyStorage(nil)

	storage.StoreKey("key1", []byte("key"), nil)

	err := storage.DeleteKey("key1")
	if err != nil {
		t.Fatalf("DeleteKey failed: %v", err)
	}

	_, _, err = storage.GetKey("key1")
	if err != ErrKeyNotFound {
		t.Error("Expected key to be deleted")
	}
}

func TestMemoryKeyStorageList(t *testing.T) {
	storage := NewMemoryKeyStorage(nil)

	storage.StoreKey("key1", []byte("key1"), nil)
	storage.StoreKey("key2", []byte("key2"), nil)
	storage.StoreKey("key3", []byte("key3"), nil)

	keys, err := storage.ListKeys()
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
}

func TestMemoryKeyStorageExists(t *testing.T) {
	storage := NewMemoryKeyStorage(nil)

	storage.StoreKey("key1", []byte("key"), nil)

	exists, err := storage.KeyExists("key1")
	if err != nil {
		t.Fatalf("KeyExists failed: %v", err)
	}

	if !exists {
		t.Error("Expected key to exist")
	}

	exists, _ = storage.KeyExists("nonexistent")
	if exists {
		t.Error("Expected key to not exist")
	}
}

func TestFileKeyStorage(t *testing.T) {
	tempDir := t.TempDir()
	masterKey := make([]byte, 32)

	storage, err := NewFileKeyStorage(tempDir, masterKey)
	if err != nil {
		t.Fatalf("Failed to create file storage: %v", err)
	}

	key := []byte("secret")
	err = storage.StoreKey("test-key", key, map[string]string{"test": "value"})
	if err != nil {
		t.Fatalf("StoreKey failed: %v", err)
	}

	retrievedKey, _, err := storage.GetKey("test-key")
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}

	if !bytes.Equal(key, retrievedKey) {
		t.Error("Retrieved key does not match")
	}
}

func TestFileKeyStoragePersistence(t *testing.T) {
	tempDir := t.TempDir()
	masterKey := make([]byte, 32)

	storage1, _ := NewFileKeyStorage(tempDir, masterKey)
	storage1.StoreKey("persistent-key", []byte("persistent"), nil)

	storage2, _ := NewFileKeyStorage(tempDir, masterKey)
	key, _, err := storage2.GetKey("persistent-key")
	if err != nil {
		t.Fatalf("Failed to load persisted key: %v", err)
	}

	if !bytes.Equal([]byte("persistent"), key) {
		t.Error("Persisted key does not match")
	}
}

func TestEncodeDecodeBase64(t *testing.T) {
	key := []byte("test key data")

	encoded := EncodeKeyToBase64(key)
	if encoded == "" {
		t.Error("Expected non-empty encoded key")
	}

	decoded, err := DecodeKeyFromBase64(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !bytes.Equal(key, decoded) {
		t.Error("Decoded key does not match")
	}
}

func TestCleanupExpiredKeys(t *testing.T) {
	storage := NewMemoryKeyStorage(nil)

	storage.StoreKey("key1", []byte("key1"), nil)

	err := storage.CleanupExpiredKeys()
	if err != nil {
		t.Errorf("CleanupExpiredKeys failed: %v", err)
	}
}
