package ui

import (
	"sync"
	"time"
)

type NotificationType string
type NotificationPriority string

const (
	NotificationTypeMessage   NotificationType = "message"
	NotificationTypeCall     NotificationType = "call"
	NotificationTypeFile    NotificationType = "file"
	NotificationTypeSystem  NotificationType = "system"
	NotificationTypeAlert   NotificationType = "alert"
)

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityHigh    NotificationPriority = "high"
	NotificationPriorityUrgent  NotificationPriority = "urgent"
)

type Notification struct {
	ID          string
	Type        NotificationType
	Title       string
	Message     string
	Icon        string
	PeerID      string
	PeerName    string
	Timestamp   time.Time
	Priority    NotificationPriority
	Read        bool
	Clicked     bool
	Actions     []NotificationAction
}

type NotificationAction struct {
	ID      string
	Label   string
	Action  string
}

type NotificationComponent struct {
	mu              sync.RWMutex
	notifications   []*Notification
	maxNotifications int
	enabled        bool
	soundEnabled    bool
	desktopEnabled  bool
	previewEnabled  bool
	onNotificationClick func(notification *Notification)
	onNotificationAction func(notification *Notification, action string)
	onMarkRead      func(notificationID string)
	onClear         func(notificationID string)
}

func NewNotificationComponent() *NotificationComponent {
	return &NotificationComponent{
		notifications:    make([]*Notification, 0),
		maxNotifications: 100,
		enabled:         true,
		soundEnabled:    true,
		desktopEnabled:  true,
		previewEnabled:  true,
	}
}

func (nc *NotificationComponent) AddNotification(notif *Notification) error {
	if !nc.enabled {
		return nil
	}

	nc.mu.Lock()
	defer nc.mu.Unlock()

	notif.Timestamp = time.Now()
	notif.ID = generateNotificationID()

	nc.notifications = append(nc.notifications, notif)

	if len(nc.notifications) > nc.maxNotifications {
		nc.notifications = nc.notifications[len(nc.notifications)-nc.maxNotifications:]
	}

	if nc.soundEnabled && notif.Type == NotificationTypeMessage {
		// Trigger sound
	}

	if nc.desktopEnabled && notif.Type == NotificationTypeMessage {
		// Trigger desktop notification
	}

	return nil
}

func (nc *NotificationComponent) AddMessageNotification(peerID, peerName, message string) error {
	notif := &Notification{
		Type:     NotificationTypeMessage,
		Title:    peerName,
		Message:  message,
		PeerID:   peerID,
		PeerName: peerName,
		Priority: NotificationPriorityNormal,
	}
	return nc.AddNotification(notif)
}

func (nc *NotificationComponent) AddCallNotification(peerID, peerName string, isVideo bool) error {
	title := "来电"
	if !isVideo {
		title = "语音来电"
	}

	notif := &Notification{
		Type:     NotificationTypeCall,
		Title:    title,
		Message:  peerName + " 呼叫你",
		PeerID:   peerID,
		PeerName: peerName,
		Priority: NotificationPriorityHigh,
		Actions: []NotificationAction{
			{ID: "accept", Label: "接听", Action: "accept"},
			{ID: "reject", Label: "拒绝", Action: "reject"},
		},
	}
	return nc.AddNotification(notif)
}

func (nc *NotificationComponent) AddFileNotification(peerID, peerName, fileName string, fileSize int64) error {
	notif := &Notification{
		Type:     NotificationTypeFile,
		Title:    "文件传输",
		Message:  peerName + " 发送文件: " + fileName,
		PeerID:   peerID,
		PeerName: peerName,
		Priority: NotificationPriorityNormal,
	}
	return nc.AddNotification(notif)
}

func (nc *NotificationComponent) AddSystemNotification(title, message string, priority NotificationPriority) error {
	notif := &Notification{
		Type:     NotificationTypeSystem,
		Title:    title,
		Message:  message,
		Priority: priority,
	}
	return nc.AddNotification(notif)
}

func (nc *NotificationComponent) AddAlertNotification(title, message string) error {
	notif := &Notification{
		Type:     NotificationTypeAlert,
		Title:    title,
		Message:  message,
		Priority: NotificationPriorityUrgent,
	}
	return nc.AddNotification(notif)
}

func (nc *NotificationComponent) RemoveNotification(id string) {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	for i, n := range nc.notifications {
		if n.ID == id {
			nc.notifications = append(nc.notifications[:i], nc.notifications[i+1:]...)
			return
		}
	}
}

func (nc *NotificationComponent) MarkAsRead(id string) {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	for _, n := range nc.notifications {
		if n.ID == id {
			n.Read = true
			if nc.onMarkRead != nil {
				nc.onMarkRead(id)
			}
			return
		}
	}
}

func (nc *NotificationComponent) MarkAllAsRead() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	for _, n := range nc.notifications {
		n.Read = true
	}
}

func (nc *NotificationComponent) ClearAll() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	nc.notifications = make([]*Notification, 0)
}

func (nc *NotificationComponent) ClearByType(notifType NotificationType) {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	result := make([]*Notification, 0)
	for _, n := range nc.notifications {
		if n.Type != notifType {
			result = append(result, n)
		}
	}
	nc.notifications = result
}

func (nc *NotificationComponent) GetNotification(id string) (*Notification, bool) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	for _, n := range nc.notifications {
		if n.ID == id {
			return n, true
		}
	}
	return nil, false
}

func (nc *NotificationComponent) GetAllNotifications() []*Notification {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	result := make([]*Notification, len(nc.notifications))
	copy(result, nc.notifications)
	return result
}

func (nc *NotificationComponent) GetUnreadCount() int {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	count := 0
	for _, n := range nc.notifications {
		if !n.Read {
			count++
		}
	}
	return count
}

func (nc *NotificationComponent) GetByType(notifType NotificationType) []*Notification {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	result := make([]*Notification, 0)
	for _, n := range nc.notifications {
		if n.Type == notifType {
			result = append(result, n)
		}
	}
	return result
}

func (nc *NotificationComponent) GetByPeer(peerID string) []*Notification {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	result := make([]*Notification, 0)
	for _, n := range nc.notifications {
		if n.PeerID == peerID {
			result = append(result, n)
		}
	}
	return result
}

func (nc *NotificationComponent) HandleClick(id string) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	for _, n := range nc.notifications {
		if n.ID == id {
			n.Clicked = true
			n.Read = true
			if nc.onNotificationClick != nil {
				nc.onNotificationClick(n)
			}
			return
		}
	}
}

func (nc *NotificationComponent) HandleAction(id, action string) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	for _, n := range nc.notifications {
		if n.ID == id {
			if nc.onNotificationAction != nil {
				nc.onNotificationAction(n, action)
			}
			return
		}
	}
}

func (nc *NotificationComponent) SetEnabled(enabled bool) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.enabled = enabled
}

func (nc *NotificationComponent) SetSoundEnabled(enabled bool) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.soundEnabled = enabled
}

func (nc *NotificationComponent) SetDesktopEnabled(enabled bool) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.desktopEnabled = enabled
}

func (nc *NotificationComponent) SetPreviewEnabled(enabled bool) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.previewEnabled = enabled
}

func (nc *NotificationComponent) SetMaxNotifications(max int) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.maxNotifications = max
}

func (nc *NotificationComponent) IsEnabled() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.enabled
}

func (nc *NotificationComponent) IsSoundEnabled() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.soundEnabled
}

func (nc *NotificationComponent) IsDesktopEnabled() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.desktopEnabled
}

func (nc *NotificationComponent) IsPreviewEnabled() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.previewEnabled
}

func (nc *NotificationComponent) SetOnNotificationClick(callback func(notification *Notification)) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.onNotificationClick = callback
}

func (nc *NotificationComponent) SetOnNotificationAction(callback func(notification *Notification, action string)) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.onNotificationAction = callback
}

func (nc *NotificationComponent) SetOnMarkRead(callback func(notificationID string)) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.onMarkRead = callback
}

func (nc *NotificationComponent) SetOnClear(callback func(notificationID string)) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.onClear = callback
}

func generateNotificationID() string {
	return "notif_" + time.Now().Format("20060102150405")
}
