package ui

import "sync"

type DefaultUIManager struct {
	config       *UIConfig
	initialized  bool
	mu           sync.RWMutex
	trayIconData []byte
	trayTooltip  string
	unreadCount  int
}

func NewUIManager() UIManager {
	return &DefaultUIManager{
		config: &UIConfig{
			Theme:          ThemeDark,
			Language:       "zh-CN",
			FontSize:       14,
			ShowAvatars:    true,
			ShowTimestamps: true,
			NotifySounds:   true,
			NotifyDesktop:  true,
			AutoStart:      false,
			MinimizeToTray: true,
		},
	}
}

func (m *DefaultUIManager) Initialize(config *UIConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return ErrAlreadyInitialized
	}

	if config != nil {
		m.config = config
	}

	m.initialized = true
	return nil
}

func (m *DefaultUIManager) ShowMainWindow() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	return nil
}

func (m *DefaultUIManager) HideMainWindow() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	return nil
}

func (m *DefaultUIManager) Quit() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	m.initialized = false
	return nil
}

func (m *DefaultUIManager) UpdateConfig(config *UIConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	if config == nil {
		return ErrInvalidConfig
	}

	m.config = config
	return nil
}

func (m *DefaultUIManager) GetConfig() (*UIConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, ErrNotInitialized
	}

	return m.config, nil
}

func (m *DefaultUIManager) ShowNotification(title, message string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	return nil
}

func (m *DefaultUIManager) PlaySound(soundType string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	return nil
}

func (m *DefaultUIManager) SetTrayIcon(iconData []byte, tooltip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	m.trayIconData = iconData
	m.trayTooltip = tooltip
	return nil
}

func (m *DefaultUIManager) UpdateUnreadCount(count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	m.unreadCount = count
	return nil
}

func (m *DefaultUIManager) Destroy() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = false
	m.trayIconData = nil
	m.config = nil
	return nil
}
