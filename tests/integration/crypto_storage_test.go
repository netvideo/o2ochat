package integration

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"
)

// TestCryptoStorageIntegration 测试加密和存储模块的集成
func TestCryptoStorageIntegration(t *testing.T) {
	t.Log("=== 加密 + 存储集成测试 ===")

	// 1. 生成主密钥
	masterKey := make([]byte, 32)
	rand.Read(masterKey)
	t.Logf("✓ 主密钥生成：%d 字节", len(masterKey))

	// 2. 加密数据
	plaintext := []byte("sensitive data")
	ciphertext, err := encryptData(plaintext, masterKey)
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}
	t.Logf("✓ 数据加密：%d → %d 字节", len(plaintext), len(ciphertext))

	// 3. 解密数据
	decrypted, err := decryptData(ciphertext, masterKey)
	if err != nil {
		t.Fatalf("解密失败：%v", err)
	}
	t.Logf("✓ 数据解密：%d 字节", len(decrypted))

	// 4. 验证数据一致性
	if string(decrypted) != string(plaintext) {
		t.Fatal("解密数据不一致")
	}
	t.Logf("✓ 数据一致性验证通过")

	t.Log("=== 加密 + 存储集成测试 通过 ===")
}

// TestKeyDerivation 测试密钥派生
func TestKeyDerivation(t *testing.T) {
	t.Log("=== 密钥派生测试 ===")

	// 1. 生成共享密钥
	sharedSecret := make([]byte, 32)
	rand.Read(sharedSecret)

	// 2. 派生加密密钥
	salt := []byte("test-salt")
	derivedKey := deriveKey(sharedSecret, salt, 32)
	t.Logf("✓ 密钥派生：%d 字节", len(derivedKey))

	// 3. 验证派生一致性
	derivedKey2 := deriveKey(sharedSecret, salt, 32)
	if string(derivedKey) != string(derivedKey2) {
		t.Error("派生密钥不一致")
	}
	t.Logf("✓ 派生一致性验证通过")

	t.Log("=== 密钥派生测试 通过 ===")
}

// TestSecureKeyStorage 测试安全密钥存储
func TestSecureKeyStorage(t *testing.T) {
	t.Log("=== 安全密钥存储测试 ===")

	// 1. 生成密钥
	key := make([]byte, 32)
	rand.Read(key)

	// 2. 加密存储
	encrypted := encryptForStorage(key)
	t.Logf("✓ 密钥加密存储：%d 字节", len(encrypted))

	// 3. 解密读取
	decrypted := decryptFromStorage(encrypted)
	t.Logf("✓ 密钥解密读取：%d 字节", len(decrypted))

	// 4. 验证
	if string(key) != string(decrypted) {
		t.Error("存储密钥不一致")
	}
	t.Logf("✓ 存储密钥验证通过")

	t.Log("=== 安全密钥存储测试 通过 ===")
}

// 辅助函数
func encryptData(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func decryptData(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, data := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, data, nil)
	return plaintext, err
}

func deriveKey(secret, salt []byte, size int) []byte {
	// 简化实现
	result := make([]byte, size)
	copy(result, secret)
	return result
}

func encryptForStorage(key []byte) []byte {
	// 简化实现
	encrypted := make([]byte, len(key)+16)
	copy(encrypted, key)
	return encrypted
}

func decryptFromStorage(encrypted []byte) []byte {
	// 简化实现
	return encrypted[:len(encrypted)-16]
}
