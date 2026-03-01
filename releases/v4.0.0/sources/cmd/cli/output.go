package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Color represents ANSI color codes
type Color int

const (
	ColorReset Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
	ColorBold
)

// ColorOutput provides colored output functionality
type ColorOutput struct {
	enabled bool
}

// NewColorOutput creates a new color output
func NewColorOutput() *ColorOutput {
	// Check if terminal supports colors
	enabled := os.Getenv("TERM") != "dumb" &&
		(os.Getenv("FORCE_COLOR") != "" || isTerminal())

	return &ColorOutput{
		enabled: enabled,
	}
}

// Colorize adds color to text
func (co *ColorOutput) Colorize(text string, color Color) string {
	if !co.enabled {
		return text
	}

	colorCodes := map[Color]string{
		ColorReset:   "\033[0m",
		ColorRed:     "\033[31m",
		ColorGreen:   "\033[32m",
		ColorYellow:  "\033[33m",
		ColorBlue:    "\033[34m",
		ColorMagenta: "\033[35m",
		ColorCyan:    "\033[36m",
		ColorWhite:   "\033[37m",
		ColorBold:    "\033[1m",
	}

	code, exists := colorCodes[color]
	if !exists {
		return text
	}

	return code + text + colorCodes[ColorReset]
}

// Success prints success message in green
func (co *ColorOutput) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(co.Colorize("✅ "+msg, ColorGreen))
}

// Error prints error message in red
func (co *ColorOutput) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(co.Colorize("❌ "+msg, ColorRed))
}

// Warning prints warning message in yellow
func (co *ColorOutput) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(co.Colorize("⚠️  "+msg, ColorYellow))
}

// Info prints info message in blue
func (co *ColorOutput) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(co.Colorize("ℹ️  "+msg, ColorBlue))
}

// Bold prints bold text
func (co *ColorOutput) Bold(text string) string {
	return co.Colorize(text, ColorBold)
}

// Disable disables color output
func (co *ColorOutput) Disable() {
	co.enabled = false
}

// Enable enables color output
func (co *ColorOutput) Enable() {
	co.enabled = true
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	total     int
	current   int
	width     int
	label     string
	color     *ColorOutput
	startTime time.Time
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, label string) *ProgressBar {
	return &ProgressBar{
		total:     total,
		current:   0,
		width:     50,
		label:     label,
		color:     NewColorOutput(),
		startTime: time.Now(),
	}
}

// Update updates progress bar
func (pb *ProgressBar) Update(current int) {
	pb.current = current
	pb.render()
}

// Increment increments progress by 1
func (pb *ProgressBar) Increment() {
	pb.current++
	pb.render()
}

// Finish finishes progress bar
func (pb *ProgressBar) Finish() {
	pb.current = pb.total
	pb.render()
	fmt.Println()
}

// SetLabel sets progress bar label
func (pb *ProgressBar) SetLabel(label string) {
	pb.label = label
}

// SetWidth sets progress bar width
func (pb *ProgressBar) SetWidth(width int) {
	pb.width = width
}

// render renders the progress bar
func (pb *ProgressBar) render() {
	percentage := float64(pb.current) / float64(pb.total)
	filledWidth := int(percentage * float64(pb.width))
	emptyWidth := pb.width - filledWidth

	// Create bar string
	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", emptyWidth)

	// Calculate elapsed time
	elapsed := time.Since(pb.startTime)
	elapsedStr := formatDuration(elapsed)

	// Calculate ETA
	if pb.current > 0 {
		rate := float64(pb.current) / elapsed.Seconds()
		remaining := float64(pb.total-pb.current) / rate
		etaStr := formatDuration(time.Duration(remaining) * time.Second)

		// Print progress bar
		fmt.Printf("\r%s %s %s/%s (%.1f%%) [%s] ETA: %s  ",
			pb.color.Colorize(pb.label, ColorCyan),
			bar,
			pb.color.Colorize(fmt.Sprintf("%d", pb.current), ColorGreen),
			pb.color.Colorize(fmt.Sprintf("%d", pb.total), ColorWhite),
			percentage*100,
			elapsedStr,
			etaStr,
		)
	} else {
		fmt.Printf("\r%s %s %s/%s (%.1f%%) [%s]  ",
			pb.color.Colorize(pb.label, ColorCyan),
			bar,
			pb.color.Colorize(fmt.Sprintf("%d", pb.current), ColorGreen),
			pb.color.Colorize(fmt.Sprintf("%d", pb.total), ColorWhite),
			percentage*100,
			elapsedStr,
		)
	}
}

// formatDuration formats duration as human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// Global color output instance
var globalColorOutput *ColorOutput

// init initializes global color output
func init() {
	globalColorOutput = NewColorOutput()
}

// Color returns global color output
func Color() *ColorOutput {
	return globalColorOutput
}

// DisableColors disables color output globally
func DisableColors() {
	globalColorOutput.Disable()
}

// EnableColors enables color output globally
func EnableColors() {
	globalColorOutput.Enable()
}

// RegisterColorCommands registers color-related commands
func RegisterColorCommands(cli *InteractiveCLI) {
	// color command
	cli.RegisterCommand(Command{
		Name:        "color",
		Description: "Test color output",
		Usage:       "color",
		Handler: func(args []string) error {
			co := Color()

			fmt.Println(co.Bold("Color Test:"))
			fmt.Println(co.Colorize("  Red text", ColorRed))
			fmt.Println(co.Colorize("  Green text", ColorGreen))
			fmt.Println(co.Colorize("  Yellow text", ColorYellow))
			fmt.Println(co.Colorize("  Blue text", ColorBlue))
			fmt.Println(co.Colorize("  Magenta text", ColorMagenta))
			fmt.Println(co.Colorize("  Cyan text", ColorCyan))
			fmt.Println(co.Colorize("  White text", ColorWhite))

			return nil
		},
	})

	// progress command
	cli.RegisterCommand(Command{
		Name:        "progress",
		Description: "Test progress bar",
		Usage:       "progress [total]",
		Handler: func(args []string) error {
			total := 100
			if len(args) > 0 {
				fmt.Sscanf(args[0], "%d", &total)
			}

			pb := NewProgressBar(total, "Processing")

			for i := 0; i <= total; i++ {
				pb.Update(i)
				time.Sleep(50 * time.Millisecond)
			}
			pb.Finish()

			Color().Success("Progress test completed!")

			return nil
		},
	})
}

// Main function for color output demo
func main() {
	// Create interactive CLI
	cli := NewInteractiveCLI()

	// Register color commands
	RegisterColorCommands(cli)

	// Run CLI
	if err := cli.Run(); err != nil {
		fmt.Printf("❌ CLI error: %v\n", err)
		os.Exit(1)
	}
}
