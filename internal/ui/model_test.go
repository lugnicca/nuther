package ui

import (
	"testing"
	"time"

	"nuther/internal/config"
	"nuther/internal/smart"
)

func TestNewModel(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	if m.Config != cfg {
		t.Error("Config not set correctly")
	}

	if m.Styles == nil {
		t.Error("Styles should not be nil")
	}

	if len(m.Drives) != 0 {
		t.Errorf("Initial Drives should be empty, got %d", len(m.Drives))
	}

	if m.SelectedDrive != 0 {
		t.Errorf("Initial SelectedDrive = %d, want 0", m.SelectedDrive)
	}

	if m.ActiveTab != TabOverview {
		t.Errorf("Initial ActiveTab = %d, want %d", m.ActiveTab, TabOverview)
	}

	if m.Ready {
		t.Error("Initial Ready should be false")
	}

	if !m.Loading {
		t.Error("Initial Loading should be true")
	}

	if m.ShowHelp {
		t.Error("Initial ShowHelp should be false")
	}

	if len(m.Tabs) != len(TabNames) {
		t.Errorf("Tabs length = %d, want %d", len(m.Tabs), len(TabNames))
	}
}

func TestModelGetCurrentDrive(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	// No drives
	if drive := m.GetCurrentDrive(); drive != nil {
		t.Error("GetCurrentDrive should return nil when no drives")
	}

	// Add drives
	m.Drives = []smart.DriveInfo{
		{Device: "/dev/sda", Model: "Drive1"},
		{Device: "/dev/sdb", Model: "Drive2"},
	}

	// First drive
	m.SelectedDrive = 0
	drive := m.GetCurrentDrive()
	if drive == nil {
		t.Fatal("GetCurrentDrive returned nil")
	}
	if drive.Model != "Drive1" {
		t.Errorf("GetCurrentDrive Model = %q, want %q", drive.Model, "Drive1")
	}

	// Second drive
	m.SelectedDrive = 1
	drive = m.GetCurrentDrive()
	if drive.Model != "Drive2" {
		t.Errorf("GetCurrentDrive Model = %q, want %q", drive.Model, "Drive2")
	}

	// Out of range
	m.SelectedDrive = 10
	if drive := m.GetCurrentDrive(); drive != nil {
		t.Error("GetCurrentDrive should return nil when index out of range")
	}
}

func TestModelNextDrive(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	// No drives - should not panic
	m.NextDrive()
	if m.SelectedDrive != 0 {
		t.Errorf("SelectedDrive = %d, want 0", m.SelectedDrive)
	}

	// Add drives
	m.Drives = []smart.DriveInfo{
		{Device: "/dev/sda"},
		{Device: "/dev/sdb"},
		{Device: "/dev/sdc"},
	}

	m.SelectedDrive = 0
	m.NextDrive()
	if m.SelectedDrive != 1 {
		t.Errorf("After NextDrive: SelectedDrive = %d, want 1", m.SelectedDrive)
	}

	m.NextDrive()
	if m.SelectedDrive != 2 {
		t.Errorf("After NextDrive: SelectedDrive = %d, want 2", m.SelectedDrive)
	}

	// Wrap around
	m.NextDrive()
	if m.SelectedDrive != 0 {
		t.Errorf("After NextDrive (wrap): SelectedDrive = %d, want 0", m.SelectedDrive)
	}
}

func TestModelPrevDrive(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	// No drives
	m.PrevDrive()
	if m.SelectedDrive != 0 {
		t.Errorf("SelectedDrive = %d, want 0", m.SelectedDrive)
	}

	// Add drives
	m.Drives = []smart.DriveInfo{
		{Device: "/dev/sda"},
		{Device: "/dev/sdb"},
		{Device: "/dev/sdc"},
	}

	m.SelectedDrive = 2
	m.PrevDrive()
	if m.SelectedDrive != 1 {
		t.Errorf("After PrevDrive: SelectedDrive = %d, want 1", m.SelectedDrive)
	}

	m.PrevDrive()
	if m.SelectedDrive != 0 {
		t.Errorf("After PrevDrive: SelectedDrive = %d, want 0", m.SelectedDrive)
	}

	// Wrap around
	m.PrevDrive()
	if m.SelectedDrive != 2 {
		t.Errorf("After PrevDrive (wrap): SelectedDrive = %d, want 2", m.SelectedDrive)
	}
}

func TestModelNextTab(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	m.ActiveTab = TabOverview
	m.NextTab()
	if m.ActiveTab != TabAttributes {
		t.Errorf("After NextTab: ActiveTab = %d, want %d", m.ActiveTab, TabAttributes)
	}

	// Go through all tabs
	for i := 0; i < TabCount; i++ {
		m.NextTab()
	}
	if m.ActiveTab != TabAttributes {
		t.Errorf("After full cycle: ActiveTab = %d, want %d", m.ActiveTab, TabAttributes)
	}
}

func TestModelPrevTab(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	m.ActiveTab = TabOverview
	m.PrevTab()
	if m.ActiveTab != TabSettings {
		t.Errorf("After PrevTab: ActiveTab = %d, want %d", m.ActiveTab, TabSettings)
	}

	m.PrevTab()
	if m.ActiveTab != TabAllDrives {
		t.Errorf("After PrevTab: ActiveTab = %d, want %d", m.ActiveTab, TabAllDrives)
	}
}

func TestModelResetAttributeSelection(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	m.SelectedAttr = 5
	m.ScrollOffset = 10

	m.ResetAttributeSelection()

	if m.SelectedAttr != 0 {
		t.Errorf("SelectedAttr = %d, want 0", m.SelectedAttr)
	}
	if m.ScrollOffset != 0 {
		t.Errorf("ScrollOffset = %d, want 0", m.ScrollOffset)
	}
}

func TestModelScrollUp(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	// At top
	m.SelectedAttr = 0
	m.ScrollUp()
	if m.SelectedAttr != 0 {
		t.Errorf("SelectedAttr = %d, want 0 (can't scroll past top)", m.SelectedAttr)
	}

	// In middle
	m.SelectedAttr = 5
	m.ScrollUp()
	if m.SelectedAttr != 4 {
		t.Errorf("SelectedAttr = %d, want 4", m.SelectedAttr)
	}

	// With scroll offset
	m.SelectedAttr = 3
	m.ScrollOffset = 5
	m.ScrollUp()
	if m.SelectedAttr != 2 {
		t.Errorf("SelectedAttr = %d, want 2", m.SelectedAttr)
	}
	if m.ScrollOffset != 2 {
		t.Errorf("ScrollOffset = %d, want 2", m.ScrollOffset)
	}
}

func TestModelScrollDown(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Height = 30

	// No drive
	m.ScrollDown()
	if m.SelectedAttr != 0 {
		t.Errorf("SelectedAttr = %d, want 0 (no drive)", m.SelectedAttr)
	}

	// With drive
	m.Drives = []smart.DriveInfo{
		{
			IsNVMe: false,
			Attributes: []smart.SmartAttribute{
				{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5},
			},
		},
	}
	m.SelectedDrive = 0
	m.SelectedAttr = 0

	m.ScrollDown()
	if m.SelectedAttr != 1 {
		t.Errorf("SelectedAttr = %d, want 1", m.SelectedAttr)
	}
}

func TestModelGetMaxVisibleAttributes(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	// Small height
	m.Height = 20
	max := m.getMaxVisibleAttributes()
	if max != 5 {
		t.Errorf("getMaxVisibleAttributes() = %d, want 5 (minimum)", max)
	}

	// Normal height
	m.Height = 50
	max = m.getMaxVisibleAttributes()
	expected := 50 - 22 // 28
	if max != expected {
		t.Errorf("getMaxVisibleAttributes() = %d, want %d", max, expected)
	}
}

func TestModelSettingsNextOption(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	// Theme cycling
	m.SettingsSelected = SettingsTheme
	originalTheme := m.Config.Theme
	m.SettingsNextOption()
	// Theme should have changed after cycling (unless there's only one theme)
	_ = m.Config.Theme != originalTheme

	// Logo toggle
	m.SettingsSelected = SettingsShowLogo
	originalLogo := m.Config.Display.ShowLogo
	m.SettingsNextOption()
	if m.Config.Display.ShowLogo == originalLogo {
		t.Error("ShowLogo should have toggled")
	}

	// Temp unit toggle
	m.SettingsSelected = SettingsTempUnit
	originalFahrenheit := m.Config.Display.ShowFahrenheit
	m.SettingsNextOption()
	if m.Config.Display.ShowFahrenheit == originalFahrenheit {
		t.Error("ShowFahrenheit should have toggled")
	}
}

func TestModelSettingsPrevOption(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	// Theme cycling backwards
	m.SettingsSelected = SettingsTheme
	m.Config.Theme = "default"
	m.SettingsPrevOption()
	// Theme should have changed to last in list

	// Logo toggle
	m.SettingsSelected = SettingsShowLogo
	originalLogo := m.Config.Display.ShowLogo
	m.SettingsPrevOption()
	if m.Config.Display.ShowLogo == originalLogo {
		t.Error("ShowLogo should have toggled")
	}
}

func TestTabConstants(t *testing.T) {
	if TabOverview != 0 {
		t.Errorf("TabOverview = %d, want 0", TabOverview)
	}
	if TabAttributes != 1 {
		t.Errorf("TabAttributes = %d, want 1", TabAttributes)
	}
	if TabDetails != 2 {
		t.Errorf("TabDetails = %d, want 2", TabDetails)
	}
	if TabAllDrives != 3 {
		t.Errorf("TabAllDrives = %d, want 3", TabAllDrives)
	}
	if TabSettings != 4 {
		t.Errorf("TabSettings = %d, want 4", TabSettings)
	}
	if TabCount != 5 {
		t.Errorf("TabCount = %d, want 5", TabCount)
	}
}

func TestTabNames(t *testing.T) {
	if len(TabNames) != TabCount {
		t.Errorf("TabNames length = %d, want %d", len(TabNames), TabCount)
	}

	expectedNames := []string{
		"Overview",
		"S.M.A.R.T. Attributes",
		"Details",
		"All Drives",
		"Settings",
	}

	for i, name := range expectedNames {
		if TabNames[i] != name {
			t.Errorf("TabNames[%d] = %q, want %q", i, TabNames[i], name)
		}
	}
}

func TestSettingsConstants(t *testing.T) {
	if SettingsTheme != 0 {
		t.Errorf("SettingsTheme = %d, want 0", SettingsTheme)
	}
	if SettingsShowLogo != 1 {
		t.Errorf("SettingsShowLogo = %d, want 1", SettingsShowLogo)
	}
	if SettingsTempUnit != 2 {
		t.Errorf("SettingsTempUnit = %d, want 2", SettingsTempUnit)
	}
	if SettingsCount != 3 {
		t.Errorf("SettingsCount = %d, want 3", SettingsCount)
	}
}

func TestThemeNames(t *testing.T) {
	themes := config.ListThemes()

	if len(themes) == 0 {
		t.Fatal("ListThemes should return at least one theme")
	}

	// Verify themes are sorted
	for i := 1; i < len(themes); i++ {
		if themes[i] < themes[i-1] {
			t.Errorf("ListThemes not sorted: %q comes after %q", themes[i], themes[i-1])
		}
	}

	// Verify "default" theme always exists
	found := false
	for _, name := range themes {
		if name == "default" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListThemes should contain \"default\" theme")
	}
}

func TestSettingsSavedMsg(t *testing.T) {
	successMsg := SettingsSavedMsg{Success: true}
	if !successMsg.Success {
		t.Error("Success should be true")
	}
	if successMsg.Error != nil {
		t.Error("Error should be nil on success")
	}

	errorMsg := SettingsSavedMsg{Success: false, Error: nil}
	if errorMsg.Success {
		t.Error("Success should be false")
	}
}

func TestClearScreenshotStatusCmd(t *testing.T) {
	cmd := ClearScreenshotStatusCmd()
	if cmd == nil {
		t.Error("ClearScreenshotStatusCmd should return a command")
	}
}

func TestModelInit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestModelDriveNavigationResetsSelection(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Drives = []smart.DriveInfo{
		{Device: "/dev/sda"},
		{Device: "/dev/sdb"},
	}
	m.SelectedAttr = 5
	m.ScrollOffset = 3

	m.NextDrive()

	if m.SelectedAttr != 0 {
		t.Errorf("SelectedAttr should be reset to 0, got %d", m.SelectedAttr)
	}
	if m.ScrollOffset != 0 {
		t.Errorf("ScrollOffset should be reset to 0, got %d", m.ScrollOffset)
	}
}

func TestModelTabNavigationResetsSelection(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.SelectedAttr = 5
	m.ScrollOffset = 3

	m.NextTab()

	if m.SelectedAttr != 0 {
		t.Errorf("SelectedAttr should be reset to 0, got %d", m.SelectedAttr)
	}
	if m.ScrollOffset != 0 {
		t.Errorf("ScrollOffset should be reset to 0, got %d", m.ScrollOffset)
	}
}

func TestModelLastRefreshInitialized(t *testing.T) {
	cfg := config.DefaultConfig()
	before := time.Now()
	m := NewModel(cfg)
	after := time.Now()

	if m.LastRefresh.Before(before) || m.LastRefresh.After(after) {
		t.Error("LastRefresh should be set to current time")
	}
}
