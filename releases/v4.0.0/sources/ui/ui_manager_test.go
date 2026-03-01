package ui

import (
	"testing"
)

func TestNewUIManager(t *testing.T) {
	manager := NewUIManager()
	if manager == nil {
		t.Error("expected non-nil UIManager")
	}

	defaultMgr, ok := manager.(*DefaultUIManager)
	if !ok {
		t.Error("expected DefaultUIManager type")
	}

	if defaultMgr.config == nil {
		t.Error("expected default config to be set")
	}
	if defaultMgr.config.Theme != ThemeDark {
		t.Errorf("expected default ThemeDark, got %v", defaultMgr.config.Theme)
	}
	if defaultMgr.config.Language != "zh-CN" {
		t.Errorf("expected default Language zh-CN, got %s", defaultMgr.config.Language)
	}
	if defaultMgr.initialized {
		t.Error("expected initialized to be false initially")
	}
}

func TestUIManagerInitialize(t *testing.T) {
	manager := NewUIManager()

	err := manager.Initialize(nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = manager.Initialize(&UIConfig{Theme: ThemeLight})
	if err != ErrAlreadyInitialized {
		t.Errorf("expected ErrAlreadyInitialized, got %v", err)
	}
}

func TestUIManagerShowHideMainWindow(t *testing.T) {
	manager := NewUIManager()

	_, err := manager.GetConfig()
	if err != ErrNotInitialized {
		t.Errorf("expected ErrNotInitialized before Initialize, got %v", err)
	}

	manager.Initialize(nil)

	err = manager.ShowMainWindow()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = manager.HideMainWindow()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUIManagerUpdateConfig(t *testing.T) {
	manager := NewUIManager()
	manager.Initialize(nil)

	newConfig := &UIConfig{
		Theme:    ThemeLight,
		Language: "en-US",
		FontSize: 16,
	}

	err := manager.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	config, _ := manager.GetConfig()
	if config.Theme != ThemeLight {
		t.Errorf("expected ThemeLight, got %v", config.Theme)
	}
	if config.FontSize != 16 {
		t.Errorf("expected FontSize 16, got %d", config.FontSize)
	}

	err = manager.UpdateConfig(nil)
	if err != ErrInvalidConfig {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestUIManagerGetConfig(t *testing.T) {
	manager := NewUIManager()

	_, err := manager.GetConfig()
	if err != ErrNotInitialized {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}

	manager.Initialize(&UIConfig{FontSize: 18})

	config, err := manager.GetConfig()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if config.FontSize != 18 {
		t.Errorf("expected FontSize 18, got %d", config.FontSize)
	}
}

func TestUIManagerShowNotification(t *testing.T) {
	manager := NewUIManager()
	manager.Initialize(nil)

	err := manager.ShowNotification("Test Title", "Test Message")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUIManagerPlaySound(t *testing.T) {
	manager := NewUIManager()
	manager.Initialize(nil)

	err := manager.PlaySound("message")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUIManagerSetTrayIcon(t *testing.T) {
	manager := NewUIManager()
	manager.Initialize(nil)

	iconData := []byte{0x00, 0x01, 0x02}
	err := manager.SetTrayIcon(iconData, "O2OChat")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mgr := manager.(*DefaultUIManager)
	if string(mgr.trayIconData) != string(iconData) {
		t.Error("tray icon data not stored correctly")
	}
	if mgr.trayTooltip != "O2OChat" {
		t.Errorf("expected tooltip O2OChat, got %s", mgr.trayTooltip)
	}
}

func TestUIManagerUpdateUnreadCount(t *testing.T) {
	manager := NewUIManager()
	manager.Initialize(nil)

	err := manager.UpdateUnreadCount(5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mgr := manager.(*DefaultUIManager)
	if mgr.unreadCount != 5 {
		t.Errorf("expected unreadCount 5, got %d", mgr.unreadCount)
	}
}

func TestUIManagerQuit(t *testing.T) {
	manager := NewUIManager()
	manager.Initialize(nil)

	err := manager.Quit()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = manager.ShowMainWindow()
	if err != ErrNotInitialized {
		t.Errorf("expected ErrNotInitialized after Quit, got %v", err)
	}
}

func TestUIManagerDestroy(t *testing.T) {
	manager := NewUIManager()
	manager.Initialize(nil)
	manager.SetTrayIcon([]byte{0x00}, "test")

	err := manager.Destroy()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mgr := manager.(*DefaultUIManager)
	if mgr.initialized {
		t.Error("expected initialized to be false after Destroy")
	}
	if mgr.trayIconData != nil {
		t.Error("expected trayIconData to be nil after Destroy")
	}
	if mgr.config != nil {
		t.Error("expected config to be nil after Destroy")
	}
}
