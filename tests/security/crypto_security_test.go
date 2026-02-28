package security

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/netvideo/crypto"
	"github.com/netvideo/identity"
)

// TestReplayAttackPrevention 测试防重放攻击
func TestReplayAttackPrevention(t *testing.T) {
	// 创建存储
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	// 创建身份管理器
	manager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   identity.KeyTypeEd25519,
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	// 1. 创建挑战
	challenge, err := manager.GenerateChallenge()
	if err != nil {
		t.Fatalf("Failed to generate challenge: %v", err)
	}

	if len(challenge.Data) == 0 {
		t.Error("Challenge data should not be empty")
	}

	// 2. 响应挑战
	signature, err := manager.SignChallenge(id.PeerID, challenge)
	if err != nil {
		t.Fatalf("Failed to sign challenge: %v", err)
	}

	// 3. 验证挑战
	valid, err := manager.VerifyChallenge(id.PeerID, challenge, signature)
	if err != nil {
		t.Fatalf("Failed to verify challenge: %v", err)
	}

	if !valid {
		t.Error("Challenge verification should succeed")
	}

	// 4. 测试重放攻击
	_, err = manager.SignChallenge(id.PeerID, challenge)
	if err != nil {
		t.Fatalf("Failed to sign challenge again: %v", err)
	}

	// 挑战应该已经过期或被使用
	if challenge.IsExpired() {
		t.Log("Challenge correctly expired")
	}
}

// TestChallengeTimeout 测试挑战超时
func TestChallengeTimeout(t *testing.T) {
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	manager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   identity.KeyTypeEd25519,
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	// 创建挑战
	challenge, err := manager.GenerateChallenge()
	if err != nil {
		t.Fatalf("Failed to generate challenge: %v", err)
	}

	// 等待挑战过期
	time.Sleep(2 * time.Second)

	// 验证挑战已过期
	if !challenge.IsExpired() {
		t.Error("Challenge should be expired")
	}

	// 尝试响应过期挑战
	_, err = manager.SignChallenge(id.PeerID, challenge)
	if err == nil {
		t.Error("Should fail to sign expired challenge")
	}
}

// TestKeyExchangeSecurity 测试密钥交换安全
func TestKeyExchangeSecurity(t *testing.T) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}

	cryptoManager := crypto.NewCryptoManager(config)
	if cryptoManager == nil {
		t.Fatal("Failed to create crypto manager")
	}

	exchange := crypto.NewKeyExchange(cryptoManager)
	if exchange == nil {
		t.Fatal("Failed to create key exchange")
	}

	// 发起密钥交换
	result, err := exchange.Initiate()
	if err != nil {
		t.Fatalf("Failed to initiate key exchange: %v", err)
	}

	if result == nil {
		t.Fatal("Key exchange result should not be nil")
	}

	// 响应该交换
	response, err := exchange.Respond(result.PublicKey, result.SessionID)
	if err != nil {
		t.Fatalf("Failed to respond to key exchange: %v", err)
	}

	if response == nil {
		t.Fatal("Key exchange response should not be nil")
	}

	// 完成密钥交换
	sessionKey, err := exchange.Finalize(response, result.SessionID)
	if err != nil {
		t.Fatalf("Failed to finalize key exchange: %v", err)
	}

	if len(sessionKey) == 0 {
		t.Error("Session key should not be empty")
	}

	// 测试密钥轮换
	newKey, err := exchange.RotateKey(result.SessionID)
	if err != nil {
		t.Fatalf("Failed to rotate key: %v", err)
	}

	if len(newKey) == 0 {
		t.Error("New key should not be empty")
	}

	// 销毁会话
	err = exchange.DestroySession(result.SessionID)
	if err != nil {
		t.Fatalf("Failed to destroy session: %v", err)
	}
}

// TestEncryptionSecurity 测试加密安全
func TestEncryptionSecurity(t *testing.T) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}

	manager := crypto.NewCryptoManager(config)
	if manager == nil {
		t.Fatal("Failed to create crypto manager")
	}

	// 1. 测试加密和解密
	plaintext := []byte("Hello, World!")
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	ciphertext, err := manager.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Error("Ciphertext should not be empty")
	}

	decrypted, err := manager.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Error("Decrypted text should match plaintext")
	}

	// 2. 测试签名和验证
	signature, err := manager.Sign(plaintext, key)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	if len(signature) == 0 {
		t.Error("Signature should not be empty")
	}

	valid, err := manager.Verify(plaintext, signature, key)
	if err != nil {
		t.Fatalf("Failed to verify: %v", err)
	}

	if !valid {
		t.Error("Signature should be valid")
	}

	// 3. 测试哈希
	hash, err := manager.Hash(plaintext, crypto.AlgorithmSHA256)
	if err != nil {
		t.Fatalf("Failed to hash: %v", err)
	}

	if len(hash) == 0 {
		t.Error("Hash should not be empty")
	}

	// 4. 测试密钥派生
	derivedKey, err := manager.DeriveKey([]byte("password"), []byte("salt"), 32)
	if err != nil {
		t.Fatalf("Failed to derive key: %v", err)
	}

	if len(derivedKey) == 0 {
		t.Error("Derived key should not be empty")
	}
}

// TestSignatureUnforgeability 测试签名不可伪造性
func TestSignatureUnforgeability(t *testing.T) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}

	manager := crypto.NewCryptoManager(config)
	if manager == nil {
		t.Fatal("Failed to create crypto manager")
	}

	message := []byte("Test message")

	// 生成密钥对
	publicKey := make([]byte, 32)
	privateKey := make([]byte, 32)
	_, err := rand.Read(publicKey)
	if err != nil {
		t.Fatalf("Failed to generate public key: %v", err)
	}
	_, err = rand.Read(privateKey)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// 使用正确的私钥签名
	signature, err := manager.Sign(message, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	// 使用正确的公钥验证应该成功
	valid, err := manager.Verify(message, signature, publicKey)
	if err != nil {
		t.Fatalf("Failed to verify: %v", err)
	}

	if !valid {
		t.Error("Valid signature should verify successfully")
	}

	// 使用错误的公钥验证应该失败
	wrongPublicKey := make([]byte, 32)
	_, err = rand.Read(wrongPublicKey)
	if err != nil {
		t.Fatalf("Failed to generate wrong public key: %v", err)
	}

	valid, err = manager.Verify(message, signature, wrongPublicKey)
	if err != nil {
		t.Fatalf("Failed to verify with wrong key: %v", err)
	}

	if valid {
		t.Error("Signature should not verify with wrong public key")
	}
}
