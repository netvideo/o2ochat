package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/netvideo/crypto"
	"github.com/netvideo/identity"
	"github.com/netvideo/signaling"
	"github.com/netvideo/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIdentityCryptoStorageIntegration 测试身份 + 加密 + 存储集成
func TestIdentityCryptoStorageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建临时目录
	testDir := t.TempDir()

	// 1. 创建存储管理器
	storageManager, err := storage.NewSQLiteStorageManager(testDir)
	require.NoError(t, err)
	defer storageManager.Close()

	// 2. 创建加密管理器
	cryptoManager := crypto.NewCryptoManager()

	// 3. 创建身份管理器
	identityManager := identity.NewIdentityManager()
	defer identityManager.Close()

	// 4. 创建身份
	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := identityManager.CreateIdentity(config)
	require.NoError(t, err)
	t.Logf("✓ 身份创建：%s", id.PeerID)

	// 5. 导出身份（加密）
	password := "integration-test-password"
	exportData, err := identityManager.ExportIdentity(id.PeerID, password)
	require.NoError(t, err)
	t.Logf("✓ 身份导出：%d 字节", len(exportData))

	// 6. 使用加密管理器加密导出数据
	encryptConfig := &crypto.EncryptionConfig{
		Algorithm: crypto.AlgorithmAESGCM,
		KeySize:   32,
		NonceSize: 12,
		TagSize:   16,
	}

	key := cryptoManager.RandomBytes(32)
	encryptedData, err := cryptoManager.Encrypt(exportData, key, encryptConfig)
	require.NoError(t, err)
	t.Logf("✓ 数据加密：%d → %d 字节", len(exportData), len(encryptedData))

	// 7. 存储到 SQLite
	err = storageManager.SaveEncryptedKey(id.PeerID, "identity", encryptedData)
	require.NoError(t, err)
	t.Logf("✓ 加密数据存储完成")

	// 8. 从存储读取
	storedData, err := storageManager.LoadEncryptedKey(id.PeerID, "identity")
	require.NoError(t, err)
	assert.Equal(t, encryptedData, storedData, "存储数据应相同")
	t.Logf("✓ 加密数据读取完成")

	// 9. 解密
	decryptedData, err := cryptoManager.Decrypt(storedData, key)
	require.NoError(t, err)
	assert.Equal(t, exportData, decryptedData, "解密数据应相同")
	t.Logf("✓ 数据解密完成")

	// 10. 导入身份
	importedID, err := identityManager.ImportIdentity(decryptedData, password)
	require.NoError(t, err)
	assert.Equal(t, id.PeerID, importedID.PeerID, "Peer ID 应相同")
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

	// 1. 创建身份
	identityManager := identity.NewIdentityManager()
	defer identityManager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	peerA, err := identityManager.CreateIdentity(config)
	require.NoError(t, err)

	peerB, err := identityManager.CreateIdentity(config)
	require.NoError(t, err)

	t.Logf("✓ 创建两个身份：%s, %s", peerA.PeerID, peerB.PeerID)

	// 2. 创建信令消息
	message := &signaling.SignalingMessage{
		Type:      signaling.MessageTypeOffer,
		From:      peerA.PeerID,
		To:        peerB.PeerID,
		Data:      map[string]interface{}{"type": "offer", "sdp": "test-sdp"},
		Timestamp: time.Now(),
	}

	// 3. 签名消息
	messageData := message.Serialize()
	signature, err := identityManager.SignMessage(messageData)
	require.NoError(t, err)
	message.Signature = signature

	t.Logf("✓ 消息签名完成")

	// 4. 验证签名
	valid := identityManager.VerifySignature(peerA.PeerID, messageData, signature)
	assert.True(t, valid, "签名验证应通过")

	t.Logf("✓ 签名验证通过")

	// 5. 模拟接收方验证
	messageStr := message.Serialize()
	valid = identityManager.VerifySignature(peerA.PeerID, []byte(messageStr), message.Signature)
	assert.True(t, valid, "接收方验证应通过")

	t.Logf("✓ 接收方验证通过")

	t.Log("=== 带身份验证的信令测试 通过 ===")
}

// TestFullStackRegistration 测试完整注册流程
func TestFullStackRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	testDir := t.TempDir()

	// 1. 初始化所有组件
	storageManager, err := storage.NewSQLiteStorageManager(testDir)
	require.NoError(t, err)
	defer storageManager.Close()

	cryptoManager := crypto.NewCryptoManager()
	identityManager := identity.NewIdentityManager()
	defer identityManager.Close()

	t.Log("✓ 组件初始化完成")

	// 2. 创建身份
	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := identityManager.CreateIdentity(config)
	require.NoError(t, err)
	t.Logf("✓ 身份创建：%s", id.PeerID)

	// 3. 生成密钥加密密钥
	kek := cryptoManager.RandomBytes(32)
	t.Logf("✓ 密钥加密密钥生成")

	// 4. 导出并加密身份
	password := "registration-password"
	exportData, err := identityManager.ExportIdentity(id.PeerID, password)
	require.NoError(t, err)

	encryptConfig := &crypto.EncryptionConfig{
		Algorithm: crypto.AlgorithmAESGCM,
		KeySize:   32,
	}

	encryptedIdentity, err := cryptoManager.Encrypt(exportData, kek, encryptConfig)
	require.NoError(t, err)
	t.Logf("✓ 身份加密完成")

	// 5. 存储
	err = storageManager.SaveEncryptedKey(id.PeerID, "identity", encryptedIdentity)
	require.NoError(t, err)
	t.Logf("✓ 加密身份存储完成")

	// 6. 验证存储
	stored, err := storageManager.LoadEncryptedKey(id.PeerID, "identity")
	require.NoError(t, err)

	// 7. 解密并导入
	decrypted, err := cryptoManager.Decrypt(stored, kek)
	require.NoError(t, err)

	importedID, err := identityManager.ImportIdentity(decrypted, password)
	require.NoError(t, err)
	assert.Equal(t, id.PeerID, importedID.PeerID)
	t.Logf("✓ 身份恢复验证成功")

	t.Log("=== 完整注册流程测试 通过 ===")
}
