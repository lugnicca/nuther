package ui

import (
	"testing"
	"time"

	"nuther/internal/config"
	"nuther/internal/smart"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateWindowSizeMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := m.Update(msg)

	model := updatedModel.(Model)
	if model.Width != 100 {
		t.Errorf("Width = %d, want 100", model.Width)
	}
	if model.Height != 50 {
		t.Errorf("Height = %d, want 50", model.Height)
	}
	if !model.Ready {
		t.Error("Ready should be true after window size message")
	}
	if cmd != nil {
		t.Error("Command should be nil for window size message")
	}
}

func TestUpdateDrivesLoadedMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Loading = true

	drives := []smart.DriveInfo{
		{Device: "/dev/sda", Model: "Drive1"},
		{Device: "/dev/sdb", Model: "Drive2"},
	}
	msg := DrivesLoadedMsg{Drives: drives, Error: nil}

	updatedModel, _ := m.Update(msg)
	model := updatedModel.(Model)

	if model.Loading {
		t.Error("Loading should be false after drives loaded")
	}
	if len(model.Drives) != 2 {
		t.Errorf("Drives length = %d, want 2", len(model.Drives))
	}
	if model.Error != nil {
		t.Error("Error should be nil")
	}
}

func TestUpdateDrivesLoadedMsgWithError(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Loading = true

	msg := DrivesLoadedMsg{Drives: nil, Error: errMock}

	updatedModel, _ := m.Update(msg)
	model := updatedModel.(Model)

	if model.Loading {
		t.Error("Loading should be false")
	}
	if model.Error == nil {
		t.Error("Error should be set")
	}
	if len(model.Drives) != 0 {
		t.Error("Drives should remain empty when no drives provided")
	}
}

func TestUpdateDrivesLoadedMsgWithDemoFallback(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Loading = true

	// Demo fallback: drives are provided alongside a non-nil error
	drives := []smart.DriveInfo{
		{Device: "demo0", Model: "Demo Drive", IsDemo: true},
	}
	msg := DrivesLoadedMsg{Drives: drives, Error: errMock}

	updatedModel, _ := m.Update(msg)
	model := updatedModel.(Model)

	if model.Loading {
		t.Error("Loading should be false")
	}
	if model.Error == nil {
		t.Error("Error should be set for demo fallback")
	}
	if len(model.Drives) != 1 {
		t.Errorf("Drives should be set from demo data, got %d", len(model.Drives))
	}
}

var errMock = &mockError{}

type mockError struct{}

func (e *mockError) Error() string { return "mock error" }

func TestUpdateDrivesLoadedMsgResetsSelectedDrive(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.SelectedDrive = 5 // Out of range

	drives := []smart.DriveInfo{
		{Device: "/dev/sda"},
	}
	msg := DrivesLoadedMsg{Drives: drives}

	updatedModel, _ := m.Update(msg)
	model := updatedModel.(Model)

	if model.SelectedDrive != 0 {
		t.Errorf("SelectedDrive should be reset to 0, got %d", model.SelectedDrive)
	}
}

func TestUpdateScreenshotMsgSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	msg := ScreenshotMsg{Success: true}

	updatedModel, cmd := m.Update(msg)
	model := updatedModel.(Model)

	if model.ScreenshotStatus != "success" {
		t.Errorf("ScreenshotStatus = %q, want %q", model.ScreenshotStatus, "success")
	}
	if model.ScreenshotMessage != "Screenshot copied to clipboard!" {
		t.Errorf("ScreenshotMessage = %q, want %q", model.ScreenshotMessage, "Screenshot copied to clipboard!")
	}
	if cmd == nil {
		t.Error("Should return a command to clear status")
	}
}

func TestUpdateScreenshotMsgError(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	msg := ScreenshotMsg{Success: false, Error: errMock}

	updatedModel, _ := m.Update(msg)
	model := updatedModel.(Model)

	if model.ScreenshotStatus != "error" {
		t.Errorf("ScreenshotStatus = %q, want %q", model.ScreenshotStatus, "error")
	}
}

func TestUpdateClearScreenshotMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ScreenshotStatus = "success"
	m.ScreenshotMessage = "Test message"
	m.ScreenshotTime = time.Now().Add(-5 * time.Second) // 5 seconds ago

	msg := clearScreenshotMsg{}
	updatedModel, _ := m.Update(msg)
	model := updatedModel.(Model)

	if model.ScreenshotStatus != "" {
		t.Errorf("ScreenshotStatus should be cleared, got %q", model.ScreenshotStatus)
	}
	if model.ScreenshotMessage != "" {
		t.Errorf("ScreenshotMessage should be cleared, got %q", model.ScreenshotMessage)
	}
}

func TestUpdateClearScreenshotMsgTooEarly(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ScreenshotStatus = "success"
	m.ScreenshotMessage = "Test message"
	m.ScreenshotTime = time.Now() // Just now

	msg := clearScreenshotMsg{}
	updatedModel, _ := m.Update(msg)
	model := updatedModel.(Model)

	// Should NOT be cleared if less than 3 seconds have passed
	if model.ScreenshotStatus == "" {
		t.Error("ScreenshotStatus should not be cleared too early")
	}
}

func TestUpdateSettingsSavedMsgSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	msg := SettingsSavedMsg{Success: true}

	updatedModel, cmd := m.Update(msg)
	model := updatedModel.(Model)

	if model.SettingsMessage != "Settings saved!" {
		t.Errorf("SettingsMessage = %q, want %q", model.SettingsMessage, "Settings saved!")
	}
	if cmd == nil {
		t.Error("Should return a command to clear message")
	}
}

func TestUpdateSettingsSavedMsgError(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	msg := SettingsSavedMsg{Success: false, Error: errMock}

	updatedModel, _ := m.Update(msg)
	model := updatedModel.(Model)

	if model.SettingsMessage == "" {
		t.Error("SettingsMessage should contain error")
	}
}

func TestUpdateClearSettingsMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.SettingsMessage = "Some message"

	msg := clearSettingsMsg{}
	updatedModel, _ := m.Update(msg)
	model := updatedModel.(Model)

	if model.SettingsMessage != "" {
		t.Errorf("SettingsMessage should be cleared, got %q", model.SettingsMessage)
	}
}

func TestHandleKeyPressQuit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.handleKeyPress(msg)

	// Quit should return tea.Quit command
	if cmd == nil {
		t.Error("Quit should return a command")
	}
}

func TestHandleKeyPressTab(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ActiveTab = TabOverview

	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := m.handleKeyPress(msg)
	model := updatedModel.(Model)

	if model.ActiveTab != TabAttributes {
		t.Errorf("ActiveTab = %d, want %d", model.ActiveTab, TabAttributes)
	}
}

func TestHandleKeyPressShiftTab(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ActiveTab = TabAttributes

	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	updatedModel, _ := m.handleKeyPress(msg)
	model := updatedModel.(Model)

	if model.ActiveTab != TabOverview {
		t.Errorf("ActiveTab = %d, want %d", model.ActiveTab, TabOverview)
	}
}

func TestHandleKeyPressNextDrive(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Drives = []smart.DriveInfo{{}, {}}
	m.SelectedDrive = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ := m.handleKeyPress(msg)
	model := updatedModel.(Model)

	if model.SelectedDrive != 1 {
		t.Errorf("SelectedDrive = %d, want 1", model.SelectedDrive)
	}
}

func TestHandleKeyPressPrevDrive(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Drives = []smart.DriveInfo{{}, {}}
	m.SelectedDrive = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	updatedModel, _ := m.handleKeyPress(msg)
	model := updatedModel.(Model)

	if model.SelectedDrive != 0 {
		t.Errorf("SelectedDrive = %d, want 0", model.SelectedDrive)
	}
}

func TestHandleKeyPressRefresh(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Loading = false
	m.LastRefresh = time.Time{} // Zero time — cache expired

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, cmd := m.handleKeyPress(msg)
	model := updatedModel.(Model)

	if !model.Loading {
		t.Error("Loading should be true after refresh with expired cache")
	}
	if cmd == nil {
		t.Error("Refresh should return a command when cache expired")
	}
}

func TestHandleKeyPressRefreshWithCache(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Loading = false
	m.Drives = []smart.DriveInfo{{Device: "/dev/sda"}}
	m.LastRefresh = time.Now() // Fresh cache

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, cmd := m.handleKeyPress(msg)
	model := updatedModel.(Model)

	if model.Loading {
		t.Error("Loading should remain false when cache is fresh")
	}
	if cmd != nil {
		t.Error("Refresh should not return a command when cache is fresh")
	}
}

func TestHandleKeyPressRefreshCacheExpired(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Loading = false
	m.Drives = []smart.DriveInfo{{Device: "/dev/sda"}}
	m.LastRefresh = time.Now().Add(-2 * time.Minute) // Expired cache

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, cmd := m.handleKeyPress(msg)
	model := updatedModel.(Model)

	if !model.Loading {
		t.Error("Loading should be true when cache is expired")
	}
	if cmd == nil {
		t.Error("Refresh should return a command when cache is expired")
	}
}

func TestHandleKeyPressForceRefresh(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Loading = false
	m.Drives = []smart.DriveInfo{{Device: "/dev/sda"}}
	m.LastRefresh = time.Now() // Fresh cache — but force should bypass

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
	updatedModel, cmd := m.handleKeyPress(msg)
	model := updatedModel.(Model)

	if !model.Loading {
		t.Error("Loading should be true after force refresh even with fresh cache")
	}
	if cmd == nil {
		t.Error("Force refresh should always return a command")
	}
}

func TestIsCacheFresh(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name     string
		setup    func(m *Model)
		expected bool
	}{
		{
			"fresh cache with drives",
			func(m *Model) {
				m.Drives = []smart.DriveInfo{{Device: "/dev/sda"}}
				m.LastRefresh = time.Now()
			},
			true,
		},
		{
			"zero time",
			func(m *Model) {
				m.Drives = []smart.DriveInfo{{Device: "/dev/sda"}}
				m.LastRefresh = time.Time{}
			},
			false,
		},
		{
			"expired cache",
			func(m *Model) {
				m.Drives = []smart.DriveInfo{{Device: "/dev/sda"}}
				m.LastRefresh = time.Now().Add(-2 * time.Minute)
			},
			false,
		},
		{
			"no drives",
			func(m *Model) {
				m.Drives = nil
				m.LastRefresh = time.Now()
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(cfg)
			tt.setup(&m)
			if got := m.isCacheFresh(); got != tt.expected {
				t.Errorf("isCacheFresh() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHandleKeyPressHelp(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ShowHelp = false

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel, _ := m.handleKeyPress(msg)
	model := updatedModel.(Model)

	if !model.ShowHelp {
		t.Error("ShowHelp should be true")
	}

	// Toggle off
	updatedModel, _ = model.handleKeyPress(msg)
	model = updatedModel.(Model)

	if model.ShowHelp {
		t.Error("ShowHelp should be false after second press")
	}
}

func TestHandleKeyPressUpDown(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.Height = 50
	m.Drives = []smart.DriveInfo{
		{
			Attributes: []smart.SmartAttribute{{}, {}, {}},
		},
	}
	m.SelectedAttr = 1

	// Up
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.handleKeyPress(upMsg)
	model := updatedModel.(Model)

	if model.SelectedAttr != 0 {
		t.Errorf("SelectedAttr = %d, want 0", model.SelectedAttr)
	}

	// Down
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = model.handleKeyPress(downMsg)
	model = updatedModel.(Model)

	if model.SelectedAttr != 1 {
		t.Errorf("SelectedAttr = %d, want 1", model.SelectedAttr)
	}
}

func TestHandleKeyPressUpDownInSettings(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ActiveTab = TabSettings
	m.SettingsSelected = 1

	// Up
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.handleKeyPress(upMsg)
	model := updatedModel.(Model)

	if model.SettingsSelected != 0 {
		t.Errorf("SettingsSelected = %d, want 0", model.SettingsSelected)
	}

	// Can't go below 0
	updatedModel, _ = model.handleKeyPress(upMsg)
	model = updatedModel.(Model)
	if model.SettingsSelected != 0 {
		t.Errorf("SettingsSelected should stay at 0, got %d", model.SettingsSelected)
	}

	// Down
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = model.handleKeyPress(downMsg)
	model = updatedModel.(Model)

	if model.SettingsSelected != 1 {
		t.Errorf("SettingsSelected = %d, want 1", model.SettingsSelected)
	}
}

func TestHandleKeyPressLeftRightInSettings(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ActiveTab = TabSettings
	m.SettingsSelected = SettingsShowLogo

	originalValue := m.Config.Display.ShowLogo

	// Right
	rightMsg := tea.KeyMsg{Type: tea.KeyRight}
	updatedModel, _ := m.handleKeyPress(rightMsg)
	model := updatedModel.(Model)

	if model.Config.Display.ShowLogo == originalValue {
		t.Error("ShowLogo should have changed with Right key")
	}

	// Left
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft}
	updatedModel, _ = model.handleKeyPress(leftMsg)
	model = updatedModel.(Model)

	if model.Config.Display.ShowLogo != originalValue {
		t.Error("ShowLogo should have toggled back with Left key")
	}
}

func TestHandleKeyPressEnterInSettings(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ActiveTab = TabSettings

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := m.handleKeyPress(enterMsg)

	if cmd == nil {
		t.Error("Enter in settings should return a command")
	}
}

func TestHandleKeyPressEnterNotInSettings(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ActiveTab = TabOverview

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := m.handleKeyPress(enterMsg)

	if cmd != nil {
		t.Error("Enter outside settings should not return a command")
	}
}
