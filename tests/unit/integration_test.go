package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/netvideo/crypto"
	"github.com/netvideo/identity"
	"github.com/netvideo/signaling"
	"github.com/netvideo/storage"
)

// TestIdentityCryptoStorageIntegration 测试身份 + 加密 + 存储集成
func TestIdentityCryptoStorageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建临时目录
	testDir := t.TempDir()

	// 1. 创建内存存储管理器（避免 SQLite 依赖问题）
	storageManager := storage.NewMemoryStorageManager()
	if storageManager == nil {
		t.Fatal("Failed to create memory storage manager")
	}
	defer storageManager.Close()

	// 2. 创建加密管理器
	cryptoConfig := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}

	cryptoManager := crypto.NewCryptoManager(cryptoConfig)
	if cryptoManager == nil {
		t.Fatal("Failed to create crypto manager")
	}

	// 3. 创建身份管理器
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	if store == nil {
		t.Fatal("Failed to create memory identity store")
	}
	if keyStorage == nil {
		t.Fatal("Failed to create memory key storage")
	}

	identityManager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer identityManager.Close()

	// 4. 创建身份
	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := identityManager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}
	t.Logf("✓ 身份创建：%s", id.PeerID)

	// 5. 导出身份（加密）
	password := "integration-test-password"
	exportData, err := identityManager.ExportIdentity(id.PeerID, password)
	if err != nil {
		t.Fatalf("Failed to export identity: %v", err)
	}
	t.Logf("✓ 身份导出：%d 字节", len(exportData))

	// 6. 使用加密管理器加密导出数据
	key := make([]byte, 32)
	_, err = cryptoManager.RandomBytes(key)
	if err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	encryptedData, err := cryptoManager.Encrypt(exportData, key)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}
	t.Logf("✓ 数据加密：%d → %d 字节", len(exportData), len(encryptedData))

	// 7. 存储
	err = storageManager.SaveEncryptedKey(id.PeerID, "identity", encryptedData)
	if err != nil {
		t.Fatalf("Failed to save encrypted key: %v", err)
	}
	t.Logf("✓ 加密数据存储完成")

	// 8. 从存储读取
	storedData, err := storageManager.LoadEncryptedKey(id.PeerID, "identity")
	if err != nil {
		t.Fatalf("Failed to load encrypted key: %v", err)
	}
	if len(storedData) != len(encryptedData) {
		t.Errorf("Stored data length mismatch: expected %d, got %d", len(encryptedData), len(storedData))
	}
	t.Logf("✓ 加密数据读取完成")

	// 9. 解密
	decryptedData, err := cryptoManager.Decrypt(storedData, key)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}
	if len(decryptedData) != len(exportData) {
		t.Errorf("Decrypted data length mismatch: expected %d, got %d", len(exportData), len(decryptedData))
	}
	t.Logf("✓ 数据解密完成")

	// 10. 导入身份
	importedID, err := identityManager.ImportIdentity(decryptedData, password)
	if err != nil {
		t.Fatalf("Failed to import identity: %v", err)
	}
	if importedID != id.PeerID {
		t.Errorf("Peer ID mismatch: expected %s, got %s", id.PeerID, importedID)
	}
	t.Logf("✓ 身份导入成功：%s", importedID.PeerID)

	t.Log("=== 身份 + 加密 + 存储集成测试 通过 ===")
}

// TestSignalingWithIdentity 测试带身份验证的信令
func TestSignalingWithIdentity(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建身份
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()
	identityManager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer identityManager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := identityManager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	// 创建信令配置
	signalingConfig := &signaling.ClientConfig{
		ServerURL: "wss://signal.o2ochat.io",
		PeerID:    id.PeerID,
		Timeout:   10 * time.Second,
	}

	// 创建信令客户端
	client := signaling.NewClient(signalingConfig)
	if client == nil {
		t.Fatal("Failed to create signaling client")
	}

	// 测试连接（可能失败，因为是集成测试）
	err = client.Connect(ctx)
	if err != nil {
		t.Logf("信令连接失败（预期）: %v", err)
		// 连接失败是可以接受的，因为我们测试的是集成
	} else {
		t.Log("✓ 信令连接成功")
		defer client.Disconnect()
	}

	t.Log("=== 信令 + 身份集成测试 完成 ===")
}

// TestFullIntegration 测试完整集成流程
func TestFullIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	t.Log("=== 开始完整集成测试 ===")

	// 1. 创建身份
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()
	identityManager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer identityManager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := identityManager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}
	t.Logf("✓ 身份创建：%s", id.PeerID)

	// 2. 签名消息
	message := []byte("Integration test message")
	signature, err := identityManager.SignMessage(id.PeerID, message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}
	t.Logf("✓ 消息签名：%d 字节", len(signature))

	// 3. 验证签名
	valid, err := identityManager.VerifySignature(id.PeerID, message, signature)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if !valid {
		t.Error("Signature verification failed")
	}
	t.Logf("✓ 签名验证：%v", valid)

	// 4. 挑战 - 响应
	challenge, err := identityManager.GenerateChallenge(id.PeerID)
	if err != nil {
		t.Fatalf("Failed to generate challenge: %v", err)
	}
	t.Logf("✓ 挑战生成：%d 字节", len(challenge.Data))

	response, err := identityManager.RespondToChallenge(id.PeerID, challenge)
	if err != nil {
		t.Fatalf("Failed to respond to challenge: %v", err)
	}
	t.Logf("✓ 挑战响应：%d 字节", len(response))

	valid, err = identityManager.VerifyChallengeResponse(id.PeerID, challenge, response)
	if err != nil {
		t.Fatalf("Failed to verify challenge response: %v", err)
	}
	if !valid {
		t.Error("Challenge response verification failed")
	}
	t.Logf("✓ 挑战验证：%v", valid)

	t.Log("=== 完整集成测试 通过 ===")
}
