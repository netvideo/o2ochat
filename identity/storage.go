package identity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type fileKeyStorage struct {
	storagePath string
	mu          sync.RWMutex
}

func NewFileKeyStorage(storagePath string) (KeyStorage, error) {
	if err := os.MkdirAll(storagePath, 0700); err != nil {
		return nil, err
	}

	return &fileKeyStorage{
		storagePath: storagePath,
	}, nil
}

func (s *fileKeyStorage) SavePrivateKey(peerID string, encryptedKey []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.storagePath, peerID+".key")
	return os.WriteFile(filename, encryptedKey, 0600)
}

func (s *fileKeyStorage) LoadPrivateKey(peerID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := filepath.Join(s.storagePath, peerID+".key")
	return os.ReadFile(filename)
}

func (s *fileKeyStorage) DeletePrivateKey(peerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.storagePath, peerID+".key")
	return os.Remove(filename)
}

func (s *fileKeyStorage) KeyExists(peerID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := filepath.Join(s.storagePath, peerID+".key")
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *fileKeyStorage) ListKeys() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.storagePath)
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".key" {
			keys = append(keys, entry.Name()[:len(entry.Name())-4])
		}
	}

	return keys, nil
}

type memoryKeyStorage struct {
	keys map[string][]byte
	mu   sync.RWMutex
}

func NewMemoryKeyStorage() KeyStorage {
	return &memoryKeyStorage{
		keys: make(map[string][]byte),
	}
}

func (s *memoryKeyStorage) SavePrivateKey(peerID string, encryptedKey []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[peerID] = encryptedKey
	return nil
}

func (s *memoryKeyStorage) LoadPrivateKey(peerID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.keys[peerID]
	if !exists {
		return nil, ErrIdentityNotFound
	}
	return key, nil
}

func (s *memoryKeyStorage) DeletePrivateKey(peerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.keys, peerID)
	return nil
}

func (s *memoryKeyStorage) KeyExists(peerID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.keys[peerID]
	return exists, nil
}

func (s *memoryKeyStorage) ListKeys() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peerIDs := make([]string, 0, len(s.keys))
	for peerID := range s.keys {
		peerIDs = append(peerIDs, peerID)
	}
	return peerIDs, nil
}

type memoryIdentityStore struct {
	identities map[string]*Identity
	metadata   map[string]*IdentityMetadata
	mu         sync.RWMutex
}

func NewMemoryIdentityStore() IdentityStore {
	return &memoryIdentityStore{
		identities: make(map[string]*Identity),
		metadata:   make(map[string]*IdentityMetadata),
	}
}

func (s *memoryIdentityStore) Save(identity *Identity, metadata *IdentityMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.identities[identity.PeerID] = identity
	if metadata != nil {
		s.metadata[identity.PeerID] = metadata
	}

	return nil
}

func (s *memoryIdentityStore) Load(peerID string) (*Identity, *IdentityMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	identity, exists := s.identities[peerID]
	if !exists {
		return nil, nil, ErrIdentityNotFound
	}

	metadata := s.metadata[peerID]
	return identity, metadata, nil
}

func (s *memoryIdentityStore) Delete(peerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.identities[peerID]; !exists {
		return ErrIdentityNotFound
	}

	delete(s.identities, peerID)
	delete(s.metadata, peerID)

	return nil
}

func (s *memoryIdentityStore) Exists(peerID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.identities[peerID]
	return exists, nil
}

func (s *memoryIdentityStore) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peerIDs := make([]string, 0, len(s.identities))
	for peerID := range s.identities {
		peerIDs = append(peerIDs, peerID)
	}

	return peerIDs, nil
}

type fileIdentityStore struct {
	storagePath string
	mu          sync.RWMutex
}

func NewFileIdentityStore(storagePath string) (IdentityStore, error) {
	if err := os.MkdirAll(storagePath, 0700); err != nil {
		return nil, err
	}

	return &fileIdentityStore{
		storagePath: storagePath,
	}, nil
}

func (s *fileIdentityStore) Save(identity *Identity, metadata *IdentityMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	identityFile := filepath.Join(s.storagePath, identity.PeerID+".json")
	data := struct {
		Identity *Identity         `json:"identity"`
		Metadata *IdentityMetadata `json:"metadata"`
	}{
		Identity: identity,
		Metadata: metadata,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(identityFile, jsonData, 0600)
}

func (s *fileIdentityStore) Load(peerID string) (*Identity, *IdentityMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	identityFile := filepath.Join(s.storagePath, peerID+".json")
	data, err := os.ReadFile(identityFile)
	if err != nil {
		return nil, nil, ErrIdentityNotFound
	}

	var stored struct {
		Identity *Identity         `json:"identity"`
		Metadata *IdentityMetadata `json:"metadata"`
	}

	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, nil, ErrInvalidFormat
	}

	return stored.Identity, stored.Metadata, nil
}

func (s *fileIdentityStore) Delete(peerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	identityFile := filepath.Join(s.storagePath, peerID+".json")
	return os.Remove(identityFile)
}

func (s *fileIdentityStore) Exists(peerID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	identityFile := filepath.Join(s.storagePath, peerID+".json")
	_, err := os.Stat(identityFile)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *fileIdentityStore) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.storagePath)
	if err != nil {
		return nil, err
	}

	var peerIDs []string
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			peerIDs = append(peerIDs, entry.Name()[:len(entry.Name())-5])
		}
	}

	return peerIDs, nil
}
