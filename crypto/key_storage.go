package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	ErrKeyExists = errors.New("crypto: key already exists")
)

type keyEntry struct {
	Key       []byte            `json:"key"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt time.Time         `json:"expires_at"`
	Encrypted bool              `json:"encrypted"`
}

type memoryKeyStorage struct {
	keys      sync.Map
	masterKey []byte
	mu        sync.RWMutex
}

type fileKeyStorage struct {
	keys      sync.Map
	dataDir   string
	masterKey []byte
	mu        sync.RWMutex
}

func NewMemoryKeyStorage(masterKey []byte) KeyStorage {
	return &memoryKeyStorage{
		masterKey: masterKey,
	}
}

func NewFileKeyStorage(dataDir string, masterKey []byte) (KeyStorage, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}

	storage := &fileKeyStorage{
		dataDir:   dataDir,
		masterKey: masterKey,
	}

	if err := storage.loadKeys(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (m *memoryKeyStorage) StoreKey(keyID string, key []byte, metadata map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.keys.Load(keyID); exists {
		return ErrKeyExists
	}

	encryptedKey, err := m.encryptKey(key)
	if err != nil {
		return err
	}

	entry := &keyEntry{
		Key:       encryptedKey,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
		Encrypted: true,
	}

	m.keys.Store(keyID, entry)
	return nil
}

func (m *memoryKeyStorage) GetKey(keyID string) ([]byte, map[string]string, error) {
	val, ok := m.keys.Load(keyID)
	if !ok {
		return nil, nil, ErrKeyNotFound
	}

	entry := val.(*keyEntry)

	if time.Now().After(entry.ExpiresAt) {
		m.keys.Delete(keyID)
		return nil, nil, ErrKeyNotFound
	}

	key, err := m.decryptKey(entry.Key)
	if err != nil {
		return nil, nil, err
	}

	return key, entry.Metadata, nil
}

func (m *memoryKeyStorage) DeleteKey(keyID string) error {
	_, ok := m.keys.Load(keyID)
	if !ok {
		return ErrKeyNotFound
	}

	m.keys.Delete(keyID)
	return nil
}

func (m *memoryKeyStorage) ListKeys() ([]string, error) {
	var keys []string
	m.keys.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys, nil
}

func (m *memoryKeyStorage) KeyExists(keyID string) (bool, error) {
	_, exists := m.keys.Load(keyID)
	return exists, nil
}

func (m *memoryKeyStorage) CleanupExpiredKeys() error {
	now := time.Now()
	m.keys.Range(func(key, value interface{}) bool {
		entry := value.(*keyEntry)
		if now.After(entry.ExpiresAt) {
			m.keys.Delete(key)
		}
		return true
	})
	return nil
}

func (m *memoryKeyStorage) encryptKey(key []byte) ([]byte, error) {
	if m.masterKey == nil {
		return key, nil
	}

	block, err := aes.NewCipher(m.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, key, nil)
	return ciphertext, nil
}

func (m *memoryKeyStorage) decryptKey(data []byte) ([]byte, error) {
	if m.masterKey == nil {
		return data, nil
	}

	block, err := aes.NewCipher(m.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (f *fileKeyStorage) StoreKey(keyID string, key []byte, metadata map[string]string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.keys.Load(keyID); exists {
		return ErrKeyExists
	}

	encryptedKey, err := f.encryptKey(key)
	if err != nil {
		return err
	}

	entry := &keyEntry{
		Key:       encryptedKey,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
		Encrypted: true,
	}

	f.keys.Store(keyID, entry)
	return f.saveKeyToFile(keyID, entry)
}

func (f *fileKeyStorage) GetKey(keyID string) ([]byte, map[string]string, error) {
	val, ok := f.keys.Load(keyID)
	if !ok {
		if err := f.loadKeyFromFile(keyID); err != nil {
			return nil, nil, ErrKeyNotFound
		}
		val, ok = f.keys.Load(keyID)
		if !ok {
			return nil, nil, ErrKeyNotFound
		}
	}

	entry := val.(*keyEntry)

	if time.Now().After(entry.ExpiresAt) {
		f.keys.Delete(keyID)
		return nil, nil, ErrKeyNotFound
	}

	key, err := f.decryptKey(entry.Key)
	if err != nil {
		return nil, nil, err
	}

	return key, entry.Metadata, nil
}

func (f *fileKeyStorage) DeleteKey(keyID string) error {
	_, ok := f.keys.Load(keyID)
	if !ok {
		return ErrKeyNotFound
	}

	f.keys.Delete(keyID)

	keyFile := filepath.Join(f.dataDir, keyID+".json")
	return os.Remove(keyFile)
}

func (f *fileKeyStorage) ListKeys() ([]string, error) {
	var keys []string
	f.keys.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys, nil
}

func (f *fileKeyStorage) KeyExists(keyID string) (bool, error) {
	_, exists := f.keys.Load(keyID)
	if !exists {
		keyFile := filepath.Join(f.dataDir, keyID+".json")
		if _, err := os.Stat(keyFile); err == nil {
			return true, nil
		}
	}
	return exists, nil
}

func (f *fileKeyStorage) CleanupExpiredKeys() error {
	now := time.Now()
	f.keys.Range(func(key, value interface{}) bool {
		entry := value.(*keyEntry)
		if now.After(entry.ExpiresAt) {
			f.keys.Delete(key)
			keyFile := filepath.Join(f.dataDir, key.(string)+".json")
			os.Remove(keyFile)
		}
		return true
	})
	return nil
}

func (f *fileKeyStorage) saveKeyToFile(keyID string, entry *keyEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	keyFile := filepath.Join(f.dataDir, keyID+".json")
	return os.WriteFile(keyFile, data, 0600)
}

func (f *fileKeyStorage) loadKeyFromFile(keyID string) error {
	keyFile := filepath.Join(f.dataDir, keyID+".json")
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return err
	}

	var entry keyEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return err
	}

	f.keys.Store(keyID, &entry)
	return nil
}

func (f *fileKeyStorage) loadKeys() error {
	files, err := os.ReadDir(f.dataDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			keyID := file.Name()[:len(file.Name())-5]
			f.loadKeyFromFile(keyID)
		}
	}
	return nil
}

func (f *fileKeyStorage) encryptKey(key []byte) ([]byte, error) {
	if f.masterKey == nil {
		return key, nil
	}

	block, err := aes.NewCipher(f.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, key, nil)
	return ciphertext, nil
}

func (f *fileKeyStorage) decryptKey(data []byte) ([]byte, error) {
	if f.masterKey == nil {
		return data, nil
	}

	block, err := aes.NewCipher(f.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func EncodeKeyToBase64(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func DecodeKeyFromBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}
