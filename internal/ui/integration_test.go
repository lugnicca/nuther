package ui

import (
	"fmt"
	"strings"
	"testing"

	"nuther/internal/config"
	"nuther/internal/smart"
)

// setupModel creates a ready Model with drives loaded via DrivesLoadedMsg,
// simulating the real pipeline: drives → message → Update → model ready for View.
func setupModel(t *testing.T, drives []smart.DriveInfo) Model {
	t.Helper()
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	msg := DrivesLoadedMsg{Drives: drives}
	updated, _ := m.Update(msg)
	return updated.(Model)
}

func TestIntegration_DemoData_AllTabs(t *testing.T) {
	drives := smart.CreateDemoData()
	m := setupModel(t, drives)

	if len(m.Drives) != len(drives) {
		t.Fatalf("expected %d drives, got %d", len(drives), len(m.Drives))
	}

	selectedModel := m.Drives[m.SelectedDrive].Model

	for tab := 0; tab < TabCount; tab++ {
		m.ActiveTab = tab
		view := m.View()

		if view == "" {
			t.Errorf("Tab %d (%s): View() returned empty string", tab, TabNames[tab])
		}

		// The selected drive model name should appear in the view for all tabs
		// except Settings (which doesn't show per-drive info)
		if tab != TabSettings && tab != TabSnapshots && !strings.Contains(view, selectedModel) {
			t.Errorf("Tab %d (%s): expected to find model %q in view output", tab, TabNames[tab], selectedModel)
		}
	}
}

func TestIntegration_NVMeDrive_OverviewContent(t *testing.T) {
	drive := smart.DriveInfo{
		Device:       "/dev/nvme0n1",
		Model:        "Samsung SSD 970 EVO Plus 1TB",
		Serial:       "S4EVNX0T123456K",
		Firmware:     "2B2QEXM7",
		Capacity:     "1.00 TB",
		Interface:    "NVMe",
		IsNVMe:       true,
		IsSSD:        true,
		FormFactor:   "M.2",
		HealthStatus: smart.HealthGood,
		HealthPassed: true,
		Temperature:  38,
		PowerOnHours: 12547,
		PowerCycles:  1823,
		NVMeAttributes: []smart.NVMeAttribute{
			{Name: "Temperature", RawValue: "38°C", NumericValue: 38, Status: smart.HealthGood},
			{Name: "Available Spare", RawValue: "100%", NumericValue: 100, Status: smart.HealthGood},
			{Name: "Percentage Used", RawValue: "3%", NumericValue: 3, Status: smart.HealthGood},
		},
		NVMeHealthLog: &smart.NVMeHealthLog{
			PercentageUsed: 3,
			Temperature:    38,
		},
	}

	m := setupModel(t, []smart.DriveInfo{drive})
	m.ActiveTab = TabOverview
	view := m.View()

	checks := []struct {
		name string
		want string
	}{
		{"model name", "Samsung SSD 970 EVO Plus 1TB"},
		{"NVMe interface", "NVMe"},
	}

	for _, c := range checks {
		if !strings.Contains(view, c.want) {
			t.Errorf("Overview should contain %s (%q)", c.name, c.want)
		}
	}
}

func TestIntegration_SATADrive_AttributesContent(t *testing.T) {
	drive := smart.DriveInfo{
		Device:       "/dev/sda",
		Model:        "WDC WD40EZRZ-00GXCB0",
		Serial:       "WD-WCC7K3LVJKEE",
		Firmware:     "80.00A80",
		Capacity:     "4.00 TB",
		Interface:    "SATA HDD (5400 RPM)",
		RotationRate: 5400,
		HealthStatus: smart.HealthGood,
		HealthPassed: true,
		Temperature:  34,
		Attributes: []smart.SmartAttribute{
			{ID: 5, Name: "Reallocated_Sector_Ct", Value: 200, Worst: 200, Threshold: 140, RawValue: 0, RawString: "0"},
			{ID: 9, Name: "Power_On_Hours", Value: 67, Worst: 67, Threshold: 0, RawValue: 28934, RawString: "28934"},
			{ID: 194, Name: "Temperature_Celsius", Value: 117, Worst: 106, Threshold: 0, RawValue: 34, RawString: "34"},
			{ID: 197, Name: "Current_Pending_Sector", Value: 200, Worst: 200, Threshold: 0, RawValue: 0, RawString: "0"},
		},
	}

	m := setupModel(t, []smart.DriveInfo{drive})
	m.ActiveTab = TabAttributes
	view := m.View()

	// SATA attributes should show attribute names
	attrNames := []string{
		"Reallocated_Sector_Ct",
		"Power_On_Hours",
		"Temperature_Celsius",
	}

	for _, name := range attrNames {
		if !strings.Contains(view, name) {
			t.Errorf("Attributes tab should contain attribute %q", name)
		}
	}
}

func TestIntegration_HDDDrive_DetailsContent(t *testing.T) {
	drive := smart.DriveInfo{
		Device:       "/dev/sdb",
		Model:        "Seagate ST2000DM008",
		Serial:       "ZFL1KXYZ",
		Firmware:     "0001",
		Capacity:     "2.00 TB",
		Interface:    "SATA HDD (7200 RPM)",
		FormFactor:   "3.5 inches",
		RotationRate: 7200,
		HealthStatus: smart.HealthCaution,
		HealthPassed: true,
		Temperature:  42,
	}

	m := setupModel(t, []smart.DriveInfo{drive})
	m.ActiveTab = TabDetails
	view := m.View()

	checks := []struct {
		name string
		want string
	}{
		{"serial", "ZFL1KXYZ"},
		{"firmware", "0001"},
		{"form factor", "3.5 inches"},
	}

	for _, c := range checks {
		if !strings.Contains(view, c.want) {
			t.Errorf("Details tab should contain %s (%q)", c.name, c.want)
		}
	}
}

func TestIntegration_MultiDrive_Navigation(t *testing.T) {
	drives := []smart.DriveInfo{
		{Device: "/dev/sda", Model: "Drive Alpha", HealthStatus: smart.HealthGood},
		{Device: "/dev/sdb", Model: "Drive Beta", HealthStatus: smart.HealthGood},
		{Device: "/dev/sdc", Model: "Drive Gamma", HealthStatus: smart.HealthCaution},
	}

	m := setupModel(t, drives)

	if m.SelectedDrive != 0 {
		t.Fatalf("initial SelectedDrive should be 0, got %d", m.SelectedDrive)
	}

	// NextDrive
	m.NextDrive()
	if m.SelectedDrive != 1 {
		t.Fatalf("after NextDrive, SelectedDrive should be 1, got %d", m.SelectedDrive)
	}

	// View on Overview should show drive 2's model name
	m.ActiveTab = TabOverview
	view := m.View()
	if !strings.Contains(view, "Drive Beta") {
		t.Error("Overview should contain the selected drive 'Drive Beta'")
	}

	// AllDrives tab should show all 3 model names
	m.ActiveTab = TabAllDrives
	view = m.View()
	for _, d := range drives {
		if !strings.Contains(view, d.Model) {
			t.Errorf("AllDrives tab should contain %q", d.Model)
		}
	}
}

func TestIntegration_DemoMode_Banner(t *testing.T) {
	drives := []smart.DriveInfo{
		{Device: "demo0", Model: "Demo Drive", IsDemo: true, HealthStatus: smart.HealthGood},
	}

	m := setupModel(t, drives)
	m.ActiveTab = TabOverview
	view := m.View()

	if !strings.Contains(view, "DEMO") {
		t.Error("View should contain 'DEMO' banner when drives are demo data")
	}
}

func TestIntegration_ErrorState_NoDrives(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50

	msg := DrivesLoadedMsg{Drives: nil, Error: fmt.Errorf("permission denied")}
	updated, _ := m.Update(msg)
	model := updated.(Model)

	view := model.View()

	// The no-drives view should show the warning about no drives
	if !strings.Contains(strings.ToLower(view), "no drives") {
		t.Error("View with no drives should contain 'No drives' message")
	}
}

func TestIntegration_DriveSelection_BoundsReset(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Ready = true
	m.Width = 120
	m.Height = 50
	m.SelectedDrive = 5 // Out of range for the upcoming 2-drive load

	drives := []smart.DriveInfo{
		{Device: "/dev/sda", Model: "Drive One", HealthStatus: smart.HealthGood},
		{Device: "/dev/sdb", Model: "Drive Two", HealthStatus: smart.HealthGood},
	}
	msg := DrivesLoadedMsg{Drives: drives}
	updated, _ := m.Update(msg)
	model := updated.(Model)

	if model.SelectedDrive != 0 {
		t.Errorf("SelectedDrive should be reset to 0 when out of bounds, got %d", model.SelectedDrive)
	}

	// Verify the view renders correctly with the reset selection
	view := model.View()
	if !strings.Contains(view, "Drive One") {
		t.Error("View should show the first drive after bounds reset")
	}
}
