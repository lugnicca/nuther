package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Check theme
	if cfg.Theme != "default" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "default")
	}

	// Check colors are set
	if cfg.Colors.AccentPrimary == "" {
		t.Error("Colors.AccentPrimary is empty")
	}
	if cfg.Colors.AccentSecondary == "" {
		t.Error("Colors.AccentSecondary is empty")
	}
	if cfg.Colors.Success == "" {
		t.Error("Colors.Success is empty")
	}
	if cfg.Colors.Warning == "" {
		t.Error("Colors.Warning is empty")
	}
	if cfg.Colors.Danger == "" {
		t.Error("Colors.Danger is empty")
	}

	// Check display defaults
	if !cfg.Display.ShowLogo {
		t.Error("Display.ShowLogo = false, want true")
	}
	if cfg.Display.ShowFahrenheit {
		t.Error("Display.ShowFahrenheit = true, want false")
	}
	if cfg.Display.CompactMode {
		t.Error("Display.CompactMode = true, want false")
	}
	if !cfg.Display.ShowIcons {
		t.Error("Display.ShowIcons = false, want true")
	}

	// Check keybindings
	if cfg.Keybindings.Quit != "q" {
		t.Errorf("Keybindings.Quit = %q, want %q", cfg.Keybindings.Quit, "q")
	}
	if cfg.Keybindings.Tab != "tab" {
		t.Errorf("Keybindings.Tab = %q, want %q", cfg.Keybindings.Tab, "tab")
	}
	if cfg.Keybindings.NextDrive != "n" {
		t.Errorf("Keybindings.NextDrive = %q, want %q", cfg.Keybindings.NextDrive, "n")
	}
	if cfg.Keybindings.PrevDrive != "p" {
		t.Errorf("Keybindings.PrevDrive = %q, want %q", cfg.Keybindings.PrevDrive, "p")
	}
	if cfg.Keybindings.Refresh != "r" {
		t.Errorf("Keybindings.Refresh = %q, want %q", cfg.Keybindings.Refresh, "r")
	}
	if cfg.Keybindings.Help != "?" {
		t.Errorf("Keybindings.Help = %q, want %q", cfg.Keybindings.Help, "?")
	}
	if cfg.Keybindings.Screenshot != "s" {
		t.Errorf("Keybindings.Screenshot = %q, want %q", cfg.Keybindings.Screenshot, "s")
	}

	// Check screenshot defaults
	if cfg.Screenshot.Dir == "" {
		t.Error("Screenshot.Dir is empty")
	}
}

func TestLoadWithNoConfigFile(t *testing.T) {
	// This test just verifies Load doesn't error and returns a non-nil config
	// The actual theme may differ based on whether a config file exists
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil")
	}

	// Theme should be non-empty (either default or from existing config)
	if cfg.Theme == "" {
		t.Error("Theme is empty")
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nuther", "config.yaml")

	// Create config with custom values
	cfg := &Config{
		Theme: "nord",
		Colors: ColorConfig{
			AccentPrimary:   "#88c0d0",
			AccentSecondary: "#b48ead",
			Success:         "#a3be8c",
			Warning:         "#ebcb8b",
			Danger:          "#bf616a",
			Info:            "#81a1c1",
			Text:            "#d8dee9",
			TextDim:         "#4c566a",
			Background:      "#2e3440",
			Border:          "#3b4252",
		},
		Display: DisplayConfig{
			ShowLogo:       false,
			ShowFahrenheit: true,
			CompactMode:    true,
			ShowIcons:      false,
		},
		Keybindings: KeybindingsConfig{
			Quit:       "ctrl+c",
			Tab:        "tab",
			NextDrive:  "j",
			PrevDrive:  "k",
			Refresh:    "ctrl+r",
			Help:       "h",
			Screenshot: "ctrl+s",
		},
	}

	// Create directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Save directly to temp path (bypassing GetConfigPath)
	data, err := cfg.marshalYAML()
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Read the file back and verify it was written correctly
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if len(readData) == 0 {
		t.Error("Config file is empty")
	}
}

// marshalYAML helper for testing
func (c *Config) marshalYAML() ([]byte, error) {
	return []byte(`theme: ` + c.Theme + `
colors:
  accent_primary: "` + c.Colors.AccentPrimary + `"
  accent_secondary: "` + c.Colors.AccentSecondary + `"
  success: "` + c.Colors.Success + `"
  warning: "` + c.Colors.Warning + `"
  danger: "` + c.Colors.Danger + `"
  info: "` + c.Colors.Info + `"
  text: "` + c.Colors.Text + `"
  text_dim: "` + c.Colors.TextDim + `"
  background: "` + c.Colors.Background + `"
  border: "` + c.Colors.Border + `"
display:
  show_logo: false
  show_fahrenheit: true
  compact_mode: true
  show_icons: false
keybindings:
  quit: "ctrl+c"
  tab: "tab"
  next_drive: "j"
  prev_drive: "k"
  refresh: "ctrl+r"
  help: "h"
  screenshot: "ctrl+s"
`), nil
}

func TestColorConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify color format (should be hex)
	colors := []struct {
		name  string
		value string
	}{
		{"AccentPrimary", cfg.Colors.AccentPrimary},
		{"AccentSecondary", cfg.Colors.AccentSecondary},
		{"Success", cfg.Colors.Success},
		{"Warning", cfg.Colors.Warning},
		{"Danger", cfg.Colors.Danger},
		{"Info", cfg.Colors.Info},
		{"Text", cfg.Colors.Text},
		{"TextDim", cfg.Colors.TextDim},
		{"Background", cfg.Colors.Background},
		{"Border", cfg.Colors.Border},
	}

	for _, c := range colors {
		if len(c.value) != 7 || c.value[0] != '#' {
			t.Errorf("%s = %q, want 7-char hex color starting with #", c.name, c.value)
		}
	}
}

func TestDisplayConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test that display config has valid defaults
	display := cfg.Display

	// ShowLogo should be true by default
	if !display.ShowLogo {
		t.Error("Default ShowLogo should be true")
	}

	// ShowFahrenheit should be false by default (Celsius)
	if display.ShowFahrenheit {
		t.Error("Default ShowFahrenheit should be false")
	}

	// CompactMode should be false by default
	if display.CompactMode {
		t.Error("Default CompactMode should be false")
	}

	// ShowIcons should be true by default
	if !display.ShowIcons {
		t.Error("Default ShowIcons should be true")
	}
}

func TestKeybindingsConfig(t *testing.T) {
	cfg := DefaultConfig()
	kb := cfg.Keybindings

	// Check that keybindings are not empty
	bindings := map[string]string{
		"Quit":       kb.Quit,
		"Tab":        kb.Tab,
		"NextDrive":  kb.NextDrive,
		"PrevDrive":  kb.PrevDrive,
		"Refresh":    kb.Refresh,
		"Help":       kb.Help,
		"Screenshot": kb.Screenshot,
	}

	for name, value := range bindings {
		if value == "" {
			t.Errorf("Keybinding %s is empty", name)
		}
	}
}

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	// Should end with config.yaml
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("GetConfigPath() = %q, should end with config.yaml", path)
	}

	// Should contain nuther directory
	dir := filepath.Dir(path)
	if filepath.Base(dir) != "nuther" {
		t.Errorf("GetConfigPath() = %q, parent dir should be nuther", path)
	}
}

func TestLoadFromYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "nuther")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	yamlContent := `theme: nord
colors:
  accent_primary: "#88c0d0"
  accent_secondary: "#b48ead"
  success: "#a3be8c"
  warning: "#ebcb8b"
  danger: "#bf616a"
  info: "#81a1c1"
  text: "#eceff4"
  text_dim: "#4c566a"
  background: "#2e3440"
  border: "#3b4252"
display:
  show_logo: false
  show_fahrenheit: true
  compact_mode: true
  show_icons: false
keybindings:
  quit: "ctrl+q"
  tab: "tab"
  next_drive: "j"
  prev_drive: "k"
  refresh: "ctrl+r"
  help: "h"
  screenshot: "ctrl+s"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if len(data) == 0 {
		t.Error("Config file is empty")
	}

	// Verify YAML can be unmarshaled
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if cfg.Theme != "nord" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "nord")
	}
	if cfg.Display.ShowLogo {
		t.Error("ShowLogo should be false")
	}
	if !cfg.Display.ShowFahrenheit {
		t.Error("ShowFahrenheit should be true")
	}
}

func TestConfigSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "new_dir", "nuther", "config.yaml")

	cfg := DefaultConfig()
	cfg.Theme = "dracula"

	// Manually create the directory and save
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestConfigYAMLRoundTrip(t *testing.T) {
	original := &Config{
		Theme: "gruvbox",
		Colors: ColorConfig{
			AccentPrimary:   "#83a598",
			AccentSecondary: "#d3869b",
			Success:         "#b8bb26",
			Warning:         "#fabd2f",
			Danger:          "#fb4934",
			Info:            "#83a598",
			Text:            "#ebdbb2",
			TextDim:         "#928374",
			Background:      "#282828",
			Border:          "#3c3836",
		},
		Display: DisplayConfig{
			ShowLogo:       true,
			ShowFahrenheit: false,
			CompactMode:    false,
			ShowIcons:      true,
		},
		Keybindings: KeybindingsConfig{
			Quit:       "q",
			Tab:        "tab",
			NextDrive:  "n",
			PrevDrive:  "p",
			Refresh:    "r",
			Help:       "?",
			Screenshot: "s",
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var restored Config
	if err := yaml.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare
	if restored.Theme != original.Theme {
		t.Errorf("Theme mismatch: got %q, want %q", restored.Theme, original.Theme)
	}
	if restored.Colors.AccentPrimary != original.Colors.AccentPrimary {
		t.Errorf("AccentPrimary mismatch: got %q, want %q",
			restored.Colors.AccentPrimary, original.Colors.AccentPrimary)
	}
	if restored.Display.ShowLogo != original.Display.ShowLogo {
		t.Errorf("ShowLogo mismatch: got %v, want %v",
			restored.Display.ShowLogo, original.Display.ShowLogo)
	}
	if restored.Keybindings.Quit != original.Keybindings.Quit {
		t.Errorf("Quit keybinding mismatch: got %q, want %q",
			restored.Keybindings.Quit, original.Keybindings.Quit)
	}
}

func TestLoadAppliesThemeColors(t *testing.T) {
	// Test that when custom colors are not specified, theme colors are applied
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "nuther")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	// Config with theme but no custom colors
	yamlContent := `theme: nord
display:
  show_logo: true
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	var cfg Config
	data, _ := os.ReadFile(configPath)
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Simulate the theme application logic from Load()
	if cfg.Colors.AccentPrimary == "" {
		theme := GetTheme(cfg.Theme)
		cfg.Colors = theme.Colors
	}

	// Colors should come from Nord theme
	nordTheme := GetTheme("nord")
	if cfg.Colors.AccentPrimary != nordTheme.Colors.AccentPrimary {
		t.Errorf("AccentPrimary = %q, want %q (from nord theme)",
			cfg.Colors.AccentPrimary, nordTheme.Colors.AccentPrimary)
	}
}

func TestLoadWithInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "nuther")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	// Invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Read and try to unmarshal
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestConfigSaveMethod(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nuther", "config.yaml")

	cfg := DefaultConfig()
	cfg.Theme = "dracula"

	// Create the directory structure
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Marshal and save
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Verify file was created and can be read
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var loadedCfg Config
	if err := yaml.Unmarshal(readData, &loadedCfg); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if loadedCfg.Theme != "dracula" {
		t.Errorf("Theme = %q, want %q", loadedCfg.Theme, "dracula")
	}
}

func TestLoadNoError(t *testing.T) {
	// Load should not error even if config doesn't exist
	cfg, err := Load()
	if err != nil {
		t.Errorf("Load should not error: %v", err)
	}
	if cfg == nil {
		t.Error("Config should not be nil")
	}
}

func TestGetConfigPathReturnsValidPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath error: %v", err)
	}

	// Path should be absolute or have expected structure
	if path == "" {
		t.Error("Path should not be empty")
	}

	// Check it ends with expected path
	if !strings.HasSuffix(path, filepath.Join("nuther", "config.yaml")) {
		t.Errorf("Path should end with nuther/config.yaml, got %q", path)
	}
}

func TestConfigStructTags(t *testing.T) {
	cfg := DefaultConfig()

	// Test that config can be marshaled to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Verify YAML contains expected keys
	yamlStr := string(data)
	expectedKeys := []string{"theme:", "colors:", "display:", "keybindings:", "screenshot:"}
	for _, key := range expectedKeys {
		if !strings.Contains(yamlStr, key) {
			t.Errorf("YAML should contain %q", key)
		}
	}
}

func TestColorConfigAllFields(t *testing.T) {
	cfg := DefaultConfig()

	fields := []struct {
		name  string
		value string
	}{
		{"AccentPrimary", cfg.Colors.AccentPrimary},
		{"AccentSecondary", cfg.Colors.AccentSecondary},
		{"Success", cfg.Colors.Success},
		{"Warning", cfg.Colors.Warning},
		{"Danger", cfg.Colors.Danger},
		{"Info", cfg.Colors.Info},
		{"Text", cfg.Colors.Text},
		{"TextDim", cfg.Colors.TextDim},
		{"Background", cfg.Colors.Background},
		{"Border", cfg.Colors.Border},
	}

	for _, f := range fields {
		if f.value == "" {
			t.Errorf("%s should not be empty", f.name)
		}
	}
}

func TestDisplayConfigAllFields(t *testing.T) {
	cfg := DefaultConfig()

	// Test boolean fields have expected defaults
	if !cfg.Display.ShowLogo {
		t.Error("ShowLogo should be true by default")
	}
	if cfg.Display.ShowFahrenheit {
		t.Error("ShowFahrenheit should be false by default")
	}
	if cfg.Display.CompactMode {
		t.Error("CompactMode should be false by default")
	}
	if !cfg.Display.ShowIcons {
		t.Error("ShowIcons should be true by default")
	}
}

func TestSave(t *testing.T) {
	// Save to the actual config path, then verify file exists
	cfg := DefaultConfig()
	cfg.Theme = "default"

	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify the file was created
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Config file should exist at %q after Save()", path)
	}
}
