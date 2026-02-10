package config

import (
	"testing"
)

func TestDefaultConfigNotNil(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
}

func TestDefaultConfigTheme(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Theme != "default" {
		t.Errorf("Default Theme = %q, want %q", cfg.Theme, "default")
	}
}

func TestDefaultConfigColorsMatchDefaultTheme(t *testing.T) {
	cfg := DefaultConfig()
	defaultTheme := GetTheme("default")

	if cfg.Colors.AccentPrimary != defaultTheme.Colors.AccentPrimary {
		t.Errorf("AccentPrimary mismatch: config=%q, theme=%q",
			cfg.Colors.AccentPrimary, defaultTheme.Colors.AccentPrimary)
	}
	if cfg.Colors.AccentSecondary != defaultTheme.Colors.AccentSecondary {
		t.Errorf("AccentSecondary mismatch: config=%q, theme=%q",
			cfg.Colors.AccentSecondary, defaultTheme.Colors.AccentSecondary)
	}
	if cfg.Colors.Success != defaultTheme.Colors.Success {
		t.Errorf("Success mismatch: config=%q, theme=%q",
			cfg.Colors.Success, defaultTheme.Colors.Success)
	}
	if cfg.Colors.Warning != defaultTheme.Colors.Warning {
		t.Errorf("Warning mismatch: config=%q, theme=%q",
			cfg.Colors.Warning, defaultTheme.Colors.Warning)
	}
	if cfg.Colors.Danger != defaultTheme.Colors.Danger {
		t.Errorf("Danger mismatch: config=%q, theme=%q",
			cfg.Colors.Danger, defaultTheme.Colors.Danger)
	}
}

func TestDefaultConfigDisplaySettings(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Display.ShowLogo {
		t.Error("Default ShowLogo should be true")
	}
	if cfg.Display.ShowFahrenheit {
		t.Error("Default ShowFahrenheit should be false")
	}
	if cfg.Display.CompactMode {
		t.Error("Default CompactMode should be false")
	}
	if !cfg.Display.ShowIcons {
		t.Error("Default ShowIcons should be true")
	}
}

func TestDefaultConfigKeybindings(t *testing.T) {
	cfg := DefaultConfig()
	kb := cfg.Keybindings

	expectedBindings := map[string]string{
		"Quit":       "q",
		"Tab":        "tab",
		"NextDrive":  "n",
		"PrevDrive":  "p",
		"Refresh":    "r",
		"Help":       "?",
		"Screenshot": "s",
	}

	actualBindings := map[string]string{
		"Quit":       kb.Quit,
		"Tab":        kb.Tab,
		"NextDrive":  kb.NextDrive,
		"PrevDrive":  kb.PrevDrive,
		"Refresh":    kb.Refresh,
		"Help":       kb.Help,
		"Screenshot": kb.Screenshot,
	}

	for name, expected := range expectedBindings {
		if actual := actualBindings[name]; actual != expected {
			t.Errorf("Keybinding %s = %q, want %q", name, actual, expected)
		}
	}
}

func TestDefaultConfigIsCopyable(t *testing.T) {
	cfg1 := DefaultConfig()
	cfg2 := DefaultConfig()

	// Modify cfg1
	cfg1.Theme = "nord"
	cfg1.Display.ShowLogo = false

	// cfg2 should be unaffected
	if cfg2.Theme != "default" {
		t.Error("DefaultConfig() should return independent copies")
	}
	if !cfg2.Display.ShowLogo {
		t.Error("DefaultConfig() should return independent copies")
	}
}
