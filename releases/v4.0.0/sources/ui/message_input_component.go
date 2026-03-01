package ui

type MessageInputComponent struct {
	onSend         func(text string, attachments []string)
	onAttach       func()
	onImageSelect  func()
	onFileSelect   func()
	maxTextLength  int
	enabled        bool
}

func NewMessageInputComponent() *MessageInputComponent {
	return &MessageInputComponent{
		maxTextLength: 4096,
		enabled:       true,
	}
}

func (mi *MessageInputComponent) SendMessage(text string) {
	if !mi.enabled || text == "" {
		return
	}

	if mi.onSend != nil {
		mi.onSend(text, nil)
	}
}

func (mi *MessageInputComponent) SendMessageWithAttachments(text string, attachments []string) {
	if !mi.enabled || text == "" {
		return
	}

	if mi.onSend != nil {
		mi.onSend(text, attachments)
	}
}

func (mi *MessageInputComponent) HandleAttach() {
	if mi.enabled && mi.onAttach != nil {
		mi.onAttach()
	}
}

func (mi *MessageInputComponent) HandleImageSelect() {
	if mi.enabled && mi.onImageSelect != nil {
		mi.onImageSelect()
	}
}

func (mi *MessageInputComponent) HandleFileSelect() {
	if mi.enabled && mi.onFileSelect != nil {
		mi.onFileSelect()
	}
}

func (mi *MessageInputComponent) SetOnSend(callback func(text string, attachments []string)) {
	mi.onSend = callback
}

func (mi *MessageInputComponent) SetOnAttach(callback func()) {
	mi.onAttach = callback
}

func (mi *MessageInputComponent) SetOnImageSelect(callback func()) {
	mi.onImageSelect = callback
}

func (mi *MessageInputComponent) SetOnFileSelect(callback func()) {
	mi.onFileSelect = callback
}

func (mi *MessageInputComponent) SetEnabled(enabled bool) {
	mi.enabled = enabled
}

func (mi *MessageInputComponent) IsEnabled() bool {
	return mi.enabled
}

func (mi *MessageInputComponent) SetMaxTextLength(length int) {
	mi.maxTextLength = length
}

func (mi *MessageInputComponent) GetMaxTextLength() int {
	return mi.maxTextLength
}

func (mi *MessageInputComponent) ValidateText(text string) (bool, string) {
	if !mi.enabled {
		return false, "input disabled"
	}
	if len(text) == 0 {
		return false, "empty message"
	}
	if len(text) > mi.maxTextLength {
		return false, "message too long"
	}
	return true, ""
}
