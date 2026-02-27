package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

type ExportData struct {
	PeerID     string            `json:"peer_id"`
	PublicKey  []byte            `json:"public_key"`
	PrivateKey []byte            `json:"private_key"`
	Metadata   *IdentityMetadata `json:"metadata"`
	ExportedAt time.Time         `json:"exported_at"`
}

type encryptedData struct {
	Ciphertext []byte `json:"ciphertext"`
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Version    int    `json:"version"`
}

const (
	pbkdf2Iterations = 100000
	pbkdf2KeyLen     = 32
	aesNonceSize     = 12
	exportVersion    = 1
)

func encryptData(data *ExportData, password string) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyLen, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, jsonData, nil)

	encrypted := &encryptedData{
		Ciphertext: ciphertext,
		Salt:       salt,
		Nonce:      nonce,
		Version:    exportVersion,
	}

	return json.Marshal(encrypted)
}

func decryptData(encryptedBytes []byte, password string) (*ExportData, error) {
	var encrypted encryptedData
	if err := json.Unmarshal(encryptedBytes, &encrypted); err != nil {
		return nil, ErrInvalidFormat
	}

	if encrypted.Version != exportVersion {
		return nil, ErrInvalidFormat
	}

	key := pbkdf2.Key([]byte(password), encrypted.Salt, pbkdf2Iterations, pbkdf2KeyLen, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	nonce := encrypted.Nonce
	ciphertext := encrypted.Ciphertext

	if len(nonce) != aesNonceSize {
		return nil, ErrDecryptionFailed
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrPasswordMismatch
	}

	var data ExportData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, ErrInvalidFormat
	}

	return &data, nil
}

func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.New("failed to generate random bytes")
	}
	return b, nil
}
