package performance

import (
	"testing"

	"github.com/netvideo/crypto"
)

func BenchmarkEncryption(b *testing.B) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}
	manager := crypto.NewCryptoManager(config)
	encConfig := &crypto.EncryptionConfig{
		Algorithm: crypto.AlgorithmAESGCM,
		KeySize:   32,
		NonceSize: 12,
		TagSize:   16,
	}

	key, err := manager.RandomBytes(32)
	if err != nil {
		b.Fatalf("生成随机密钥失败: %v", err)
	}
	data := make([]byte, 1024)

	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		encrypted, err := manager.Encrypt(data, key, encConfig)
		if err != nil {
			b.Fatalf("加密失败: %v", err)
		}

		_, err = manager.Decrypt(encrypted, key)
		if err != nil {
			b.Fatalf("解密失败: %v", err)
		}
	}
}

func BenchmarkSigning(b *testing.B) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}
	manager := crypto.NewCryptoManager(config)
	keyPair, err := manager.GenerateKeyPair(crypto.AlgorithmEd25519)
	if err != nil {
		b.Fatalf("生成密钥对失败: %v", err)
	}

	message := make([]byte, 256)

	b.ResetTimer()
	b.SetBytes(int64(len(message)))

	for i := 0; i < b.N; i++ {
		signedMsg, err := manager.Sign(message, keyPair.PrivateKey)
		if err != nil {
			b.Fatalf("签名失败: %v", err)
		}

		valid, err := manager.Verify(signedMsg, keyPair.PublicKey)
		if err != nil || !valid {
			b.Fatalf("验证失败: %v", err)
		}
	}
}

func BenchmarkHashCalculation(b *testing.B) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}
	manager := crypto.NewCryptoManager(config)
	data := make([]byte, 1024*1024)

	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		_, err := manager.Hash(data, crypto.AlgorithmSHA256)
		if err != nil {
			b.Fatalf("哈希计算失败: %v", err)
		}
	}
}

func BenchmarkKeyGeneration(b *testing.B) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}
	manager := crypto.NewCryptoManager(config)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateKeyPair(crypto.AlgorithmEd25519)
		if err != nil {
			b.Fatalf("生成密钥对失败: %v", err)
		}
	}
}

func BenchmarkKeyExchange(b *testing.B) {
	config := &crypto.SecurityConfig{
		EncryptionAlgorithm:  crypto.AlgorithmAESGCM,
		EncryptionKeySize:    256,
		SignatureAlgorithm:   crypto.AlgorithmEd25519,
		KeyExchangeAlgorithm: crypto.AlgorithmX25519,
		HashAlgorithm:        crypto.AlgorithmSHA256,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		exchange := crypto.NewKeyExchange(config)
		result, err := exchange.Initiate()
		if err != nil {
			b.Fatalf("密钥交换初始化失败: %v", err)
		}

		_, err = exchange.Finalize(result.PublicKey, result.SessionID)
		if err != nil {
			b.Fatalf("密钥交换完成失败: %v", err)
		}

		exchange.DestroySession(result.SessionID)
	}
}
