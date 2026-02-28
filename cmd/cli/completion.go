package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// CompletionManager manages command auto-completion
type CompletionManager struct {
	commands   []string
	aliases    map[string]string
	history    []string
	maxHistory int
}

// NewCompletionManager creates a new completion manager
func NewCompletionManager() *CompletionManager {
	return &CompletionManager{
		commands:   make([]string, 0),
		aliases:    make(map[string]string),
		history:    make([]string, 0),
		maxHistory: 1000,
	}
}

// RegisterCommand registers a command for completion
func (cm *CompletionManager) RegisterCommand(name string, aliases []string) {
	cm.commands = append(cm.commands, name)
	for _, alias := range aliases {
		cm.aliases[alias] = name
	}
}

// Complete completes a partial command
func (cm *CompletionManager) Complete(partial string) []string {
	if partial == "" {
		return cm.commands
	}

	matches := make([]string, 0)
	partial = strings.ToLower(partial)

	// Search commands
	for _, cmd := range cm.commands {
		if strings.HasPrefix(strings.ToLower(cmd), partial) {
			matches = append(matches, cmd)
		}
	}

	// Search aliases
	for alias, name := range cm.aliases {
		if strings.HasPrefix(strings.ToLower(alias), partial) {
			matches = append(matches, name)
		}
	}

	// Sort and remove duplicates
	sort.Strings(matches)
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			result = append(result, match)
		}
	}

	return result
}

// AddHistory adds a command to history
func (cm *CompletionManager) AddHistory(cmd string) {
	cm.history = append(cm.history, cmd)
	if len(cm.history) > cm.maxHistory {
		cm.history = cm.history[1:]
	}
}

// GetHistory returns command history
func (cm *CompletionManager) GetHistory() []string {
	return cm.history
}

// SetupBashCompletion sets up bash completion
func (cm *CompletionManager) SetupBashCompletion() error {
	completion := `# O2OChat CLI bash completion
_o2ochat_completion() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=($(compgen -W "$1" -- "$cur"))
}
complete -F _o2ochat_completion o2ochat
`

	// Write to bash completion directory
	completionDir := "/etc/bash_completion.d"
	if _, err := os.Stat(completionDir); err == nil {
		return os.WriteFile(completionDir+"/o2ochat", []byte(completion), 0644)
	}

	return nil
}

// SetupZshCompletion sets up zsh completion
func (cm *CompletionManager) SetupZshCompletion() error {
	completion := `# O2OChat CLI zsh completion
#compdef o2ochat

_o2ochat() {
    local -a cmd
    cmd=(
        'help:Show help information'
        'exit:Exit the CLI'
        'history:Show command history'
        'clear:Clear the screen'
        'status:Show system status'
    )
    _describe 'commands' cmd
}

_o2ochat "$@"
`

	// Write to zsh completion directory
	completionDir := "/usr/local/share/zsh/site-functions"
	if _, err := os.Stat(completionDir); err == nil {
		return os.WriteFile(completionDir+"/_o2ochat", []byte(completion), 0644)
	}

	return nil
}

// GenerateCompletionScript generates a completion script
func (cm *CompletionManager) GenerateCompletionScript(shell string) string {
	switch shell {
	case "bash":
		return cm.generateBashScript()
	case "zsh":
		return cm.generateZshScript()
	case "fish":
		return cm.generateFishScript()
	default:
		return "# Unsupported shell: " + shell
	}
}

func (cm *CompletionManager) generateBashScript() string {
	commands := strings.Join(cm.commands, " ")
	return fmt.Sprintf(`# O2OChat CLI bash completion script
# Source this file in your .bashrc

_o2ochat_completion() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local commands="%s"
    COMPREPLY=($(compgen -W "$commands" -- "$cur"))
}

complete -F _o2ochat_completion o2ochat
`, commands)
}

func (cm *CompletionManager) generateZshScript() string {
	var cmdList strings.Builder
	for _, cmd := range cm.commands {
		cmdList.WriteString(fmt.Sprintf("        '%s:O2OChat command'\n", cmd))
	}

	return fmt.Sprintf(`# O2OChat CLI zsh completion script
# Source this file in your .zshrc

#compdef o2ochat

_o2ochat() {
    local -a commands
    commands=(
%s
    )
    _describe 'commands' commands
}

_o2ochat "$@"
`, cmdList.String())
}

func (cm *CompletionManager) generateFishScript() string {
	var completions strings.Builder
	for _, cmd := range cm.commands {
		completions.WriteString(fmt.Sprintf("complete -c o2ochat -n '__fish_use_subcommand' -a %s\n", cmd))
	}

	return fmt.Sprintf(`# O2OChat CLI fish completion script
# Source this file in your fish config

%s
`, completions.String())
}

// ShowCompletions shows possible completions for a partial command
func (cm *CompletionManager) ShowCompletions(partial string) {
	matches := cm.Complete(partial)

	if len(matches) == 0 {
		fmt.Println("No matching commands")
		return
	}

	if len(matches) == 1 {
		fmt.Printf("Command: %s\n", matches[0])
		return
	}

	fmt.Println("Possible commands:")
	for _, match := range matches {
		fmt.Printf("  %s\n", match)
	}
}

// InstallCompletion installs completion for current shell
func (cm *CompletionManager) InstallCompletion() error {
	shell := os.Getenv("SHELL")

	if strings.Contains(shell, "bash") {
		return cm.SetupBashCompletion()
	} else if strings.Contains(shell, "zsh") {
		return cm.SetupZshCompletion()
	} else if strings.Contains(shell, "fish") {
		// Fish completion is typically done via fish_completions
		fmt.Println("For fish shell, please source the generated script")
		fmt.Println(cm.GenerateCompletionScript("fish"))
		return nil
	}

	return fmt.Errorf("unsupported shell: %s", shell)
}

// SetupCompletion sets up completion for the CLI
func SetupCompletion(cli *InteractiveCLI) *CompletionManager {
	cm := NewCompletionManager()
	
	// Register CLI commands
	for name, cmd := range cli.commands {
		aliases := cmd.Aliases
		if aliases == nil {
			aliases = []string{}
		}
		cm.RegisterCommand(name, aliases)
	}
	
	return cm
}

// GenerateCompletionScript generates completion script for shell
func GenerateCompletionScript() string {
	return BashCompletionScript()
}

// BashCompletionScript returns bash completion script
func BashCompletionScript() string {
	return `# O2OChat CLI bash completion
_o2ochat_completion() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=( $(compgen -W "help exit history clear status config diagnostic" -- "$cur") )
}
complete -F _o2ochat_completion o2ochat
`
}

		cm.RegisterCommand(name, aliases)
	}
	
	return cm
}

// GenerateCompletionScript generates completion script for shell
func GenerateCompletionScript() string {
	return BashCompletionScript()
}

// BashCompletionScript returns bash completion script
func BashCompletionScript() string {
	return `# O2OChat CLI bash completion
_o2ochat_completion() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=( $(compgen -W "help exit history clear status config diagnostic" -- "$cur") )
}
complete -F _o2ochat_completion o2ochat
`
}


	cm := NewCompletionManager()

	// Register default commands
	cm.RegisterCommand("help", []string{"h"})
	cm.RegisterCommand("exit", []string{"quit", "q"})
	cm.RegisterCommand("history", []string{"hist"})
	cm.RegisterCommand("clear", []string{"cls"})
	cm.RegisterCommand("status", []string{"stat"})
	cm.RegisterCommand("config", []string{"cfg"})
	cm.RegisterCommand("diagnostic", []string{"diag"})

	switch os.Args[1] {
	case "install":
		if err := cm.InstallCompletion(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Shell completion installed successfully")
	case "bash":
		fmt.Println(cm.GenerateCompletionScript("bash"))
	case "zsh":
		fmt.Println(cm.GenerateCompletionScript("zsh"))
	case "fish":
		fmt.Println(cm.GenerateCompletionScript("fish"))
	case "complete":
		partial := ""
		if len(os.Args) > 2 {
			partial = os.Args[2]
		}
		cm.ShowCompletions(partial)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
