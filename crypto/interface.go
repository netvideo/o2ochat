package crypto

type CryptoManager interface {
	GenerateKeyPair(algorithm AlgorithmType) (*KeyPair, error)
	Encrypt(plaintext []byte, key []byte, config *EncryptionConfig) (*EncryptedMessage, error)
	Decrypt(msg *EncryptedMessage, key []byte) ([]byte, error)
	Sign(message []byte, privateKey []byte) (*SignedMessage, error)
	Verify(signedMsg *SignedMessage, publicKey []byte) (bool, error)
	Hash(data []byte, algorithm AlgorithmType) ([]byte, error)
	RandomBytes(size int) ([]byte, error)
	DeriveKey(secret []byte, salt []byte, info []byte, size int) ([]byte, error)
	Cleanup() error
	ConstantTimeCompare(a, b []byte) bool
	SecureZeroMemory(data []byte)
	SecureRandomInt(min, max int) (int, error)
	GenerateKeyID(publicKey []byte) string
}

type KeyExchange interface {
	Initiate() (*KeyExchangeResult, error)
	Respond(peerPublicKey []byte) (*KeyExchangeResult, error)
	Finalize(peerPublicKey []byte, sessionID string) ([]byte, error)
	VerifyExchange(sharedSecret []byte, proof []byte) (bool, error)
	RotateKey(sessionID string) ([]byte, error)
	DestroySession(sessionID string) error
}

type KeyStorage interface {
	StoreKey(keyID string, key []byte, metadata map[string]string) error
	GetKey(keyID string) ([]byte, map[string]string, error)
	DeleteKey(keyID string) error
	ListKeys() ([]string, error)
	KeyExists(keyID string) (bool, error)
	CleanupExpiredKeys() error
}

type CryptoUtil interface {
	ConstantTimeCompare(a, b []byte) bool
	SecureZeroMemory(data []byte)
	SecureRandomInt(min, max int) (int, error)
	HashPassword(password string) ([]byte, error)
	VerifyPassword(password string, hash []byte) (bool, error)
	GenerateKeyID(publicKey []byte) string
}

type ForwardSecrecyProtocol interface {
	Handshake(peerPublicKey []byte) (*SessionKeys, error)
	VerifyHandshake(peerPublicKey []byte, sessionKeys *SessionKeys) (bool, error)
}
