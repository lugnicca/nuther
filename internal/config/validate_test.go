package config

import (
	"testing"
)

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		color string
		valid bool
	}{
		{"#000000", true},
		{"#FFFFFF", true},
		{"#ffffff", true},
		{"#00d7ff", true},
		{"#ABC123", true},
		{"#abc123", true},
		{"000000", false},
		{"#00000", false},
		{"#0000000", false},
		{"#GGGGGG", false},
		{"", false},
		{"red", false},
		{"rgb(0,0,0)", false},
		{"#FFF", false},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := isValidHexColor(tt.color)
			if result != tt.valid {
				t.Errorf("isValidHexColor(%q) = %v, want %v", tt.color, result, tt.valid)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	// Valid config
	cfg := DefaultConfig()
	errs := cfg.Validate()
	if errs.HasErrors() {
		t.Errorf("Default config should be valid, got errors: %v", errs)
	}

	// Invalid theme
	cfg = DefaultConfig()
	cfg.Theme = "nonexistent"
	errs = cfg.Validate()
	if !errs.HasErrors() {
		t.Error("Config with invalid theme should have errors")
	}

	// Invalid color
	cfg = DefaultConfig()
	cfg.Colors.AccentPrimary = "not-a-color"
	errs = cfg.Validate()
	if !errs.HasErrors() {
		t.Error("Config with invalid color should have errors")
	}
}

func TestConfigValidateAndFix(t *testing.T) {
	// Invalid theme gets fixed
	cfg := &Config{Theme: "nonexistent"}
	cfg.ValidateAndFix()
	if cfg.Theme != "default" {
		t.Errorf("Theme should be fixed to default, got %q", cfg.Theme)
	}

	// Empty theme gets fixed
	cfg = &Config{}
	cfg.ValidateAndFix()
	if cfg.Theme != "default" {
		t.Errorf("Empty theme should be fixed to default, got %q", cfg.Theme)
	}

	// Colors get applied from theme
	cfg = &Config{Theme: "nord"}
	cfg.ValidateAndFix()
	nordTheme := GetTheme("nord")
	if cfg.Colors.AccentPrimary != nordTheme.Colors.AccentPrimary {
		t.Errorf("Colors should be applied from nord theme")
	}

	// Invalid color gets replaced
	cfg = DefaultConfig()
	cfg.Colors.AccentPrimary = "invalid"
	cfg.ValidateAndFix()
	if cfg.Colors.AccentPrimary == "invalid" {
		t.Error("Invalid color should be replaced")
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{Field: "theme", Message: "unknown theme"}
	expected := "theme: unknown theme"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestValidationErrors(t *testing.T) {
	// Empty errors
	var errs ValidationErrors
	if errs.HasErrors() {
		t.Error("Empty ValidationErrors should not have errors")
	}
	if errs.Error() != "" {
		t.Error("Empty ValidationErrors should return empty string")
	}

	// With errors
	errs = append(errs, ValidationError{Field: "a", Message: "error1"})
	errs = append(errs, ValidationError{Field: "b", Message: "error2"})

	if !errs.HasErrors() {
		t.Error("ValidationErrors with items should have errors")
	}

	errorStr := errs.Error()
	if errorStr == "" {
		t.Error("ValidationErrors.Error() should not be empty")
	}
}

func TestValidateAllThemes(t *testing.T) {
	themes := ListThemes()
	for _, themeName := range themes {
		cfg := DefaultConfig()
		cfg.Theme = themeName
		theme := GetTheme(themeName)
		cfg.Colors = theme.Colors

		errs := cfg.Validate()
		if errs.HasErrors() {
			t.Errorf("Config with theme %q should be valid, got errors: %v", themeName, errs)
		}
	}
}

func TestValidateEmptyColors(t *testing.T) {
	cfg := &Config{
		Theme: "default",
		Colors: ColorConfig{
			// All empty - should be valid since empty means "use theme defaults"
		},
	}

	errs := cfg.Validate()
	if errs.HasErrors() {
		t.Errorf("Config with empty colors should be valid, got errors: %v", errs)
	}
}

func TestMergeWithThemeColors(t *testing.T) {
	theme := GetTheme("default")

	// Partial colors: user sets only Warning, rest should come from theme
	colors := ColorConfig{
		Warning: "#ff0000",
	}
	MergeWithThemeColors(&colors, theme.Colors)

	if colors.Warning != "#ff0000" {
		t.Errorf("Warning should be preserved as user value, got %q", colors.Warning)
	}
	if colors.AccentPrimary != theme.Colors.AccentPrimary {
		t.Errorf("AccentPrimary should come from theme, got %q", colors.AccentPrimary)
	}
	if colors.Text != theme.Colors.Text {
		t.Errorf("Text should come from theme, got %q", colors.Text)
	}
	if colors.TempHot != theme.Colors.TempHot {
		t.Errorf("TempHot should come from theme, got %q", colors.TempHot)
	}

	// All fields set: nothing should change
	full := theme.Colors
	full.Warning = "#aabbcc"
	MergeWithThemeColors(&full, theme.Colors)
	if full.Warning != "#aabbcc" {
		t.Errorf("All-set colors should not be overwritten, got %q", full.Warning)
	}
}

func TestValidateAndFixPartialColorMerge(t *testing.T) {
	// User sets only Warning — other colors should be filled from theme
	cfg := &Config{
		Theme: "default",
		Colors: ColorConfig{
			Warning: "#ff0000",
		},
	}
	cfg.ValidateAndFix()

	theme := GetTheme("default")
	if cfg.Colors.Warning != "#ff0000" {
		t.Errorf("User Warning should be preserved, got %q", cfg.Colors.Warning)
	}
	if cfg.Colors.AccentPrimary != theme.Colors.AccentPrimary {
		t.Errorf("AccentPrimary should come from theme, got %q", cfg.Colors.AccentPrimary)
	}
}

func TestValidateAndFixReturnsErrors(t *testing.T) {
	// Invalid theme should produce an error
	cfg := &Config{Theme: "nonexistent"}
	errs := cfg.ValidateAndFix()
	if !errs.HasErrors() {
		t.Error("ValidateAndFix should return errors when theme is invalid")
	}

	// Invalid color should produce an error
	cfg = DefaultConfig()
	cfg.Colors.AccentPrimary = "bad-color"
	errs = cfg.ValidateAndFix()
	if !errs.HasErrors() {
		t.Error("ValidateAndFix should return errors when color is invalid")
	}
}
