package signaling

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Signer interface {
	SignMessage(message []byte) ([]byte, error)
	GetPublicKey() []byte
	GetPeerID() string
}

type MessageSigner struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	peerID     string
	nonceStore *NonceStore
}

type NonceStore struct {
	nonces map[string]time.Time
	mu     sync.Mutex
}

func NewNonceStore() *NonceStore {
	return &NonceStore{
		nonces: make(map[string]time.Time),
	}
}

func (ns *NonceStore) Add(nonce string, expiry time.Duration) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.nonces[nonce]; exists {
		return false
	}

	ns.nonces[nonce] = time.Now().Add(expiry)
	return true
}

func (ns *NonceStore) Validate(nonce string) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	expiry, exists := ns.nonces[nonce]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		delete(ns.nonces, nonce)
		return false
	}

	return true
}

func (ns *NonceStore) Cleanup() {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	now := time.Now()
	for nonce, expiry := range ns.nonces {
		if now.After(expiry) {
			delete(ns.nonces, nonce)
		}
	}
}

func NewMessageSigner(privateKey ed25519.PrivateKey, peerID string) *MessageSigner {
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return &MessageSigner{
		privateKey: privateKey,
		publicKey:  publicKey,
		peerID:     peerID,
		nonceStore: NewNonceStore(),
	}
}

func (s *MessageSigner) SignMessage(message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)
	signature := ed25519.Sign(s.privateKey, hash[:])
	return signature, nil
}

func (s *MessageSigner) GetPublicKey() []byte {
	return s.publicKey
}

func (s *MessageSigner) GetPeerID() string {
	return s.peerID
}

func (s *MessageSigner) AddNonce(nonce string, expiry time.Duration) bool {
	return s.nonceStore.Add(nonce, expiry)
}

func (s *MessageSigner) ValidateNonce(nonce string) bool {
	return s.nonceStore.Validate(nonce)
}

func SignMessage(msg *SignalingMessage, privateKey ed25519.PrivateKey) ([]byte, error) {
	signData, err := GetSignableData(msg)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(signData)
	signature := ed25519.Sign(privateKey, hash[:])
	return signature, nil
}

func VerifySignature(msg *SignalingMessage, publicKey ed25519.PublicKey) bool {
	if len(msg.Signature) == 0 {
		return false
	}

	signData, err := GetSignableData(msg)
	if err != nil {
		return false
	}

	hash := sha256.Sum256(signData)
	return ed25519.Verify(publicKey, hash[:], msg.Signature)
}

func GetSignableData(msg *SignalingMessage) ([]byte, error) {
	signable := struct {
		Type      MessageType `json:"type"`
		From      string      `json:"from"`
		To        string      `json:"to"`
		Data      interface{} `json:"data"`
		Timestamp int64       `json:"timestamp"`
		Nonce     string      `json:"nonce"`
	}{
		Type:      msg.Type,
		From:      msg.From,
		To:        msg.To,
		Data:      msg.Data,
		Timestamp: msg.Timestamp.Unix(),
		Nonce:     msg.Nonce,
	}

	return json.Marshal(signable)
}

func GenerateNonce() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func ValidateNonceFormat(nonce string) bool {
	if len(nonce) < 16 {
		return false
	}

	_, err := base64.URLEncoding.DecodeString(nonce)
	return err == nil
}

func CreateSignableMessage(msg *SignalingMessage) ([]byte, error) {
	data := map[string]interface{}{
		"type":      msg.Type,
		"from":      msg.From,
		"to":        msg.To,
		"timestamp": msg.Timestamp.UnixNano(),
		"nonce":     msg.Nonce,
	}

	if msg.Data != nil {
		data["data"] = msg.Data
	}

	return json.Marshal(data)
}

type SignedMessage struct {
	SignalingMessage
	Signature string `json:"signature"`
	SignedAt  int64  `json:"signed_at"`
	PublicKey string `json:"public_key"`
}

func (s *SignedMessage) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *SignedMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

func SignMessageToBase64(msg *SignalingMessage, privateKey ed25519.PrivateKey) (string, error) {
	signature, err := SignMessage(msg, privateKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func VerifySignatureFromBase64(msg *SignalingMessage, publicKey ed25519.PublicKey, signatureBase64 string) bool {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}

	msg.Signature = signature
	return VerifySignature(msg, publicKey)
}

func HashMessage(msg *SignalingMessage) (string, error) {
	data, err := CreateSignableMessage(msg)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

func ValidateMessageTimestamp(msg *SignalingMessage, maxAge time.Duration) bool {
	now := time.Now()
	age := now.Sub(msg.Timestamp)
	return age >= 0 && age <= maxAge
}
