package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用程序主配置结构
type Config struct {
	Version  string         `json:"version"`
	App      AppConfig      `json:"app"`
	Identity IdentityConfig `json:"identity"`
	Network  NetworkConfig  `json:"network"`
	Storage  StorageConfig  `json:"storage"`
	Security SecurityConfig `json:"security"`
	Log      LogConfig      `json:"log"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	DataDir    string `json:"data_dir"`
	ConfigPath string `json:"config_path"`
	Debug      bool   `json:"debug"`
	AutoStart  bool   `json:"auto_start"`
}

// IdentityConfig 身份管理配置
type IdentityConfig struct {
	KeyType        string `json:"key_type"`
	KeyLength      int    `json:"key_length"`
	PeerIDEncoding string `json:"peer_id_encoding"`
	StoragePath    string `json:"storage_path"`
	AutoCreate     bool   `json:"auto_create"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ListenAddr         string   `json:"listen_addr"`
	ListenPort         int      `json:"listen_port"`
	EnableIPv6         bool     `json:"enable_ipv6"`
	EnableIPv4         bool     `json:"enable_ipv4"`
	EnableUPnP         bool     `json:"enable_upnp"`
	EnableNATTraversal bool     `json:"enable_nat_traversal"`
	STUNServers        []string `json:"stun_servers"`
	TURNServers        []string `json:"turn_servers"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type            string `json:"type"`
	Path            string `json:"path"`
	MaxSize         int64  `json:"max_size"`
	CacheSize       int    `json:"cache_size"`
	AutoCleanup     bool   `json:"auto_cleanup"`
	CleanupInterval string `json:"cleanup_interval"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EncryptionAlgorithm         string `json:"encryption_algorithm"`
	KeyExchangeMethod           string `json:"key_exchange_method"`
	EnablePerfectForwardSecrecy bool   `json:"enable_pfs"`
	CertificatePath             string `json:"certificate_path"`
	PrivateKeyPath              string `json:"private_key_path"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"`
	Output     string `json:"output"`
	FilePath   string `json:"file_path"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0.0",
		App: AppConfig{
			Name:       "O2OChat",
			Version:    "1.0.0",
			DataDir:    "./data",
			ConfigPath: "./config.json",
			Debug:      false,
			AutoStart:  false,
		},
		Identity: IdentityConfig{
			KeyType:        "ed25519",
			KeyLength:      256,
			PeerIDEncoding: "base58",
			StoragePath:    "./data/identity",
			AutoCreate:     true,
		},
		Network: NetworkConfig{
			ListenAddr:         "0.0.0.0",
			ListenPort:         0,
			EnableIPv6:         true,
			EnableIPv4:         true,
			EnableUPnP:         true,
			EnableNATTraversal: true,
			STUNServers:        []string{"stun:stun.l.google.com:19302"},
			TURNServers:        []string{},
		},
		Storage: StorageConfig{
			Type:            "sqlite",
			Path:            "./data/storage",
			MaxSize:         10 * 1024 * 1024 * 1024, // 10GB
			CacheSize:       100,
			AutoCleanup:     true,
			CleanupInterval: "24h",
		},
		Security: SecurityConfig{
			EncryptionAlgorithm:         "AES-256-GCM",
			KeyExchangeMethod:           "X25519",
			EnablePerfectForwardSecrecy: true,
			CertificatePath:             "",
			PrivateKeyPath:              "",
		},
		Log: LogConfig{
			Level:      "info",
			Output:     "file",
			FilePath:   "./data/logs/o2ochat.log",
			MaxSize:    100, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		},
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
