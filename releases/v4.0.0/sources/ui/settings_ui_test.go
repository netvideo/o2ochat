package ui

import (
	"testing"
)

func TestNewSettingsUI(t *testing.T) {
	settings := NewSettingsUI()
	if settings == nil {
		t.Error("expected non-nil SettingsUI")
	}

	defaultSettings, ok := settings.(*DefaultSettingsUI)
	if !ok {
		t.Error("expected DefaultSettingsUI type")
	}

	if defaultSettings.settings == nil {
		t.Error("expected settings map to be initialized")
	}
}

func TestSettingsUIShowSettings(t *testing.T) {
	settings := NewSettingsUI()

	err := settings.ShowSettings()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSettingsUIUpdateSetting(t *testing.T) {
	settings := NewSettingsUI()

	err := settings.UpdateSetting("general", "theme", "dark")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = settings.UpdateSetting("", "theme", "dark")
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}

	err = settings.UpdateSetting("general", "", "dark")
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestSettingsUIGetSetting(t *testing.T) {
	settings := NewSettingsUI()

	settings.UpdateSetting("general", "theme", "dark")

	value, err := settings.GetSetting("general", "theme")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "dark" {
		t.Errorf("expected dark, got %v", value)
	}

	_, err = settings.GetSetting("", "theme")
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}

	_, err = settings.GetSetting("general", "")
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}

	_, err = settings.GetSetting("nonexistent", "theme")
	if err != ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}

	_, err = settings.GetSetting("general", "nonexistent")
	if err != ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestSettingsUIResetSettings(t *testing.T) {
	settings := NewSettingsUI()

	settings.UpdateSetting("general", "theme", "dark")
	settings.UpdateSetting("general", "language", "zh-CN")

	err := settings.ResetSettings()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = settings.GetSetting("general", "theme")
	if err != ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestSettingsUISetSaveCallback(t *testing.T) {
	settings := NewSettingsUI()

	callbackCalled := false
	callback := func(config *UIConfig) {
		callbackCalled = true
	}

	err := settings.SetSaveCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultSettings := settings.(*DefaultSettingsUI)
	defaultSettings.saveCallback(&UIConfig{Theme: ThemeDark})
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}

func TestSettingsUISetTestCallback(t *testing.T) {
	settings := NewSettingsUI()

	callbackCalled := false
	callback := func(testType string) {
		callbackCalled = true
	}

	err := settings.SetTestCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultSettings := settings.(*DefaultSettingsUI)
	defaultSettings.testCallback("network")
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}
