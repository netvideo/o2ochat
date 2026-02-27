package identity

import "errors"

var (
	ErrIdentityNotFound    = errors.New("identity not found")
	ErrIdentityExists      = errors.New("identity already exists")
	ErrInvalidPeerID       = errors.New("invalid peer ID")
	ErrInvalidPublicKey    = errors.New("invalid public key")
	ErrInvalidPrivateKey   = errors.New("invalid private key")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrSignatureFailed     = errors.New("signature failed")
	ErrVerificationFailed  = errors.New("signature verification failed")
	ErrEncryptionFailed    = errors.New("encryption failed")
	ErrDecryptionFailed    = errors.New("decryption failed")
	ErrPasswordRequired    = errors.New("password required")
	ErrPasswordMismatch    = errors.New("password mismatch")
	ErrPasswordTooShort    = errors.New("password too short")
	ErrInvalidFormat       = errors.New("invalid format")
	ErrStorageFailed       = errors.New("storage operation failed")
	ErrChallengeExpired    = errors.New("challenge expired")
	ErrChallengeInvalid    = errors.New("invalid challenge")
	ErrMetadataNotFound    = errors.New("metadata not found")
)

type IdentityError struct {
	Code    string
	Message string
	Err     error
}

func (e *IdentityError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *IdentityError) Unwrap() error {
	return e.Err
}

func NewIdentityError(code, message string, err error) *IdentityError {
	return &IdentityError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
