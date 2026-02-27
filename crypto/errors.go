package crypto

import "errors"

var (
	ErrInvalidKeySize       = errors.New("invalid key size")
	ErrInvalidNonceSize    = errors.New("invalid nonce size")
	ErrEncryptionFailed    = errors.New("encryption failed")
	ErrDecryptionFailed    = errors.New("decryption failed")
	ErrSignatureFailed     = errors.New("signature failed")
	ErrVerificationFailed  = errors.New("signature verification failed")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrKeyExchangeFailed   = errors.New("key exchange failed")
	ErrKeyNotFound         = errors.New("key not found")
	ErrKeyAlreadyExists    = errors.New("key already exists")
	ErrKeyExpired          = errors.New("key expired")
	ErrInvalidAlgorithm    = errors.New("invalid algorithm")
	ErrInvalidPublicKey    = errors.New("invalid public key")
	ErrInvalidPrivateKey   = errors.New("invalid private key")
	ErrRandomGeneration    = errors.New("random number generation failed")
	ErrKeyDerivationFailed = errors.New("key derivation failed")
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrPasswordMismatch    = errors.New("password mismatch")
	ErrPasswordTooShort    = errors.New("password too short")
	ErrInvalidHash         = errors.New("invalid hash")
	ErrMemoryCleanup       = errors.New("memory cleanup failed")
)

type CryptoError struct {
	Code    string
	Message string
	Err     error
}

func (e *CryptoError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *CryptoError) Unwrap() error {
	return e.Err
}

func NewCryptoError(code, message string, err error) *CryptoError {
	return &CryptoError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
