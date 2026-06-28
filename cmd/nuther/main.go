package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"nuther/internal/config"
	"nuther/internal/platform"
	"nuther/internal/smartwatch"
	"nuther/internal/ui"
	"nuther/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

// version is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=1.0.0" ./cmd/nuther
var version = "dev"

func main() {
	// Propagate build-time version to UI header
	components.AppVersion = version

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

	if len(os.Args) > 1 && (os.Args[1] == "watch-smart" || os.Args[1] == "watch") {
		if err := runSmartWatcher(os.Args[1], os.Args[2:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return
			}
			fmt.Fprintf(os.Stderr, "watch-smart: %v\n", err)
			os.Exit(1)
		}
		return
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

func runSmartWatcher(name string, args []string) error {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	defaultStore, err := defaultSmartSnapshotStore()
	if err != nil {
		return err
	}

	interval := flags.Duration("interval", 30*time.Second, "device detection polling interval")
	snapshotInterval := flags.Duration("snapshot-interval", 0, "periodic snapshot interval; disabled when 0")
	storePath := flags.String("store", defaultStore, "snapshot store directory")
	listenAddr := flags.String("listen", "127.0.0.1:0", "HTTP listen address for events and snapshot retrieval")
	webhookURL := flags.String("webhook-url", "", "optional outbound webhook URL for snapshot events")
	once := flags.Bool("once", false, "take one manual snapshot of detected drives and exit")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *interval <= 0 {
		return fmt.Errorf("--interval must be greater than 0")
	}
	if *snapshotInterval < 0 {
		return fmt.Errorf("--snapshot-interval cannot be negative")
	}
	if *webhookURL != "" {
		parsed, err := url.Parse(*webhookURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("--webhook-url must be an absolute URL")
		}
	}

	store := smartwatch.NewStore(*storePath)
	hub := smartwatch.NewEventHub(*webhookURL)
	server := smartwatch.NewServer(store, hub)

	ln, baseURL, err := smartwatch.Listen(*listenAddr)
	if err != nil {
		return err
	}
	httpServer := &http.Server{Handler: server.Handler()}
	serverErr := make(chan error, 1)
	go func() {
		err := httpServer.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	fmt.Printf("SMART snapshot store: %s\n", *storePath)
	fmt.Printf("Event subscription URL: %s/events\n", baseURL)
	fmt.Printf("Snapshot index URL: %s/snapshots\n", baseURL)
	if *webhookURL != "" {
		fmt.Printf("Webhook delivery: %s\n", redactURL(*webhookURL))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	watcher := smartwatch.NewWatcher(smartwatch.WatcherConfig{
		Store:            store,
		Hub:              hub,
		PollInterval:     *interval,
		SnapshotInterval: *snapshotInterval,
	})

	if *once {
		records, err := watcher.SnapshotOnce(ctx, smartwatch.ReasonManual)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		<-serverErr
		if err != nil {
			return err
		}
		fmt.Printf("Snapshots written: %d\n", len(records))
		for _, record := range records {
			fmt.Printf("- %s %s %s\n", record.ID, record.Device.Device, record.Device.HealthStatus)
		}
		return nil
	}

	runErr := make(chan error, 1)
	go func() {
		runErr <- watcher.Run(ctx)
	}()

	select {
	case err := <-serverErr:
		stop()
		if err != nil {
			return err
		}
		return nil
	case err := <-runErr:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		<-serverErr
		if err == context.Canceled {
			return nil
		}
		return err
	}
}

func defaultSmartSnapshotStore() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "nuther", "smart-snapshots"), nil
}

func redactURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "<invalid>"
	}
	if parsed.User != nil {
		parsed.User = url.User("***")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	if parsed.Path != "" {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) > 1 {
			parsed.Path = "/" + parts[0] + "/..."
		}
	}
	return parsed.String()
}
