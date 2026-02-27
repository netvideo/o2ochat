package security

import (
	"testing"

	"github.com/netvideo/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIdentityValidation 测试身份验证安全性
func TestIdentityValidation(t *testing.T) {
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	// 1. 验证有效 Peer ID
	assert.True(t, identity.ValidatePeerID(id.PeerID), "有效 Peer ID 应通过验证")

	// 2. 验证无效 Peer ID
	invalidIDs := []string{
		"",
		"invalid",
		"Qm",
		"QmInvalidPeerIDFormat",
	}

	for _, invalidID := range invalidIDs {
		assert.False(t, identity.ValidatePeerID(invalidID), "无效 Peer ID 应验证失败：%s", invalidID)
	}

	t.Logf("身份验证测试通过")
}

// TestSignatureForgery 测试签名伪造防护
func TestSignatureForgery(t *testing.T) {
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	// 1. 正常签名
	message := []byte("测试消息")
	signature, err := manager.SignMessage(message)
	require.NoError(t, err)

	// 2. 验证正常签名
	assert.True(t, manager.VerifySignature(id.PeerID, message, signature), "正常签名应验证通过")

	// 3. 篡改消息
	tamperedMessage := []byte("篡改后的消息")
	assert.False(t, manager.VerifySignature(id.PeerID, tamperedMessage, signature), "篡改消息应验证失败")

	// 4. 篡改签名
	tamperedSignature := make([]byte, len(signature))
	copy(tamperedSignature, signature)
	tamperedSignature[0] ^= 0xFF // 翻转第一位

	assert.False(t, manager.VerifySignature(id.PeerID, message, tamperedSignature), "篡改签名应验证失败")

	// 5. 使用其他身份的签名
	id2, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	signature2, err := manager.SignMessage(message)
	require.NoError(t, err)

	assert.False(t, manager.VerifySignature(id.PeerID, message, signature2), "其他身份的签名应验证失败")

	t.Logf("签名伪造防护测试通过")
}

// TestKeyStorageSecurity 测试密钥存储安全
func TestKeyStorageSecurity(t *testing.T) {
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	// 1. 导出身份（加密）
	password := "secure-password-123"
	exportData, err := manager.ExportIdentity(id.PeerID, password)
	require.NoError(t, err)

	// 2. 验证导出数据已加密
	assert.NotEmpty(t, exportData)
	// 导出数据应包含加密标识和密文
	assert.Greater(t, len(exportData), 100, "导出数据应有合理长度")

	// 3. 错误密码无法解密
	_, err = manager.ImportIdentity(exportData, "wrong-password")
	assert.Error(t, err, "错误密码应导致解密失败")

	// 4. 空密码测试
	_, err = manager.ImportIdentity(exportData, "")
	assert.Error(t, err, "空密码应导致解密失败")

	t.Logf("密钥存储安全测试通过")
}

// TestMetadataSecurity 测试元数据安全
func TestMetadataSecurity(t *testing.T) {
	manager := identity.NewIdentityManager()
	defer manager.Close()

	config := &identity.IdentityConfig{
		KeyType:   "ed25519",
		KeyLength: 256,
	}

	id, err := manager.CreateIdentity(config)
	require.NoError(t, err)

	// 1. 设置元数据
	metadata := map[string]interface{}{
		"name":      "测试用户",
		"email":     "test@example.com",
		"createdAt": "2026-02-27",
	}

	err = manager.UpdateMetadata(id.PeerID, metadata)
	require.NoError(t, err)

	// 2. 获取元数据
	retrieved, err := manager.GetMetadata(id.PeerID)
	require.NoError(t, err)
	assert.Equal(t, metadata["name"], retrieved["name"], "元数据应相同")

	// 3. 更新元数据
	newMetadata := map[string]interface{}{
		"name": "更新后的名称",
	}

	err = manager.UpdateMetadata(id.PeerID, newMetadata)
	require.NoError(t, err)

	// 4. 验证更新
	updated, err := manager.GetMetadata(id.PeerID)
	require.NoError(t, err)
	assert.Equal(t, newMetadata["name"], updated["name"], "元数据应已更新")

	t.Logf("元数据安全测试通过")
}
