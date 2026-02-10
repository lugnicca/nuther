package ui

import (
	"testing"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test that all bindings are set
	bindings := []struct {
		name    string
		binding interface{}
	}{
		{"Quit", km.Quit},
		{"Tab", km.Tab},
		{"ShiftTab", km.ShiftTab},
		{"NextDrive", km.NextDrive},
		{"PrevDrive", km.PrevDrive},
		{"Up", km.Up},
		{"Down", km.Down},
		{"Left", km.Left},
		{"Right", km.Right},
		{"Refresh", km.Refresh},
		{"Help", km.Help},
		{"Screenshot", km.Screenshot},
		{"Enter", km.Enter},
	}

	for _, b := range bindings {
		if b.binding == nil {
			t.Errorf("KeyMap.%s is nil", b.name)
		}
	}
}

func TestKeyMapShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	shortHelp := km.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("ShortHelp should return bindings")
	}

	// Should include essential bindings
	if len(shortHelp) != 6 {
		t.Errorf("ShortHelp returned %d bindings, want 6", len(shortHelp))
	}
}

func TestKeyMapFullHelp(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("FullHelp should return binding groups")
	}

	if len(fullHelp) != 3 {
		t.Errorf("FullHelp returned %d groups, want 3", len(fullHelp))
	}

	// Each group should have bindings
	for i, group := range fullHelp {
		if len(group) == 0 {
			t.Errorf("FullHelp group %d is empty", i)
		}
	}
}

func TestKeyMapImplementsHelpKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test interface compliance by calling methods
	shortHelp := km.ShortHelp()
	if shortHelp == nil {
		t.Error("ShortHelp should not return nil")
	}

	fullHelp := km.FullHelp()
	if fullHelp == nil {
		t.Error("FullHelp should not return nil")
	}
}
