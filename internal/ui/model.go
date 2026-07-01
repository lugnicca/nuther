package ui

import (
	"os"
	"path/filepath"
	"time"

	"nuther/internal/config"
	"nuther/internal/smart"
	"nuther/internal/smartwatch"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Tab constants
const (
	TabOverview   = 0
	TabAttributes = 1
	TabDetails    = 2
	TabAllDrives  = 3
	TabSectorGrid = 4
	TabSnapshots  = 5
	TabSettings   = 6
	TabCount      = 7
)

// UIChrome is the vertical space consumed by header + tabs + drive selector + status bar + margins
const UIChrome = 22

// TabNames contains the names of all tabs
var TabNames = []string{
	"Overview",
	"S.M.A.R.T. Attributes",
	"Details",
	"All Drives",
	"Sector Grid",
	"Snapshots",
	"Settings",
}

// Settings option indices
const (
	SettingsTheme = iota
	SettingsShowLogo
	SettingsTempUnit
	SettingsCount
)

// Model is the main Bubble Tea model
type Model struct {
	// Drive data
	Drives        []smart.DriveInfo
	SelectedDrive int

	// UI state
	ActiveTab        int
	SelectedAttr     int
	SelectedSnapshot int
	ScrollOffset     int
	ShowHelp         bool

	// Terminal dimensions
	Width  int
	Height int

	// Runtime state
	Ready           bool
	Loading         bool
	LastRefresh     time.Time
	CacheDuration   time.Duration
	Error           error
	SnapshotIndex   smartwatch.Index
	SnapshotError   error
	SnapshotStore   string
	ViewingSnapshot bool

	// Screenshot state
	ScreenshotStatus  string // "", "capturing", "success", "error"
	ScreenshotMessage string
	ScreenshotTime    time.Time

	// Settings state
	SettingsSelected int // Which setting is selected (0=theme, 1=show_logo, etc.)
	SettingsMessage  string

	// Bubble Tea components
	Spinner spinner.Model
	Help    help.Model
	KeyMap  KeyMap

	// Configuration
	Config *config.Config
	Styles *styles.Styles

	// Tab names
	Tabs []string
}

// NewModel creates a new model with default values
func NewModel(cfg *config.Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		Drives:           make([]smart.DriveInfo, 0),
		SelectedDrive:    0,
		ActiveTab:        TabOverview,
		SelectedAttr:     0,
		SelectedSnapshot: 0,
		ScrollOffset:     0,
		ShowHelp:         false,
		Ready:            false,
		Loading:          true,
		LastRefresh:      time.Now(),
		CacheDuration:    60 * time.Second,
		SnapshotStore:    defaultSnapshotStore(),
		ViewingSnapshot:  false,
		ScreenshotStatus: "",
		Spinner:          s,
		Help:             help.New(),
		KeyMap:           NewKeyMapFromConfig(cfg),
		Config:           cfg,
		Styles:           styles.NewStyles(cfg),
		Tabs:             TabNames,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		LoadDrivesCmd(),
		LoadSnapshotsCmd(m.SnapshotStore),
	)
}

func defaultSnapshotStore() string {
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil || home == "" {
			return filepath.Join(".", "smart-snapshots")
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "nuther", "smart-snapshots")
}

// isCacheFresh returns true if cached drive data is still valid
func (m *Model) isCacheFresh() bool {
	return len(m.Drives) > 0 && time.Since(m.LastRefresh) < m.CacheDuration
}

// GetCurrentDrive returns the currently selected drive, or nil if none
func (m *Model) GetCurrentDrive() *smart.DriveInfo {
	if len(m.Drives) == 0 || m.SelectedDrive >= len(m.Drives) {
		return nil
	}
	return &m.Drives[m.SelectedDrive]
}

// NextDrive selects the next drive in the list
func (m *Model) NextDrive() {
	if len(m.Drives) > 0 {
		m.SelectedDrive = (m.SelectedDrive + 1) % len(m.Drives)
		m.ResetAttributeSelection()
	}
}

// PrevDrive selects the previous drive in the list
func (m *Model) PrevDrive() {
	if len(m.Drives) > 0 {
		m.SelectedDrive = (m.SelectedDrive - 1 + len(m.Drives)) % len(m.Drives)
		m.ResetAttributeSelection()
	}
}

// NextTab switches to the next tab
func (m *Model) NextTab() {
	m.ActiveTab = (m.ActiveTab + 1) % TabCount
	m.ResetAttributeSelection()
}

// PrevTab switches to the previous tab
func (m *Model) PrevTab() {
	m.ActiveTab = (m.ActiveTab - 1 + TabCount) % TabCount
	m.ResetAttributeSelection()
}

// ResetAttributeSelection resets the attribute selection state
func (m *Model) ResetAttributeSelection() {
	m.SelectedAttr = 0
	m.ScrollOffset = 0
}

// ScrollUp moves the selection up in the attribute list
func (m *Model) ScrollUp() {
	if m.ActiveTab == TabSnapshots {
		if m.SelectedSnapshot > 0 {
			m.SelectedSnapshot--
		}
		return
	}

	if m.SelectedAttr > 0 {
		m.SelectedAttr--
		if m.SelectedAttr < m.ScrollOffset {
			m.ScrollOffset = m.SelectedAttr
		}
	}
}

// ScrollDown moves the selection down in the attribute list
func (m *Model) ScrollDown() {
	if m.ActiveTab == TabSnapshots {
		if m.SelectedSnapshot < len(m.SnapshotIndex.Snapshots)-1 {
			m.SelectedSnapshot++
		}
		return
	}

	drive := m.GetCurrentDrive()
	if drive == nil {
		return
	}

	maxAttr := drive.GetAttributeCount()
	maxVisible := m.getMaxVisibleAttributes()

	if m.SelectedAttr < maxAttr-1 {
		m.SelectedAttr++
		if m.SelectedAttr >= m.ScrollOffset+maxVisible {
			m.ScrollOffset = m.SelectedAttr - maxVisible + 1
		}
	}
}

// getMaxVisibleAttributes returns the maximum number of visible attributes
func (m *Model) getMaxVisibleAttributes() int {
	maxVisible := m.Height - UIChrome
	if maxVisible < 5 {
		maxVisible = 5
	}
	return maxVisible
}

// ClearScreenshotStatus clears the screenshot status after a delay
func ClearScreenshotStatusCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearScreenshotMsg{}
	})
}

type clearScreenshotMsg struct{}

// Settings methods

// SettingsNextOption cycles through options for the selected setting
func (m *Model) SettingsNextOption() {
	themeNames := config.ListThemes()
	switch m.SettingsSelected {
	case SettingsTheme:
		// Cycle through themes
		currentIdx := 0
		for i, name := range themeNames {
			if name == m.Config.Theme {
				currentIdx = i
				break
			}
		}
		nextIdx := (currentIdx + 1) % len(themeNames)
		m.Config.Theme = themeNames[nextIdx]
		// Apply theme colors
		theme := config.GetTheme(m.Config.Theme)
		m.Config.Colors = theme.Colors
		m.Styles = styles.NewStyles(m.Config)
	case SettingsShowLogo:
		m.Config.Display.ShowLogo = !m.Config.Display.ShowLogo
	case SettingsTempUnit:
		m.Config.Display.ShowFahrenheit = !m.Config.Display.ShowFahrenheit
		m.Styles = styles.NewStyles(m.Config)
	}
}

// SettingsPrevOption cycles backwards through options for the selected setting
func (m *Model) SettingsPrevOption() {
	themeNames := config.ListThemes()
	switch m.SettingsSelected {
	case SettingsTheme:
		// Cycle through themes backwards
		currentIdx := 0
		for i, name := range themeNames {
			if name == m.Config.Theme {
				currentIdx = i
				break
			}
		}
		prevIdx := (currentIdx - 1 + len(themeNames)) % len(themeNames)
		m.Config.Theme = themeNames[prevIdx]
		// Apply theme colors
		theme := config.GetTheme(m.Config.Theme)
		m.Config.Colors = theme.Colors
		m.Styles = styles.NewStyles(m.Config)
	case SettingsShowLogo:
		m.Config.Display.ShowLogo = !m.Config.Display.ShowLogo
	case SettingsTempUnit:
		m.Config.Display.ShowFahrenheit = !m.Config.Display.ShowFahrenheit
		m.Styles = styles.NewStyles(m.Config)
	}
}

// SettingsApply saves the current settings to config file
func (m *Model) SettingsApply() tea.Cmd {
	return func() tea.Msg {
		err := m.Config.Save()
		if err != nil {
			return SettingsSavedMsg{Success: false, Error: err}
		}
		return SettingsSavedMsg{Success: true}
	}
}

// SettingsSavedMsg is sent when settings are saved
type SettingsSavedMsg struct {
	Success bool
	Error   error
}
