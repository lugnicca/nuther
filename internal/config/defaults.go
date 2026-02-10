package config

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
	}
}
