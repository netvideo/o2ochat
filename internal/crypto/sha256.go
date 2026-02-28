package sha256

import (
	"crypto/sha256"
	"hash"
)

// Sum256 计算 SHA-256 校验和
func Sum256(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// New 创建新的 SHA-256 hash.Hash
func New() hash.Hash {
	return sha256.New()
}
