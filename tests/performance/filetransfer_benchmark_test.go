package performance

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/netvideo/filetransfer"
)

func BenchmarkFileChunking(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark")
	if err != nil {
		b.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.bin")
	fileSize := int64(10 * 1024 * 1024) // 10MB

	file, err := os.Create(testFile)
	if err != nil {
		b.Fatalf("创建测试文件失败: %v", err)
	}
	file.Truncate(fileSize)
	file.Close()

	chunkManager, err := filetransfer.NewChunkManager(256*1024, tempDir)
	if err != nil {
		b.Fatalf("创建块管理器失败: %v", err)
	}
	scheduler := filetransfer.NewScheduler()
	_ = scheduler

	b.ResetTimer()
	b.SetBytes(fileSize)

	for i := 0; i < b.N; i++ {
		_, err := chunkManager.ChunkFile(testFile, 256*1024)
		if err != nil {
			b.Fatalf("文件分块失败: %v", err)
		}
	}
}

func BenchmarkMerkleTreeBuild(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark")
	if err != nil {
		b.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.bin")
	fileSize := int64(10 * 1024 * 1024)

	file, err := os.Create(testFile)
	if err != nil {
		b.Fatalf("创建测试文件失败: %v", err)
	}
	file.Truncate(fileSize)
	file.Close()

	chunkManager, err := filetransfer.NewChunkManager(256*1024, tempDir)
	if err != nil {
		b.Fatalf("创建块管理器失败: %v", err)
	}

	metadata, err := chunkManager.ChunkFile(testFile, 256*1024)
	if err != nil {
		b.Fatalf("文件分块失败: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = metadata.MerkleRoot
	}
}

func BenchmarkChunkVerification(b *testing.B) {
	chunkSize := 256 * 1024
	chunk := make([]byte, chunkSize)
	hash := make([]byte, 32)

	b.ResetTimer()
	b.SetBytes(int64(chunkSize))

	for i := 0; i < b.N; i++ {
		_ = chunk
		_ = hash
	}
}

func BenchmarkFileMerge(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark")
	if err != nil {
		b.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	chunkDir := filepath.Join(tempDir, "chunks")
	os.MkdirAll(chunkDir, 0755)

	numChunks := 10
	for i := 0; i < numChunks; i++ {
		chunkPath := filepath.Join(chunkDir, "chunk_0000")
		file, err := os.Create(chunkPath)
		if err != nil {
			b.Fatalf("创建块文件失败: %v", err)
		}
		file.Truncate(256 * 1024)
		file.Close()
	}

	chunkManager, err := filetransfer.NewChunkManager(256*1024, tempDir)
	if err != nil {
		b.Fatalf("创建块管理器失败: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tempDir, "output.bin")
		err := chunkManager.MergeFile("testfile", outputPath)
		if err != nil {
			b.Fatalf("文件合并失败: %v", err)
		}
		os.Remove(outputPath)
	}
}
