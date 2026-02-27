package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// identityManager implements IdentityManager interface for managing user identities.
type identityManager struct {
	store     IdentityStore
	keyStore  KeyStorage
	peerIDGen PeerIDUtil
	mu        sync.RWMutex
}

// NewIdentityManager creates a new IdentityManager with the given stores.
func NewIdentityManager(store IdentityStore, keyStore KeyStorage) (IdentityManager, error) {
	return &identityManager{
		store:     store,
		keyStore:  keyStore,
		peerIDGen: NewPeerIDGenerator(),
	}, nil
}

// CreateIdentity creates a new identity with Ed25519 key pair.
func (m *identityManager) CreateIdentity(config *IdentityConfig) (*Identity, error) {
	if config == nil {
		config = DefaultIdentityConfig()
	}

	var publicKey ed25519.PublicKey
	var privateKey ed25519.PrivateKey
	var err error

	switch config.KeyType {
	case KeyTypeEd25519:
		publicKey, privateKey, err = ed25519.GenerateKey(rand.Reader)
	case KeyTypeRSA:
		return nil, ErrKeyGenerationFailed
	default:
		publicKey, privateKey, err = ed25519.GenerateKey(rand.Reader)
	}

	if err != nil {
		return nil, ErrKeyGenerationFailed
	}

	peerID := m.peerIDGen.EncodePeerID(publicKey, config.PeerIDEncoding)

	exists, err := m.store.Exists(peerID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrIdentityExists
	}

	identity := &Identity{
		PeerID:     peerID,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		CreatedAt:  time.Now(),
	}

	metadata := &IdentityMetadata{
		PeerID:      peerID,
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		DisplayName: "",
		AvatarURL:   "",
		DeviceInfo:  "",
	}

	if err := m.store.Save(identity, metadata); err != nil {
		return nil, err
	}

	if err := m.keyStore.SavePrivateKey(peerID, privateKey); err != nil {
		return nil, err
	}

	return identity, nil
}

// LoadIdentity loads an existing identity by Peer ID.
func (m *identityManager) LoadIdentity(peerID string) (*Identity, error) {
	identity, metadata, err := m.store.Load(peerID)
	if err != nil {
		return nil, ErrIdentityNotFound
	}

	privateKey, err := m.keyStore.LoadPrivateKey(peerID)
	if err == nil && len(privateKey) > 0 {
		identity.PrivateKey = privateKey
	}

	if metadata != nil {
		metadata.LastUsedAt = time.Now()
		m.store.Save(identity, metadata)
	}

	return identity, nil
}

// VerifyIdentity verifies if the given public key matches the identity.
func (m *identityManager) VerifyIdentity(peerID string, publicKey []byte) bool {
	if !m.peerIDGen.ValidatePeerID(peerID) {
		return false
	}

	identity, _, err := m.store.Load(peerID)
	if err != nil {
		return false
	}

	if len(identity.PublicKey) != len(publicKey) {
		return false
	}

	for i := range identity.PublicKey {
		if identity.PublicKey[i] != publicKey[i] {
			return false
		}
	}

	return true
}

// SignMessage signs a message using the first available identity.
func (m *identityManager) SignMessage(message []byte) ([]byte, error) {
	identities, err := m.store.List()
	if err != nil || len(identities) == 0 {
		return nil, ErrIdentityNotFound
	}
	return m.SignMessageForIdentity(identities[0], message)
}

// SignMessageForIdentity signs a message using a specific identity.
func (m *identityManager) SignMessageForIdentity(peerID string, message []byte) ([]byte, error) {
	privateKey, err := m.keyStore.LoadPrivateKey(peerID)
	if err != nil {
		return nil, ErrSignatureFailed
	}

	signature := ed25519.Sign(privateKey, message)
	return signature, nil
}

// VerifySignature verifies a message signature using the identity's public key.
func (m *identityManager) VerifySignature(peerID string, message, signature []byte) bool {
	identity, _, err := m.store.Load(peerID)
	if err != nil {
		return false
	}

	return ed25519.Verify(identity.PublicKey, message, signature)
}

// ExportIdentity exports the first identity as encrypted data.
func (m *identityManager) ExportIdentity(password string) ([]byte, error) {
	if password == "" {
		return nil, ErrPasswordRequired
	}

	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	identities, err := m.store.List()
	if err != nil || len(identities) == 0 {
		return nil, ErrIdentityNotFound
	}

	peerID := identities[0]
	identity, metadata, err := m.store.Load(peerID)
	if err != nil {
		return nil, ErrIdentityNotFound
	}

	privateKey, err := m.keyStore.LoadPrivateKey(peerID)
	if err != nil {
		return nil, ErrStorageFailed
	}

	exportData := &ExportData{
		PeerID:     peerID,
		PublicKey:  identity.PublicKey,
		PrivateKey: privateKey,
		Metadata:   metadata,
		ExportedAt: time.Now(),
	}

	encrypted, err := encryptData(exportData, password)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	return encrypted, nil
}

// ImportIdentity imports an identity from encrypted data.
func (m *identityManager) ImportIdentity(data []byte, password string) (*Identity, error) {
	if password == "" {
		return nil, ErrPasswordRequired
	}

	exportData, err := decryptData(data, password)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	exists, err := m.store.Exists(exportData.PeerID)
	if err == nil && exists {
		return nil, ErrIdentityExists
	}

	identity := &Identity{
		PeerID:     exportData.PeerID,
		PublicKey:  exportData.PublicKey,
		PrivateKey: exportData.PrivateKey,
		CreatedAt:  exportData.ExportedAt,
	}

	if err := m.store.Save(identity, exportData.Metadata); err != nil {
		return nil, err
	}

	if err := m.keyStore.SavePrivateKey(exportData.PeerID, exportData.PrivateKey); err != nil {
		return nil, err
	}

	return identity, nil
}

// DeleteIdentity deletes an identity by Peer ID.
func (m *identityManager) DeleteIdentity(peerID string) error {
	exists, err := m.store.Exists(peerID)
	if err != nil {
		return ErrIdentityNotFound
	}
	if !exists {
		return ErrIdentityNotFound
	}

	if err := m.store.Delete(peerID); err != nil {
		return ErrStorageFailed
	}

	if err := m.keyStore.DeletePrivateKey(peerID); err != nil {
		return ErrStorageFailed
	}

	return nil
}

// ListIdentities returns all identity Peer IDs.
func (m *identityManager) ListIdentities() ([]string, error) {
	return m.store.List()
}

// GetMetadata returns the metadata for an identity.
func (m *identityManager) GetMetadata(peerID string) (*IdentityMetadata, error) {
	_, metadata, err := m.store.Load(peerID)
	if err != nil {
		return nil, ErrMetadataNotFound
	}
	return metadata, nil
}

// UpdateMetadata updates the metadata for an identity.
func (m *identityManager) UpdateMetadata(peerID string, metadata *IdentityMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	identity, existingMetadata, err := m.store.Load(peerID)
	if err != nil {
		return ErrMetadataNotFound
	}

	if metadata.DisplayName != "" {
		existingMetadata.DisplayName = metadata.DisplayName
	}
	if metadata.AvatarURL != "" {
		existingMetadata.AvatarURL = metadata.AvatarURL
	}
	if metadata.DeviceInfo != "" {
		existingMetadata.DeviceInfo = metadata.DeviceInfo
	}

	return m.store.Save(identity, existingMetadata)
}

// GenerateChallenge generates a new challenge for identity verification.
func (m *identityManager) GenerateChallenge(peerID string) (*Challenge, error) {
	exists, err := m.store.Exists(peerID)
	if err != nil || !exists {
		return nil, ErrIdentityNotFound
	}

	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}

	challenge := &Challenge{
		Challenge: base64.StdEncoding.EncodeToString(randomBytes),
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	return challenge, nil
}

// VerifyChallenge verifies a challenge response for identity verification.
func (m *identityManager) VerifyChallenge(peerID string, challenge *Challenge, response *ChallengeResponse) (bool, error) {
	if time.Now().After(challenge.ExpiresAt) {
		return false, ErrChallengeExpired
	}

	identity, _, err := m.store.Load(peerID)
	if err != nil {
		return false, ErrIdentityNotFound
	}

	challengeData, err := base64.StdEncoding.DecodeString(response.Challenge)
	if err != nil {
		return false, ErrChallengeInvalid
	}

	signature := ed25519.Sign(identity.PrivateKey, challengeData)
	expectedResponse := base64.StdEncoding.EncodeToString(signature)

	if expectedResponse != response.Response {
		return false, ErrVerificationFailed
	}

	return true, nil
}
