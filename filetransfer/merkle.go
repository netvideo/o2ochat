package filetransfer

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

type MerkleTreeImpl struct {
	chunks   [][]byte
	tree     [][]byte
	rootHash []byte
}

func NewMerkleTree() MerkleTree {
	return &MerkleTreeImpl{}
}

func (m *MerkleTreeImpl) BuildTree(chunks [][]byte) ([][]byte, error) {
	if len(chunks) == 0 {
		return nil, errors.New("no chunks provided")
	}

	m.chunks = make([][]byte, len(chunks))
	copy(m.chunks, chunks)

	level := make([][]byte, len(chunks))
	for i, chunk := range chunks {
		h := sha256.Sum256(chunk)
		level[i] = h[:]
		m.tree = append(m.tree, h[:])
	}

	for len(level) > 1 {
		nextLevel := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			left := level[i]
			var right []byte
			if i+1 < len(level) {
				right = level[i+1]
			} else {
				right = left
			}

			combined := make([]byte, len(left)+len(right))
			copy(combined, left)
			copy(combined[len(left):], right)

			h := sha256.Sum256(combined)
			nextLevel = append(nextLevel, h[:])
			m.tree = append(m.tree, h[:])
		}
		level = nextLevel
	}

	m.rootHash = level[0]
	return m.tree, nil
}

func (m *MerkleTreeImpl) GetRootHash() []byte {
	if m.rootHash == nil {
		return nil
	}
	result := make([]byte, len(m.rootHash))
	copy(result, m.rootHash)
	return result
}

func (m *MerkleTreeImpl) VerifyChunk(index int, chunkData []byte, proof [][]byte) (bool, error) {
	if index < 0 || index >= len(m.chunks) {
		return false, ErrInvalidChunkIndex
	}

	h := sha256.Sum256(chunkData)
	currentHash := h[:]

	currentIdx := index

	for i := 0; i < len(proof); i++ {
		var combined []byte
		if currentIdx%2 == 0 {
			combined = append(currentHash, proof[i]...)
		} else {
			combined = append(proof[i], currentHash...)
		}

		h = sha256.Sum256(combined)
		currentHash = h[:]

		currentIdx = currentIdx / 2
	}

	for i := range currentHash {
		if currentHash[i] != m.rootHash[i] {
			return false, nil
		}
	}

	return true, nil
}

func (m *MerkleTreeImpl) GenerateProof(index int) ([][]byte, error) {
	if index < 0 || index >= len(m.chunks) {
		return nil, ErrInvalidChunkIndex
	}

	var proof [][]byte
	level := make([][]byte, len(m.chunks))
	for i, chunk := range m.chunks {
		h := sha256.Sum256(chunk)
		level[i] = h[:]
	}

	currentIdx := index

	for len(level) > 1 {
		var siblingHash []byte

		if currentIdx%2 == 0 {
			if currentIdx+1 < len(level) {
				siblingHash = level[currentIdx+1]
			} else {
				siblingHash = level[currentIdx]
			}
		} else {
			siblingHash = level[currentIdx-1]
		}

		proof = append(proof, siblingHash)

		nextLevel := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			left := level[i]
			var right []byte
			if i+1 < len(level) {
				right = level[i+1]
			} else {
				right = left
			}

			combined := make([]byte, len(left)+len(right))
			copy(combined, left)
			copy(combined[len(left):], right)

			h := sha256.Sum256(combined)
			nextLevel = append(nextLevel, h[:])
		}

		level = nextLevel
		currentIdx = currentIdx / 2
	}

	return proof, nil
}

func (m *MerkleTreeImpl) Serialize() ([]byte, error) {
	if m.rootHash == nil {
		return nil, ErrMerkleTreeBuild
	}

	data := make([]byte, 0)

	rootLen := make([]byte, 4)
	binary.BigEndian.PutUint32(rootLen, uint32(len(m.rootHash)))
	data = append(data, rootLen...)
	data = append(data, m.rootHash...)

	chunksLen := make([]byte, 4)
	binary.BigEndian.PutUint32(chunksLen, uint32(len(m.chunks)))
	data = append(data, chunksLen...)

	for _, chunk := range m.chunks {
		chunkLen := make([]byte, 4)
		binary.BigEndian.PutUint32(chunkLen, uint32(len(chunk)))
		data = append(data, chunkLen...)
		data = append(data, chunk...)
	}

	return data, nil
}

func (m *MerkleTreeImpl) Deserialize(data []byte) error {
	if len(data) < 4 {
		return ErrMerkleTreeVerify
	}

	offset := 0

	rootLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	if len(data) < offset+int(rootLen) {
		return ErrMerkleTreeVerify
	}
	m.rootHash = make([]byte, rootLen)
	copy(m.rootHash, data[offset:offset+int(rootLen)])
	offset += int(rootLen)

	if len(data) < offset+4 {
		return ErrMerkleTreeVerify
	}
	numChunks := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	m.chunks = make([][]byte, 0, numChunks)
	for i := 0; i < int(numChunks); i++ {
		if len(data) < offset+4 {
			return ErrMerkleTreeVerify
		}
		chunkLen := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4

		if len(data) < offset+int(chunkLen) {
			return ErrMerkleTreeVerify
		}
		chunk := make([]byte, chunkLen)
		copy(chunk, data[offset:offset+int(chunkLen)])
		m.chunks = append(m.chunks, chunk)
		offset += int(chunkLen)
	}

	return nil
}
