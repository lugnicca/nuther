package components

import (
	"strings"
	"testing"
	"time"
)

func TestRenderDemoBanner(t *testing.T) {
	s := newTestStyles()

	result := RenderDemoBanner(80, s)
	if result == "" {
		t.Error("Demo banner should not be empty")
	}
	if !strings.Contains(result, "DEMO") {
		t.Error("Demo banner should contain DEMO")
	}
}

func TestRenderStatusBarWithDemo(t *testing.T) {
	s := newTestStyles()
	now := time.Now()

	result := RenderStatusBar(0, 2, now, true, false, 80, testKeyMap{}, s)
	if result == "" {
		t.Error("Status bar with demo should not be empty")
	}
	if !strings.Contains(result, "[DEMO]") {
		t.Error("Status bar should contain [DEMO] indicator when isDemo is true")
	}
}

func TestRenderStatusBarWithoutDemo(t *testing.T) {
	s := newTestStyles()
	now := time.Now()

	result := RenderStatusBar(0, 2, now, false, false, 80, testKeyMap{}, s)
	if strings.Contains(result, "[DEMO]") {
		t.Error("Status bar should not contain [DEMO] when isDemo is false")
	}
}
