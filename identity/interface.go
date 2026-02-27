package identity

type IdentityManager interface {
	CreateIdentity(config *IdentityConfig) (*Identity, error)
	LoadIdentity(peerID string) (*Identity, error)
	VerifyIdentity(peerID string, publicKey []byte) bool
	SignMessage(message []byte) ([]byte, error)
	VerifySignature(peerID string, message, signature []byte) bool
	ExportIdentity(password string) ([]byte, error)
	ImportIdentity(data []byte, password string) (*Identity, error)
	DeleteIdentity(peerID string) error
	ListIdentities() ([]string, error)
	GetMetadata(peerID string) (*IdentityMetadata, error)
	UpdateMetadata(peerID string, metadata *IdentityMetadata) error
	GenerateChallenge(peerID string) (*Challenge, error)
	VerifyChallenge(peerID string, challenge *Challenge, response *ChallengeResponse) (bool, error)
}

type PeerIDUtil interface {
	GeneratePeerID(publicKey []byte) string
	ValidatePeerID(peerID string) bool
	ExtractPublicKeyHash(peerID string) ([]byte, error)
	EncodePeerID(publicKey []byte, encoding PeerIDEncoding) string
	DecodePeerID(peerID string, encoding PeerIDEncoding) ([]byte, error)
}

type KeyStorage interface {
	SavePrivateKey(peerID string, encryptedKey []byte) error
	LoadPrivateKey(peerID string) ([]byte, error)
	DeletePrivateKey(peerID string) error
	KeyExists(peerID string) (bool, error)
	ListKeys() ([]string, error)
}

type IdentityStore interface {
	Save(identity *Identity, metadata *IdentityMetadata) error
	Load(peerID string) (*Identity, *IdentityMetadata, error)
	Delete(peerID string) error
	Exists(peerID string) (bool, error)
	List() ([]string, error)
}
