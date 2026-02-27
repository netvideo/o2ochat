package security

import (
	"testing"

	"github.com/netvideo/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentityValidation(t *testing.T) {
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

	peerIDGen := identity.NewPeerIDGenerator()
	assert.True(t, peerIDGen.ValidatePeerID(id.PeerID), "有效 Peer ID 应通过验证")

	invalidIDs := []string{
		"",
		"invalid",
		"Qm",
		"QmInvalidPeerIDFormat",
	}

	for _, invalidID := range invalidIDs {
		assert.False(t, peerIDGen.ValidatePeerID(invalidID), "无效 Peer ID 应验证失败: %s", invalidID)
	}

	t.Logf("身份验证测试通过")
}

func TestSignatureForgery(t *testing.T) {
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

	message := []byte("测试消息")
	signature, err := manager.SignMessage(message)
	require.NoError(t, err)

	assert.True(t, manager.VerifySignature(id.PeerID, message, signature), "正常签名应验证通过")

	tamperedMessage := []byte("篡改后的消息")
	assert.False(t, manager.VerifySignature(id.PeerID, tamperedMessage, signature), "篡改消息应验证失败")

	tamperedSignature := make([]byte, len(signature))
	copy(tamperedSignature, signature)
	tamperedSignature[0] ^= 0xFF

	assert.False(t, manager.VerifySignature(id.PeerID, message, tamperedSignature), "篡改签名应验证失败")

	t.Logf("签名伪造防护测试通过")
}

func TestKeyStorageSecurity(t *testing.T) {
	store := identity.NewMemoryIdentityStore()
	keyStorage := identity.NewMemoryKeyStorage()

	manager, err := identity.NewIdentityManager(store, keyStorage)
	require.NoError(t, err)

	config := &identity.IdentityConfig{
		KeyType:   identity.KeyTypeEd25519,
		KeyLength: 256,
	}

	_, err = manager.CreateIdentity(config)
	require.NoError(t, err)

	password := "secure-password-123"
	exportData, err := manager.ExportIdentity(password)
	require.NoError(t, err)

	assert.NotEmpty(t, exportData)
	assert.Greater(t, len(exportData), 100, "导出数据应有合理长度")

	t.Logf("密钥存储安全测试通过")
}

func TestMetadataSecurity(t *testing.T) {
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

	metadata := &identity.IdentityMetadata{
		PeerID:      id.PeerID,
		DisplayName: "测试用户",
		DeviceInfo:  "test-device",
	}

	err = manager.UpdateMetadata(id.PeerID, metadata)
	require.NoError(t, err)

	retrieved, err := manager.GetMetadata(id.PeerID)
	require.NoError(t, err)
	assert.Equal(t, metadata.DisplayName, retrieved.DisplayName, "元数据应相同")

	t.Logf("元数据安全测试通过")
}
