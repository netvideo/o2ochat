package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// InteractiveCLI represents an interactive command-line interface
type InteractiveCLI struct {
	reader   *bufio.Reader
	commands map[string]Command
	history  []string
	prompt   string
	isActive bool
}

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Usage       string
	Handler     func(args []string) error
	Aliases     []string
}

// NewInteractiveCLI creates a new interactive CLI
func NewInteractiveCLI() *InteractiveCLI {
	return &InteractiveCLI{
		reader:   bufio.NewReader(os.Stdin),
		commands: make(map[string]Command),
		history:  make([]string, 0),
		prompt:   "o2ochat> ",
		isActive: false,
	}
}

// RegisterCommand registers a command
func (cli *InteractiveCLI) RegisterCommand(cmd Command) {
	cli.commands[cmd.Name] = cmd
	for _, alias := range cmd.Aliases {
		cli.commands[alias] = cmd
	}
}

// Start starts the interactive CLI
func (cli *InteractiveCLI) Start() error {
	cli.isActive = true

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║   O2OChat Interactive CLI v3.0.0      ║")
	fmt.Println("║   Type 'help' for available commands  ║")
	fmt.Println("║   Type 'exit' to quit                 ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	for cli.isActive {
		fmt.Print(cli.prompt)

		input, err := cli.reader.ReadString('\n')
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Add to history
		cli.history = append(cli.history, input)

		// Parse command
		parts := strings.Fields(input)
		cmdName := parts[0]
		args := parts[1:]

		// Execute command
		if err := cli.executeCommand(cmdName, args); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
	}

	return nil
}

// executeCommand executes a command
func (cli *InteractiveCLI) executeCommand(name string, args []string) error {
	cmd, exists := cli.commands[name]
	if !exists {
		fmt.Printf("❌ Unknown command: %s\n", name)
		fmt.Println("Type 'help' for available commands")
		return nil
	}

	return cmd.Handler(args)
}

// Stop stops the interactive CLI
func (cli *InteractiveCLI) Stop() {
	cli.isActive = false
	fmt.Println("👋 Goodbye!")
}

// GetHistory returns command history
func (cli *InteractiveCLI) GetHistory() []string {
	return cli.history
}

// ClearHistory clears command history
func (cli *InteractiveCLI) ClearHistory() {
	cli.history = make([]string, 0)
}

// SetPrompt sets the CLI prompt
func (cli *InteractiveCLI) SetPrompt(prompt string) {
	cli.prompt = prompt
}

// RegisterDefaultCommands registers default commands
func (cli *InteractiveCLI) RegisterDefaultCommands() {
	// help command
	cli.RegisterCommand(Command{
		Name:        "help",
		Description: "Show help information",
		Usage:       "help [command]",
		Handler: func(args []string) error {
			if len(args) > 0 {
				// Show specific command help
				cmdName := args[0]
				cmd, exists := cli.commands[cmdName]
				if !exists {
					fmt.Printf("❌ Unknown command: %s\n", cmdName)
					return nil
				}
				fmt.Printf("Command: %s\n", cmd.Name)
				fmt.Printf("Description: %s\n", cmd.Description)
				fmt.Printf("Usage: %s\n", cmd.Usage)
				if len(cmd.Aliases) > 0 {
					fmt.Printf("Aliases: %s\n", strings.Join(cmd.Aliases, ", "))
				}
			} else {
				// Show all commands
				fmt.Println("Available commands:")
				for name, cmd := range cli.commands {
					if name == cmd.Name { // Only show main commands, not aliases
						fmt.Printf("  %-20s - %s\n", name, cmd.Description)
					}
				}
			}
			return nil
		},
	})

	// exit command
	cli.RegisterCommand(Command{
		Name:        "exit",
		Description: "Exit the CLI",
		Usage:       "exit",
		Aliases:     []string{"quit", "q"},
		Handler: func(args []string) error {
			cli.Stop()
			return nil
		},
	})

	// history command
	cli.RegisterCommand(Command{
		Name:        "history",
		Description: "Show command history",
		Usage:       "history [limit]",
		Aliases:     []string{"hist"},
		Handler: func(args []string) error {
			limit := 10
			if len(args) > 0 {
				fmt.Sscanf(args[0], "%d", &limit)
			}

			history := cli.GetHistory()
			if len(history) == 0 {
				fmt.Println("No command history")
				return nil
			}

			start := 0
			if len(history) > limit {
				start = len(history) - limit
			}

			for i := start; i < len(history); i++ {
				fmt.Printf("  %d: %s\n", i+1, history[i])
			}
			return nil
		},
	})

	// clear command
	cli.RegisterCommand(Command{
		Name:        "clear",
		Description: "Clear the screen",
		Usage:       "clear",
		Aliases:     []string{"cls"},
		Handler: func(args []string) error {
			fmt.Print("\033[H\033[2J")
			return nil
		},
	})

	// status command
	cli.RegisterCommand(Command{
		Name:        "status",
		Description: "Show system status",
		Usage:       "status",
		Handler: func(args []string) error {
			fmt.Println("📊 O2OChat System Status:")
			fmt.Println("  Version: v3.0.0-beta")
			fmt.Println("  Phase: Phase 2 - Function Enhancement")
			fmt.Println("  Status: Running")
			fmt.Println("  CLI: Interactive Mode")
			return nil
		},
	})
}

// Run runs the interactive CLI with default commands
func (cli *InteractiveCLI) Run() error {
	cli.RegisterDefaultCommands()
	return cli.Start()
}

// Main function for interactive CLI
func main() {
	cli := NewInteractiveCLI()
	if err := cli.Run(); err != nil {
		fmt.Printf("❌ CLI error: %v\n", err)
		os.Exit(1)
	}
}
