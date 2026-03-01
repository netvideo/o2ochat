package ui

type UIManager interface {
	Initialize(config *UIConfig) error
	ShowMainWindow() error
	HideMainWindow() error
	Quit() error
	UpdateConfig(config *UIConfig) error
	GetConfig() (*UIConfig, error)
	ShowNotification(title, message string) error
	PlaySound(soundType string) error
	SetTrayIcon(iconData []byte, tooltip string) error
	UpdateUnreadCount(count int) error
	Destroy() error
}

type ChatUI interface {
	OpenChat(peerID string) error
	CloseChat(peerID string) error
	AddMessage(message *MessageItem) error
	UpdateMessageStatus(messageID string, status MessageStatus) error
	ClearChat(peerID string) error
	SearchMessages(peerID, query string) ([]*MessageItem, error)
	GetChatHistory(peerID string, limit int) ([]*MessageItem, error)
	SetInputCallback(callback func(text string, attachments []string)) error
	SetReactionCallback(callback func(messageID string, reaction string)) error
}

type ContactUI interface {
	AddContact(contact *ContactInfo) error
	RemoveContact(peerID string) error
	UpdateContact(contact *ContactInfo) error
	SearchContacts(query string) ([]*ContactInfo, error)
	GetAllContacts() ([]*ContactInfo, error)
	GetOnlineContacts() ([]*ContactInfo, error)
	SetContactSelectCallback(callback func(peerID string)) error
	SetAddContactCallback(callback func(peerID, name string)) error
}

type FileTransferUI interface {
	ShowFileTransfer() error
	AddTransferTask(task *TransferTaskUI) error
	UpdateTransferProgress(taskID string, progress float64, speed float64) error
	CompleteTransferTask(taskID string, success bool, errorMsg string) error
	CancelTransferTask(taskID string) error
	OpenFileLocation(filePath string) error
	SetFileSelectCallback(callback func(filePaths []string)) error
	SetFolderSelectCallback(callback func(folderPath string)) error
}

type CallUI interface {
	ShowIncomingCall(callInfo *CallInfo) error
	ShowOutgoingCall(callInfo *CallInfo) error
	UpdateCallState(state *CallUIState) error
	EndCall(sessionID string) error
	SetVideoFrameCallback(callback func(frame []byte, width, height int)) error
	SetAudioDataCallback(callback func(data []byte, sampleRate int)) error
	SetCallControlCallback(callback func(action CallAction)) error
}

type SettingsUI interface {
	ShowSettings() error
	UpdateSetting(section, key string, value interface{}) error
	GetSetting(section, key string) (interface{}, error)
	ResetSettings() error
	SetSaveCallback(callback func(config *UIConfig)) error
	SetTestCallback(callback func(testType string)) error
}
