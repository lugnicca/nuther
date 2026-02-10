package views

import (
	"strings"
	"testing"

	"nuther/internal/config"
	"nuther/internal/smart"
	"nuther/internal/ui/components"
	"nuther/internal/ui/styles"
)

func newTestStyles() *styles.Styles {
	cfg := config.DefaultConfig()
	return styles.NewStyles(cfg)
}

func createTestNVMeDrive() smart.DriveInfo {
	return smart.DriveInfo{
		Device:       "/dev/nvme0n1",
		Model:        "Samsung 970 EVO Plus",
		Serial:       "S123456",
		Firmware:     "1.0",
		Capacity:     "1.00 TB",
		Interface:    "NVMe",
		IsNVMe:       true,
		IsSSD:        true,
		HealthStatus: smart.HealthGood,
		Temperature:  35,
		PowerOnHours: 1000,
		PowerCycles:  100,
		FormFactor:   "M.2",
		NVMeAttributes: []smart.NVMeAttribute{
			{Name: "Temperature", RawValue: "35°C", Status: smart.HealthGood},
			{Name: "Available Spare", RawValue: "100%", Status: smart.HealthGood},
		},
		NVMeHealthLog: &smart.NVMeHealthLog{
			PercentageUsed: 5,
		},
	}
}

func createTestSATADrive() smart.DriveInfo {
	return smart.DriveInfo{
		Device:       "/dev/sda",
		Model:        "Samsung 870 EVO",
		Serial:       "S654321",
		Firmware:     "SVT02B6Q",
		Capacity:     "500 GB",
		Interface:    "SATA SSD",
		IsNVMe:       false,
		IsSSD:        true,
		HealthStatus: smart.HealthGood,
		Temperature:  32,
		PowerOnHours: 5000,
		PowerCycles:  500,
		FormFactor:   "2.5 inches",
		Attributes: []smart.SmartAttribute{
			{ID: 5, Name: "Reallocated_Sector_Ct", Value: 100, Worst: 100, Threshold: 10, RawValue: 0},
			{ID: 9, Name: "Power_On_Hours", Value: 99, Worst: 99, Threshold: 0, RawValue: 5000},
			{ID: 194, Name: "Temperature_Celsius", Value: 68, Worst: 60, Threshold: 0, RawValue: 32},
		},
	}
}

func createTestHDDDrive() smart.DriveInfo {
	return smart.DriveInfo{
		Device:       "/dev/sdb",
		Model:        "WD Blue 4TB",
		Serial:       "WD12345",
		Firmware:     "80.00A80",
		Capacity:     "4.00 TB",
		Interface:    "SATA HDD (5400 RPM)",
		IsNVMe:       false,
		IsSSD:        false,
		RotationRate: 5400,
		HealthStatus: smart.HealthCaution,
		Temperature:  40,
		PowerOnHours: 20000,
		PowerCycles:  1000,
		FormFactor:   "3.5 inches",
		Attributes: []smart.SmartAttribute{
			{ID: 5, Name: "Reallocated_Sector_Ct", Value: 100, Worst: 100, Threshold: 10, RawValue: 5},
			{ID: 9, Name: "Power_On_Hours", Value: 70, Worst: 70, Threshold: 0, RawValue: 20000},
		},
	}
}

func TestRenderOverview(t *testing.T) {
	s := newTestStyles()

	tests := []struct {
		name  string
		drive smart.DriveInfo
	}{
		{"NVMe drive", createTestNVMeDrive()},
		{"SATA SSD", createTestSATADrive()},
		{"HDD", createTestHDDDrive()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderOverview(tt.drive, 0, 0, 120, 50, s)

			if result == "" {
				t.Error("Overview should not be empty")
			}

			// Should contain model name
			if !strings.Contains(result, tt.drive.Model) {
				t.Errorf("Overview should contain model %q", tt.drive.Model)
			}
		})
	}
}

func TestRenderOverviewWithScroll(t *testing.T) {
	s := newTestStyles()
	drive := createTestSATADrive()

	// Add more attributes for scrolling
	for i := 0; i < 20; i++ {
		drive.Attributes = append(drive.Attributes, smart.SmartAttribute{
			ID:   200 + i,
			Name: "Test_Attribute",
		})
	}

	result := RenderOverview(drive, 10, 5, 120, 30, s)
	if result == "" {
		t.Error("Overview with scroll should not be empty")
	}
}

func TestRenderAttributes(t *testing.T) {
	s := newTestStyles()

	tests := []struct {
		name  string
		drive smart.DriveInfo
	}{
		{"NVMe drive", createTestNVMeDrive()},
		{"SATA drive", createTestSATADrive()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderAttributes(tt.drive, 0, 0, 120, 50, s)

			if result == "" {
				t.Error("Attributes should not be empty")
			}
		})
	}
}

func TestRenderAttributesEmpty(t *testing.T) {
	s := newTestStyles()
	drive := smart.DriveInfo{
		Model: "Empty Drive",
	}

	result := RenderAttributes(drive, 0, 0, 120, 50, s)
	if result == "" {
		t.Error("Attributes for empty drive should show a message")
	}
}

func TestRenderDetails(t *testing.T) {
	s := newTestStyles()

	tests := []struct {
		name  string
		drive smart.DriveInfo
	}{
		{"NVMe drive", createTestNVMeDrive()},
		{"SATA SSD", createTestSATADrive()},
		{"HDD", createTestHDDDrive()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderDetails(tt.drive, 120, s)

			if result == "" {
				t.Error("Details should not be empty")
			}

			// Should contain device info
			if !strings.Contains(result, tt.drive.Device) && tt.drive.Device != "" {
				t.Errorf("Details should contain device %q", tt.drive.Device)
			}
		})
	}
}

func TestRenderAllDrives(t *testing.T) {
	s := newTestStyles()
	drives := []smart.DriveInfo{
		createTestNVMeDrive(),
		createTestSATADrive(),
		createTestHDDDrive(),
	}

	result := RenderAllDrives(drives, 0, 120, 50, s)

	if result == "" {
		t.Error("All drives should not be empty")
	}

	// Should contain all drive models
	for _, drive := range drives {
		if !strings.Contains(result, drive.Model) {
			t.Errorf("All drives should contain model %q", drive.Model)
		}
	}
}

func TestRenderAllDrivesEmpty(t *testing.T) {
	s := newTestStyles()
	var drives []smart.DriveInfo

	result := RenderAllDrives(drives, 0, 120, 50, s)
	if result == "" {
		t.Error("All drives for empty list should show a message")
	}
}

func TestRenderAllDrivesSelected(t *testing.T) {
	s := newTestStyles()
	drives := []smart.DriveInfo{
		createTestNVMeDrive(),
		createTestSATADrive(),
	}

	// Test different selection indices
	for i := 0; i < len(drives); i++ {
		result := RenderAllDrives(drives, i, 120, 50, s)
		if result == "" {
			t.Errorf("All drives with selected=%d should not be empty", i)
		}
	}
}

func TestRenderSettings(t *testing.T) {
	s := newTestStyles()
	cfg := config.DefaultConfig()

	tests := []struct {
		name            string
		selectedSetting int
		message         string
	}{
		{"theme selected", 0, ""},
		{"logo selected", 1, ""},
		{"temp unit selected", 2, ""},
		{"with message", 0, "Settings saved!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderSettings(cfg, tt.selectedSetting, tt.message, 120, s)

			if result == "" {
				t.Error("Settings should not be empty")
			}

			if tt.message != "" && !strings.Contains(result, tt.message) {
				t.Errorf("Settings should contain message %q", tt.message)
			}
		})
	}
}

func TestRenderSettingsThemes(t *testing.T) {
	s := newTestStyles()

	themes := []string{"default", "dracula", "gruvbox", "nord", "rose-petale", "solarized-dark", "sous-bois"}
	for _, theme := range themes {
		cfg := config.DefaultConfig()
		cfg.Theme = theme

		result := RenderSettings(cfg, 0, "", 120, s)
		if result == "" {
			t.Errorf("Settings with theme %q should not be empty", theme)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := components.Truncate(tt.str, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.str, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestRenderOverviewWithNilNVMeHealthLog(t *testing.T) {
	s := newTestStyles()
	drive := createTestNVMeDrive()
	drive.NVMeHealthLog = nil

	result := RenderOverview(drive, 0, 0, 120, 50, s)
	if result == "" {
		t.Error("Overview with nil NVMeHealthLog should not be empty")
	}
}

func TestRenderOverviewHighTemperature(t *testing.T) {
	s := newTestStyles()
	drive := createTestNVMeDrive()
	drive.Temperature = 75 // Hot temperature

	result := RenderOverview(drive, 0, 0, 120, 50, s)
	if result == "" {
		t.Error("Overview with high temperature should not be empty")
	}
}

func TestRenderOverviewHealthStatuses(t *testing.T) {
	s := newTestStyles()

	statuses := []smart.HealthStatus{smart.HealthGood, smart.HealthCaution, smart.HealthBad}
	for _, status := range statuses {
		drive := createTestNVMeDrive()
		drive.HealthStatus = status

		result := RenderOverview(drive, 0, 0, 120, 50, s)
		if result == "" {
			t.Errorf("Overview with health %q should not be empty", status)
		}
	}
}
