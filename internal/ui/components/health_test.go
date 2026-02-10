package components

import (
	"strings"
	"testing"

	"nuther/internal/smart"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderStyledStatus(t *testing.T) {
	result := RenderStyledStatus("GOOD", lipgloss.Color("#00ff00"))
	if result == "" {
		t.Error("RenderStyledStatus should not return empty string")
	}
	if !strings.Contains(result, "GOOD") {
		t.Error("RenderStyledStatus should contain the text")
	}
}

func TestRenderHealthBadge(t *testing.T) {
	result := RenderHealthBadge(smart.HealthGood, "●", lipgloss.Color("#00ff00"), 6)
	if result == "" {
		t.Error("RenderHealthBadge should not return empty string")
	}
	if !strings.Contains(result, "GOOD") {
		t.Error("RenderHealthBadge should contain the status text")
	}
}
