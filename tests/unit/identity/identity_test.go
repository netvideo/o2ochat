package identity_test

import (
	"testing"

	"github.com/netvideo/identity"
)

func TestCreateIdentity(t *testing.T) {
	// 使用内存存储创建管理器
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	if store == nil {
		t.Fatal("Failed to create memory identity store")
	}
	if keyStorage == nil {
		t.Fatal("Failed to create memory key storage")
	}

	manager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	// 执行
	id, err := manager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	// 验证
	if id.PeerID == "" {
		t.Error("Peer ID should not be empty")
	}
	if len(id.PublicKey) == 0 {
		t.Error("PublicKey should not be empty")
	}
	if id.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestSignAndVerify(t *testing.T) {
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	manager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	message := []byte("测试消息")

	// 执行
	signature, err := manager.SignMessage(id.PeerID, message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	if len(signature) == 0 {
		t.Fatal("Signature should not be empty")
	}

	// 验证签名
	valid, err := manager.VerifySignature(id.PeerID, message, signature)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	if !valid {
		t.Error("Signature should be valid")
	}

	// 验证错误的签名
	wrongMessage := []byte("错误的消息")
	valid, err = manager.VerifySignature(id.PeerID, wrongMessage, signature)
	if err != nil {
		t.Fatalf("Failed to verify wrong signature: %v", err)
	}

	if valid {
		t.Error("Signature should not be valid for wrong message")
	}
}

func TestExportImportIdentity(t *testing.T) {
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	manager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	// 导出身份
	password := "test-password"
	exportData, err := manager.ExportIdentity(id.PeerID, password)
	if err != nil {
		t.Fatalf("Failed to export identity: %v", err)
	}

	if len(exportData) == 0 {
		t.Fatal("Export data should not be empty")
	}

	// 导入身份
	newPeerID, err := manager.ImportIdentity(exportData, password)
	if err != nil {
		t.Fatalf("Failed to import identity: %v", err)
	}

	if newPeerID != id.PeerID {
		t.Errorf("Imported Peer ID mismatch: expected %s, got %s", id.PeerID, newPeerID)
	}
}

func TestChallengeResponse(t *testing.T) {
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	manager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	// 生成挑战
	challenge, err := manager.GenerateChallenge(id.PeerID)
	if err != nil {
		t.Fatalf("Failed to generate challenge: %v", err)
	}

	if len(challenge.Data) == 0 {
		t.Error("Challenge data should not be empty")
	}

	// 响应挑战
	response, err := manager.RespondToChallenge(id.PeerID, challenge)
	if err != nil {
		t.Fatalf("Failed to respond to challenge: %v", err)
	}

	if len(response) == 0 {
		t.Error("Challenge response should not be empty")
	}

	// 验证挑战响应
	valid, err := manager.VerifyChallengeResponse(id.PeerID, challenge, response)
	if err != nil {
		t.Fatalf("Failed to verify challenge response: %v", err)
	}

	if !valid {
		t.Error("Challenge response should be valid")
	}
}

func TestMultipleIdentities(t *testing.T) {
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	manager, err := identity.NewIdentityManager(store, keyStorage)
	if err != nil {
		t.Fatalf("Failed to create identity manager: %v", err)
	}
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	// 创建多个身份
	identities := make([]*identity.Identity, 5)
	for i := 0; i < 5; i++ {
		id, err := manager.CreateIdentity(config)
		if err != nil {
			t.Fatalf("Failed to create identity %d: %v", i, err)
		}
		identities[i] = id
	}

	// 验证所有身份都是唯一的
	peerIDs := make(map[string]bool)
	for _, id := range identities {
		if peerIDs[id.PeerID] {
			t.Errorf("Duplicate Peer ID found: %s", id.PeerID)
		}
		peerIDs[id.PeerID] = true
	}

	if len(peerIDs) != 5 {
		t.Errorf("Expected 5 unique identities, got %d", len(peerIDs))
	}
}
