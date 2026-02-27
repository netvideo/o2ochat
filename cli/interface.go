package cli

type CLIManager interface {
	Initialize(config *CLIConfig) error
	RegisterCommand(cmd *CommandConfig, handler CommandHandler) error
	Execute(args []string) (*CommandResult, error)
	GetHelp(command string) (string, error)
	GetCommands() ([]*CommandConfig, error)
	SetOutputFormat(format OutputFormat) error
	RunInteractive() error
	Close() error
}

type CommandHandler interface {
	Execute(ctx *CommandContext) (*CommandResult, error)
	Validate(ctx *CommandContext) error
	Autocomplete(ctx *CommandContext, word string) ([]string, error)
	Help() string
}

type ConfigCLI interface {
	ShowConfig(section string) (*CommandResult, error)
	SetConfig(section, key, value string) (*CommandResult, error)
	ResetConfig() (*CommandResult, error)
	ImportConfig(path string) (*CommandResult, error)
	ExportConfig(path string) (*CommandResult, error)
	ValidateConfig() (*CommandResult, error)
}

type NetworkCLI interface {
	TestConnection(peerID string) (*CommandResult, error)
	DiagnoseNetwork() (*CommandResult, error)
	ShowConnections() (*CommandResult, error)
	ShowNetworkStats() (*CommandResult, error)
	TestNATTraversal() (*CommandResult, error)
	ShowRoutingTable() (*CommandResult, error)
}

type DataCLI interface {
	BackupData(path string) (*CommandResult, error)
	RestoreData(path string) (*CommandResult, error)
	CleanupData(olderThan string) (*CommandResult, error)
	ExportMessages(peerID, path string) (*CommandResult, error)
	ImportMessages(path string) (*CommandResult, error)
	ShowStorageStats() (*CommandResult, error)
}

type DebugCLI interface {
	ShowLogs(level string, lines int) (*CommandResult, error)
	MonitorPerformance(interval int) (*CommandResult, error)
	ProfileMemory(duration int) (*CommandResult, error)
	ProfileCPU(duration int) (*CommandResult, error)
	TraceCall(function string) (*CommandResult, error)
	StressTest(target string, duration int) (*CommandResult, error)
}

type AppCLI interface {
	Start(daemon bool, configPath string) (*CommandResult, error)
	Stop() (*CommandResult, error)
	Restart() (*CommandResult, error)
	Status() (*CommandResult, error)
	Version() (*CommandResult, error)
}
