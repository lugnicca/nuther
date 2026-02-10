package styles

import (
	"testing"

	"nuther/internal/config"
	"nuther/internal/smart"
)

func TestNewStyles(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewStyles(cfg)

	if s == nil {
		t.Fatal("NewStyles returned nil")
	}

	// Check colors are set
	if s.AccentPrimary == "" {
		t.Error("AccentPrimary should be set")
	}
	if s.AccentSecondary == "" {
		t.Error("AccentSecondary should be set")
	}
	if s.Success == "" {
		t.Error("Success should be set")
	}
	if s.Warning == "" {
		t.Error("Warning should be set")
	}
	if s.Danger == "" {
		t.Error("Danger should be set")
	}
}

func TestNewStylesWithFahrenheit(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Display.ShowFahrenheit = true

	s := NewStyles(cfg)

	if !s.ShowFahrenheit {
		t.Error("ShowFahrenheit should be true")
	}
}

func TestGetHealthStyle(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewStyles(cfg)

	tests := []struct {
		status smart.HealthStatus
	}{
		{smart.HealthGood},
		{smart.HealthCaution},
		{smart.HealthBad},
		{"unknown"},
		{""},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			style := s.GetHealthStyle(tt.status)
			// Just ensure it doesn't panic and returns a style
			_ = style.Render("test")
		})
	}
}

func TestGetHealthColor(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewStyles(cfg)

	tests := []struct {
		status   smart.HealthStatus
		expected string
	}{
		{smart.HealthGood, string(s.Success)},
		{smart.HealthCaution, string(s.Warning)},
		{smart.HealthBad, string(s.Danger)},
		{"unknown", string(s.TextDim)},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			color := s.GetHealthColor(tt.status)
			if string(color) != tt.expected {
				t.Errorf("GetHealthColor(%q) = %q, want %q", tt.status, color, tt.expected)
			}
		})
	}
}

func TestGetHealthIcon(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewStyles(cfg)

	tests := []struct {
		status   smart.HealthStatus
		expected string
	}{
		{smart.HealthGood, IconSuccess},
		{smart.HealthCaution, IconWarning},
		{smart.HealthBad, IconError},
		{"unknown", IconInfo},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			icon := s.GetHealthIcon(tt.status)
			if icon != tt.expected {
				t.Errorf("GetHealthIcon(%q) = %q, want %q", tt.status, icon, tt.expected)
			}
		})
	}
}

func TestGetDriveIcon(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewStyles(cfg)

	tests := []struct {
		name     string
		drive    smart.DriveInfo
		expected string
	}{
		{"NVMe", smart.DriveInfo{IsNVMe: true}, IconNVMe},
		{"SSD", smart.DriveInfo{IsSSD: true}, IconSSD},
		{"USB", smart.DriveInfo{IsUSB: true}, IconUSB},
		{"HDD", smart.DriveInfo{}, IconHDD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := s.GetDriveIcon(&tt.drive)
			if icon != tt.expected {
				t.Errorf("GetDriveIcon() = %q, want %q", icon, tt.expected)
			}
		})
	}
}

func TestGetTemperatureColor(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewStyles(cfg)

	tests := []struct {
		name string
		temp int
	}{
		{"cold", 20},
		{"normal", 35},
		{"warm", 45},
		{"hot", 55},
		{"critical", 65},
		{"danger", 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := s.GetTemperatureColor(tt.temp)
			if color == "" {
				t.Errorf("GetTemperatureColor(%d) returned empty", tt.temp)
			}
		})
	}
}

func TestFormatTemperature(t *testing.T) {
	cfg := config.DefaultConfig()

	// Celsius
	cfg.Display.ShowFahrenheit = false
	s := NewStyles(cfg)

	result := s.FormatTemperature(25)
	if result != "25°C" {
		t.Errorf("FormatTemperature(25) in Celsius = %q, want %q", result, "25°C")
	}

	// Fahrenheit
	cfg.Display.ShowFahrenheit = true
	s = NewStyles(cfg)

	result = s.FormatTemperature(25)
	expected := "77°F" // 25 * 9/5 + 32 = 77
	if result != expected {
		t.Errorf("FormatTemperature(25) in Fahrenheit = %q, want %q", result, expected)
	}
}

func TestFormatTemperatureEdgeCases(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Display.ShowFahrenheit = true
	s := NewStyles(cfg)

	tests := []struct {
		celsius    int
		fahrenheit string
	}{
		{0, "32°F"},
		{100, "212°F"},
		{-40, "-40°F"},
	}

	for _, tt := range tests {
		t.Run(tt.fahrenheit, func(t *testing.T) {
			result := s.FormatTemperature(tt.celsius)
			if result != tt.fahrenheit {
				t.Errorf("FormatTemperature(%d) = %q, want %q", tt.celsius, result, tt.fahrenheit)
			}
		})
	}
}

func TestIconConstants(t *testing.T) {
	icons := []struct {
		name  string
		value string
	}{
		{"IconSuccess", IconSuccess},
		{"IconWarning", IconWarning},
		{"IconError", IconError},
		{"IconInfo", IconInfo},
		{"IconBullet", IconBullet},
		{"IconHDD", IconHDD},
		{"IconSSD", IconSSD},
		{"IconNVMe", IconNVMe},
		{"IconUSB", IconUSB},
	}

	for _, icon := range icons {
		t.Run(icon.name, func(t *testing.T) {
			if icon.value == "" {
				t.Errorf("%s should not be empty", icon.name)
			}
		})
	}
}

func TestBoxDrawingConstants(t *testing.T) {
	chars := []struct {
		name  string
		value string
	}{
		{"BoxTopLeft", BoxTopLeft},
		{"BoxTopRight", BoxTopRight},
		{"BoxBottomLeft", BoxBottomLeft},
		{"BoxBottomRight", BoxBottomRight},
		{"BoxHorizontal", BoxHorizontal},
		{"BoxVertical", BoxVertical},
		{"ProgressFull", ProgressFull},
		{"ProgressEmpty", ProgressEmpty},
	}

	for _, c := range chars {
		t.Run(c.name, func(t *testing.T) {
			if c.value == "" {
				t.Errorf("%s should not be empty", c.name)
			}
		})
	}
}

func TestStylesAreRenderable(t *testing.T) {
	cfg := config.DefaultConfig()
	s := NewStyles(cfg)

	// Test that all styles can render without panic
	// Each style is tested individually
	testCases := []struct {
		name   string
		render func() string
	}{
		{"Base", func() string { return s.Base.Render("test") }},
		{"Bold", func() string { return s.Bold.Render("test") }},
		{"Italic", func() string { return s.Italic.Render("test") }},
		{"Dim", func() string { return s.Dim.Render("test") }},
		{"Logo", func() string { return s.Logo.Render("test") }},
		{"Subtitle", func() string { return s.Subtitle.Render("test") }},
		{"Tab", func() string { return s.Tab.Render("test") }},
		{"ActiveTab", func() string { return s.ActiveTab.Render("test") }},
		{"DriveButton", func() string { return s.DriveButton.Render("test") }},
		{"DriveButtonSelected", func() string { return s.DriveButtonSelected.Render("test") }},
		{"HealthGood", func() string { return s.HealthGood.Render("test") }},
		{"HealthCaution", func() string { return s.HealthCaution.Render("test") }},
		{"HealthBad", func() string { return s.HealthBad.Render("test") }},
		{"TableHeader", func() string { return s.TableHeader.Render("test") }},
		{"TableRow", func() string { return s.TableRow.Render("test") }},
		{"TableRowAlt", func() string { return s.TableRowAlt.Render("test") }},
		{"TableRowSelected", func() string { return s.TableRowSelected.Render("test") }},
		{"Card", func() string { return s.Card.Render("test") }},
		{"CardTitle", func() string { return s.CardTitle.Render("test") }},
		{"CardValue", func() string { return s.CardValue.Render("test") }},
		{"CardLabel", func() string { return s.CardLabel.Render("test") }},
		{"StatusBar", func() string { return s.StatusBar.Render("test") }},
		{"HelpKey", func() string { return s.HelpKey.Render("test") }},
		{"HelpDesc", func() string { return s.HelpDesc.Render("test") }},
		{"HelpOverlay", func() string { return s.HelpOverlay.Render("test") }},
		{"Box", func() string { return s.Box.Render("test") }},
		{"BoxTitle", func() string { return s.BoxTitle.Render("test") }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.render()
			if result == "" {
				t.Errorf("%s.Render() returned empty string", tc.name)
			}
		})
	}
}

func TestNewStylesWithDifferentThemes(t *testing.T) {
	themes := []string{"default", "dracula", "gruvbox", "nord", "rose-petale", "solarized-dark", "sous-bois"}

	for _, themeName := range themes {
		t.Run(themeName, func(t *testing.T) {
			cfg := config.DefaultConfig()
			theme := config.GetTheme(themeName)
			cfg.Theme = themeName
			cfg.Colors = theme.Colors

			s := NewStyles(cfg)

			if s == nil {
				t.Errorf("NewStyles with theme %q returned nil", themeName)
			}

			// Verify colors match theme
			if string(s.AccentPrimary) != theme.Colors.AccentPrimary {
				t.Errorf("AccentPrimary mismatch for theme %q", themeName)
			}
		})
	}
}
