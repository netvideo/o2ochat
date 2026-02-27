package ui

import (
	"sync"
)

type DefaultSettingsUI struct {
	mu            sync.RWMutex
	settings      map[string]map[string]interface{}
	saveCallback  func(config *UIConfig)
	testCallback  func(testType string)
}

func NewSettingsUI() SettingsUI {
	return &DefaultSettingsUI{
		settings: make(map[string]map[string]interface{}),
	}
}

func (s *DefaultSettingsUI) ShowSettings() error {
	return nil
}

func (s *DefaultSettingsUI) UpdateSetting(section, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if section == "" || key == "" {
		return ErrInvalidParameter
	}

	if s.settings[section] == nil {
		s.settings[section] = make(map[string]interface{})
	}

	s.settings[section][key] = value
	return nil
}

func (s *DefaultSettingsUI) GetSetting(section, key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if section == "" || key == "" {
		return nil, ErrInvalidParameter
	}

	if s.settings[section] == nil {
		return nil, ErrResourceNotFound
	}

	value, ok := s.settings[section][key]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return value, nil
}

func (s *DefaultSettingsUI) ResetSettings() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings = make(map[string]map[string]interface{})
	return nil
}

func (s *DefaultSettingsUI) SetSaveCallback(callback func(config *UIConfig)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.saveCallback = callback
	return nil
}

func (s *DefaultSettingsUI) SetTestCallback(callback func(testType string)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.testCallback = callback
	return nil
}
