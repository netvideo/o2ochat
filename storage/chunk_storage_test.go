package storage

import (
	"testing"
)

func TestChunkStorage_BasicOperations(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	chunkStorage := manager.GetChunkStorage()

	fileID := "file-001"
	chunkData := []byte("This is chunk data for testing purposes. It should be stored and retrieved correctly.")

	t.Run("StoreChunk", func(t *testing.T) {
		err := chunkStorage.StoreChunk(fileID, 0, chunkData)
		if err != nil {
			t.Fatalf("Failed to store chunk: %v", err)
		}
	})

	t.Run("GetChunk", func(t *testing.T) {
		retrieved, err := chunkStorage.GetChunk(fileID, 0)
		if err != nil {
			t.Fatalf("Failed to get chunk: %v", err)
		}

		if string(retrieved) != string(chunkData) {
			t.Errorf("Chunk data mismatch")
		}
	})

	t.Run("ChunkExists", func(t *testing.T) {
		exists, err := chunkStorage.ChunkExists(fileID, 0)
		if err != nil {
			t.Fatalf("Failed to check chunk existence: %v", err)
		}

		if !exists {
			t.Error("Chunk should exist")
		}

		exists, err = chunkStorage.ChunkExists(fileID, 999)
		if err != nil {
			t.Fatalf("Failed to check chunk existence: %v", err)
		}

		if exists {
			t.Error("Non-existent chunk should not exist")
		}
	})
}

func TestChunkStorage_MultipleChunks(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	chunkStorage := manager.GetChunkStorage()

	fileID := "file-multi"
	chunks := [][]byte{
		[]byte("Chunk 0"),
		[]byte("Chunk 1"),
		[]byte("Chunk 2"),
		[]byte("Chunk 3"),
		[]byte("Chunk 4"),
	}

	for i, data := range chunks {
		err := chunkStorage.StoreChunk(fileID, i, data)
		if err != nil {
			t.Fatalf("Failed to store chunk %d: %v", i, err)
		}
	}

	t.Run("GetChunkIndices", func(t *testing.T) {
		indices, err := chunkStorage.GetChunkIndices(fileID)
		if err != nil {
			t.Fatalf("Failed to get chunk indices: %v", err)
		}

		if len(indices) != 5 {
			t.Errorf("Expected 5 indices, got %d", len(indices))
		}

		for i, idx := range indices {
			if idx != i {
				t.Errorf("Expected index %d, got %d", i, idx)
			}
		}
	})

	t.Run("GetChunkStats", func(t *testing.T) {
		stats, err := chunkStorage.GetChunkStats(fileID)
		if err != nil {
			t.Fatalf("Failed to get chunk stats: %v", err)
		}

		if stats.ChunkCount != 5 {
			t.Errorf("Expected 5 chunks, got %d", stats.ChunkCount)
		}

		if stats.TotalCount != 5 {
			t.Errorf("Expected total count 5, got %d", stats.TotalCount)
		}
	})
}

func TestChunkStorage_DeleteChunk(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	chunkStorage := manager.GetChunkStorage()

	fileID := "file-delete"
	chunkData := []byte("Chunk to delete")

	chunkStorage.StoreChunk(fileID, 0, chunkData)

	t.Run("DeleteChunk", func(t *testing.T) {
		err := chunkStorage.DeleteChunk(fileID, 0)
		if err != nil {
			t.Fatalf("Failed to delete chunk: %v", err)
		}

		exists, _ := chunkStorage.ChunkExists(fileID, 0)
		if exists {
			t.Error("Chunk should not exist after deletion")
		}
	})

	t.Run("DeleteNonExistent", func(t *testing.T) {
		err := chunkStorage.DeleteChunk(fileID, 999)
		if err != ErrChunkNotFound {
			t.Errorf("Expected ErrChunkNotFound, got %v", err)
		}
	})
}

func TestChunkStorage_DeleteAllChunks(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	chunkStorage := manager.GetChunkStorage()

	fileID := "file-delete-all"
	for i := 0; i < 5; i++ {
		chunkStorage.StoreChunk(fileID, i, []byte("Chunk "+string(rune('0'+i))))
	}

	t.Run("DeleteAllChunks", func(t *testing.T) {
		err := chunkStorage.DeleteAllChunks(fileID)
		if err != nil {
			t.Fatalf("Failed to delete all chunks: %v", err)
		}

		indices, _ := chunkStorage.GetChunkIndices(fileID)
		if len(indices) != 0 {
			t.Errorf("Expected 0 indices after delete all, got %d", len(indices))
		}
	})
}

func TestChunkStorage_GetChunkNotFound(t *testing.T) {
	manager, cleanup := setupTestStorage(t)
	defer cleanup()

	chunkStorage := manager.GetChunkStorage()

	t.Run("GetNonExistentChunk", func(t *testing.T) {
		_, err := chunkStorage.GetChunk("nonexistent", 0)
		if err != ErrChunkNotFound {
			t.Errorf("Expected ErrChunkNotFound, got %v", err)
		}
	})
}
