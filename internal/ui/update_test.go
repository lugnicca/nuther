package ui

import (
	"testing"
	"time"

	"nuther/internal/config"
	"nuther/internal/smart"
	"nuther/internal/smartwatch"

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

func TestHandleKeyPressEnterOnScreenshotDirEntersEditMode(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Screenshot.Dir = "/home/user/pics"
	m := NewModel(cfg)
	m.ActiveTab = TabSettings
	m.SettingsSelected = SettingsScreenshotDir

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := m.handleKeyPress(enterMsg)
	model := updatedModel.(Model)

	if cmd != nil {
		t.Error("Entering edit mode should not return a command")
	}
	if !model.SettingsEditingDir {
		t.Error("SettingsEditingDir should be true after Enter on dir row")
	}
	if model.SettingsDirBuffer != "/home/user/pics" {
		t.Errorf("SettingsDirBuffer = %q, want %q", model.SettingsDirBuffer, "/home/user/pics")
	}
}

func TestHandleKeyPressScreenshotDirEditFlow(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Screenshot.Dir = "/old"
	m := NewModel(cfg)
	m.ActiveTab = TabSettings
	m.SettingsSelected = SettingsScreenshotDir
	m.SettingsEditingDir = true
	m.SettingsDirBuffer = "/old"

	// Printable runes append to the buffer
	runesMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-', 'n', 'e', 'w'}}
	updatedModel, _ := m.handleKeyPress(runesMsg)
	model := updatedModel.(Model)
	if model.SettingsDirBuffer != "/old-new" {
		t.Fatalf("buffer after typing = %q, want %q", model.SettingsDirBuffer, "/old-new")
	}

	// Backspace deletes the last rune
	bsMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ = model.handleKeyPress(bsMsg)
	model = updatedModel.(Model)
	if model.SettingsDirBuffer != "/old-ne" {
		t.Fatalf("buffer after backspace = %q, want %q", model.SettingsDirBuffer, "/old-ne")
	}

	// Esc cancels and discards the buffer
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = model.handleKeyPress(escMsg)
	model = updatedModel.(Model)
	if model.SettingsEditingDir {
		t.Error("SettingsEditingDir should be false after Esc")
	}
	if model.Config.Screenshot.Dir != "/old" {
		t.Errorf("Esc should not commit, Config.Screenshot.Dir = %q, want %q", model.Config.Screenshot.Dir, "/old")
	}
}

func TestHandleKeyPressScreenshotDirEnterCommits(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Screenshot.Dir = "/old"
	m := NewModel(cfg)
	m.ActiveTab = TabSettings
	m.SettingsSelected = SettingsScreenshotDir
	m.SettingsEditingDir = true
	m.SettingsDirBuffer = "/new/dir"

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := m.handleKeyPress(enterMsg)
	model := updatedModel.(Model)

	if cmd != nil {
		t.Error("Committing the dir should not save yet (no command)")
	}
	if model.SettingsEditingDir {
		t.Error("SettingsEditingDir should be false after commit")
	}
	if model.Config.Screenshot.Dir != "/new/dir" {
		t.Errorf("Config.Screenshot.Dir = %q, want %q", model.Config.Screenshot.Dir, "/new/dir")
	}
}

func TestHandleKeyPressScreenshotDirUpCommitsAndMoves(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Screenshot.Dir = "/old"
	m := NewModel(cfg)
	m.ActiveTab = TabSettings
	m.SettingsSelected = SettingsScreenshotDir
	m.SettingsEditingDir = true
	m.SettingsDirBuffer = "/new/dir"

	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.handleKeyPress(upMsg)
	model := updatedModel.(Model)

	if model.SettingsEditingDir {
		t.Error("SettingsEditingDir should be false after navigation")
	}
	if model.Config.Screenshot.Dir != "/new/dir" {
		t.Errorf("navigation should commit, Config.Screenshot.Dir = %q", model.Config.Screenshot.Dir)
	}
	if model.SettingsSelected != SettingsTempUnit {
		t.Errorf("SettingsSelected = %d, want %d (moved up)", model.SettingsSelected, SettingsTempUnit)
	}
}

func TestHandleKeyPressEnterInSnapshotsOpensSelectedSnapshot(t *testing.T) {
	store := smartwatch.NewStore(t.TempDir())
	now := time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC)
	olderDrive := smart.DriveInfo{Device: "/dev/sda", Model: "Older Archive", Serial: "OLD123", HealthStatus: smart.HealthGood}
	newerDrive := smart.DriveInfo{Device: "/dev/sdb", Model: "Selected Archive", Serial: "NEW123", HealthStatus: smart.HealthCaution}
	olderRecord, err := store.SaveSnapshot(now, smartwatch.ReasonStartup, olderDrive)
	if err != nil {
		t.Fatalf("SaveSnapshot older: %v", err)
	}
	newerRecord, err := store.SaveSnapshot(now.Add(time.Hour), smartwatch.ReasonManual, newerDrive)
	if err != nil {
		t.Fatalf("SaveSnapshot newer: %v", err)
	}
	index, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}

	cfg := config.DefaultConfig()
	m := NewModel(cfg)
	m.ActiveTab = TabSnapshots
	m.SnapshotStore = store.Dir()
	m.SnapshotIndex = index
	m.SelectedSnapshot = 0 // newest first in the UI

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := m.handleKeyPress(enterMsg)
	if cmd == nil {
		t.Fatal("Enter in snapshots should return an open snapshot command")
	}
	msg := cmd()
	opened, ok := msg.(SnapshotOpenedMsg)
	if !ok {
		t.Fatalf("command returned %T, want SnapshotOpenedMsg", msg)
	}
	if opened.Error != nil {
		t.Fatalf("open snapshot error: %v", opened.Error)
	}
	if opened.Snapshot.ID != newerRecord.ID || opened.Snapshot.ID == olderRecord.ID {
		t.Fatalf("opened snapshot ID = %q, want newest selected %q", opened.Snapshot.ID, newerRecord.ID)
	}

	modelAfterOpen, _ := updatedModel.(Model).Update(opened)
	openedModel := modelAfterOpen.(Model)
	if openedModel.ActiveTab != TabOverview {
		t.Fatalf("ActiveTab after opening snapshot = %d, want %d", openedModel.ActiveTab, TabOverview)
	}
	if !openedModel.ViewingSnapshot {
		t.Fatal("ViewingSnapshot should be true after opening archived snapshot")
	}
	if len(openedModel.Drives) != 1 || openedModel.Drives[0].Model != newerDrive.Model {
		t.Fatalf("opened drive = %+v, want %q", openedModel.Drives, newerDrive.Model)
	}
}
