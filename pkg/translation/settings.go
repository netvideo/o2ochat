// Package translation provides message translation settings
package translation

import (
	"sync"

	"github.com/netvideo/o2ochat/pkg/ai"
)

// TranslationDirection represents the direction of translation
type TranslationDirection string

const (
	// DirectionIncoming for incoming (received) messages
	DirectionIncoming TranslationDirection = "incoming"
	// DirectionOutgoing for outgoing (sent) messages
	DirectionOutgoing TranslationDirection = "outgoing"
)

// TranslationSettings holds translation settings for a chat
type TranslationSettings struct {
	// EnableIncomingTranslation enables translation for received messages
	EnableIncomingTranslation bool `json:"enable_incoming_translation"`

	// EnableOutgoingTranslation enables translation for sent messages
	EnableOutgoingTranslation bool `json:"enable_outgoing_translation"`

	// SourceLanguage is the source language code (e.g., "zh-CN")
	SourceLanguage string `json:"source_language"`

	// TargetLanguage is the target language code (e.g., "en")
	TargetLanguage string `json:"target_language"`

	// AutoDetectSource enables automatic source language detection
	AutoDetectSource bool `json:"auto_detect_source"`

	// Provider is the AI provider to use for translation
	Provider ai.ProviderType `json:"provider"`

	// ShowOriginal shows original text alongside translation
	ShowOriginal bool `json:"show_original"`

	// CacheTranslations caches translation results
	CacheTranslations bool `json:"cache_translations"`
}

// ChatTranslationSettings holds translation settings for a specific chat
type ChatTranslationSettings struct {
	// ChatID is the unique identifier for the chat
	ChatID string `json:"chat_id"`

	// PeerID is the peer's identifier
	PeerID string `json:"peer_id"`

	// Settings is the translation settings
	Settings *TranslationSettings `json:"settings"`

	// Enabled indicates if translation is enabled for this chat
	Enabled bool `json:"enabled"`
}

// TranslationManager manages translation settings for chats
type TranslationManager struct {
	settings       map[string]*ChatTranslationSettings
	globalSettings *TranslationSettings
	mu             sync.RWMutex
}

// NewTranslationManager creates a new translation manager
func NewTranslationManager() *TranslationManager {
	return &TranslationManager{
		settings: make(map[string]*ChatTranslationSettings),
		globalSettings: &TranslationSettings{
			EnableIncomingTranslation: false,
			EnableOutgoingTranslation: false,
			SourceLanguage:            "auto",
			TargetLanguage:            "en",
			AutoDetectSource:          true,
			Provider:                  ai.ProviderOllama,
			ShowOriginal:              true,
			CacheTranslations:         true,
		},
	}
}

// GetSettings returns translation settings for a chat
func (m *TranslationManager) GetSettings(chatID string) (*TranslationSettings, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if chatSettings, exists := m.settings[chatID]; exists {
		if chatSettings.Enabled && chatSettings.Settings != nil {
			return chatSettings.Settings, true
		}
	}

	// Return global settings if chat-specific not found
	return m.globalSettings, m.globalSettings.EnableIncomingTranslation || m.globalSettings.EnableOutgoingTranslation
}

// SetSettings sets translation settings for a chat
func (m *TranslationManager) SetSettings(chatID, peerID string, settings *TranslationSettings) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings[chatID] = &ChatTranslationSettings{
		ChatID:   chatID,
		PeerID:   peerID,
		Settings: settings,
		Enabled:  settings.EnableIncomingTranslation || settings.EnableOutgoingTranslation,
	}
}

// SetIncomingTranslation enables/disables incoming message translation
func (m *TranslationManager) SetIncomingTranslation(chatID string, enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if settings, exists := m.settings[chatID]; exists {
		settings.Settings.EnableIncomingTranslation = enabled
		settings.Enabled = enabled || settings.Settings.EnableOutgoingTranslation
	} else {
		// Create new settings
		newSettings := &TranslationSettings{
			EnableIncomingTranslation: enabled,
			EnableOutgoingTranslation: false,
			SourceLanguage:            m.globalSettings.SourceLanguage,
			TargetLanguage:            m.globalSettings.TargetLanguage,
			AutoDetectSource:          m.globalSettings.AutoDetectSource,
			Provider:                  m.globalSettings.Provider,
			ShowOriginal:              m.globalSettings.ShowOriginal,
			CacheTranslations:         m.globalSettings.CacheTranslations,
		}

		m.settings[chatID] = &ChatTranslationSettings{
			ChatID:   chatID,
			Settings: newSettings,
			Enabled:  enabled,
		}
	}
}

// SetOutgoingTranslation enables/disables outgoing message translation
func (m *TranslationManager) SetOutgoingTranslation(chatID string, enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if settings, exists := m.settings[chatID]; exists {
		settings.Settings.EnableOutgoingTranslation = enabled
		settings.Enabled = enabled || settings.Settings.EnableIncomingTranslation
	} else {
		// Create new settings
		newSettings := &TranslationSettings{
			EnableIncomingTranslation: false,
			EnableOutgoingTranslation: enabled,
			SourceLanguage:            m.globalSettings.SourceLanguage,
			TargetLanguage:            m.globalSettings.TargetLanguage,
			AutoDetectSource:          m.globalSettings.AutoDetectSource,
			Provider:                  m.globalSettings.Provider,
			ShowOriginal:              m.globalSettings.ShowOriginal,
			CacheTranslations:         m.globalSettings.CacheTranslations,
		}

		m.settings[chatID] = &ChatTranslationSettings{
			ChatID:   chatID,
			Settings: newSettings,
			Enabled:  enabled,
		}
	}
}

// GetIncomingTranslationStatus returns incoming translation status for a chat
func (m *TranslationManager) GetIncomingTranslationStatus(chatID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if settings, exists := m.settings[chatID]; exists {
		return settings.Settings.EnableIncomingTranslation
	}

	return m.globalSettings.EnableIncomingTranslation
}

// GetOutgoingTranslationStatus returns outgoing translation status for a chat
func (m *TranslationManager) GetOutgoingTranslationStatus(chatID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if settings, exists := m.settings[chatID]; exists {
		return settings.Settings.EnableOutgoingTranslation
	}

	return m.globalSettings.EnableOutgoingTranslation
}

// SetGlobalSettings sets global translation settings
func (m *TranslationManager) SetGlobalSettings(settings *TranslationSettings) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.globalSettings = settings
}

// GetGlobalSettings returns global translation settings
func (m *TranslationManager) GetGlobalSettings() *TranslationSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.globalSettings
}

// RemoveSettings removes translation settings for a chat
func (m *TranslationManager) RemoveSettings(chatID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.settings, chatID)
}

// ListSettings returns all chat translation settings
func (m *TranslationManager) ListSettings() []*ChatTranslationSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ChatTranslationSettings, 0, len(m.settings))
	for _, settings := range m.settings {
		result = append(result, settings)
	}

	return result
}

// ShouldTranslate checks if a message should be translated based on direction
func (m *TranslationManager) ShouldTranslate(chatID string, direction TranslationDirection) bool {
	settings, enabled := m.GetSettings(chatID)
	if !enabled {
		return false
	}

	switch direction {
	case DirectionIncoming:
		return settings.EnableIncomingTranslation
	case DirectionOutgoing:
		return settings.EnableOutgoingTranslation
	default:
		return false
	}
}

// IsTranslationEnabled returns true if translation is enabled for any direction
func (m *TranslationManager) IsTranslationEnabled(chatID string) bool {
	settings, enabled := m.GetSettings(chatID)
	if !enabled {
		return false
	}

	return settings.EnableIncomingTranslation || settings.EnableOutgoingTranslation
}
