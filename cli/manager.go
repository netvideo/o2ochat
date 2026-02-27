package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type commandEntry struct {
	config  *CommandConfig
	handler CommandHandler
}

type DefaultCLIManager struct {
	config          *CLIConfig
	commands        map[string]*commandEntry
	outputFormat    OutputFormat
	output          io.Writer
	error           io.Writer
	input           io.Reader
	mu              sync.RWMutex
	version         string
	buildTime       string
	interactiveMode bool
}

func NewCLIManager() CLIManager {
	return &DefaultCLIManager{
		commands:     make(map[string]*commandEntry),
		output:       os.Stdout,
		error:        os.Stderr,
		input:        os.Stdin,
		outputFormat: OutputFormatText,
		version:      "1.0.0",
		buildTime:    time.Now().Format(time.RFC3339),
	}
}

func (m *DefaultCLIManager) Initialize(config *CLIConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		m.config = DefaultCLIConfig()
	} else {
		m.config = config
	}

	if m.config.Timeout <= 0 {
		m.config.Timeout = 30
	}

	if m.config.Version != "" {
		m.version = m.config.Version
	}

	if m.config.BuildTime != "" {
		m.buildTime = m.config.BuildTime
	}

	return nil
}

func (m *DefaultCLIManager) RegisterCommand(cmd *CommandConfig, handler CommandHandler) error {
	if cmd == nil || handler == nil {
		return fmt.Errorf("command config and handler cannot be nil")
	}

	if cmd.Name == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &commandEntry{
		config:  cmd,
		handler: handler,
	}

	m.commands[cmd.Name] = entry

	for _, alias := range cmd.Aliases {
		m.commands[alias] = entry
	}

	return nil
}

func (m *DefaultCLIManager) Execute(args []string) (*CommandResult, error) {
	startTime := time.Now()

	if len(args) == 0 {
		return &CommandResult{
			Success:  false,
			Message:  "no command provided",
			ExitCode: 1,
		}, nil
	}

	m.mu.RLock()
	entry, exists := m.commands[args[0]]
	m.mu.RUnlock()

	if !exists {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("unknown command: %s", args[0]),
			ExitCode: 127,
		}, nil
	}

	flags, err := m.parseFlags(args[1:], entry.config.Flags)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("flag parsing error: %v", err),
			ExitCode: 1,
		}, nil
	}

	ctx := &CommandContext{
		Args:    args[1:],
		Flags:   flags,
		Config:  m.config,
		Output:  m.output,
		Error:   m.error,
		Input:   m.input,
		Handler: entry.handler,
	}

	if err := entry.handler.Validate(ctx); err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("validation error: %v", err),
			ExitCode: 1,
		}, nil
	}

	result, err := entry.handler.Execute(ctx)
	if err != nil {
		return &CommandResult{
			Success:  false,
			Message:  fmt.Sprintf("execution error: %v", err),
			ExitCode: 1,
		}, nil
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

func (m *DefaultCLIManager) parseFlags(args []string, flagDefs []Flag) (map[string]interface{}, error) {
	flags := make(map[string]interface{})

	for _, fd := range flagDefs {
		flags[fd.Name] = fd.Default
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			i++
			continue
		}

		name := strings.TrimLeft(arg, "-")
		var flagDef *Flag

		for _, fd := range flagDefs {
			if fd.Name == name || fd.Short == name {
				flagDef = &fd
				break
			}
		}

		if flagDef == nil {
			return nil, fmt.Errorf("unknown flag: %s", name)
		}

		if flagDef.Type == "bool" {
			flags[flagDef.Name] = true
		} else if i+1 < len(args) {
			i++
			flags[flagDef.Name] = args[i]
		}

		i++
	}

	return flags, nil
}

func (m *DefaultCLIManager) GetHelp(command string) (string, error) {
	m.mu.RLock()
	entry, exists := m.commands[command]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("unknown command: %s", command)
	}

	return entry.handler.Help(), nil
}

func (m *DefaultCLIManager) GetCommands() ([]*CommandConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := make(map[string]bool)
	var commands []*CommandConfig

	for name, entry := range m.commands {
		if !seen[name] {
			seen[name] = true
			commands = append(commands, entry.config)
		}
	}

	return commands, nil
}

func (m *DefaultCLIManager) SetOutputFormat(format OutputFormat) error {
	validFormats := map[OutputFormat]bool{
		OutputFormatText:  true,
		OutputFormatJSON:  true,
		OutputFormatYAML:  true,
		OutputFormatTable: true,
	}

	if !validFormats[format] {
		return fmt.Errorf("invalid output format: %s", format)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.outputFormat = format
	return nil
}

func (m *DefaultCLIManager) RunInteractive() error {
	m.interactiveMode = true

	fmt.Fprintln(m.output, "O2OChat CLI - Interactive Mode")
	fmt.Fprintln(m.output, "Type 'help' for available commands, 'exit' to quit")
	fmt.Fprintln(m.output)

	scanner := NewScanner(m.input)

	for {
		fmt.Fprint(m.output, "> ")

		line, err := scanner.Scan()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(m.error, "Error reading input: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "exit" || line == "quit" {
			break
		}

		if line == "help" {
			commands, _ := m.GetCommands()
			fmt.Fprintln(m.output, "Available commands:")
			for _, cmd := range commands {
				if !cmd.Hidden {
					fmt.Fprintf(m.output, "  %s - %s\n", cmd.Name, cmd.Description)
				}
			}
			continue
		}

		args := strings.Fields(line)
		result, err := m.Execute(args)
		if err != nil {
			fmt.Fprintf(m.error, "Error: %v\n", err)
			continue
		}

		if !result.Success {
			fmt.Fprintf(m.output, "Error: %s\n", result.Message)
		} else if result.Message != "" {
			fmt.Fprintln(m.output, result.Message)
		}

		if result.Data != nil {
			data, _ := json.MarshalIndent(result.Data, "", "  ")
			fmt.Fprintln(m.output, string(data))
		}
	}

	m.interactiveMode = false
	return nil
}

func (m *DefaultCLIManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commands = make(map[string]*commandEntry)
	return nil
}

func (m *DefaultCLIManager) GetVersion() string {
	return m.version
}

func (m *DefaultCLIManager) GetBuildTime() string {
	return m.buildTime
}

func FormatOutput(result *CommandResult, format OutputFormat) (string, error) {
	switch format {
	case OutputFormatJSON:
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil

	case OutputFormatYAML:
		data, err := json.Marshal(result)
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		err = json.Indent(&buf, data, "", "  ")
		if err != nil {
			return "", err
		}
		return buf.String(), nil

	case OutputFormatText:
		fallthrough
	default:
		if result.Success {
			return result.Message, nil
		}
		return fmt.Sprintf("Error: %s", result.Message), nil
	}
}

type OutputFormatter interface {
	Format(result *CommandResult) (string, error)
}

type TextFormatter struct{}

func (f *TextFormatter) Format(result *CommandResult) (string, error) {
	if result.Success {
		return result.Message, nil
	}
	return fmt.Sprintf("Error: %s", result.Message), nil
}

type JSONFormatter struct{}

func (f *JSONFormatter) Format(result *CommandResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
