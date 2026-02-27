package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"
)

var (
	ErrSignatureInvalid      = errors.New("crypto: invalid signature")
	ErrAlgorithmNotSupported = errors.New("crypto: algorithm not supported")
)

type cryptoManager struct {
	config   *SecurityConfig
	sessions sync.Map
	mu       sync.RWMutex
}

func NewCryptoManager(config *SecurityConfig) CryptoManager {
	if config == nil {
		config = DefaultSecurityConfig()
	}
	return &cryptoManager{
		config: config,
	}
}

func (c *cryptoManager) GenerateKeyPair(algorithm AlgorithmType) (*KeyPair, error) {
	var publicKey, privateKey []byte

	switch algorithm {
	case AlgorithmEd25519:
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		publicKey = pub
		privateKey = priv
	case AlgorithmX25519:
		var err error
		publicKey = make([]byte, 32)
		privateKey = make([]byte, 32)
		if _, err = io.ReadFull(rand.Reader, privateKey); err != nil {
			return nil, err
		}
		publicKey = x25519BasePointMul(privateKey)
	default:
		return nil, ErrAlgorithmNotSupported
	}

	now := time.Now()
	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Algorithm:  algorithm,
		CreatedAt:  now,
		ExpiresAt:  now.Add(365 * 24 * time.Hour),
	}, nil
}

func (c *cryptoManager) Encrypt(plaintext []byte, key []byte, config *EncryptionConfig) (*EncryptedMessage, error) {
	if config == nil {
		config = DefaultEncryptionConfig()
	}

	if len(key) != config.KeySize {
		return nil, ErrInvalidKeySize
	}

	var nonce []byte
	var ciphertext []byte

	switch config.Algorithm {
	case AlgorithmAESGCM:
		var err error
		nonce = make([]byte, config.NonceSize)
		if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}

		aesgcm, err := cipher.NewGCMWithNonceSize(block, config.NonceSize)
		if err != nil {
			return nil, err
		}

		ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)

	case AlgorithmChaCha20:
		return nil, ErrAlgorithmNotSupported
	default:
		return nil, ErrAlgorithmNotSupported
	}

	tagSize := config.TagSize
	if tagSize == 0 {
		tagSize = 16
	}

	tag := ciphertext[len(ciphertext)-tagSize:]
	ciphertext = ciphertext[:len(ciphertext)-tagSize]

	return &EncryptedMessage{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Tag:        tag,
		Algorithm:  config.Algorithm,
		Version:    "1.0",
		Timestamp:  time.Now(),
	}, nil
}

func (c *cryptoManager) Decrypt(msg *EncryptedMessage, key []byte) ([]byte, error) {
	if msg == nil {
		return nil, ErrDecryptionFailed
	}

	switch msg.Algorithm {
	case AlgorithmAESGCM:
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}

		nonceSize := 12
		if len(msg.Nonce) > 0 {
			nonceSize = len(msg.Nonce)
		}

		aesgcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
		if err != nil {
			return nil, err
		}

		ciphertextWithTag := append(msg.Ciphertext, msg.Tag...)
		plaintext, err := aesgcm.Open(nil, msg.Nonce, ciphertextWithTag, nil)
		if err != nil {
			return nil, ErrDecryptionFailed
		}

		return plaintext, nil

	case AlgorithmChaCha20:
		return nil, ErrAlgorithmNotSupported

	default:
		return nil, ErrAlgorithmNotSupported
	}
}

func (c *cryptoManager) Sign(message []byte, privateKey []byte) (*SignedMessage, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, ErrInvalidKeySize
	}

	privKey := ed25519.PrivateKey(privateKey)
	signature := ed25519.Sign(privKey, message)

	return &SignedMessage{
		Message:   message,
		Signature: signature,
		PublicKey: privKey.Public().(ed25519.PublicKey),
		Algorithm: AlgorithmEd25519,
	}, nil
}

func (c *cryptoManager) Verify(signedMsg *SignedMessage, publicKey []byte) (bool, error) {
	if signedMsg == nil || len(publicKey) == 0 {
		return false, ErrSignatureInvalid
	}

	if len(signedMsg.Signature) != ed25519.SignatureSize {
		return false, ErrSignatureInvalid
	}

	switch signedMsg.Algorithm {
	case AlgorithmEd25519:
		return ed25519.Verify(publicKey, signedMsg.Message, signedMsg.Signature), nil
	default:
		return false, ErrAlgorithmNotSupported
	}
}

func (c *cryptoManager) Hash(data []byte, algorithm AlgorithmType) ([]byte, error) {
	switch algorithm {
	case AlgorithmSHA256:
		hash := sha256.Sum256(data)
		return hash[:], nil
	case AlgorithmSHA3_256, AlgorithmBLAKE2b:
		return nil, ErrAlgorithmNotSupported
	default:
		hash := sha256.Sum256(data)
		return hash[:], nil
	}
}

func (c *cryptoManager) RandomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (c *cryptoManager) DeriveKey(secret []byte, salt []byte, info []byte, size int) ([]byte, error) {
	if len(secret) == 0 {
		return nil, ErrInvalidKeySize
	}

	if len(salt) == 0 {
		salt = make([]byte, 32)
	}

	_, err := c.HMACSHA256(salt, secret)
	if err != nil {
		return nil, err
	}

	okm := make([]byte, size)
	t := make([]byte, 0, 32)
	counter := byte(1)

	for i := 0; i < size; i += 32 {
		h := sha256.New()
		h.Write(t)
		h.Write(info)
		h.Write([]byte{counter})
		t = h.Sum(nil)

		copy(okm[i:], t)
		counter++
	}

	return okm, nil
}

func (c *cryptoManager) HMACSHA256(key, data []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(key)
	h.Write(data)
	return h.Sum(nil), nil
}

func (c *cryptoManager) Cleanup() error {
	c.sessions.Range(func(key, value interface{}) bool {
		c.sessions.Delete(key)
		return true
	})
	return nil
}

func (c *cryptoManager) ConstantTimeCompare(a, b []byte) bool {
	return sha256.Sum256(a) == sha256.Sum256(b)
}

func (c *cryptoManager) SecureZeroMemory(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

func (c *cryptoManager) SecureRandomInt(min, max int) (int, error) {
	if min >= max {
		return min, nil
	}

	rangeSize := max - min
	bytes := make([]byte, 4)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return 0, err
	}

	val := int(uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3]))
	if val < 0 {
		val = -val
	}
	return min + (val % rangeSize), nil
}

func (c *cryptoManager) GenerateKeyID(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:16])
}
