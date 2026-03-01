// Package filetransfer provides file transfer capabilities for P2P communications
package filetransfer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestTransferStateString 测试传输状态字符串表示
func TestTransferStateString(t *testing.T) {
	tests := []struct {
		state    TransferState
		expected string
	}{
		{StatePending, "pending"},
		{StateInProgress, "in_progress"},
		{StatePaused, "paused"},
		{StateCompleted, "completed"},
		{StateFailed, "failed"},
		{StateCancelled, "cancelled"},
		{TransferState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("TransferState.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestDefaultTransferConfig 测试默认配置
func TestDefaultTransferConfig(t *testing.T) {
	config := DefaultTransferConfig()

	if config.ChunkSize != 256*1024 {
		t.Errorf("DefaultTransferConfig.ChunkSize = %d, want %d", config.ChunkSize, 256*1024)
	}

	if config.ParallelChunks != 5 {
		t.Errorf("DefaultTransferConfig.ParallelChunks = %d, want 5", config.ParallelChunks)
	}

	if config.DownloadPath != "./downloads" {
		t.Errorf("DefaultTransferConfig.DownloadPath = %s, want ./downloads", config.DownloadPath)
	}

	if config.MaxRetries != 3 {
		t.Errorf("DefaultTransferConfig.MaxRetries = %d, want 3", config.MaxRetries)
	}

	if config.RetryInterval != 5*time.Second {
		t.Errorf("DefaultTransferConfig.RetryInterval = %v, want 5s", config.RetryInterval)
	}

	if !config.ResumeEnabled {
		t.Error("DefaultTransferConfig.ResumeEnabled should be true")
	}
}

// TestNewManager 测试创建管理器
func TestNewManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filetransfer_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		ChunkSize:      1024,
		ParallelChunks: 2,
		DownloadPath:   tempDir,
		MaxRetries:     1,
		RetryInterval:  1 * time.Second,
		ResumeEnabled:  true,
	}

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	defer func() {
		if err := mgr.Close(); err != nil {
			t.Errorf("Close() error: %v", err)
		}
	}()
}

// TestBuildMerkleTree 测试构建Merkle树
func TestBuildMerkleTree(t *testing.T) {
	tests := []struct {
		name   string
		chunks [][]byte
	}{
		{
			name:   "Empty chunks",
			chunks: [][]byte{},
		},
		{
			name:   "Single chunk",
			chunks: [][]byte{[]byte("chunk1")},
		},
		{
			name:   "Two chunks",
			chunks: [][]byte{[]byte("chunk1"), []byte("chunk2")},
		},
		{
			name: "Multiple chunks",
			chunks: [][]byte{
				[]byte("chunk1"),
				[]byte("chunk2"),
				[]byte("chunk3"),
				[]byte("chunk4"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := BuildMerkleTree(tt.chunks)

			if len(tt.chunks) == 0 {
				if tree != nil {
					t.Error("BuildMerkleTree() expected nil for empty chunks")
				}
				return
			}

			if tree == nil {
				t.Fatal("BuildMerkleTree() returned nil")
			}

			if tree.Root == "" {
				t.Error("MerkleTree.Root is empty")
			}

			if len(tree.Leaves) != len(tt.chunks) {
				t.Errorf("MerkleTree has %d leaves, want %d", len(tree.Leaves), len(tt.chunks))
			}
		})
	}
}

// BenchmarkBuildMerkleTree 基准测试构建Merkle树
func BenchmarkBuildMerkleTree(b *testing.B) {
	chunks := make([][]byte, 100)
	for i := range chunks {
		chunks[i] = make([]byte, 1024)
		for j := range chunks[i] {
			chunks[i][j] = byte(i + j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := BuildMerkleTree(chunks)
		if tree == nil {
			b.Fatal("BuildMerkleTree returned nil")
		}
	}
}
