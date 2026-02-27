package filetransfer

import (
	"os"
	"testing"
)

func TestIntegration_ChunkFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpFile, err := os.CreateTemp("", "test_chunk_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	content := make([]byte, 5*1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}

	if err := os.WriteFile(tmpFile.Name(), content, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	chunkMgr, err := NewChunkManager(1024*1024, "./test_integration_storage")
	if err != nil {
		t.Fatalf("Failed to create chunk manager: %v", err)
	}
	defer os.RemoveAll("./test_integration_storage")

	metadata, err := chunkMgr.ChunkFile(tmpFile.Name(), 1024*1024)
	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}

	if metadata.TotalChunks != 5 {
		t.Errorf("Expected 5 chunks, got %d", metadata.TotalChunks)
	}

	if metadata.FileSize != int64(len(content)) {
		t.Errorf("Expected file size %d, got %d", len(content), metadata.FileSize)
	}
}

func TestIntegration_MultipleChunks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpFile, err := os.CreateTemp("", "test_multi_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	content := make([]byte, 10*1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}

	if err := os.WriteFile(tmpFile.Name(), content, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	chunkMgr, err := NewChunkManager(512*1024, "./test_integration_storage2")
	if err != nil {
		t.Fatalf("Failed to create chunk manager: %v", err)
	}
	defer os.RemoveAll("./test_integration_storage2")

	metadata, err := chunkMgr.ChunkFile(tmpFile.Name(), 512*1024)
	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}

	if metadata.TotalChunks != 20 {
		t.Errorf("Expected 20 chunks, got %d", metadata.TotalChunks)
	}

	chunks, err := chunkMgr.GetAllChunks(metadata.FileID)
	if err != nil {
		t.Fatalf("GetAllChunks failed: %v", err)
	}

	if len(chunks) != 20 {
		t.Errorf("Expected 20 chunks info, got %d", len(chunks))
	}
}

func TestIntegration_VerifyFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpFile, err := os.CreateTemp("", "test_verify_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	content := make([]byte, 3*1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}

	if err := os.WriteFile(tmpFile.Name(), content, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	chunkMgr, err := NewChunkManager(1024*1024, "./test_integration_storage3")
	if err != nil {
		t.Fatalf("Failed to create chunk manager: %v", err)
	}
	defer os.RemoveAll("./test_integration_storage3")

	metadata, err := chunkMgr.ChunkFile(tmpFile.Name(), 1024*1024)
	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}

	chunks, err := chunkMgr.GetAllChunks(metadata.FileID)
	if err != nil {
		t.Fatalf("GetAllChunks failed: %v", err)
	}

	for _, chunk := range chunks {
		chunkData := make([]byte, chunk.Size)
		for i := range chunkData {
			chunkData[i] = byte((int(chunk.Offset) + i) % 256)
		}
		err = chunkMgr.SaveChunk(metadata.FileID, chunk.Index, chunkData)
		if err != nil {
			t.Fatalf("SaveChunk failed for chunk %d: %v", chunk.Index, err)
		}
	}

	valid, err := chunkMgr.VerifyFile(metadata.FileID)
	if err != nil {
		t.Logf("VerifyFile returned error (expected for modified data): %v", err)
	}

	if valid {
		t.Error("Expected file verification to fail for modified data")
	}
}

func TestIntegration_Scheduler(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	scheduler := NewScheduler()

	availableChunks := map[int][]string{
		0: {"peer1", "peer2", "peer3"},
		1: {"peer1", "peer2"},
		2: {"peer1"},
		3: {"peer1", "peer2", "peer3", "peer4"},
		4: {"peer2", "peer3"},
	}

	completedCount := 0
	for i := 0; i < 10; i++ {
		chunkIdx, peers, err := scheduler.SelectNextChunk("testfile", availableChunks)
		if err != nil {
			if err == ErrInsufficientPeers {
				break
			}
			t.Fatalf("SelectNextChunk failed: %v", err)
		}

		if chunkIdx < 0 {
			break
		}

		if len(peers) == 0 {
			t.Error("No peers returned")
		}

		scheduler.UpdateChunkStatus("testfile", chunkIdx, true, peers[0])
		completedCount++
	}

	stats, err := scheduler.GetSchedulerStats("testfile")
	if err != nil {
		t.Fatalf("GetSchedulerStats failed: %v", err)
	}

	if stats.Successful == 0 {
		t.Error("Expected at least one successful chunk")
	}

	t.Logf("Completed %d chunks successfully", completedCount)
}
