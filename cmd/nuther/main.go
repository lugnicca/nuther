package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"nuther/internal/config"
	"nuther/internal/platform"
	"nuther/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// version is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=1.0.0" ./cmd/nuther
var version = "dev"

func main() {
	// Handle --version / -v
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Printf("nuther %s\n", version)
			os.Exit(0)
		}
	}

	// Initialize logging
	if logFile := initLogging(); logFile != nil {
		defer logFile.Close()
	}

	// Check for privileges
	if !platform.IsElevated() {
		printStartupWarning()
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Create model
	model := ui.NewModel(cfg)

	// Create program with alt screen
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// initLogging sets up slog to write to a log file, or stderr if --debug is passed.
// Returns the log file handle if one was opened, so the caller can defer Close().
func initLogging() *os.File {
	debug := false
	for _, arg := range os.Args[1:] {
		if arg == "--debug" {
			debug = true
			break
		}
	}

	if debug {
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
		slog.SetDefault(slog.New(handler))
		return nil
	}

	// Log to file at ~/.config/nuther/nuther.log
	logPath := getLogPath()
	if logPath == "" {
		// Cannot determine log path, discard logs
		handler := slog.NewTextHandler(io.Discard, nil)
		slog.SetDefault(slog.New(handler))
		return nil
	}

	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		handler := slog.NewTextHandler(io.Discard, nil)
		slog.SetDefault(slog.New(handler))
		return nil
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		handler := slog.NewTextHandler(io.Discard, nil)
		slog.SetDefault(slog.New(handler))
		return nil
	}

	handler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))
	return f
}

func getLogPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "nuther", "nuther.log")
}

func printStartupWarning() {
	fmt.Println()
	fmt.Println("  ⚠ WARNING: Running without root privileges")
	fmt.Println("     S.M.A.R.T. data requires root access.")
	fmt.Println("     Run with: sudo nuther")
	fmt.Println()
	fmt.Println("     Demo mode will be activated with sample data.")
	fmt.Println()
	fmt.Println("     Press Enter to continue...")

	var input string
	_, _ = fmt.Scanln(&input)
}
