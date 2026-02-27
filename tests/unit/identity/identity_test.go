package identity_test

import (
	"testing"

	"github.com/netvideo/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIdentity(t *testing.T) {
	// 准备
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	// 执行
	id, err := manager.CreateIdentity(config)

	// 验证
	require.NoError(t, err, "创建身份不应出错")
	assert.NotEmpty(t, id.PeerID, "Peer ID 不应为空")
	assert.NotEmpty(t, id.PublicKey, "公钥不应为空")
	assert.Equal(t, 32, len(id.PublicKey), "Ed25519 公钥应为 32 字节")
	assert.NotZero(t, id.CreatedAt, "创建时间不应为零")
}

func TestSignAndVerify(t *testing.T) {
	// 准备
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	message := []byte("测试消息")

	// 执行
	signature, err := manager.SignMessage(message)
	require.NoError(t, err, "签名不应出错")

	valid := manager.VerifySignature(id.PeerID, message, signature)

	// 验证
	assert.True(t, valid, "签名验证应通过")
	assert.NotEmpty(t, signature, "签名不应为空")
	assert.Equal(t, 64, len(signature), "Ed25519 签名应为 64 字节")
}

func TestExportImportIdentity(t *testing.T) {
	// 准备
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	originalID, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	password := "test-password"

	// 执行
	exportData, err := manager.ExportIdentity(originalID.PeerID, password)
	require.NoError(t, err, "导出不应出错")

	importedID, err := manager.ImportIdentity(exportData, password)
	require.NoError(t, err, "导入不应出错")

	// 验证
	assert.Equal(t, originalID.PeerID, importedID.PeerID, "Peer ID 应相同")
	assert.Equal(t, originalID.PublicKey, importedID.PublicKey, "公钥应相同")
}

func TestInvalidPassword(t *testing.T) {
	// 准备
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	exportData, err := manager.ExportIdentity(id.PeerID, "correct-password")
	require.NoError(t, err)

	// 执行和验证
	_, err = manager.ImportIdentity(exportData, "wrong-password")
	assert.Error(t, err, "错误密码应导致导入失败")
}

func TestDeleteIdentity(t *testing.T) {
	// 准备
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	// 执行
	err = manager.DeleteIdentity(id.PeerID)
	require.NoError(t, err, "删除身份不应出错")

	// 验证
	_, err = manager.LoadIdentity(id.PeerID)
	assert.Error(t, err, "删除后加载身份应失败")
}

func TestListIdentities(t *testing.T) {
	// 准备
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	// 创建多个身份
	numIdentities := 3
	expectedIDs := make(map[string]bool)

	for i := 0; i < numIdentities; i++ {
		id, err := manager.CreateIdentity(config)
		require.NoError(t, err)
		expectedIDs[id.PeerID] = true
	}

	// 执行
	list, err := manager.ListIdentities()
	require.NoError(t, err)

	// 验证
	assert.Equal(t, numIdentities, len(list), "身份列表长度应正确")

	for _, peerID := range list {
		assert.True(t, expectedIDs[peerID], "Peer ID 应在预期列表中")
	}
}
