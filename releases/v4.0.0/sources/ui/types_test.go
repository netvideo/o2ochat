package ui

import (
	"testing"
	"time"
)

func TestUITheme(t *testing.T) {
	tests := []struct {
		name     string
		theme    UITheme
		expected string
	}{
		{"Light theme", ThemeLight, "light"},
		{"Dark theme", ThemeDark, "dark"},
		{"Auto theme", ThemeAuto, "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.theme) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.theme)
			}
		})
	}
}

func TestMessageType(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{"Text message", MessageTypeText, "text"},
		{"Image message", MessageTypeImage, "image"},
		{"File message", MessageTypeFile, "file"},
		{"Voice message", MessageTypeVoice, "voice"},
		{"Video message", MessageTypeVideo, "video"},
		{"System message", MessageTypeSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.msgType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.msgType)
			}
		})
	}
}

func TestMessageStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   MessageStatus
		expected string
	}{
		{"Sending", MessageStatusSending, "sending"},
		{"Sent", MessageStatusSent, "sent"},
		{"Delivered", MessageStatusDelivered, "delivered"},
		{"Read", MessageStatusRead, "read"},
		{"Failed", MessageStatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}

func TestCallAction(t *testing.T) {
	tests := []struct {
		name     string
		action   CallAction
		expected string
	}{
		{"Accept", CallActionAccept, "accept"},
		{"Reject", CallActionReject, "reject"},
		{"Mute", CallActionMute, "mute"},
		{"Unmute", CallActionUnmute, "unmute"},
		{"VideoOn", CallActionVideoOn, "video_on"},
		{"VideoOff", CallActionVideoOff, "video_off"},
		{"ScreenShare", CallActionScreenShare, "screen_share"},
		{"End", CallActionEnd, "end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.action) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.action)
			}
		})
	}
}

func TestUIConfig(t *testing.T) {
	config := &UIConfig{
		Theme:          ThemeDark,
		Language:       "zh-CN",
		FontSize:       14,
		ShowAvatars:    true,
		ShowTimestamps: true,
		NotifySounds:   true,
		NotifyDesktop:  true,
		AutoStart:      false,
		MinimizeToTray: true,
	}

	if config.Theme != ThemeDark {
		t.Errorf("expected ThemeDark, got %v", config.Theme)
	}
	if config.Language != "zh-CN" {
		t.Errorf("expected zh-CN, got %s", config.Language)
	}
	if config.FontSize != 14 {
		t.Errorf("expected 14, got %d", config.FontSize)
	}
}

func TestContactInfo(t *testing.T) {
	now := time.Now()
	contact := &ContactInfo{
		PeerID:      "QmPeer123",
		Name:        "张三",
		Avatar:      "avatar.png",
		LastSeen:    now,
		Online:      true,
		UnreadCount: 5,
		IsFavorite:  true,
		Groups:      []string{"朋友", "同事"},
	}

	if contact.PeerID != "QmPeer123" {
		t.Errorf("expected QmPeer123, got %s", contact.PeerID)
	}
	if contact.Name != "张三" {
		t.Errorf("expected 张三, got %s", contact.Name)
	}
	if !contact.Online {
		t.Error("expected online to be true")
	}
	if contact.UnreadCount != 5 {
		t.Errorf("expected 5, got %d", contact.UnreadCount)
	}
}

func TestMessageItem(t *testing.T) {
	now := time.Now()
	msg := &MessageItem{
		ID:        "msg123",
		From:      "QmPeer456",
		To:        "QmPeer123",
		Content:   "你好",
		Type:      MessageTypeText,
		Timestamp: now,
		Status:    MessageStatusSent,
		IsOwn:     true,
	}

	if msg.ID != "msg123" {
		t.Errorf("expected msg123, got %s", msg.ID)
	}
	if msg.Content != "你好" {
		t.Errorf("expected 你好, got %s", msg.Content)
	}
	if msg.Type != MessageTypeText {
		t.Errorf("expected MessageTypeText, got %v", msg.Type)
	}
	if !msg.IsOwn {
		t.Error("expected IsOwn to be true")
	}
}

func TestAttachment(t *testing.T) {
	attachment := &Attachment{
		ID:        "att123",
		FileName:  "document.pdf",
		FileSize:  1024 * 1024,
		MimeType:  "application/pdf",
		Progress:  50.0,
		Status:    "downloading",
		LocalPath: "/downloads/document.pdf",
	}

	if attachment.FileName != "document.pdf" {
		t.Errorf("expected document.pdf, got %s", attachment.FileName)
	}
	if attachment.FileSize != 1024*1024 {
		t.Errorf("expected 1048576, got %d", attachment.FileSize)
	}
	if attachment.Progress != 50.0 {
		t.Errorf("expected 50.0, got %f", attachment.Progress)
	}
}

func TestReaction(t *testing.T) {
	reaction := &Reaction{
		Emoji: "👍",
		Users: []string{"user1", "user2"},
		Count: 2,
	}

	if reaction.Emoji != "👍" {
		t.Errorf("expected 👍, got %s", reaction.Emoji)
	}
	if reaction.Count != 2 {
		t.Errorf("expected 2, got %d", reaction.Count)
	}
	if len(reaction.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(reaction.Users))
	}
}

func TestNetworkStats(t *testing.T) {
	stats := &NetworkStats{
		Jitter:     10.5,
		Latency:    50.0,
		PacketLoss: 0.1,
		Bandwidth:  1000000.0,
	}

	if stats.Jitter != 10.5 {
		t.Errorf("expected 10.5, got %f", stats.Jitter)
	}
	if stats.Latency != 50.0 {
		t.Errorf("expected 50.0, got %f", stats.Latency)
	}
	if stats.PacketLoss != 0.1 {
		t.Errorf("expected 0.1, got %f", stats.PacketLoss)
	}
}

func TestCallUIState(t *testing.T) {
	state := &CallUIState{
		SessionID:       "call123",
		PeerID:          "QmPeer456",
		PeerName:        "张三",
		IsIncoming:      true,
		HasVideo:        true,
		IsMuted:         false,
		IsVideoOff:      false,
		IsScreenSharing: false,
		Duration:        60 * time.Second,
		Quality:         0.95,
	}

	if state.SessionID != "call123" {
		t.Errorf("expected call123, got %s", state.SessionID)
	}
	if state.PeerName != "张三" {
		t.Errorf("expected 张三, got %s", state.PeerName)
	}
	if !state.IsIncoming {
		t.Error("expected IsIncoming to be true")
	}
	if state.Duration != 60*time.Second {
		t.Errorf("expected 60s, got %v", state.Duration)
	}
}

func TestCallInfo(t *testing.T) {
	info := &CallInfo{
		SessionID:  "call123",
		PeerID:     "QmPeer456",
		PeerName:   "张三",
		PeerAvatar: "avatar.png",
		HasVideo:   true,
		IsIncoming: true,
	}

	if info.PeerID != "QmPeer456" {
		t.Errorf("expected QmPeer456, got %s", info.PeerID)
	}
	if !info.HasVideo {
		t.Error("expected HasVideo to be true")
	}
}

func TestTransferTaskUI(t *testing.T) {
	task := &TransferTaskUI{
		TaskID:    "task123",
		FileName:  "video.mp4",
		FileSize:  1024 * 1024 * 100,
		Direction: "download",
		PeerID:    "QmPeer456",
		PeerName:  "张三",
		Progress:  25.0,
		Speed:     1024 * 1024,
		Status:    "downloading",
		StartTime: time.Now().Unix(),
	}

	if task.TaskID != "task123" {
		t.Errorf("expected task123, got %s", task.TaskID)
	}
	if task.Progress != 25.0 {
		t.Errorf("expected 25.0, got %f", task.Progress)
	}
	if task.Status != "downloading" {
		t.Errorf("expected downloading, got %s", task.Status)
	}
}
