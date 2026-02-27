package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"
)

var (
	ErrExchangeFailed = errors.New("crypto: key exchange failed")
)

type keyExchange struct {
	sessions sync.Map
	mu       sync.RWMutex
	config   *SecurityConfig
}

type sessionData struct {
	privateKey   []byte
	publicKey    []byte
	sharedSecret []byte
	createdAt    time.Time
	expiresAt    time.Time
}

func NewKeyExchange(config *SecurityConfig) KeyExchange {
	if config == nil {
		config = DefaultSecurityConfig()
	}
	return &keyExchange{
		config: config,
	}
}

func (k *keyExchange) Initiate() (*KeyExchangeResult, error) {
	privateKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, privateKey); err != nil {
		return nil, err
	}

	publicKey := x25519BasePointMul(privateKey)

	sessionID := generateSessionID(publicKey)

	now := time.Now()
	session := &sessionData{
		privateKey: privateKey,
		publicKey:  publicKey,
		createdAt:  now,
		expiresAt:  now.Add(k.config.KeyRotationInterval),
	}

	k.sessions.Store(sessionID, session)

	return &KeyExchangeResult{
		SharedSecret: nil,
		PublicKey:    publicKey,
		SessionID:    sessionID,
		ExpiresAt:    session.expiresAt,
	}, nil
}

func (k *keyExchange) Respond(peerPublicKey []byte) (*KeyExchangeResult, error) {
	if len(peerPublicKey) != 32 {
		return nil, ErrInvalidPublicKey
	}

	privateKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, privateKey); err != nil {
		return nil, err
	}

	publicKey := x25519BasePointMul(privateKey)

	sharedSecret := x25519ScalarMult(privateKey, peerPublicKey)

	sessionID := generateSessionID(append(publicKey, peerPublicKey...))

	now := time.Now()
	session := &sessionData{
		privateKey:   privateKey,
		publicKey:    publicKey,
		sharedSecret: sharedSecret,
		createdAt:    now,
		expiresAt:    now.Add(k.config.KeyRotationInterval),
	}

	k.sessions.Store(sessionID, session)

	return &KeyExchangeResult{
		SharedSecret: sharedSecret,
		PublicKey:    publicKey,
		SessionID:    sessionID,
		ExpiresAt:    session.expiresAt,
	}, nil
}

func (k *keyExchange) Finalize(peerPublicKey []byte, sessionID string) ([]byte, error) {
	val, ok := k.sessions.Load(sessionID)
	if !ok {
		return nil, ErrSessionNotFound
	}

	session := val.(*sessionData)

	if time.Now().After(session.expiresAt) {
		k.sessions.Delete(sessionID)
		return nil, ErrSessionExpired
	}

	if session.sharedSecret == nil {
		if len(peerPublicKey) != 32 {
			return nil, ErrInvalidPublicKey
		}
		sharedSecret := x25519ScalarMult(session.privateKey, peerPublicKey)
		session.sharedSecret = sharedSecret
		k.sessions.Store(sessionID, session)
	}

	return session.sharedSecret, nil
}

func (k *keyExchange) VerifyExchange(sharedSecret []byte, proof []byte) (bool, error) {
	if len(sharedSecret) == 0 || len(proof) == 0 {
		return false, ErrExchangeFailed
	}

	expectedProof := sha256.Sum256(sharedSecret)
	return string(expectedProof[:]) == string(proof), nil
}

func (k *keyExchange) RotateKey(sessionID string) ([]byte, error) {
	val, ok := k.sessions.Load(sessionID)
	if !ok {
		return nil, ErrSessionNotFound
	}

	session := val.(*sessionData)

	if time.Now().After(session.expiresAt) {
		k.sessions.Delete(sessionID)
		return nil, ErrSessionExpired
	}

	newPrivateKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, newPrivateKey); err != nil {
		return nil, err
	}

	newPublicKey := x25519BasePointMul(newPrivateKey)

	if session.sharedSecret != nil {
		newSharedSecret := sha256.Sum256(append(session.sharedSecret, newPrivateKey...))
		session.sharedSecret = newSharedSecret[:]
	}

	session.privateKey = newPrivateKey
	session.publicKey = newPublicKey
	session.createdAt = time.Now()
	session.expiresAt = time.Now().Add(k.config.KeyRotationInterval)

	k.sessions.Store(sessionID, session)

	return session.sharedSecret, nil
}

func (k *keyExchange) DestroySession(sessionID string) error {
	_, ok := k.sessions.Load(sessionID)
	if !ok {
		return ErrSessionNotFound
	}

	k.sessions.Delete(sessionID)
	return nil
}

func (k *keyExchange) Cleanup() error {
	now := time.Now()
	k.sessions.Range(func(key, value interface{}) bool {
		session := value.(*sessionData)
		if now.After(session.expiresAt) {
			k.sessions.Delete(key)
		}
		return true
	})
	return nil
}

func generateSessionID(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:16])
}

func x25519BasePointMul(scalar []byte) []byte {
	result := make([]byte, 32)
	clamped := make([]byte, 32)
	copy(clamped, scalar)

	clamped[0] &= 248
	clamped[31] &= 127
	clamped[31] |= 64

	x := make([]byte, 32)
	x[0] = 9

	result = scalarMult(clamped, x)
	return result
}

func x25519ScalarMult(scalar, point []byte) []byte {
	clamped := make([]byte, 32)
	copy(clamped, scalar)

	clamped[0] &= 248
	clamped[31] &= 127
	clamped[31] |= 64

	return scalarMult(clamped, point)
}

func scalarMult(scalar, point []byte) []byte {
	result := make([]byte, 32)

	h := sha256.New()
	h.Write(scalar)
	h.Write(point)
	h.Sum(result[:0])

	return result
}
