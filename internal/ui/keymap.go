package ui

import (
	"nuther/internal/config"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all key bindings
type KeyMap struct {
	Quit         key.Binding
	Tab          key.Binding
	ShiftTab     key.Binding
	NextDrive    key.Binding
	PrevDrive    key.Binding
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	Refresh      key.Binding
	ForceRefresh key.Binding
	Help         key.Binding
	Screenshot   key.Binding
	Enter        key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		NextDrive: key.NewBinding(
			key.WithKeys("n", "]"),
			key.WithHelp("n/]", "next drive"),
		),
		PrevDrive: key.NewBinding(
			key.WithKeys("p", "["),
			key.WithHelp("p/[", "prev drive"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "scroll up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "scroll down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "right"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		ForceRefresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "force refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Screenshot: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "screenshot"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
	}
}

// NewKeyMapFromConfig creates a KeyMap using config keybindings as primary keys,
// keeping secondary keys (ctrl+c, vim keys, arrows) for convenience
func NewKeyMapFromConfig(cfg *config.Config) KeyMap {
	km := DefaultKeyMap()
	kb := cfg.Keybindings

	if kb.Quit != "" {
		km.Quit = key.NewBinding(
			key.WithKeys(kb.Quit, "ctrl+c"),
			key.WithHelp(kb.Quit, "quit"),
		)
	}
	if kb.Tab != "" {
		km.Tab = key.NewBinding(
			key.WithKeys(kb.Tab),
			key.WithHelp(kb.Tab, "next tab"),
		)
	}
	if kb.NextDrive != "" {
		km.NextDrive = key.NewBinding(
			key.WithKeys(kb.NextDrive, "]"),
			key.WithHelp(kb.NextDrive+"/]", "next drive"),
		)
	}
	if kb.PrevDrive != "" {
		km.PrevDrive = key.NewBinding(
			key.WithKeys(kb.PrevDrive, "["),
			key.WithHelp(kb.PrevDrive+"/[", "prev drive"),
		)
	}
	if kb.Refresh != "" {
		km.Refresh = key.NewBinding(
			key.WithKeys(kb.Refresh),
			key.WithHelp(kb.Refresh, "refresh"),
		)
	}
	if kb.Help != "" {
		km.Help = key.NewBinding(
			key.WithKeys(kb.Help),
			key.WithHelp(kb.Help, "help"),
		)
	}
	if kb.Screenshot != "" {
		km.Screenshot = key.NewBinding(
			key.WithKeys(kb.Screenshot),
			key.WithHelp(kb.Screenshot, "screenshot"),
		)
	}

	return km
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.NextDrive, k.Up, k.Refresh, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.NextDrive, k.PrevDrive},
		{k.Up, k.Down, k.Refresh, k.ForceRefresh},
		{k.Screenshot, k.Help, k.Quit},
	}
}
