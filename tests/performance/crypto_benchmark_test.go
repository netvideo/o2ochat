package performance

import (
	"testing"

	"github.com/netvideo/crypto"
)

func BenchmarkEncryption(b *testing.B) {
	manager := crypto.NewCryptoManager()
	config := &crypto.EncryptionConfig{
		Algorithm: crypto.AlgorithmAESGCM,
		KeySize:   32,
		NonceSize: 12,
		TagSize:   16,
	}

	key := manager.RandomBytes(32)
	data := make([]byte, 1024) // 1KB 数据

	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		encrypted, err := manager.Encrypt(data, key, config)
		if err != nil {
			b.Fatalf("加密失败：%v", err)
		}

		_, err = manager.Decrypt(encrypted, key)
		if err != nil {
			b.Fatalf("解密失败：%v", err)
		}
	}
}

func BenchmarkSigning(b *testing.B) {
	manager := crypto.NewCryptoManager()
	keyPair, err := manager.GenerateKeyPair(crypto.AlgorithmEd25519)
	if err != nil {
		b.Fatalf("生成密钥对失败：%v", err)
	}

	message := make([]byte, 256)

	b.ResetTimer()
	b.SetBytes(int64(len(message)))

	for i := 0; i < b.N; i++ {
		signedMsg, err := manager.Sign(message, keyPair.PrivateKey)
		if err != nil {
			b.Fatalf("签名失败：%v", err)
		}

		valid, err := manager.Verify(signedMsg, keyPair.PublicKey)
		if err != nil || !valid {
			b.Fatalf("验证失败：%v", err)
		}
	}
}

func BenchmarkHashCalculation(b *testing.B) {
	manager := crypto.NewCryptoManager()
	data := make([]byte, 1024*1024) // 1MB 数据

	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		_, err := manager.Hash(data, crypto.AlgorithmSHA256)
		if err != nil {
			b.Fatalf("哈希计算失败：%v", err)
		}
	}
}

func BenchmarkKeyGeneration(b *testing.B) {
	manager := crypto.NewCryptoManager()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateKeyPair(crypto.AlgorithmEd25519)
		if err != nil {
			b.Fatalf("生成密钥对失败：%v", err)
		}
	}
}

func BenchmarkKeyExchange(b *testing.B) {
	manager := crypto.NewCryptoManager()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		exchange := crypto.NewKeyExchange(manager)
		sessionID, err := exchange.Initiate()
		if err != nil {
			b.Fatalf("密钥交换初始化失败：%v", err)
		}

		exchange.Finalize(sessionID)
		exchange.DestroySession(sessionID)
	}
}
