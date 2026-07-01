package ui

import (
	"strings"
	"testing"
	"time"

	"nuther/internal/config"
	"nuther/internal/smart"
	"nuther/internal/smartwatch"
)

func TestViewNotReady(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = false

	result := m.View()

	if result == "" {
		t.Error("View when not ready should show loading")
	}
	if !strings.Contains(result, "Initializing") {
		t.Error("View when not ready should show initializing message")
	}
}

func TestViewReady(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50

	result := m.View()

	if result == "" {
		t.Error("View when ready should not be empty")
	}
}

func TestViewNoDrives(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	m.Drives = []smart.DriveInfo{}

	result := m.View()

	if result == "" {
		t.Error("View with no drives should not be empty")
	}
	if !strings.Contains(result, "No drives") {
		t.Error("View with no drives should show warning")
	}
}

func TestViewWithDrives(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	m.Drives = []smart.DriveInfo{
		{
			Device:       "/dev/sda",
			Model:        "Test Drive",
			HealthStatus: smart.HealthGood,
		},
	}

	result := m.View()

	if result == "" {
		t.Error("View with drives should not be empty")
	}
}

func TestViewDifferentTabs(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	m.Drives = []smart.DriveInfo{
		{
			Device:       "/dev/sda",
			Model:        "Test Drive",
			HealthStatus: smart.HealthGood,
			Attributes: []smart.SmartAttribute{
				{ID: 1, Name: "Test"},
			},
		},
	}

	tabs := []int{TabOverview, TabAttributes, TabDetails, TabAllDrives, TabSectorGrid, TabSnapshots, TabSettings}
	for _, tab := range tabs {
		m.ActiveTab = tab
		result := m.View()

		if result == "" {
			t.Errorf("View with tab %d should not be empty", tab)
		}
	}
}

func TestViewWithScreenshotStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	m.Drives = []smart.DriveInfo{{Model: "Test"}}
	m.ScreenshotStatus = "success"
	m.ScreenshotMessage = "Screenshot saved!"

	result := m.View()

	if result == "" {
		t.Error("View with screenshot status should not be empty")
	}
}

func TestViewWithHelp(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	m.Drives = []smart.DriveInfo{{Model: "Test"}}
	m.ShowHelp = true

	result := m.View()

	if result == "" {
		t.Error("View with help should not be empty")
	}
}

func TestOverlayHelp(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Width = 80
	m.Height = 30

	screen := strings.Repeat("Line content\n", 30)
	helpContent := "Help line 1\nHelp line 2\nHelp line 3"

	result := m.overlayHelp(screen, helpContent)

	if result == "" {
		t.Error("overlayHelp should not return empty")
	}

	// Help content should be in result
	if !strings.Contains(result, "Help line") {
		t.Error("overlayHelp should contain help content")
	}
}

func TestOverlayHelpSmallScreen(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Width = 40
	m.Height = 5

	screen := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	helpContent := "Help 1\nHelp 2\nHelp 3\nHelp 4\nHelp 5\nHelp 6\nHelp 7"

	result := m.overlayHelp(screen, helpContent)

	if result == "" {
		t.Error("overlayHelp on small screen should not return empty")
	}
}

func TestViewSettingsTab(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	m.Drives = []smart.DriveInfo{{Model: "Test"}}
	m.ActiveTab = TabSettings
	m.SettingsMessage = "Settings saved!"

	result := m.View()

	if result == "" {
		t.Error("View on settings tab should not be empty")
	}
}

func TestViewAllDrivesTab(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	m.Drives = []smart.DriveInfo{
		{Model: "Drive 1", HealthStatus: smart.HealthGood},
		{Model: "Drive 2", HealthStatus: smart.HealthCaution},
	}
	m.ActiveTab = TabAllDrives

	result := m.View()

	if result == "" {
		t.Error("View on all drives tab should not be empty")
	}
}

func TestViewSectorGridTab(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	m.Drives = []smart.DriveInfo{
		{Model: "Drive 1", HealthStatus: smart.HealthGood},
		{Model: "Drive 2", HealthStatus: smart.HealthGood, PendingSectors: 1},
	}
	m.ActiveTab = TabSectorGrid
	m.SelectedDrive = 1

	result := m.View()

	if result == "" {
		t.Error("View on sector grid tab should not be empty")
	}
	if !strings.Contains(result, "Sector Grid") {
		t.Error("View on sector grid tab should contain title")
	}
	if !strings.Contains(result, "Disk 2/2") {
		t.Error("View on sector grid tab should show selected disk index")
	}
	if !strings.Contains(result, "Drive 2") || strings.Contains(result, "Drive 1 /dev") {
		t.Error("View on sector grid tab should focus the selected disk")
	}
}

func TestViewSnapshotsTabShowsKnownDisksAndHistory(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 140
	m.Height = 50
	m.Drives = []smart.DriveInfo{{Model: "Live Drive", HealthStatus: smart.HealthGood}}
	m.ActiveTab = TabSnapshots
	m.SelectedSnapshot = 1

	first := time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC)
	second := first.Add(2 * time.Hour)
	m.SnapshotIndex = smartwatch.Index{
		Version:   1,
		UpdatedAt: second,
		Devices: map[string]smartwatch.DeviceRecord{
			"serial-sn123": {
				Key:          "serial-sn123",
				FirstSeen:    first,
				LastSeen:     second,
				SnapshotIDs:  []string{"snap-1", "snap-2"},
				LastSnapshot: "snap-2",
				Summary: smartwatch.DeviceSummary{
					Key:          "serial-sn123",
					Device:       "/dev/sda",
					Serial:       "SN123",
					Model:        "Archive Drive",
					Capacity:     "1 TB",
					HealthStatus: smart.HealthCaution,
				},
			},
		},
		Snapshots: []smartwatch.SnapshotRecord{
			{ID: "snap-1", Timestamp: first, Reason: smartwatch.ReasonStartup, Path: "snapshots/snap-1.json", Device: smartwatch.DeviceSummary{Model: "Archive Drive", HealthStatus: smart.HealthGood}},
			{ID: "snap-2", Timestamp: second, Reason: smartwatch.ReasonManual, Path: "snapshots/snap-2.json", Device: smartwatch.DeviceSummary{Model: "Archive Drive", HealthStatus: smart.HealthCaution}},
		},
	}

	result := m.View()

	for _, want := range []string{"Snapshots", "Known disks", "Snapshot history", "Archive Drive", "snap-2", "GET /snapshots/snap-1", "Enter: open overview"} {
		if !strings.Contains(result, want) {
			t.Fatalf("snapshots tab missing %q in view:\n%s", want, result)
		}
	}
}
