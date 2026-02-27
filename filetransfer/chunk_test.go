package filetransfer

import (
	"os"
	"testing"
)

func TestChunkManager_New(t *testing.T) {
	chunkMgr, err := NewChunkManager(1024, "./test_storage")
	if err != nil {
		t.Fatalf("NewChunkManager failed: %v", err)
	}

	if chunkMgr == nil {
		t.Fatal("NewChunkManager returned nil")
	}

	os.RemoveAll("./test_storage")
}

func TestChunkManager_New_DefaultChunkSize(t *testing.T) {
	chunkMgr, err := NewChunkManager(0, "./test_storage")
	if err != nil {
		t.Fatalf("NewChunkManager failed: %v", err)
	}

	if chunkMgr == nil {
		t.Fatal("NewChunkManager returned nil")
	}

	os.RemoveAll("./test_storage")
}

func TestChunkManager_GetChunkInfo_NotFound(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	defer os.RemoveAll("./test_storage")

	_, err := chunkMgr.GetChunkInfo("nonexistent", 0)
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestChunkManager_GetAllChunks_NotFound(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	defer os.RemoveAll("./test_storage")

	_, err := chunkMgr.GetAllChunks("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestChunkManager_VerifyChunk_InvalidIndex(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	defer os.RemoveAll("./test_storage")

	_, err := chunkMgr.VerifyChunk("nonexistent", 0, []byte("data"))
	if err == nil {
		t.Fatal("Expected error for invalid index")
	}
}

func TestChunkManager_ReadChunk_NotFound(t *testing.T) {
	chunkMgr, _ := NewChunkManager(1024, "./test_storage")
	defer os.RemoveAll("./test_storage")

	_, err := chunkMgr.ReadChunk("nonexistent", 0)
	if err == nil {
		t.Fatal("Expected error for nonexistent chunk")
	}
}

func TestChunkManager_SaveAndRead(t *testing.T) {
	chunkMgr, err := NewChunkManager(1024, "./test_storage")
	if err != nil {
		t.Fatalf("NewChunkManager failed: %v", err)
	}
	defer os.RemoveAll("./test_storage")

	fileID := "testfile123"
	chunkData := []byte("test chunk data")

	err = chunkMgr.SaveChunk(fileID, 0, chunkData)
	if err != nil {
		t.Fatalf("SaveChunk failed: %v", err)
	}

	readData, err := chunkMgr.ReadChunk(fileID, 0)
	if err != nil {
		t.Fatalf("ReadChunk failed: %v", err)
	}

	if string(readData) != string(chunkData) {
		t.Errorf("Read data mismatch: got %s, want %s", readData, chunkData)
	}
}

func TestChunkManager_VerifyChunk_Valid(t *testing.T) {
	chunkMgr, err := NewChunkManager(1024, "./test_storage")
	if err != nil {
		t.Fatalf("NewChunkManager failed: %v", err)
	}
	defer os.RemoveAll("./test_storage")

	fileID := "testfile123"
	chunkData := []byte("test chunk data")

	err = chunkMgr.SaveChunk(fileID, 0, chunkData)
	if err != nil {
		t.Fatalf("SaveChunk failed: %v", err)
	}

	valid, err := chunkMgr.VerifyChunk(fileID, 0, chunkData)
	if err != nil {
		t.Fatalf("VerifyChunk failed: %v", err)
	}

	if !valid {
		t.Error("Expected chunk to be valid")
	}
}

func TestChunkManager_VerifyChunk_Invalid(t *testing.T) {
	chunkMgr, err := NewChunkManager(1024, "./test_storage")
	if err != nil {
		t.Fatalf("NewChunkManager failed: %v", err)
	}
	defer os.RemoveAll("./test_storage")

	fileID := "testfile123"
	chunkData := []byte("test chunk data")

	err = chunkMgr.SaveChunk(fileID, 0, chunkData)
	if err != nil {
		t.Fatalf("SaveChunk failed: %v", err)
	}

	valid, err := chunkMgr.VerifyChunk(fileID, 0, []byte("wrong data"))
	if err != nil {
		t.Fatalf("VerifyChunk failed: %v", err)
	}

	if valid {
		t.Error("Expected chunk to be invalid")
	}
}
