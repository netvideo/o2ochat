package cli

import (
	"io"
	"time"
)

type OutputFormat string

const (
	OutputFormatText  OutputFormat = "text"
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatYAML  OutputFormat = "yaml"
	OutputFormatTable OutputFormat = "table"
)

type FlagType string

const (
	FlagTypeString   FlagType = "string"
	FlagTypeInt      FlagType = "int"
	FlagTypeBool     FlagType = "bool"
	FlagTypeFloat    FlagType = "float"
	FlagTypeDuration FlagType = "duration"
	FlagTypePath     FlagType = "path"
)

type CommandConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Usage       string   `json:"usage"`
	Aliases     []string `json:"aliases"`
	Flags       []Flag   `json:"flags"`
	Subcommands []string `json:"subcommands"`
	Hidden      bool     `json:"hidden"`
}

type Flag struct {
	Name        string      `json:"name"`
	Short       string      `json:"short"`
	Description string      `json:"description"`
	Type        FlagType    `json:"type"`
	Default     interface{} `json:"default"`
	Required    bool        `json:"required"`
	Hidden      bool        `json:"hidden"`
}

type CommandContext struct {
	Args    []string               `json:"args"`
	Flags   map[string]interface{} `json:"flags"`
	Config  *CLIConfig             `json:"config"`
	Output  io.Writer              `json:"output"`
	Error   io.Writer              `json:"error"`
	Input   io.Reader              `json:"input"`
	Handler CommandHandler         `json:"-"`
}

type CLIConfig struct {
	LogLevel    string `json:"log_level"`
	LogFormat   string `json:"log_format"`
	ColorOutput bool   `json:"color_output"`
	Interactive bool   `json:"interactive"`
	Timeout     int    `json:"timeout"`
	ConfigPath  string `json:"config_path"`
	Version     string `json:"version"`
	BuildTime   string `json:"build_time"`
	DataPath    string `json:"data_path"`
	EnableDebug bool   `json:"enable_debug"`
}

type CommandResult struct {
	Success  bool          `json:"success"`
	Message  string        `json:"message"`
	Data     interface{}   `json:"data"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type NetworkStats struct {
	TotalConnections int              `json:"total_connections"`
	Connections      []ConnectionInfo `json:"connections"`
	Bandwidth        int64            `json:"bandwidth"`
	Latency          time.Duration    `json:"latency"`
}

type ConnectionInfo struct {
	PeerID    string        `json:"peer_id"`
	Type      string        `json:"type"`
	State     string        `json:"state"`
	Duration  time.Duration `json:"duration"`
	BytesSent int64         `json:"bytes_sent"`
	BytesRecv int64         `json:"bytes_received"`
}

type StorageStats struct {
	TotalSize    int64 `json:"total_size"`
	UsedSize     int64 `json:"used_size"`
	MessageCount int64 `json:"message_count"`
	FileCount    int64 `json:"file_count"`
}

func DefaultCLIConfig() *CLIConfig {
	return &CLIConfig{
		LogLevel:    "info",
		LogFormat:   "text",
		ColorOutput: true,
		Interactive: false,
		Timeout:     30,
		ConfigPath:  "~/.o2ochat/cli.yaml",
	}
}

func DefaultCommandConfig() *CommandConfig {
	return &CommandConfig{
		Name:        "",
		Description: "",
		Usage:       "",
		Aliases:     []string{},
		Flags:       []Flag{},
		Subcommands: []string{},
		Hidden:      false,
	}
}
