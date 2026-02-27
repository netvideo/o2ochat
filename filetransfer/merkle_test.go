package filetransfer

import (
	"reflect"
	"testing"
)

func TestMerkleTree_BuildTree(t *testing.T) {
	merkle := NewMerkleTree()

	chunks := [][]byte{
		[]byte("chunk1"),
		[]byte("chunk2"),
		[]byte("chunk3"),
		[]byte("chunk4"),
	}

	tree, err := merkle.BuildTree(chunks)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}

	if tree == nil {
		t.Fatal("BuildTree returned nil tree")
	}

	rootHash := merkle.GetRootHash()
	if rootHash == nil {
		t.Fatal("GetRootHash returned nil")
	}

	if len(rootHash) != 32 {
		t.Errorf("Expected root hash length 32, got %d", len(rootHash))
	}
}

func TestMerkleTree_BuildTree_Empty(t *testing.T) {
	merkle := NewMerkleTree()

	_, err := merkle.BuildTree([][]byte{})
	if err == nil {
		t.Fatal("Expected error for empty chunks")
	}
}

func TestMerkleTree_GetRootHash_Empty(t *testing.T) {
	merkle := NewMerkleTree()

	rootHash := merkle.GetRootHash()
	if rootHash != nil {
		t.Errorf("Expected nil root hash for empty tree, got %v", rootHash)
	}
}

func TestMerkleTree_VerifyChunk(t *testing.T) {
	merkle := NewMerkleTree()

	chunks := [][]byte{
		[]byte("chunk1"),
		[]byte("chunk2"),
		[]byte("chunk3"),
	}

	_, err := merkle.BuildTree(chunks)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}

	proof, err := merkle.GenerateProof(0)
	if err != nil {
		t.Fatalf("GenerateProof failed: %v", err)
	}

	valid, err := merkle.VerifyChunk(0, chunks[0], proof)
	if err != nil {
		t.Fatalf("VerifyChunk failed: %v", err)
	}

	if !valid {
		t.Error("Chunk verification failed for valid chunk")
	}
}

func TestMerkleTree_VerifyChunk_InvalidIndex(t *testing.T) {
	merkle := NewMerkleTree()

	chunks := [][]byte{
		[]byte("chunk1"),
		[]byte("chunk2"),
	}

	merkle.BuildTree(chunks)

	_, err := merkle.VerifyChunk(5, chunks[0], [][]byte{})
	if err == nil {
		t.Fatal("Expected error for invalid index")
	}
}

func TestMerkleTree_VerifyChunk_WrongData(t *testing.T) {
	merkle := NewMerkleTree()

	chunks := [][]byte{
		[]byte("chunk1"),
		[]byte("chunk2"),
	}

	merkle.BuildTree(chunks)

	proof, _ := merkle.GenerateProof(0)

	valid, err := merkle.VerifyChunk(0, []byte("wrongdata"), proof)
	if err != nil {
		t.Fatalf("VerifyChunk failed: %v", err)
	}

	if valid {
		t.Error("Expected verification to fail for wrong data")
	}
}

func TestMerkleTree_GenerateProof(t *testing.T) {
	merkle := NewMerkleTree()

	chunks := [][]byte{
		[]byte("chunk1"),
		[]byte("chunk2"),
		[]byte("chunk3"),
		[]byte("chunk4"),
	}

	merkle.BuildTree(chunks)

	proof, err := merkle.GenerateProof(0)
	if err != nil {
		t.Fatalf("GenerateProof failed: %v", err)
	}

	if proof == nil {
		t.Fatal("GenerateProof returned nil")
	}

	expectedProofLen := 2
	if len(proof) != expectedProofLen {
		t.Errorf("Expected proof length %d, got %d", expectedProofLen, len(proof))
	}
}

func TestMerkleTree_GenerateProof_InvalidIndex(t *testing.T) {
	merkle := NewMerkleTree()

	chunks := [][]byte{
		[]byte("chunk1"),
	}

	merkle.BuildTree(chunks)

	_, err := merkle.GenerateProof(5)
	if err == nil {
		t.Fatal("Expected error for invalid index")
	}
}

func TestMerkleTree_Serialize(t *testing.T) {
	merkle := NewMerkleTree()

	chunks := [][]byte{
		[]byte("chunk1"),
		[]byte("chunk2"),
	}

	merkle.BuildTree(chunks)

	data, err := merkle.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	if data == nil {
		t.Fatal("Serialize returned nil")
	}

	if len(data) == 0 {
		t.Error("Serialize returned empty data")
	}
}

func TestMerkleTree_Deserialize(t *testing.T) {
	merkle1 := NewMerkleTree()
	chunks := [][]byte{
		[]byte("chunk1"),
		[]byte("chunk2"),
	}
	merkle1.BuildTree(chunks)

	data, err := merkle1.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	merkle2 := NewMerkleTree()
	err = merkle2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if !reflect.DeepEqual(merkle1.GetRootHash(), merkle2.GetRootHash()) {
		t.Error("Root hashes don't match after deserialize")
	}
}

func TestMerkleTree_SingleChunk(t *testing.T) {
	merkle := NewMerkleTree()

	chunks := [][]byte{
		[]byte("single chunk"),
	}

	_, err := merkle.BuildTree(chunks)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}

	rootHash := merkle.GetRootHash()
	if rootHash == nil || len(rootHash) != 32 {
		t.Errorf("Invalid root hash for single chunk")
	}
}

func BenchmarkMerkleTree_BuildTree(b *testing.B) {
	chunks := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		chunks[i] = make([]byte, 1024)
		for j := 0; j < 1024; j++ {
			chunks[i][byte(j)] = byte(i + j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merkle := NewMerkleTree()
		_, _ = merkle.BuildTree(chunks)
	}
}

func BenchmarkMerkleTree_GenerateProof(b *testing.B) {
	chunks := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		chunks[i] = make([]byte, 1024)
		for j := 0; j < 1024; j++ {
			chunks[i][byte(j)] = byte(i + j)
		}
	}

	merkle := NewMerkleTree()
	_, _ = merkle.BuildTree(chunks)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = merkle.GenerateProof(50)
	}
}

func BenchmarkMerkleTree_VerifyChunk(b *testing.B) {
	chunks := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		chunks[i] = make([]byte, 1024)
		for j := 0; j < 1024; j++ {
			chunks[i][byte(j)] = byte(i + j)
		}
	}

	merkle := NewMerkleTree()
	_, _ = merkle.BuildTree(chunks)
	proof, _ := merkle.GenerateProof(50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = merkle.VerifyChunk(50, chunks[50], proof)
	}
}
