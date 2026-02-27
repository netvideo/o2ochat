package ui

import (
	"time"
)

type UITheme string

const (
	ThemeLight UITheme = "light"
	ThemeDark  UITheme = "dark"
	ThemeAuto  UITheme = "auto"
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeFile     MessageType = "file"
	MessageTypeVoice    MessageType = "voice"
	MessageTypeVideo    MessageType = "video"
	MessageTypeSystem   MessageType = "system"
)

type MessageStatus string

const (
	MessageStatusSending   MessageStatus = "sending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

type CallAction string

const (
	CallActionAccept       CallAction = "accept"
	CallActionReject       CallAction = "reject"
	CallActionMute         CallAction = "mute"
	CallActionUnmute       CallAction = "unmute"
	CallActionVideoOn      CallAction = "video_on"
	CallActionVideoOff     CallAction = "video_off"
	CallActionScreenShare  CallAction = "screen_share"
	CallActionEnd          CallAction = "end"
)

type UIConfig struct {
	Theme          UITheme `json:"theme"`
	Language       string  `json:"language"`
	FontSize       int     `json:"font_size"`
	ShowAvatars    bool    `json:"show_avatars"`
	ShowTimestamps bool    `json:"show_timestamps"`
	NotifySounds   bool    `json:"notify_sounds"`
	NotifyDesktop  bool    `json:"notify_desktop"`
	AutoStart      bool    `json:"auto_start"`
	MinimizeToTray bool    `json:"minimize_to_tray"`
}

type ContactInfo struct {
	PeerID      string    `json:"peer_id"`
	Name        string    `json:"name"`
	Avatar      string    `json:"avatar"`
	LastSeen    time.Time `json:"last_seen"`
	Online      bool      `json:"online"`
	UnreadCount int       `json:"unread_count"`
	IsFavorite  bool      `json:"is_favorite"`
	Groups      []string  `json:"groups"`
}

type MessageItem struct {
	ID          string          `json:"id"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Content     string          `json:"content"`
	Type        MessageType     `json:"type"`
	Timestamp   time.Time       `json:"timestamp"`
	Status      MessageStatus   `json:"status"`
	IsOwn       bool            `json:"is_own"`
	Attachments []*Attachment   `json:"attachments"`
	Reactions   []*Reaction     `json:"reactions"`
}

type Attachment struct {
	ID        string    `json:"id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	Progress  float64   `json:"progress"`
	Status    string    `json:"status"`
	LocalPath string    `json:"local_path"`
}

type Reaction struct {
	Emoji   string   `json:"emoji"`
	Users   []string `json:"users"`
	Count   int      `json:"count"`
}

type NetworkStats struct {
	Jitter    float64 `json:"jitter"`
	Latency   float64 `json:"latency"`
	PacketLoss float64 `json:"packet_loss"`
	Bandwidth float64 `json:"bandwidth"`
}

type CallUIState struct {
	SessionID       string        `json:"session_id"`
	PeerID          string        `json:"peer_id"`
	PeerName        string        `json:"peer_name"`
	IsIncoming      bool          `json:"is_incoming"`
	HasVideo        bool          `json:"has_video"`
	IsMuted         bool          `json:"is_muted"`
	IsVideoOff      bool          `json:"is_video_off"`
	IsScreenSharing bool          `json:"is_screen_sharing"`
	Duration        time.Duration `json:"duration"`
	Quality        float64       `json:"quality"`
	NetworkStats   *NetworkStats `json:"network_stats"`
}

type CallInfo struct {
	SessionID  string `json:"session_id"`
	PeerID     string `json:"peer_id"`
	PeerName   string `json:"peer_name"`
	PeerAvatar string `json:"peer_avatar"`
	HasVideo   bool   `json:"has_video"`
	IsIncoming bool   `json:"is_incoming"`
}

type TransferTaskUI struct {
	TaskID     string  `json:"task_id"`
	FileName   string  `json:"file_name"`
	FileSize   int64   `json:"file_size"`
	Direction  string  `json:"direction"`
	PeerID     string  `json:"peer_id"`
	PeerName   string  `json:"peer_name"`
	Progress   float64 `json:"progress"`
	Speed      float64 `json:"speed"`
	Status     string  `json:"status"`
	StartTime  int64   `json:"start_time"`
	EndTime    int64   `json:"end_time"`
}
