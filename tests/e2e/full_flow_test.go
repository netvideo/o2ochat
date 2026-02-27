package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/netvideo/crypto"
	"github.com/netvideo/filetransfer"
	"github.com/netvideo/identity"
	"github.com/netvideo/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndFileTransfer 测试完整文件传输流程
func TestEndToEndFileTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过端到端测试")
	}

	// 创建测试目录
	testDir := utils.CreateTestDirectory(t, "e2e-filetransfer")
	defer os.RemoveAll(testDir)

	sourceDir := filepath.Join(testDir, "source")
	destDir := filepath.Join(testDir, "dest")
	os.MkdirAll(sourceDir, 0755)
	os.MkdirAll(destDir, 0755)

	// 创建测试文件
	testFile := filepath.Join(sourceDir, "test_file.bin")
	testData := make([]byte, 1024*1024) // 1MB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	err := os.WriteFile(testFile, testData, 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 发送方：分块文件
	manager := filetransfer.NewFileTransferManager()
	defer manager.Close()

	metadata, err := manager.ChunkFile(testFile, 256*1024) // 256KB 块
	require.NoError(t, err)

	t.Logf("文件分块完成：%s (%d bytes, %d chunks)",
		metadata.FileName, metadata.FileSize, metadata.TotalChunks)

	// 验证 Merkle 树
	assert.NotEmpty(t, metadata.MerkleRoot)
	assert.Equal(t, metadata.TotalChunks, len(metadata.Chunks))

	// 验证块文件
	for i, chunk := range metadata.Chunks {
		chunkPath := filepath.Join(metadata.ChunkDir, chunk.FileName)
		_, err := os.Stat(chunkPath)
		assert.NoError(t, err, "块文件%d应存在", i)

		data, err := os.ReadFile(chunkPath)
		assert.NoError(t, err)
		assert.Equal(t, chunk.Hash, crypto.SHA256Hash(data), "块%d哈希应匹配", i)
	}

	// 合并文件
	outputFile := filepath.Join(destDir, "restored_file.bin")
	err = manager.MergeFile(outputFile, metadata)
	require.NoError(t, err)

	// 验证文件完整性
	restoredData, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, len(testData), len(restoredData), "文件大小应相同")
	assert.Equal(t, testData, restoredData, "文件内容应相同")

	t.Logf("端到端文件传输测试通过")
}

// TestEndToEndIdentityFlow 测试完整身份管理流程
func TestEndToEndIdentityFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过端到端测试")
	}

	testDir := utils.CreateTestDirectory(t, "e2e-identity")
	defer os.RemoveAll(testDir)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// 创建身份管理器
	manager := identity.NewIdentityManager()
	defer manager.Close()

	// 1. 创建身份
	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	require.NoError(t, err)
	assert.NotEmpty(t, id.PeerID)
	assert.NotEmpty(t, id.PublicKey)

	t.Logf("身份创建成功：%s", id.PeerID)

	// 2. 签名验证
	message := []byte("测试消息")
	signature, err := manager.SignMessage(message)
	require.NoError(t, err)

	valid := manager.VerifySignature(id.PeerID, message, signature)
	assert.True(t, valid, "签名验证应通过")

	t.Logf("签名验证通过")

	// 3. 导出身份
	password := "test-password-123"
	exportData, err := manager.ExportIdentity(id.PeerID, password)
	require.NoError(t, err)
	assert.NotEmpty(t, exportData)

	t.Logf("身份导出成功：%d 字节", len(exportData))

	// 4. 导入身份（新管理器）
	manager2 := identity.NewIdentityManager()
	defer manager2.Close()

	importedID, err := manager2.ImportIdentity(exportData, password)
	require.NoError(t, err)
	assert.Equal(t, id.PeerID, importedID.PeerID, "Peer ID 应相同")

	t.Logf("身份导入成功：%s", importedID.PeerID)

	// 5. 验证导入的身份可以签名
	signature2, err := manager2.SignMessage(message)
	require.NoError(t, err)

	valid2 := manager2.VerifySignature(importedID.PeerID, message, signature2)
	assert.True(t, valid2, "导入身份的签名验证应通过")

	t.Logf("导入身份签名验证通过")

	// 6. 错误密码测试
	_, err = manager2.ImportIdentity(exportData, "wrong-password")
	assert.Error(t, err, "错误密码应导致导入失败")

	t.Logf("错误密码测试通过")
}

// TestEndToEndCryptoFlow 测试完整加密流程
func TestEndToEndCryptoFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过端到端测试")
	}

	testDir := utils.CreateTestDirectory(t, "e2e-crypto")
	defer os.RemoveAll(testDir)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	manager := crypto.NewCryptoManager()

	// 1. 生成密钥对
	keyPair, err := manager.GenerateKeyPair(crypto.AlgorithmEd25519)
	require.NoError(t, err)
	assert.NotEmpty(t, keyPair.PublicKey)
	assert.NotEmpty(t, keyPair.PrivateKey)

	t.Logf("密钥对生成成功")

	// 2. 签名验证
	message := []byte("测试加密消息")
	signedMsg, err := manager.Sign(message, keyPair.PrivateKey)
	require.NoError(t, err)

	valid, err := manager.Verify(signedMsg, keyPair.PublicKey)
	require.NoError(t, err)
	assert.True(t, valid, "签名验证应通过")

	t.Logf("签名验证通过")

	// 3. AES-GCM 加密
	config := &crypto.EncryptionConfig{
		Algorithm: crypto.AlgorithmAESGCM,
		KeySize:   32,
		NonceSize: 12,
		TagSize:   16,
	}

	key := manager.RandomBytes(32)
	plaintext := []byte("敏感数据")

	ciphertext, err := manager.Encrypt(plaintext, key, config)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext, "密文应不同于明文")

	t.Logf("加密成功：%d → %d 字节", len(plaintext), len(ciphertext))

	// 4. 解密
	decrypted, err := manager.Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted, "解密数据应相同")

	t.Logf("解密成功")

	// 5. 密钥交换
	exchange := crypto.NewKeyExchange(manager)

	sessionID1, err := exchange.Initiate()
	require.NoError(t, err)

	sessionID2, err := exchange.Respond(sessionID1)
	require.NoError(t, err)

	sharedKey1, err := exchange.Finalize(sessionID1)
	require.NoError(t, err)

	sharedKey2, err := exchange.Finalize(sessionID2)
	require.NoError(t, err)

	assert.Equal(t, sharedKey1, sharedKey2, "共享密钥应相同")

	t.Logf("密钥交换成功")

	// 6. 清理
	exchange.DestroySession(sessionID1)
	exchange.DestroySession(sessionID2)
}
