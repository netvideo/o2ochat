// Package pbkdf2 provides PBKDF2 key derivation
package pbkdf2

import (
	"crypto/sha256"
	"hash"
)

// Key derives a key from password and salt using PBKDF2
func Key(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	if h == nil {
		h = sha256.New
	}

	prf := h()
	prfLen := prf.Size()
	var buf [4]byte
	dk := make([]byte, keyLen)
	u := make([]byte, prfLen)

	for i := 1; keyLen > 0; i++ {
		// Compute U_1
		prf.Reset()
		prf.Write(salt)
		buf[0] = byte(i >> 24)
		buf[1] = byte(i >> 16)
		buf[2] = byte(i >> 8)
		buf[3] = byte(i)
		prf.Write(buf[:4])
		u = prf.Sum(u[:0])

		// Compute U_2 through U_c
		dkSlice := dk
		if keyLen < prfLen {
			dkSlice = dk[:keyLen]
		}
		copy(dkSlice, u)
		for j := 2; j <= iter; j++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(u[:0])
			for k := range dkSlice {
				dkSlice[k] ^= u[k]
			}
		}

		keyLen -= prfLen
		if keyLen > 0 {
			dk = dk[prfLen:]
		}
	}

	return dk[:keyLen+prfLen]
}
