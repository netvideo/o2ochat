package cli

import (
	"testing"
	"time"
)

func TestFlagTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected FlagType
	}{
		{"String", FlagTypeString},
		{"Int", FlagTypeInt},
		{"Bool", FlagTypeBool},
		{"Float", FlagTypeFloat},
		{"Duration", FlagTypeDuration},
		{"Path", FlagTypePath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("flag type should not be empty")
			}
		})
	}
}

func TestOutputFormats(t *testing.T) {
	tests := []struct {
		name     string
		expected OutputFormat
	}{
		{"Text", OutputFormatText},
		{"JSON", OutputFormatJSON},
		{"YAML", OutputFormatYAML},
		{"Table", OutputFormatTable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("output format should not be empty")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		expected LogLevel
	}{
		{"Debug", LogLevelDebug},
		{"Info", LogLevelInfo},
		{"Warn", LogLevelWarn},
		{"Error", LogLevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("log level should not be empty")
			}
		})
	}
}

func TestCommandConfig(t *testing.T) {
	config := DefaultCommandConfig()

	if config.Aliases == nil {
		t.Error("aliases should not be nil")
	}
	if config.Flags == nil {
		t.Error("flags should not be nil")
	}
	if config.Subcommands == nil {
		t.Error("subcommands should not be nil")
	}
}

func TestCommandResult(t *testing.T) {
	result := &CommandResult{
		Success:  true,
		Message:  "Command executed successfully",
		Data:     nil,
		ExitCode: 0,
		Duration: 100 * time.Millisecond,
	}

	if !result.Success && result.ExitCode == 0 {
		t.Error("success should be true when exit code is 0")
	}
	if result.Duration <= 0 {
		t.Error("duration should be positive")
	}
}

func TestCommandContext(t *testing.T) {
	ctx := &CommandContext{
		Args:   []string{"arg1", "arg2"},
		Flags:  map[string]interface{}{"verbose": true},
		Config: DefaultCLIConfig(),
	}

	if ctx.Args == nil {
		t.Error("args should not be nil")
	}
	if ctx.Flags == nil {
		t.Error("flags should not be nil")
	}
	if ctx.Config == nil {
		t.Error("config should not be nil")
	}
}

func TestCLIConfig(t *testing.T) {
	config := DefaultCLIConfig()

	if config.LogLevel == "" {
		t.Error("log level should not be empty")
	}
	if config.LogFormat == "" {
		t.Error("log format should not be empty")
	}
	if config.Timeout <= 0 {
		t.Error("timeout should be positive")
	}
}

func TestFlag(t *testing.T) {
	flag := &Flag{
		Name:        "verbose",
		Short:      "v",
		Description: "Enable verbose output",
		Type:       FlagTypeBool,
		Default:    false,
		Required:   false,
		Hidden:     false,
	}

	if flag.Name == "" {
		t.Error("flag name should not be empty")
	}
	if flag.Type == "" {
		t.Error("flag type should not be empty")
	}
}

func TestNetworkStats(t *testing.T) {
	stats := &NetworkStats{
		TotalConnections: 2,
		Connections: []ConnectionInfo{
			{
				PeerID:    "QmPeer123",
				Type:      "quic",
				State:     "connected",
				Duration:  5 * time.Minute,
				BytesSent: 1024,
			},
		},
		Bandwidth: 1000000,
		Latency:   50 * time.Millisecond,
	}

	if stats.TotalConnections < 0 {
		t.Error("total connections should not be negative")
	}
}

func TestConnectionInfo(t *testing.T) {
	info := &ConnectionInfo{
		PeerID:    "QmPeer123",
		Type:      "quic",
		State:     "connected",
		Duration:  5 * time.Minute,
		BytesSent: 1024,
		BytesRecv: 512,
	}

	if info.PeerID == "" {
		t.Error("peer ID should not be empty")
	}
	if info.Type == "" {
		t.Error("type should not be empty")
	}
	if info.State == "" {
		t.Error("state should not be empty")
	}
}

func TestStorageStats(t *testing.T) {
	stats := &StorageStats{
		TotalSize:    1000000000,
		UsedSize:     500000000,
		MessageCount: 1000,
		FileCount:    100,
	}

	if stats.TotalSize <= 0 {
		t.Error("total size should be positive")
	}
	if stats.UsedSize < 0 {
		t.Error("used size should not be negative")
	}
}

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		err  error
		name string
	}{
		{ErrCommandNotFound, "ErrCommandNotFound"},
		{ErrInvalidArgs, "ErrInvalidArgs"},
		{ErrMissingRequiredArg, "ErrMissingRequiredArg"},
		{ErrInvalidFlag, "ErrInvalidFlag"},
		{ErrFlagNotFound, "ErrFlagNotFound"},
		{ErrCommandFailed, "ErrCommandFailed"},
		{ErrConfigNotFound, "ErrConfigNotFound"},
		{ErrConfigInvalid, "ErrConfigInvalid"},
		{ErrPermissionDenied, "ErrPermissionDenied"},
		{ErrFileNotFound, "ErrFileNotFound"},
		{ErrFileReadFailed, "ErrFileReadFailed"},
		{ErrFileWriteFailed, "ErrFileWriteFailed"},
		{ErrTimeout, "ErrTimeout"},
		{ErrAlreadyRunning, "ErrAlreadyRunning"},
		{ErrNotRunning, "ErrNotRunning"},
		{ErrInvalidOutputFormat, "ErrInvalidOutputFormat"},
		{ErrCLINotInitialized, "ErrCLINotInitialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("error message should not be empty")
			}
		})
	}
}

func TestCLIError(t *testing.T) {
	innerErr := ErrCommandNotFound
	cliErr := NewCLIError("CMD_NOT_FOUND", "command not found", innerErr)

	if cliErr.Code != "CMD_NOT_FOUND" {
		t.Errorf("expected code CMD_NOT_FOUND, got %s", cliErr.Code)
	}
	if cliErr.Message != "command not found" {
		t.Errorf("expected message 'command not found', got %s", cliErr.Message)
	}
	if cliErr.Unwrap() != innerErr {
		t.Error("unwrap should return inner error")
	}
	if cliErr.Error() == "" {
		t.Error("error should not be empty")
	}
}

func TestInterfaceCompatibility(t *testing.T) {
	var _ CLIManager = nil
	var _ CommandHandler = nil
	var _ ConfigCLI = nil
	var _ NetworkCLI = nil
	var _ DataCLI = nil
	var _ DebugCLI = nil
	var _ AppCLI = nil
}
