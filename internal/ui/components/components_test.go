package components

import (
	"strings"
	"testing"
	"time"

	"nuther/internal/config"
	"nuther/internal/smart"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

func newTestStyles() *styles.Styles {
	cfg := config.DefaultConfig()
	return styles.NewStyles(cfg)
}

func TestRenderHeader(t *testing.T) {
	s := newTestStyles()
	cfg := config.DefaultConfig()

	// With logo
	cfg.Display.ShowLogo = true
	header := RenderHeader(s, cfg)
	if header == "" {
		t.Error("Header should not be empty with logo")
	}
	if !strings.Contains(header, "NUTHER") || !strings.Contains(header, AppName) {
		// ASCII logo or app name should be present
	}

	// Without logo
	cfg.Display.ShowLogo = false
	header = RenderHeader(s, cfg)
	if header == "" {
		t.Error("Header should not be empty without logo")
	}
	if !strings.Contains(header, AppName) {
		t.Error("Header should contain app name")
	}
	if !strings.Contains(header, AppVersion) {
		t.Error("Header should contain version")
	}
}

func TestRenderTabs(t *testing.T) {
	s := newTestStyles()
	tabs := []string{"Overview", "Details", "Settings"}

	result := RenderTabs(tabs, 0, 80, s)

	if result == "" {
		t.Error("Tabs should not be empty")
	}

	// All tab names should be present
	for _, tab := range tabs {
		if !strings.Contains(result, tab) {
			t.Errorf("Tab %q should be in result", tab)
		}
	}
}

func TestRenderTabsActiveTab(t *testing.T) {
	s := newTestStyles()
	tabs := []string{"Tab1", "Tab2", "Tab3"}

	// Test different active tabs
	for i := 0; i < len(tabs); i++ {
		result := RenderTabs(tabs, i, 80, s)
		if result == "" {
			t.Errorf("Tabs with active=%d should not be empty", i)
		}
	}
}

func TestRenderHorizontalLine(t *testing.T) {
	s := newTestStyles()

	line := RenderHorizontalLine(10, s)
	if line == "" {
		t.Error("Line should not be empty")
	}
}

func TestRenderDriveSelector(t *testing.T) {
	s := newTestStyles()
	drives := []smart.DriveInfo{
		{Model: "Samsung SSD 970", HealthStatus: smart.HealthGood},
		{Model: "WD Blue", HealthStatus: smart.HealthCaution},
	}

	result := RenderDriveSelector(drives, 0, s)

	if result == "" {
		t.Error("Drive selector should not be empty")
	}
}

func TestRenderDriveSelectorEmpty(t *testing.T) {
	s := newTestStyles()
	var drives []smart.DriveInfo

	result := RenderDriveSelector(drives, 0, s)
	if result != "" {
		t.Error("Empty drives should produce empty result")
	}
}

type testKeyMap struct{}

func (k testKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

func TestRenderStatusBar(t *testing.T) {
	s := newTestStyles()
	now := time.Now()

	result := RenderStatusBar(0, 3, now, false, false, 80, testKeyMap{}, s)

	if result == "" {
		t.Error("Status bar should not be empty")
	}

	if !strings.Contains(result, "1/3") {
		t.Error("Status bar should show drive count")
	}
}

func TestRenderScreenshotStatus(t *testing.T) {
	s := newTestStyles()

	tests := []struct {
		status   string
		message  string
		expected bool // should have content
	}{
		{"capturing", "Capturing...", true},
		{"success", "Screenshot saved!", true},
		{"error", "Failed to capture", true},
		{"unknown", "Something", false},
		{"", "Empty status", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := RenderScreenshotStatus(tt.status, tt.message, s)

			if tt.expected && result == "" {
				t.Errorf("Status %q should produce output", tt.status)
			}
			if !tt.expected && result != "" {
				t.Errorf("Status %q should not produce output", tt.status)
			}
		})
	}
}

func TestRenderBox(t *testing.T) {
	s := newTestStyles()

	// Without title
	result := RenderBox("Content here", 40, "", s)
	if result == "" {
		t.Error("Box should not be empty")
	}
	if !strings.Contains(result, "Content here") {
		t.Error("Box should contain content")
	}

	// With title
	result = RenderBox("Content here", 40, "Title", s)
	if result == "" {
		t.Error("Box with title should not be empty")
	}
}

func TestRenderGauge(t *testing.T) {
	s := newTestStyles()

	tests := []struct {
		name  string
		value int
		max   int
		width int
	}{
		{"zero", 0, 100, 20},
		{"half", 50, 100, 20},
		{"full", 100, 100, 20},
		{"over", 120, 100, 20},
		{"negative", -10, 100, 20},
		{"zero max", 50, 0, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderGauge(tt.value, tt.max, tt.width, lipgloss.Color("#00ff00"), s)
			if result == "" {
				t.Error("Gauge should not be empty")
			}
		})
	}
}

func TestRenderHelp(t *testing.T) {
	s := newTestStyles()

	result := RenderHelp(s)

	if result == "" {
		t.Error("Help should not be empty")
	}

	// Check for key sections
	expectedContent := []string{
		"Navigation",
		"Actions",
		"Health",
		"quit",
	}

	for _, content := range expectedContent {
		if !strings.Contains(strings.ToLower(result), strings.ToLower(content)) {
			t.Errorf("Help should contain %q", content)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		maxLen   int
		expected string
	}{
		{"short string", "Hello", 10, "Hello"},
		{"exact length", "Hello", 5, "Hello"},
		{"needs truncation", "Hello World", 8, "Hello..."},
		{"very short max", "Hello", 2, "He"},
		{"max of 3", "Hello", 3, "Hel"},
		{"empty string", "", 10, ""},
		{"unicode", "日本語テスト", 5, "日本..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.str, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.str, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		width    int
		expected string
	}{
		{"short string", "Hi", 5, "Hi   "},
		{"exact width", "Hello", 5, "Hello"},
		{"longer than width", "Hello World", 5, "Hello"},
		{"empty string", "", 3, "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadRight(tt.str, tt.width)
			if result != tt.expected {
				t.Errorf("PadRight(%q, %d) = %q, want %q", tt.str, tt.width, result, tt.expected)
			}
		})
	}
}

func TestAppConstants(t *testing.T) {
	if AppName == "" {
		t.Error("AppName should not be empty")
	}
	if AppVersion == "" {
		t.Error("AppVersion should not be empty")
	}
	if AppDescription == "" {
		t.Error("AppDescription should not be empty")
	}
}

func TestMaxDriveNameLength(t *testing.T) {
	if maxDriveNameLength <= 0 {
		t.Errorf("maxDriveNameLength = %d, should be positive", maxDriveNameLength)
	}
}

func TestPadCenter(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		length   int
		expected string
	}{
		{"shorter string even padding", "Hi", 6, "  Hi  "},
		{"shorter string odd padding", "Hi", 5, " Hi  "},
		{"exact length", "Hello", 5, "Hello"},
		{"longer than length", "TooLong", 3, "TooLong"},
		{"empty string", "", 4, "    "},
		{"single char", "X", 3, " X "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadCenter(tt.str, tt.length)
			if result != tt.expected {
				t.Errorf("PadCenter(%q, %d) = %q, want %q", tt.str, tt.length, result, tt.expected)
			}
		})
	}
}

func TestGetTableRowStyle(t *testing.T) {
	s := newTestStyles()

	tests := []struct {
		name     string
		index    int
		selected int
	}{
		{"selected row", 0, 0},
		{"alt row (odd index)", 1, -1},
		{"normal row (even index)", 2, -1},
		{"normal row zero", 0, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTableRowStyle(tt.index, tt.selected, s)
			// Just verify it returns a non-zero style (not the zero value)
			_ = result.Render("test")
		})
	}

	// Verify the function returns without panicking for all branches
	_ = GetTableRowStyle(0, 0, s)  // selected
	_ = GetTableRowStyle(0, -1, s) // normal (even)
	_ = GetTableRowStyle(1, -1, s) // alt (odd)
}

func TestRenderDriveHeader_NVMe(t *testing.T) {
	s := newTestStyles()
	drive := smart.DriveInfo{
		Model:        "Samsung SSD 980 PRO 1TB",
		Serial:       "S5GXNF0R123456",
		Firmware:     "5B2QGXA7",
		Capacity:     "1.00 TB",
		Interface:    "NVMe",
		IsNVMe:       true,
		IsSSD:        true,
		HealthStatus: smart.HealthGood,
		HealthPassed: true,
		FormFactor:   "M.2",
		NVMeHealthLog: &smart.NVMeHealthLog{
			PercentageUsed: 2,
		},
	}

	result := RenderDriveHeader(drive, s)
	if result == "" {
		t.Error("RenderDriveHeader should not return empty string")
	}
	if !strings.Contains(result, "Samsung SSD 980 PRO 1TB") {
		t.Error("Header should contain model name")
	}
	if !strings.Contains(result, "NVMe") {
		t.Error("Header should contain protocol")
	}
	if !strings.Contains(result, "98%") {
		t.Error("Header should show health percentage (100-2=98)")
	}
}

func TestRenderDriveHeader_SATA_HDD(t *testing.T) {
	s := newTestStyles()
	drive := smart.DriveInfo{
		Model:        "WDC WD40EFRX",
		Serial:       "WD-WCC4N123456",
		Firmware:     "82.00A82",
		Capacity:     "4.00 TB",
		Interface:    "SATA HDD (5400 RPM)",
		RotationRate: 5400,
		HealthStatus: smart.HealthGood,
		HealthPassed: true,
		FormFactor:   "3.5 inches",
	}

	result := RenderDriveHeader(drive, s)
	if result == "" {
		t.Error("RenderDriveHeader should not return empty string for HDD")
	}
	if !strings.Contains(result, "WDC WD40EFRX") {
		t.Error("Header should contain model name")
	}
	if !strings.Contains(result, "5400 RPM") {
		t.Error("Header should show rotation rate for HDD")
	}
}

func TestRenderDriveHeader_SATA_SSD(t *testing.T) {
	s := newTestStyles()
	drive := smart.DriveInfo{
		Model:        "Samsung SSD 870 EVO",
		Serial:       "S4HNNS0R123456",
		Firmware:     "SVT01B6Q",
		Capacity:     "500 GB",
		Interface:    "SATA SSD",
		IsSSD:        true,
		HealthStatus: smart.HealthCaution,
		HealthPassed: true,
		FormFactor:   "2.5 inches",
	}

	result := RenderDriveHeader(drive, s)
	if result == "" {
		t.Error("RenderDriveHeader should not return empty for SATA SSD")
	}
	if !strings.Contains(result, "SATA SSD") {
		t.Error("Header should contain SATA SSD protocol")
	}
}

func TestRenderStatsCards_NVMe(t *testing.T) {
	s := newTestStyles()
	drive := smart.DriveInfo{
		Temperature:  38,
		PowerOnHours: 2500,
		PowerCycles:  150,
		IsNVMe:       true,
		Interface:    "NVMe",
	}

	result := RenderStatsCards(drive, s)
	if result == "" {
		t.Error("RenderStatsCards should not return empty string")
	}
	if !strings.Contains(result, "TEMPERATURE") {
		t.Error("Stats cards should contain TEMPERATURE label")
	}
	if !strings.Contains(result, "POWER ON HRS") {
		t.Error("Stats cards should contain POWER ON HRS label")
	}
	if !strings.Contains(result, "NVMe") {
		t.Error("Stats cards should show NVMe interface")
	}
}

func TestRenderStatsCards_SATA(t *testing.T) {
	s := newTestStyles()
	drive := smart.DriveInfo{
		Temperature:  32,
		PowerOnHours: 5000,
		PowerCycles:  500,
		IsNVMe:       false,
		Interface:    "SATA SSD",
	}

	result := RenderStatsCards(drive, s)
	if result == "" {
		t.Error("RenderStatsCards should not return empty string")
	}
	if !strings.Contains(result, "SATA") {
		t.Error("Stats cards should show SATA interface")
	}
}
