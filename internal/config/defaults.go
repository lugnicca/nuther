package config

import "os"

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Theme: "default",
		Colors: GetTheme("default").Colors,
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
		Screenshot: ScreenshotConfig{
			Dir: defaultScreenshotDir(),
		},
	}
}

// defaultScreenshotDir returns the home directory, falling back to the temp dir
func defaultScreenshotDir() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return home
	}
	return os.TempDir()
}
