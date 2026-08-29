package ui

import (
	"fmt"
	"time"

	"nuther/internal/smart"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type clearSettingsMsg struct{}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true
		return m, nil

	case DrivesLoadedMsg:
		m.Loading = false
		m.Error = msg.Error
		if len(msg.Drives) > 0 {
			m.Drives = msg.Drives
			m.LastRefresh = time.Now()
			if m.SelectedDrive >= len(m.Drives) {
				m.SelectedDrive = 0
			}
		}
		return m, nil

	case SnapshotsLoadedMsg:
		m.SnapshotError = msg.Error
		if msg.Error == nil {
			m.SnapshotIndex = msg.Index
			if m.SelectedSnapshot >= len(m.SnapshotIndex.Snapshots) {
				m.SelectedSnapshot = 0
			}
		}
		return m, nil

	case SnapshotOpenedMsg:
		m.SnapshotError = msg.Error
		if msg.Error == nil {
			m.Drives = []smart.DriveInfo{msg.Snapshot.SMART}
			m.SelectedDrive = 0
			m.SelectedAttr = 0
			m.ScrollOffset = 0
			m.ActiveTab = TabOverview
			m.ViewingSnapshot = true
			m.LastRefresh = msg.Snapshot.Timestamp
		}
		return m, nil

	case ScreenshotMsg:
		if msg.Success {
			m.ScreenshotStatus = "success"
			if msg.Path != "" {
				m.ScreenshotMessage = fmt.Sprintf("Overview image saved: %s", msg.Path)
			} else {
				m.ScreenshotMessage = "Screenshot copied to clipboard!"
			}
		} else {
			m.ScreenshotStatus = "error"
			if msg.Error != nil {
				m.ScreenshotMessage = fmt.Sprintf("Screenshot failed: %v", msg.Error)
			} else {
				m.ScreenshotMessage = "Screenshot failed"
			}
		}
		m.ScreenshotTime = time.Now()
		return m, ClearScreenshotStatusCmd()

	case clearScreenshotMsg:
		// Clear screenshot status after 3 seconds
		if time.Since(m.ScreenshotTime) >= 3*time.Second {
			m.ScreenshotStatus = ""
			m.ScreenshotMessage = ""
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case SettingsSavedMsg:
		if msg.Success {
			m.SettingsMessage = "Settings saved!"
		} else {
			m.SettingsMessage = fmt.Sprintf("Failed to save: %v", msg.Error)
		}
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearSettingsMsg{}
		})

	case clearSettingsMsg:
		m.SettingsMessage = ""
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ActiveTab == TabSettings && m.SettingsEditingDir {
		switch msg.Type {
		case tea.KeyEsc:
			// Cancel and discard the working buffer
			m.SettingsEditingDir = false
			return m, nil

		case tea.KeyEnter:
			// Commit the buffer to config; a later Enter elsewhere saves it
			m.Config.Screenshot.Dir = m.SettingsDirBuffer
			m.SettingsEditingDir = false
			return m, nil

		case tea.KeyBackspace:
			runes := []rune(m.SettingsDirBuffer)
			if len(runes) > 0 {
				m.SettingsDirBuffer = string(runes[:len(runes)-1])
			}
			return m, nil

		case tea.KeyUp, tea.KeyDown:
			// Commit on navigation, then fall through to normal selection movement
			m.Config.Screenshot.Dir = m.SettingsDirBuffer
			m.SettingsEditingDir = false

		case tea.KeyRunes:
			m.SettingsDirBuffer += string(msg.Runes)
			return m, nil

		default:
			// Any other key (tab, arrows, etc.) cancels the edit and is
			// handled normally below
			m.SettingsEditingDir = false
		}
	}

	switch {
	case key.Matches(msg, m.KeyMap.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.KeyMap.Tab):
		m.ViewingSnapshot = false
		m.NextTab()
		return m, nil

	case key.Matches(msg, m.KeyMap.ShiftTab):
		m.ViewingSnapshot = false
		m.PrevTab()
		return m, nil

	case key.Matches(msg, m.KeyMap.NextDrive):
		m.NextDrive()
		return m, nil

	case key.Matches(msg, m.KeyMap.PrevDrive):
		m.PrevDrive()
		return m, nil

	case key.Matches(msg, m.KeyMap.Up):
		if m.ActiveTab == TabSettings {
			if m.SettingsSelected > 0 {
				m.SettingsSelected--
			}
		} else {
			m.ScrollUp()
		}
		return m, nil

	case key.Matches(msg, m.KeyMap.Down):
		if m.ActiveTab == TabSettings {
			if m.SettingsSelected < SettingsCount-1 {
				m.SettingsSelected++
			}
		} else {
			m.ScrollDown()
		}
		return m, nil

	case key.Matches(msg, m.KeyMap.ForceRefresh):
		m.Loading = true
		m.ViewingSnapshot = false
		return m, tea.Batch(LoadDrivesCmd(), LoadSnapshotsCmd(m.SnapshotStore))

	case key.Matches(msg, m.KeyMap.Refresh):
		if m.isCacheFresh() {
			return m, nil
		}
		m.Loading = true
		m.ViewingSnapshot = false
		return m, tea.Batch(LoadDrivesCmd(), LoadSnapshotsCmd(m.SnapshotStore))

	case key.Matches(msg, m.KeyMap.Help):
		m.ShowHelp = !m.ShowHelp
		return m, nil

	case key.Matches(msg, m.KeyMap.Screenshot):
		m.ScreenshotStatus = "capturing"
		if m.ActiveTab == TabOverview {
			m.ScreenshotMessage = "Saving overview image..."
			if drive := m.GetCurrentDrive(); drive != nil {
				return m, SaveOverviewImageCmd(*drive, m.SelectedAttr, m.ScrollOffset, m.Width, m.Height, m.Styles, m.Config)
			}
		}
		m.ScreenshotMessage = "Capturing..."
		return m, CaptureScreenshotCmd()

	case key.Matches(msg, m.KeyMap.Left):
		if m.ActiveTab == TabSettings {
			m.SettingsPrevOption()
		}
		return m, nil

	case key.Matches(msg, m.KeyMap.Right):
		if m.ActiveTab == TabSettings {
			m.SettingsNextOption()
		}
		return m, nil

	case key.Matches(msg, m.KeyMap.Enter):
		if m.ActiveTab == TabSettings {
			if m.SettingsSelected == SettingsScreenshotDir {
				// Enter toggles text-edit mode for the screenshot dir row
				m.SettingsEditingDir = true
				m.SettingsDirBuffer = m.Config.Screenshot.Dir
				return m, nil
			}
			return m, m.SettingsApply()
		}
		if m.ActiveTab == TabSnapshots {
			return m, OpenSelectedSnapshotCmd(m.SnapshotStore, m.SnapshotIndex, m.SelectedSnapshot)
		}
		return m, nil
	}

	return m, nil
}
