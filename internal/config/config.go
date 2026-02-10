package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Theme       string            `yaml:"theme"`
	Colors      ColorConfig       `yaml:"colors"`
	Display     DisplayConfig     `yaml:"display"`
	Keybindings KeybindingsConfig `yaml:"keybindings"`
}

// ColorConfig allows custom color overrides
type ColorConfig struct {
	AccentPrimary      string `yaml:"accent_primary"`
	AccentSecondary    string `yaml:"accent_secondary"`
	Success            string `yaml:"success"`
	Warning            string `yaml:"warning"`
	Danger             string `yaml:"danger"`
	Info               string `yaml:"info"`
	Text               string `yaml:"text"`
	TextDim            string `yaml:"text_dim"`
	Background         string `yaml:"background"`
	Border             string `yaml:"border"`
	BadgeForeground    string `yaml:"badge_foreground"`
	BadgeForegroundAlt string `yaml:"badge_foreground_alt"`
	SurfaceAlt         string `yaml:"surface_alt"`
	SurfaceHighlight   string `yaml:"surface_highlight"`
	TempHot            string `yaml:"temp_hot"`
}

// DisplayConfig controls visual display options
type DisplayConfig struct {
	ShowLogo       bool `yaml:"show_logo"`
	ShowFahrenheit bool `yaml:"show_fahrenheit"`
	CompactMode    bool `yaml:"compact_mode"`
	ShowIcons      bool `yaml:"show_icons"`
}

// KeybindingsConfig allows custom key bindings
type KeybindingsConfig struct {
	Quit       string `yaml:"quit"`
	Tab        string `yaml:"tab"`
	NextDrive  string `yaml:"next_drive"`
	PrevDrive  string `yaml:"prev_drive"`
	Refresh    string `yaml:"refresh"`
	Help       string `yaml:"help"`
	Screenshot string `yaml:"screenshot"`
}

// GetConfigPath returns the config file path
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}

	return filepath.Join(configDir, "nuther", "config.yaml"), nil
}

// Load loads configuration from file, returning defaults if not found
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Load user themes before resolving theme selection
	LoadUserThemes()

	path, err := GetConfigPath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Merge theme colors field-by-field (preserves user overrides, fills gaps)
	theme := GetTheme(cfg.Theme)
	MergeWithThemeColors(&cfg.Colors, theme.Colors)

	return cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
