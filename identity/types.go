package identity

import (
	"time"
)

type KeyType string

const (
	KeyTypeEd25519 KeyType = "ed25519"
	KeyTypeRSA     KeyType = "rsa"
)

type PeerIDEncoding string

const (
	PeerIDEncodingBase58 PeerIDEncoding = "base58"
	PeerIDEncodingHex    PeerIDEncoding = "hex"
)

type Identity struct {
	PeerID     string    `json:"peer_id"`
	PublicKey  []byte    `json:"public_key"`
	PrivateKey []byte    `json:"private_key"`
	CreatedAt  time.Time `json:"created_at"`
}

type IdentityConfig struct {
	KeyType        KeyType        `json:"key_type"`
	KeyLength      int            `json:"key_length"`
	PeerIDEncoding PeerIDEncoding `json:"peer_id_encoding"`
}

type IdentityMetadata struct {
	PeerID      string    `json:"peer_id"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	DeviceInfo  string    `json:"device_info"`
}

type Challenge struct {
	Challenge string    `json:"challenge"`
	Timestamp time.Time `json:"timestamp"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ChallengeResponse struct {
	Challenge string `json:"challenge"`
	Response  string `json:"response"`
	PeerID    string `json:"peer_id"`
}

func DefaultIdentityConfig() *IdentityConfig {
	return &IdentityConfig{
		KeyType:        KeyTypeEd25519,
		KeyLength:      256,
		PeerIDEncoding: PeerIDEncodingBase58,
	}
}
