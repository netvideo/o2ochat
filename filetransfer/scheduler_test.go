package filetransfer

import (
	"testing"
)

func TestScheduler_SelectNextChunk(t *testing.T) {
	scheduler := NewScheduler()

	availableChunks := map[int][]string{
		0: {"peer1", "peer2"},
		1: {"peer1"},
		2: {"peer1", "peer2", "peer3"},
	}

	chunkIdx, peers, err := scheduler.SelectNextChunk("file1", availableChunks)
	if err != nil {
		t.Fatalf("SelectNextChunk failed: %v", err)
	}

	if chunkIdx < 0 || chunkIdx > 2 {
		t.Errorf("Invalid chunk index: %d", chunkIdx)
	}

	if len(peers) == 0 {
		t.Error("No peers returned")
	}
}

func TestScheduler_SelectNextChunk_Empty(t *testing.T) {
	scheduler := NewScheduler()

	_, _, err := scheduler.SelectNextChunk("file1", map[int][]string{})
	if err == nil {
		t.Fatal("Expected error for empty chunks")
	}
}

func TestScheduler_SelectNextChunk_AllCompleted(t *testing.T) {
	scheduler := NewScheduler()

	availableChunks := map[int][]string{
		0: {"peer1"},
		1: {"peer2"},
	}

	scheduler.SelectNextChunk("file1", availableChunks)
	scheduler.UpdateChunkStatus("file1", 0, true, "peer1")
	scheduler.UpdateChunkStatus("file1", 1, true, "peer2")

	chunkIdx, peers, err := scheduler.SelectNextChunk("file1", availableChunks)
	if err != nil {
		t.Fatalf("SelectNextChunk failed: %v", err)
	}

	if chunkIdx != -1 {
		t.Errorf("Expected -1 for all completed chunks, got %d", chunkIdx)
	}

	if peers != nil {
		t.Error("Expected nil peers for all completed")
	}
}

func TestScheduler_AssignDownloadTasks(t *testing.T) {
	scheduler := NewScheduler()

	availableChunks := map[int][]string{
		0: {"peer1", "peer2"},
		1: {"peer1"},
		2: {"peer2", "peer3"},
		3: {"peer1", "peer3"},
	}

	scheduler.SelectNextChunk("file1", availableChunks)

	assignments, err := scheduler.AssignDownloadTasks("file1", 2)
	if err != nil {
		t.Fatalf("AssignDownloadTasks failed: %v", err)
	}

	if len(assignments) == 0 {
		t.Error("No assignments returned")
	}
}

func TestScheduler_UpdateChunkStatus_Success(t *testing.T) {
	scheduler := NewScheduler()

	availableChunks := map[int][]string{
		0: {"peer1", "peer2"},
	}

	scheduler.SelectNextChunk("file1", availableChunks)

	err := scheduler.UpdateChunkStatus("file1", 0, true, "peer1")
	if err != nil {
		t.Fatalf("UpdateChunkStatus failed: %v", err)
	}

	stats, _ := scheduler.GetSchedulerStats("file1")
	if stats.Successful != 1 {
		t.Errorf("Expected 1 successful, got %d", stats.Successful)
	}
}

func TestScheduler_UpdateChunkStatus_Failure(t *testing.T) {
	scheduler := NewScheduler()

	availableChunks := map[int][]string{
		0: {"peer1", "peer2"},
	}

	scheduler.SelectNextChunk("file1", availableChunks)

	err := scheduler.UpdateChunkStatus("file1", 0, false, "peer1")
	if err != nil {
		t.Fatalf("UpdateChunkStatus failed: %v", err)
	}

	stats, _ := scheduler.GetSchedulerStats("file1")
	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", stats.Failed)
	}
}

func TestScheduler_GetSchedulerStats(t *testing.T) {
	scheduler := NewScheduler()

	availableChunks := map[int][]string{
		0: {"peer1"},
		1: {"peer2"},
	}

	scheduler.SelectNextChunk("file1", availableChunks)
	scheduler.UpdateChunkStatus("file1", 0, true, "peer1")

	stats, err := scheduler.GetSchedulerStats("file1")
	if err != nil {
		t.Fatalf("GetSchedulerStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("GetSchedulerStats returned nil")
	}

	if stats.TotalScheduled == 0 {
		t.Error("Expected non-zero total scheduled")
	}
}

func TestScheduler_GetSchedulerStats_Empty(t *testing.T) {
	scheduler := NewScheduler()

	stats, err := scheduler.GetSchedulerStats("nonexistent")
	if err != nil {
		t.Fatalf("GetSchedulerStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("GetSchedulerStats returned nil for nonexistent file")
	}
}

func TestScheduler_RetryOnFailure(t *testing.T) {
	scheduler := NewScheduler()

	availableChunks := map[int][]string{
		0: {"peer1", "peer2"},
	}

	scheduler.SelectNextChunk("file1", availableChunks)
	scheduler.UpdateChunkStatus("file1", 0, false, "peer1")

	chunkIdx, _, err := scheduler.SelectNextChunk("file1", availableChunks)
	if err != nil {
		t.Fatalf("SelectNextChunk failed: %v", err)
	}

	if chunkIdx != 0 {
		t.Errorf("Expected chunk 0 to be retried, got %d", chunkIdx)
	}
}
