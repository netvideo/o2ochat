package filetransfer

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ChunkManagerImpl struct {
	chunkSize   int
	storageDir  string
	metadata    map[string]*FileMetadata
	chunks      map[string]map[int]*ChunkInfo
	mu          sync.RWMutex
	merkleTrees map[string]MerkleTree
}

func NewChunkManager(chunkSize int, storageDir string) (ChunkManager, error) {
	if chunkSize <= 0 {
		chunkSize = 1024 * 1024
	}

	if storageDir == "" {
		storageDir = "./filetransfer_storage"
	}

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, err
	}

	return &ChunkManagerImpl{
		chunkSize:   chunkSize,
		storageDir:  storageDir,
		metadata:    make(map[string]*FileMetadata),
		chunks:      make(map[string]map[int]*ChunkInfo),
		merkleTrees: make(map[string]MerkleTree),
	}, nil
}

func NewChunkManagerImpl() ChunkManager {
	cm, _ := NewChunkManager(1024*1024, "/tmp/chunks")
	return cm
}

func (c *ChunkManagerImpl) ChunkFile(filePath string, chunkSize int) (*FileMetadata, error) {
	if chunkSize <= 0 {
		chunkSize = c.chunkSize
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, ErrFileNotFound
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := stat.Size()
	totalChunks := int((fileSize + int64(chunkSize) - 1) / int64(chunkSize))

	chunks := make([][]byte, 0, totalChunks)
	chunkInfos := make(map[int]*ChunkInfo)

	for i := 0; i < totalChunks; i++ {
		chunkData := make([]byte, chunkSize)
		n, err := file.Read(chunkData)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if n < chunkSize {
			chunkData = chunkData[:n]
		}

		hash := sha256.Sum256(chunkData)
		chunkInfo := &ChunkInfo{
			FileID:    "",
			Index:     i,
			Offset:    int64(i * chunkSize),
			Size:      n,
			Hash:      hash[:],
			Completed: false,
			Verified:  false,
		}

		chunks = append(chunks, chunkData)
		chunkInfos[i] = chunkInfo
	}

	merkle := NewMerkleTree()
	_, err = merkle.BuildTree(chunks)
	if err != nil {
		return nil, ErrMerkleTreeBuild
	}

	fileID := hex.EncodeToString(merkle.GetRootHash())

	metadata := &FileMetadata{
		FileID:      fileID,
		FileName:    filepath.Base(filePath),
		FileSize:    fileSize,
		TotalChunks: totalChunks,
		ChunkSize:   chunkSize,
		MerkleRoot:  merkle.GetRootHash(),
		CreatedAt:   time.Now(),
		ModifiedAt:  stat.ModTime(),
	}

	for i := range chunkInfos {
		chunkInfos[i].FileID = fileID
	}

	c.mu.Lock()
	c.metadata[fileID] = metadata
	c.chunks[fileID] = chunkInfos
	c.merkleTrees[fileID] = merkle
	c.mu.Unlock()

	return metadata, nil
}

func (c *ChunkManagerImpl) MergeFile(fileID, outputPath string) error {
	c.mu.RLock()
	metadata, ok := c.metadata[fileID]
	if !ok {
		c.mu.RUnlock()
		return ErrFileNotFound
	}
	c.mu.RUnlock()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for i := 0; i < metadata.TotalChunks; i++ {
		chunkData, err := c.ReadChunk(fileID, i)
		if err != nil {
			return err
		}

		if _, err := outputFile.Write(chunkData); err != nil {
			return err
		}
	}

	return nil
}

func (c *ChunkManagerImpl) GetChunkInfo(fileID string, index int) (*ChunkInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chunks, ok := c.chunks[fileID]
	if !ok {
		return nil, ErrFileNotFound
	}

	chunkInfo, ok := chunks[index]
	if !ok {
		return nil, ErrInvalidChunkIndex
	}

	return chunkInfo, nil
}

func (c *ChunkManagerImpl) GetAllChunks(fileID string) ([]*ChunkInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chunks, ok := c.chunks[fileID]
	if !ok {
		return nil, ErrFileNotFound
	}

	result := make([]*ChunkInfo, 0, len(chunks))
	for _, chunkInfo := range chunks {
		result = append(result, chunkInfo)
	}

	return result, nil
}

func (c *ChunkManagerImpl) VerifyChunk(fileID string, index int, data []byte) (bool, error) {
	c.mu.RLock()
	chunkInfo, ok := c.chunks[fileID][index]
	if !ok {
		c.mu.RUnlock()
		return false, ErrInvalidChunkIndex
	}
	expectedHash := chunkInfo.Hash
	c.mu.RUnlock()

	actualHash := sha256.Sum256(data)

	for i := range expectedHash {
		if expectedHash[i] != actualHash[i] {
			return false, nil
		}
	}

	return true, nil
}

func (c *ChunkManagerImpl) VerifyFile(fileID string) (bool, error) {
	c.mu.RLock()
	metadata, ok := c.metadata[fileID]
	if !ok {
		c.mu.RUnlock()
		return false, ErrFileNotFound
	}
	_, ok = c.merkleTrees[fileID]
	c.mu.RUnlock()

	if !ok {
		return false, ErrMerkleTreeBuild
	}

	chunkHashes := make([][]byte, 0, metadata.TotalChunks)
	for i := 0; i < metadata.TotalChunks; i++ {
		chunkData, err := c.ReadChunk(fileID, i)
		if err != nil {
			return false, err
		}

		hash := sha256.Sum256(chunkData)
		chunkHashes = append(chunkHashes, hash[:])
	}

	newMerkle := NewMerkleTree()
	if _, err := newMerkle.BuildTree(chunkHashes); err != nil {
		return false, err
	}

	for i := range newMerkle.GetRootHash() {
		if newMerkle.GetRootHash()[i] != metadata.MerkleRoot[i] {
			return false, ErrFileVerification
		}
	}

	return true, nil
}

func (c *ChunkManagerImpl) SaveChunk(fileID string, index int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	chunkDir := filepath.Join(c.storageDir, fileID)
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return err
	}

	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", index))
	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		return err
	}

	if c.chunks[fileID] == nil {
		c.chunks[fileID] = make(map[int]*ChunkInfo)
	}

	hash := sha256.Sum256(data)
	c.chunks[fileID][index] = &ChunkInfo{
		FileID:    fileID,
		Index:     index,
		Offset:    int64(index * c.chunkSize),
		Size:      len(data),
		Hash:      hash[:],
		Completed: true,
		Verified:  false,
	}

	return nil
}

func (c *ChunkManagerImpl) ReadChunk(fileID string, index int) ([]byte, error) {
	chunkPath := filepath.Join(c.storageDir, fileID, fmt.Sprintf("chunk_%d", index))

	data, err := os.ReadFile(chunkPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	return data, nil
}
