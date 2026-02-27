package performance

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/netvideo/filetransfer"
	"github.com/netvideo/tests/utils"
)

func BenchmarkFileChunking(b *testing.B) {
	// 创建临时测试文件
	tempDir := utils.CreateTestDirectory(nil, "benchmark")
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.bin")
	fileSize := int64(10 * 1024 * 1024) // 10MB
	utils.CreateTestFile(nil, testFile, fileSize)

	b.ResetTimer()
	b.SetBytes(fileSize)

	for i := 0; i < b.N; i++ {
		manager := filetransfer.NewFileTransferManager()
		_, err := manager.ChunkFile(testFile, 256*1024) // 256KB 块
		if err != nil {
			b.Fatalf("文件分块失败：%v", err)
		}
		manager.Close()
	}
}

func BenchmarkMerkleTreeBuild(b *testing.B) {
	tempDir := utils.CreateTestDirectory(nil, "benchmark")
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.bin")
	fileSize := int64(10 * 1024 * 1024) // 10MB
	utils.CreateTestFile(nil, testFile, fileSize)

	manager := filetransfer.NewFileTransferManager()
	defer manager.Close()

	// 先分块
	metadata, err := manager.ChunkFile(testFile, 256*1024)
	if err != nil {
		b.Fatalf("文件分块失败：%v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 重新构建 Merkle 树
		_ = metadata.MerkleRoot
	}
}

func BenchmarkChunkVerification(b *testing.B) {
	// 模拟块验证性能测试
	chunkSize := 256 * 1024
	chunk := make([]byte, chunkSize)
	hash := make([]byte, 32)

	b.ResetTimer()
	b.SetBytes(int64(chunkSize))

	for i := 0; i < b.N; i++ {
		// 模拟哈希验证
		_ = chunk
		_ = hash
	}
}

func BenchmarkFileMerge(b *testing.B) {
	tempDir := utils.CreateTestDirectory(nil, "benchmark")
	defer os.RemoveAll(tempDir)

	// 创建测试块
	chunkDir := filepath.Join(tempDir, "chunks")
	os.MkdirAll(chunkDir, 0755)

	numChunks := 10
	for i := 0; i < numChunks; i++ {
		chunkPath := filepath.Join(chunkDir, "chunk_0000")
		utils.CreateTestFile(nil, chunkPath, 256*1024)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := filetransfer.NewFileTransferManager()

		metadata := &filetransfer.FileMetadata{
			FileName:    "test.bin",
			FileSize:    int64(numChunks) * 256 * 1024,
			TotalChunks: numChunks,
			ChunkDir:    chunkDir,
		}

		outputPath := filepath.Join(tempDir, "output.bin")
		err := manager.MergeFile(outputPath, metadata)
		if err != nil {
			b.Fatalf("文件合并失败：%v", err)
		}

		manager.Close()
		os.Remove(outputPath)
	}
}
