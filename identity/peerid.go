package identity

import (
	"crypto/sha256"
	"math/big"
)

type peerIDGenerator struct{}

var base58Chars = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var leadingZeros int
	for _, b := range data {
		if b == 0 {
			leadingZeros++
		} else {
			break
		}
	}

	result := ""
	number := new(big.Int).SetBytes(data)
	divisor := big.NewInt(58)

	for number.Sign() > 0 {
		remainder := new(big.Int)
		number, remainder = number.DivMod(number, divisor, remainder)
		result = string(base58Chars[remainder.Int64()]) + result
	}

	for i := 0; i < leadingZeros; i++ {
		result = "1" + result
	}

	return result
}

func base58Decode(s string) []byte {
	if s == "" {
		return nil
	}

	result := big.NewInt(0)
	fiftyEight := big.NewInt(58)

	for _, c := range []byte(s) {
		index := int64(-1)
		for i, b := range []byte(base58Chars) {
			if b == c {
				index = int64(i)
				break
			}
		}
		if index < 0 {
			return nil
		}
		result.Mul(result, fiftyEight)
		result.Add(result, big.NewInt(index))
	}

	leadingOnes := 0
	for _, c := range []byte(s) {
		if c == '1' {
			leadingOnes++
		} else {
			break
		}
	}

	bytes := result.Bytes()
	if leadingOnes > len(bytes) {
		padded := make([]byte, leadingOnes)
		copy(padded[leadingOnes-len(bytes):], bytes)
		return padded
	}

	return bytes
}

func NewPeerIDGenerator() PeerIDUtil {
	return &peerIDGenerator{}
}

func (p *peerIDGenerator) GeneratePeerID(publicKey []byte) string {
	return p.EncodePeerID(publicKey, PeerIDEncodingBase58)
}

func (p *peerIDGenerator) ValidatePeerID(peerID string) bool {
	if peerID == "" {
		return false
	}

	decoded := base58Decode(peerID)
	if len(decoded) == 0 {
		return false
	}

	return true
}

func (p *peerIDGenerator) ExtractPublicKeyHash(peerID string) ([]byte, error) {
	decoded := base58Decode(peerID)
	if len(decoded) == 0 {
		return nil, ErrInvalidPeerID
	}

	if len(decoded) > 16 {
		return decoded[:16], nil
	}

	return decoded, nil
}

func (p *peerIDGenerator) EncodePeerID(publicKey []byte, encoding PeerIDEncoding) string {
	hash := sha256.Sum256(publicKey)
	peerIDBytes := hash[:16]

	switch encoding {
	case PeerIDEncodingHex:
		return encodeHex(peerIDBytes)
	case PeerIDEncodingBase58:
		fallthrough
	default:
		return base58Encode(peerIDBytes)
	}
}

func (p *peerIDGenerator) DecodePeerID(peerID string, encoding PeerIDEncoding) ([]byte, error) {
	switch encoding {
	case PeerIDEncodingHex:
		return decodeHex(peerID)
	case PeerIDEncodingBase58:
		fallthrough
	default:
		decoded := base58Decode(peerID)
		if len(decoded) == 0 {
			return nil, ErrInvalidPeerID
		}
		return decoded, nil
	}
}

func encodeHex(data []byte) string {
	hexChars := "0123456789abcdef"
	result := make([]byte, len(data)*2)
	for i, b := range data {
		result[i*2] = hexChars[b>>4]
		result[i*2+1] = hexChars[b&0x0f]
	}
	return string(result)
}

func decodeHex(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, ErrInvalidFormat
	}

	result := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		high := hexCharToNibble(s[i])
		low := hexCharToNibble(s[i+1])
		if high < 0 || low < 0 {
			return nil, ErrInvalidFormat
		}
		result[i/2] = byte(high<<4 | low)
	}

	return result, nil
}

func hexCharToNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return -1
	}
}
