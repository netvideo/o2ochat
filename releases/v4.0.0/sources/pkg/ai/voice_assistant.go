package ai

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// VoiceAssistant represents an AI voice assistant
type VoiceAssistant struct {
	commands    map[string]CommandHandler
	isListening bool
	language    string
}

// CommandHandler represents a command handler function
type CommandHandler func(ctx context.Context, args []string) (string, error)

// VoiceCommand represents a voice command
type VoiceCommand struct {
	Name        string
	TriggerWords []string
	Handler     CommandHandler
	Description string
}

// NewVoiceAssistant creates a new voice assistant
func NewVoiceAssistant() *VoiceAssistant {
	va := &VoiceAssistant{
		commands:    make(map[string]CommandHandler),
		isListening: false,
		language:    "en",
	}

	// Register default commands
	va.registerDefaultCommands()

	return va
}

// registerDefaultCommands registers default voice commands
func (va *VoiceAssistant) registerDefaultCommands() {
	// Send message command
	va.commands["send_message"] = func(ctx context.Context, args []string) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("need recipient and message")
		}
		recipient := args[0]
		message := strings.Join(args[1:], " ")
		return fmt.Sprintf("Sending message to %s: %s", recipient, message), nil
	}

	// Call command
	va.commands["call"] = func(ctx context.Context, args []string) (string, error) {
		if len(args) < 1 {
			return "", fmt.Errorf("need recipient name")
		}
		recipient := args[0]
		return fmt.Sprintf("Calling %s...", recipient), nil
	}

	// Search command
	va.commands["search"] = func(ctx context.Context, args []string) (string, error) {
		if len(args) < 1 {
			return "", fmt.Errorf("need search query")
		}
		query := strings.Join(args, " ")
		return fmt.Sprintf("Searching for: %s", query), nil
	}

	// Translate command
	va.commands["translate"] = func(ctx context.Context, args []string) (string, error) {
		if len(args) < 3 {
			return "", fmt.Errorf("need text, from language, and to language")
		}
		text := args[0]
		fromLang := args[1]
		toLang := args[2]
		return fmt.Sprintf("Translating '%s' from %s to %s", text, fromLang, toLang), nil
	}

	// Remind command
	va.commands["remind"] = func(ctx context.Context, args []string) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("need time and reminder text")
		}
		reminderTime := args[0]
		reminder := strings.Join(args[1:], " ")
		return fmt.Sprintf("Setting reminder for %s: %s", reminderTime, reminder), nil
	}

	// Weather command
	va.commands["weather"] = func(ctx context.Context, args []string) (string, error) {
		location := "your location"
		if len(args) > 0 {
			location = args[0]
		}
		return fmt.Sprintf("Weather in %s: Sunny, 25°C", location), nil
	}

	// Time command
	va.commands["time"] = func(ctx context.Context, args []string) (string, error) {
		return fmt.Sprintf("Current time is %s", time.Now().Format("3:04 PM")), nil
	}

	// Help command
	va.commands["help"] = func(ctx context.Context, args []string) (string, error) {
		help := "Available commands:\n"
		help += "- send_message [recipient] [message]\n"
		help += "- call [recipient]\n"
		help += "- search [query]\n"
		help += "- translate [text] [from] [to]\n"
		help += "- remind [time] [reminder]\n"
		help += "- weather [location]\n"
		help += "- time\n"
		return help, nil
	}
}

// ProcessVoice processes voice input
func (va *VoiceAssistant) ProcessVoice(ctx context.Context, voiceInput string) (string, error) {
	// Convert voice to text (simplified - would use speech-to-text API)
	text := strings.ToLower(voiceInput)

	// Find matching command
	for cmdName, handler := range va.commands {
		triggerWords := va.getTriggerWords(cmdName)
		for _, trigger := range triggerWords {
			if strings.Contains(text, trigger) {
				// Extract arguments
				args := va.extractArguments(text, trigger)
				return handler(ctx, args)
			}
		}
	}

	return "", fmt.Errorf("command not recognized")
}

// getTriggerWords gets trigger words for a command
func (va *VoiceAssistant) getTriggerWords(cmdName string) []string {
	switch cmdName {
	case "send_message":
		return []string{"send", "message", "text"}
	case "call":
		return []string{"call", "phone", "video call"}
	case "search":
		return []string{"search", "find", "look up"}
	case "translate":
		return []string{"translate", "convert"}
	case "remind":
		return []string{"remind", "reminder", "alarm"}
	case "weather":
		return []string{"weather", "forecast"}
	case "time":
		return []string{"time", "clock"}
	case "help":
		return []string{"help", "commands"}
	default:
		return []string{cmdName}
	}
}

// extractArguments extracts arguments from voice input
func (va *VoiceAssistant) extractArguments(input, trigger string) []string {
	// Simple argument extraction (would use NLP in production)
	idx := strings.Index(strings.ToLower(input), trigger)
	if idx == -1 {
		return []string{}
	}

	remaining := input[idx+len(trigger):]
	remaining = strings.TrimSpace(remaining)

	if remaining == "" {
		return []string{}
	}

	// Split by common delimiters
	args := strings.FieldsFunc(remaining, func(r rune) bool {
		return r == ',' || r == ';' || r == ':'
	})

	// Clean up arguments
	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
	}

	return args
}

// StartListening starts voice listening
func (va *VoiceAssistant) StartListening() {
	va.isListening = true
	fmt.Println("Voice assistant is now listening...")
}

// StopListening stops voice listening
func (va *VoiceAssistant) StopListening() {
	va.isListening = false
	fmt.Println("Voice assistant stopped listening")
}

// IsListening checks if assistant is listening
func (va *VoiceAssistant) IsListening() bool {
	return va.isListening
}

// SetLanguage sets assistant language
func (va *VoiceAssistant) SetLanguage(lang string) {
	va.language = lang
}

// GetLanguage gets assistant language
func (va *VoiceAssistant) GetLanguage() string {
	return va.language
}

// AddCommand adds a custom command
func (va *VoiceAssistant) AddCommand(cmd VoiceCommand) {
	va.commands[cmd.Name] = cmd.Handler
}

// RemoveCommand removes a command
func (va *VoiceAssistant) RemoveCommand(cmdName string) {
	delete(va.commands, cmdName)
}

// GetCommands gets all available commands
func (va *VoiceAssistant) GetCommands() []string {
	cmds := make([]string, 0, len(va.commands))
	for cmd := range va.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// TestVoiceAssistant tests the voice assistant
func TestVoiceAssistant() {
	va := NewVoiceAssistant()

	// Test commands
	commands := []string{
		"send message to Alice Hello there!",
		"call Bob",
		"search for restaurants nearby",
		"translate hello to Spanish",
		"remind me tomorrow to buy groceries",
		"weather in New York",
		"what time is it",
		"help",
	}

	ctx := context.Background()
	for _, cmd := range commands {
		result, err := va.ProcessVoice(ctx, cmd)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("✅ %s\n", result)
		}
	}
}
