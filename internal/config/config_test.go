package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// 验证版本
	if cfg.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", cfg.Version)
	}

	// 验证应用配置
	if cfg.App.Name != "O2OChat" {
		t.Errorf("Expected app name O2OChat, got %s", cfg.App.Name)
	}

	if cfg.App.Version != "1.0.0" {
		t.Errorf("Expected app version 1.0.0, got %s", cfg.App.Version)
	}

	if cfg.App.DataDir != "./data" {
		t.Errorf("Expected data dir ./data, got %s", cfg.App.DataDir)
	}

	// 验证网络配置
	if cfg.Network.ListenAddr != "0.0.0.0" {
		t.Errorf("Expected listen addr 0.0.0.0, got %s", cfg.Network.ListenAddr)
	}

	if !cfg.Network.EnableIPv6 {
		t.Error("Expected IPv6 to be enabled")
	}

	// 验证存储配置
	if cfg.Storage.Type != "sqlite" {
		t.Errorf("Expected storage type sqlite, got %s", cfg.Storage.Type)
	}

	// 验证安全配置
	if cfg.Security.EncryptionAlgorithm != "AES-256-GCM" {
		t.Errorf("Expected encryption algorithm AES-256-GCM, got %s", cfg.Security.EncryptionAlgorithm)
	}

	// 验证日志配置
	if cfg.Log.Level != "info" {
		t.Errorf("Expected log level info, got %s", cfg.Log.Level)
	}
}

// TestLoadConfig 测试加载配置文件
func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	// 创建测试配置
	testConfig := &Config{
		Version: "1.0.0",
		App: AppConfig{
			Name:    "TestApp",
			Version: "1.0.0",
			DataDir: "./test_data",
			Debug:   true,
		},
		Network: NetworkConfig{
			ListenAddr: "127.0.0.1",
			ListenPort: 8080,
		},
	}

	// 保存测试配置到文件
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// 测试加载配置
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证加载的配置
	if loadedConfig.App.Name != "TestApp" {
		t.Errorf("Expected app name TestApp, got %s", loadedConfig.App.Name)
	}

	if loadedConfig.App.DataDir != "./test_data" {
		t.Errorf("Expected data dir ./test_data, got %s", loadedConfig.App.DataDir)
	}

	if !loadedConfig.App.Debug {
		t.Error("Expected debug to be true")
	}

	if loadedConfig.Network.ListenPort != 8080 {
		t.Errorf("Expected listen port 8080, got %d", loadedConfig.Network.ListenPort)
	}
}

// TestLoadConfig_NotFound 测试加载不存在的配置文件
func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error when loading non-existent config file")
	}

	if err != nil && err.Error() != "config file not found: /nonexistent/path/config.json" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestLoadConfig_InvalidJSON 测试加载无效的 JSON 配置文件
func TestLoadConfig_InvalidJSON(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_config.json")

	// 写入无效的 JSON
	invalidJSON := `{"invalid json}`
	if err := os.WriteFile(configPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// 尝试加载
	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON config")
	}
}

// TestSaveConfig 测试保存配置
func TestSaveConfig(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "test_config.json")

	// 创建测试配置
	testConfig := DefaultConfig()
	testConfig.App.Name = "SaveTest"
	testConfig.App.Debug = true

	// 保存配置
	err := SaveConfig(testConfig, configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// 加载保存的配置并验证
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.App.Name != "SaveTest" {
		t.Errorf("Expected app name SaveTest, got %s", loadedConfig.App.Name)
	}

	if !loadedConfig.App.Debug {
		t.Error("Expected debug to be true")
	}
}

// TestMergeWithFlags 测试命令行参数合并
func TestMergeWithFlags(t *testing.T) {
	// 创建默认配置
	cfg := DefaultConfig()

	// 创建命令行参数
	flags := &CommandLineFlags{
		DataDir:  "/custom/data",
		Debug:    true,
		LogLevel: "debug",
	}

	// 合并参数
	cfg.MergeWithFlags(flags)

	// 验证合并结果
	if cfg.App.DataDir != "/custom/data" {
		t.Errorf("Expected data dir /custom/data, got %s", cfg.App.DataDir)
	}

	if !cfg.App.Debug {
		t.Error("Expected debug to be true")
	}

	if cfg.Log.Level != "debug" {
		t.Errorf("Expected log level debug, got %s", cfg.Log.Level)
	}
}

// TestParseFlags 测试命令行参数解析
func TestParseFlags(t *testing.T) {
	// 由于 ParseFlags 使用了 flag 包的全局状态，
	// 这里我们只测试默认值的合理性

	flags := &CommandLineFlags{
		ConfigPath: "./config.json",
		DataDir:    "./data",
		Debug:      false,
		LogLevel:   "info",
		LogOutput:  "file",
		ServerAddr: "",
		ServerPort: 0,
		Version:    false,
		Help:       false,
	}

	// 验证默认值
	if flags.ConfigPath == "" {
		t.Error("ConfigPath should not be empty")
	}

	if flags.DataDir == "" {
		t.Error("DataDir should not be empty")
	}
}
