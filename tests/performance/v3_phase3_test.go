package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/netvideo/webrtc"
	"github.com/netvideo/group"
	"github.com/netvideo/filetransfer"
)

// TestWebRTCCallPerformance tests WebRTC call performance
func TestWebRTCCallPerformance(t *testing.T) {
	config := webrtc.DefaultCallConfig()
	manager, err := webrtc.NewCallManager(config)
	if err != nil {
		t.Fatalf("Failed to create call manager: %v", err)
	}

	iterations := 100
	startTime := time.Now()

	for i := 0; i < iterations; i++ {
		peerID := fmt.Sprintf("peer-%d", i)
		call, err := manager.StartCall(context.Background(), peerID, webrtc.CallTypeAudio)
		if err != nil {
			t.Errorf("Failed to start call %d: %v", i, err)
			continue
		}

		// End call
		err = manager.EndCall(call.ID)
		if err != nil {
			t.Errorf("Failed to end call %d: %v", i, err)
		}
	}

	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)

	t.Logf("WebRTC Call Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average time: %v", avgTime)

	if avgTime > 2*time.Second {
		t.Errorf("Call setup too slow: %v (target: <2s)", avgTime)
	}
}

// TestGroupChatPerformance tests group chat performance
func TestGroupChatPerformance(t *testing.T) {
	groupManager := group.NewGroupManager()
	messageManager := group.NewGroupMessageManager()

	// Create group
	groupObj, err := groupManager.CreateGroup("owner", "Test Group", "Test Description")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Add members
	memberCount := 100
	for i := 0; i < memberCount; i++ {
		userID := fmt.Sprintf("user-%d", i)
		err := groupManager.AddMember(groupObj.ID, userID, fmt.Sprintf("User %d", i))
		if err != nil {
			t.Errorf("Failed to add member %d: %v", i, err)
		}
	}

	// Send messages
	messageCount := 1000
	startTime := time.Now()

	for i := 0; i < messageCount; i++ {
		_, err := messageManager.SendMessage(groupObj.ID, fmt.Sprintf("user-%d", i%memberCount), fmt.Sprintf("Message %d", i), group.MessageTypeText)
		if err != nil {
			t.Errorf("Failed to send message %d: %v", i, err)
		}
	}

	duration := time.Since(startTime)
	avgTime := duration / time.Duration(messageCount)

	t.Logf("Group Chat Performance:")
	t.Logf("  Members: %d", memberCount)
	t.Logf("  Messages: %d", messageCount)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average time: %v", avgTime)

	if avgTime > 500*time.Millisecond {
		t.Errorf("Message sync too slow: %v (target: <500ms)", avgTime)
	}
}

// TestFileTransferPerformance tests file transfer performance
func TestFileTransferPerformance(t *testing.T) {
	ftm := filetransfer.NewFileTransferManager()
	ptm := filetransfer.NewParallelTransferManager(4)

	// Create test file
	testSize := int64(10 * 1024 * 1024) // 10MB
	testFile := "/tmp/test_transfer.bin"

	// Generate test data
	data := make([]byte, testSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Write test file
	err := writeFile(testFile, data)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer removeFile(testFile)

	// Split file
	startTime := time.Now()
	fileInfo, err := ftm.SplitFile(testFile, filetransfer.DefaultChunkSize)
	if err != nil {
		t.Fatalf("Failed to split file: %v", err)
	}

	splitTime := time.Since(startTime)

	// Start transfer
	transfer, err := ftm.StartTransfer("test-file", testFile, "/remote/test.bin")
	if err != nil {
		t.Fatalf("Failed to start transfer: %v", err)
	}

	// Simulate parallel transfer
	parallelTransfer, err := ptm.StartParallelTransfer("test-file", testFile, "/remote/test.bin", fileInfo.ChunkCount)
	if err != nil {
		t.Fatalf("Failed to start parallel transfer: %v", err)
	}

	// Update progress
	err = ftm.UpdateTransferProgress("test-file", testSize)
	if err != nil {
		t.Errorf("Failed to update progress: %v", err)
	}

	// Get speed
	speed, err := ftm.GetTransferSpeed("test-file")
	if err != nil {
		t.Errorf("Failed to get speed: %v", err)
	}

	transferTime := time.Since(startTime)

	t.Logf("File Transfer Performance:")
	t.Logf("  File size: %d MB", testSize/1024/1024)
	t.Logf("  Chunks: %d", fileInfo.ChunkCount)
	t.Logf("  Split time: %v", splitTime)
	t.Logf("  Transfer time: %v", transferTime)
	t.Logf("  Speed: %.2f MB/s", speed/1024/1024)

	if speed < 50*1024*1024 {
		t.Errorf("Transfer speed too slow: %.2f MB/s (target: >50 MB/s)", speed/1024/1024)
	}

	_ = parallelTransfer
}

// TestConcurrentCalls tests concurrent WebRTC calls
func TestConcurrentCalls(t *testing.T) {
	config := webrtc.DefaultCallConfig()
	manager, err := webrtc.NewCallManager(config)
	if err != nil {
		t.Fatalf("Failed to create call manager: %v", err)
	}

	concurrent := 10
	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			peerID := fmt.Sprintf("peer-%d", id)
			call, err := manager.StartCall(context.Background(), peerID, webrtc.CallTypeAudio)
			if err != nil {
				t.Errorf("Failed to start call %d: %v", id, err)
				return
			}

			time.Sleep(100 * time.Millisecond)

			err = manager.EndCall(call.ID)
			if err != nil {
				t.Errorf("Failed to end call %d: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Concurrent Calls:")
	t.Logf("  Concurrent: %d", concurrent)
	t.Logf("  Total time: %v", duration)

	if duration > 5*time.Second {
		t.Errorf("Concurrent calls too slow: %v", duration)
	}
}

// TestLargeGroupChat tests large group chat
func TestLargeGroupChat(t *testing.T) {
	groupManager := group.NewGroupManager()

	// Create large group
	groupObj, err := groupManager.CreateGroup("owner", "Large Group", "Test")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Add 100 members
	for i := 0; i < 100; i++ {
		userID := fmt.Sprintf("user-%d", i)
		err := groupManager.AddMember(groupObj.ID, userID, fmt.Sprintf("User %d", i))
		if err != nil {
			t.Fatalf("Failed to add member %d: %v", i, err)
		}
	}

	// Get members
	members, err := groupManager.GetMembers(groupObj.ID)
	if err != nil {
		t.Fatalf("Failed to get members: %v", err)
	}

	if len(members) != 101 { // 100 members + 1 owner
		t.Errorf("Expected 101 members, got %d", len(members))
	}

	t.Logf("Large Group Chat:")
	t.Logf("  Total members: %d", len(members))
	t.Logf("  Group supports 100+ users: ✓")
}

// Helper functions
func writeFile(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func removeFile(path string) {
	os.Remove(path)
}
