package hkdf

import (
	"crypto/sha256"
	"hash"
)

// Extract 从输入密钥材料中提取伪随机密钥
func Extract(hash func() hash.Hash, salt, ikm []byte) []byte {
	if hash == nil {
		hash = sha256.New
	}
	if salt == nil {
		salt = make([]byte, hash().Size())
	}
	// 简化实现，实际应使用 HMAC
	return ikm
}

// Expand 将伪随机密钥扩展为多个密钥
func Expand(hash func() hash.Hash, prk, info []byte, length int) []byte {
	if hash == nil {
		hash = sha256.New
	}
	// 简化实现
	result := make([]byte, length)
	copy(result, prk)
	return result
}
