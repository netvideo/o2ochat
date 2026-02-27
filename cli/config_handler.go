package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ConfigHandler struct {
	configPath string
	configData map[string]interface{}
	mu         *mockMutex
}

type mockMutex struct{}

func (m *mockMutex) Lock()   {}
func (m *mockMutex) Unlock() {}

func NewConfigHandler(configPath string) *ConfigHandler {
	return &ConfigHandler{
		configPath: configPath,
		configData: make(map[string]interface{}),
		mu:         &mockMutex{},
	}
}

func (h *ConfigHandler) LoadConfig() error {
	if h.configPath == "" {
		h.configPath = os.Getenv("HOME") + "/.o2ochat/config.yaml"
	}

	data, err := os.ReadFile(h.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			h.configData = getDefaultConfig()
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &h.configData)
}

func (h *ConfigHandler) SaveConfig() error {
	data, err := json.MarshalIndent(h.configData, "", "  ")
	if err != nil {
		return err
	}

	dir := strings.TrimRight(h.configPath, "/config.yaml")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(h.configPath, data, 0644)
}

func (h *ConfigHandler) ShowConfig(section string) (*CommandResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.LoadConfig(); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to load config: %v", err),
			ExitCode: 1,
		}, nil
	}

	if section == "" {
		return &CommandResult{
			Success:  true,
			Message:  "current configuration",
			Data:     h.configData,
			ExitCode: 0,
		}, nil
	}

	parts := strings.SplitN(section, ".", 2)
	if len(parts) != 2 {
		if val, ok := h.configData[section]; ok {
			return &CommandResult{
				Success:  true,
				Message:  fmt.Sprintf("section: %s", section),
				Data:     val,
				ExitCode: 0,
			}, nil
		}
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("section not found: %s", section),
			ExitCode: 1,
		}, nil
	}

	sectionName := parts[0]
	key := parts[1]

	if sectionData, ok := h.configData[sectionName].(map[string]interface{}); ok {
		if val, ok := sectionData[key]; ok {
			return &CommandResult{
				Success:  true,
				Message:  fmt.Sprintf("%s.%s", sectionName, key),
				Data:     val,
				ExitCode: 0,
			}, nil
		}
	}

	return &CommandResult{
		Success:  false,
		Message:  fmt.Sprintf("key not found: %s.%s", sectionName, key),
		ExitCode: 1,
	}, nil
}

func (h *ConfigHandler) SetConfig(section, key, value string) (*CommandResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.LoadConfig(); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to load config: %v", err),
			ExitCode: 1,
		}, nil
	}

	fullKey := section + "." + key
	parts := strings.SplitN(fullKey, ".", 2)

	if len(parts) != 2 {
		return &CommandResult{
			Success:  false,
			Message:  "invalid key format, use section.key",
			ExitCode: 1,
		}, nil
	}

	sectionName := parts[0]
	keyName := parts[1]

	if _, ok := h.configData[sectionName]; !ok {
		h.configData[sectionName] = make(map[string]interface{})
	}

	sectionData, ok := h.configData[sectionName].(map[string]interface{})
	if !ok {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("section %s is not a map", sectionName),
			ExitCode: 1,
		}, nil
	}

	parsedValue := h.parseValue(value)
	sectionData[keyName] = parsedValue

	if err := h.SaveConfig(); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to save config: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success:  true,
		Message:  fmt.Sprintf("set %s = %v", fullKey, parsedValue),
		ExitCode: 0,
	}, nil
}

func (h *ConfigHandler) ResetConfig() (*CommandResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.configData = getDefaultConfig()

	if err := h.SaveConfig(); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to reset config: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success:  true,
		Message:  "configuration reset to defaults",
		ExitCode: 0,
	}, nil
}

func (h *ConfigHandler) ImportConfig(path string) (*CommandResult, error) {
	if path == "" {
		return &CommandResult{
			Success:  false,
			Message:  "import path is required",
			ExitCode: 1,
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to read file: %v", err),
			ExitCode: 1,
		}, nil
	}

	var imported map[string]interface{}
	if err := json.Unmarshal(data, &imported); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to parse config: %v", err),
			ExitCode: 1,
		}, nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.configData = imported

	if err := h.SaveConfig(); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to save config: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success:  true,
		Message:  fmt.Sprintf("configuration imported from %s", path),
		ExitCode: 0,
	}, nil
}

func (h *ConfigHandler) ExportConfig(path string) (*CommandResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.LoadConfig(); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to load config: %v", err),
			ExitCode: 1,
		}, nil
	}

	if path == "" {
		path = "config.export.json"
	}

	data, err := json.MarshalIndent(h.configData, "", "  ")
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to marshal config: %v", err),
			ExitCode: 1,
		}, nil
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("failed to write file: %v", err),
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success:  true,
		Message:  fmt.Sprintf("configuration exported to %s", path),
		ExitCode: 0,
	}, nil
}

func (h *ConfigHandler) ValidateConfig() (*CommandResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.LoadConfig(); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("config validation failed: %v", err),
			ExitCode: 1,
		}, nil
	}

	errors := []string{}

	if network, ok := h.configData["network"].(map[string]interface{}); ok {
		if maxConn, ok := network["max_connections"].(float64); ok {
			if maxConn <= 0 || maxConn > 10000 {
				errors = append(errors, "network.max_connections must be between 1 and 10000")
			}
		}
	}

	if timeout, ok := h.configData["timeout"].(float64); ok {
		if timeout <= 0 || timeout > 300 {
			errors = append(errors, "timeout must be between 1 and 300 seconds")
		}
	}

	if len(errors) > 0 {
		return &CommandResult{
			Success:  false,
			Message:  "config validation failed",
			Data:     map[string][]string{"errors": errors},
			ExitCode: 1,
		}, nil
	}

	return &CommandResult{
		Success:  true,
		Message:  "configuration is valid",
		ExitCode: 0,
	}, nil
}

func (h *ConfigHandler) parseValue(value string) interface{} {
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	var intVal int
	if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
		return intVal
	}

	var floatVal float64
	if _, err := fmt.Sscanf(value, "%f", &floatVal); err == nil {
		return floatVal
	}

	return value
}

func getDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"network": map[string]interface{}{
			"max_connections": 100,
			"timeout":         30,
			"retry_count":     3,
		},
		"storage": map[string]interface{}{
			"path":     "~/.o2ochat/data",
			"max_size": 1073741824,
		},
		"ui": map[string]interface{}{
			"theme":     "light",
			"language":  "en",
			"font_size": 14,
		},
		"logging": map[string]interface{}{
			"level":  "info",
			"format": "text",
			"output": "stdout",
		},
		"security": map[string]interface{}{
			"enable_encryption": true,
			"verify_peer":       true,
		},
	}
}

type ShowConfigCommandHandler struct {
	configHandler *ConfigHandler
}

func NewShowConfigCommandHandler(configPath string) *ShowConfigCommandHandler {
	return &ShowConfigCommandHandler{
		configHandler: NewConfigHandler(configPath),
	}
}

func (h *ShowConfigCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	section := ""
	if v, ok := ctx.Flags["section"].(string); ok {
		section = v
	}
	return h.configHandler.ShowConfig(section)
}

func (h *ShowConfigCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *ShowConfigCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--section", "--help"}, nil
	}
	return nil, nil
}

func (h *ShowConfigCommandHandler) Help() string {
	return `显示O2OChat配置信息。

用法:
  o2ochat config show [选项]

选项:
  -s, --section    配置章节名称 (如: network, storage, ui)
  -h, --help       显示帮助信息

示例:
  o2ochat config show
  o2ochat config show --section network`
}

type SetConfigCommandHandler struct {
	configHandler *ConfigHandler
}

func NewSetConfigCommandHandler(configPath string) *SetConfigCommandHandler {
	return &SetConfigCommandHandler{
		configHandler: NewConfigHandler(configPath),
	}
}

func (h *SetConfigCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	section := ""
	key := ""
	value := ""

	if v, ok := ctx.Flags["section"].(string); ok {
		section = v
	}
	if v, ok := ctx.Flags["key"].(string); ok {
		key = v
	}
	if v, ok := ctx.Flags["value"].(string); ok {
		value = v
	}

	if section == "" || key == "" {
		return &CommandResult{
			Success:  false,
			Message:  "section and key are required",
			ExitCode: 1,
		}, nil
	}

	return h.configHandler.SetConfig(section, key, value)
}

func (h *SetConfigCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *SetConfigCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	if len(word) > 1 && word[0] == '-' {
		return []string{"--section", "--key", "--value", "--help"}, nil
	}
	return nil, nil
}

func (h *SetConfigCommandHandler) Help() string {
	return `设置O2OChat配置项。

用法:
  o2ochat config set [选项]

选项:
  -s, --section    配置章节名称
  -k, --key        配置键名
  -v, --value      配置值
  -h, --help       显示帮助信息

示例:
  o2ochat config set --section network --key max_connections --value 200
  o2ochat config set --section ui --key theme --value dark`
}

type ResetConfigCommandHandler struct {
	configHandler *ConfigHandler
}

func NewResetConfigCommandHandler(configPath string) *ResetConfigCommandHandler {
	return &ResetConfigCommandHandler{
		configHandler: NewConfigHandler(configPath),
	}
}

func (h *ResetConfigCommandHandler) Execute(ctx *CommandContext) (*CommandResult, error) {
	return h.configHandler.ResetConfig()
}

func (h *ResetConfigCommandHandler) Validate(ctx *CommandContext) error {
	return nil
}

func (h *ResetConfigCommandHandler) Autocomplete(ctx *CommandContext, word string) ([]string, error) {
	return nil, nil
}

func (h *ResetConfigCommandHandler) Help() string {
	return `重置O2OChat配置为默认值。

用法:
  o2ochat config reset [选项]

选项:
  -h, --help       显示帮助信息

示例:
  o2ochat config reset`
}
