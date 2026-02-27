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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEndFileTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过端到端测试")
	}

	tempDir, err := os.MkdirTemp("", "e2e-filetransfer")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")
	os.MkdirAll(sourceDir, 0755)
	os.MkdirAll(destDir, 0755)

	testFile := filepath.Join(sourceDir, "test_file.bin")
	testData := make([]byte, 1024*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	err = os.WriteFile(testFile, testData, 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	_ = ctx

	chunkManager, err := filetransfer.NewChunkManager(256*1024, tempDir)
	require.NoError(t, err)

	scheduler := filetransfer.NewScheduler()
	transferManager := filetransfer.NewFileTransferManager(chunkManager, scheduler, 4)
	_ = transferManager

	metadata, err := chunkManager.ChunkFile(testFile, 256*1024)
	require.NoError(t, err)

	t.Logf("文件分块完成：%s (%d bytes, %d chunks)",
		metadata.FileName, metadata.FileSize, metadata.TotalChunks)

	assert.NotEmpty(t, metadata.MerkleRoot)

	cryptoManager := crypto.NewCryptoManager(&crypto.SecurityConfig{
		HashAlgorithm: crypto.AlgorithmSHA256,
	})
	hash, err := cryptoManager.Hash(testData, crypto.AlgorithmSHA256)
	require.NoError(t, err)
	assert.Equal(t, hash, metadata.MerkleRoot, "Merkle根哈希应匹配")

	outputFile := filepath.Join(destDir, "restored_file.bin")
	err = chunkManager.MergeFile(metadata.FileID, outputFile)
	require.NoError(t, err)

	restoredData, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, len(testData), len(restoredData), "文件大小应相同")
	assert.Equal(t, testData, restoredData, "文件内容应相同")
}

func TestIdentityCreation(t *testing.T) {
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	manager, err := identity.NewIdentityManager(store, keyStorage)
	require.NoError(t, err)

	config := &identity.IdentityConfig{
		KeyType:   identity.KeyTypeEd25519,
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	assert.NotEmpty(t, id.PeerID)
	assert.NotEmpty(t, id.PublicKey)
	assert.NotEmpty(t, id.PrivateKey)

	t.Logf("创建身份成功: %s", id.PeerID)
}

func TestKeyExchange(t *testing.T) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}

	alice := crypto.NewKeyExchange(config)
	bob := crypto.NewKeyExchange(config)

	aliceResult, err := alice.Initiate()
	require.NoError(t, err)

	bobResult, err := bob.Respond(aliceResult.PublicKey)
	require.NoError(t, err)

	aliceSharedSecret, err := alice.Finalize(bobResult.PublicKey, aliceResult.SessionID)
	require.NoError(t, err)

	bobSharedSecret, err := bob.Finalize(aliceResult.PublicKey, bobResult.SessionID)
	require.NoError(t, err)

	assert.Equal(t, aliceSharedSecret, bobSharedSecret, "共享密钥应匹配")

	t.Logf("密钥交换成功")
}

func TestEncryptionDecryption(t *testing.T) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm: crypto.AlgorithmAESGCM,
		EncryptionKeySize:   256,
		HashAlgorithm:       crypto.AlgorithmSHA256,
	}

	manager := crypto.NewCryptoManager(config)

	plaintext := []byte("测试加密数据")
	key := make([]byte, 32)
	randomBytes, err := manager.RandomBytes(32)
	require.NoError(t, err)
	copy(key, randomBytes)

	encConfig := &crypto.EncryptionConfig{
		Algorithm: crypto.AlgorithmAESGCM,
		KeySize:   32,
		NonceSize: 12,
		TagSize:   16,
	}

	encrypted, err := manager.Encrypt(plaintext, key, encConfig)
	require.NoError(t, err)

	decrypted, err := manager.Decrypt(encrypted, key)
	require.NoError(t, err)

	assert.Equal(t, plaintext, decrypted, "解密后应得到原始数据")

	t.Logf("加密解密成功")
}

func TestSigningVerification(t *testing.T) {
	config := &crypto.SecurityConfig{
		SignatureAlgorithm: crypto.AlgorithmEd25519,
		HashAlgorithm:      crypto.AlgorithmSHA256,
	}

	manager := crypto.NewCryptoManager(config)

	keyPair, err := manager.GenerateKeyPair(crypto.AlgorithmEd25519)
	require.NoError(t, err)

	message := []byte("测试签名数据")

	signedMsg, err := manager.Sign(message, keyPair.PrivateKey)
	require.NoError(t, err)

	valid, err := manager.Verify(signedMsg, keyPair.PublicKey)
	require.NoError(t, err)
	assert.True(t, valid, "签名验证应通过")

	t.Logf("签名验证成功")
}
