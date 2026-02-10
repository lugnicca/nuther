package config

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetThemeExistingThemes(t *testing.T) {
	themes := []string{"atomic-tangerine", "catppuccin", "default", "dracula", "everforest", "gruvbox", "nord", "rose-petale", "solarized-dark", "sous-bois", "tokyo-night"}

	for _, themeName := range themes {
		t.Run(themeName, func(t *testing.T) {
			theme := GetTheme(themeName)

			if theme.Name == "" {
				t.Errorf("Theme %q has empty Name", themeName)
			}

			// Verify colors are set
			colors := theme.Colors
			if colors.AccentPrimary == "" {
				t.Errorf("Theme %q missing AccentPrimary", themeName)
			}
			if colors.AccentSecondary == "" {
				t.Errorf("Theme %q missing AccentSecondary", themeName)
			}
			if colors.Success == "" {
				t.Errorf("Theme %q missing Success", themeName)
			}
			if colors.Warning == "" {
				t.Errorf("Theme %q missing Warning", themeName)
			}
			if colors.Danger == "" {
				t.Errorf("Theme %q missing Danger", themeName)
			}
			if colors.Info == "" {
				t.Errorf("Theme %q missing Info", themeName)
			}
			if colors.Text == "" {
				t.Errorf("Theme %q missing Text", themeName)
			}
			if colors.TextDim == "" {
				t.Errorf("Theme %q missing TextDim", themeName)
			}
			if colors.Background == "" {
				t.Errorf("Theme %q missing Background", themeName)
			}
			if colors.Border == "" {
				t.Errorf("Theme %q missing Border", themeName)
			}
		})
	}
}

func TestGetThemeUnknownFallsBackToDefault(t *testing.T) {
	unknownThemes := []string{"unknown", "invalid", "fake", "", "NORD", "Default"}

	defaultTheme := GetTheme("default")

	for _, themeName := range unknownThemes {
		t.Run(themeName, func(t *testing.T) {
			theme := GetTheme(themeName)

			// Should fall back to default
			if theme.Colors.AccentPrimary != defaultTheme.Colors.AccentPrimary {
				t.Errorf("Unknown theme %q should fall back to default, got AccentPrimary %q, want %q",
					themeName, theme.Colors.AccentPrimary, defaultTheme.Colors.AccentPrimary)
			}
		})
	}
}

func TestListThemes(t *testing.T) {
	themes := ListThemes()

	if len(themes) == 0 {
		t.Fatal("ListThemes() returned empty slice")
	}

	// Check expected themes are present
	expectedThemes := map[string]bool{
		"atomic-tangerine": true,
		"catppuccin":       true,
		"default":          true,
		"dracula":          true,
		"everforest":       true,
		"gruvbox":          true,
		"nord":             true,
		"rose-petale":      true,
		"solarized-dark":   true,
		"sous-bois":        true,
		"tokyo-night":      true,
	}

	foundThemes := make(map[string]bool)
	for _, name := range themes {
		foundThemes[name] = true
	}

	for expected := range expectedThemes {
		if !foundThemes[expected] {
			t.Errorf("Expected theme %q not found in ListThemes()", expected)
		}
	}
}

func TestListThemesMatchesThemesMap(t *testing.T) {
	themes := ListThemes()

	if len(themes) != len(Themes) {
		t.Errorf("ListThemes() returned %d themes, but Themes map has %d", len(themes), len(Themes))
	}

	for _, name := range themes {
		if _, ok := Themes[name]; !ok {
			t.Errorf("Theme %q from ListThemes() not found in Themes map", name)
		}
	}
}

func TestThemeColorsAreValidHex(t *testing.T) {
	for name, theme := range Themes {
		t.Run(name, func(t *testing.T) {
			colors := []struct {
				field string
				value string
			}{
				{"AccentPrimary", theme.Colors.AccentPrimary},
				{"AccentSecondary", theme.Colors.AccentSecondary},
				{"Success", theme.Colors.Success},
				{"Warning", theme.Colors.Warning},
				{"Danger", theme.Colors.Danger},
				{"Info", theme.Colors.Info},
				{"Text", theme.Colors.Text},
				{"TextDim", theme.Colors.TextDim},
				{"Background", theme.Colors.Background},
				{"Border", theme.Colors.Border},
			}

			for _, c := range colors {
				if len(c.value) != 7 || c.value[0] != '#' {
					t.Errorf("Theme %q: %s = %q is not a valid 7-char hex color", name, c.field, c.value)
				}
			}
		})
	}
}

func TestThemeHasDistinctColors(t *testing.T) {
	for name, theme := range Themes {
		t.Run(name, func(t *testing.T) {
			colors := theme.Colors

			// Success, Warning, and Danger should be distinct
			if colors.Success == colors.Warning {
				t.Errorf("Theme %q: Success and Warning colors are the same", name)
			}
			if colors.Warning == colors.Danger {
				t.Errorf("Theme %q: Warning and Danger colors are the same", name)
			}
			if colors.Success == colors.Danger {
				t.Errorf("Theme %q: Success and Danger colors are the same", name)
			}

			// Text and TextDim should be different
			if colors.Text == colors.TextDim {
				t.Errorf("Theme %q: Text and TextDim colors are the same", name)
			}

			// Background and Border should be different
			if colors.Background == colors.Border {
				t.Errorf("Theme %q: Background and Border colors are the same", name)
			}
		})
	}
}

func TestDefaultThemeValues(t *testing.T) {
	theme := Themes["default"]

	expectedColors := ColorConfig{
		AccentPrimary:      "#00d7ff",
		AccentSecondary:    "#af5fff",
		Success:            "#5fff5f",
		Warning:            "#ffd700",
		Danger:             "#ff0000",
		Info:               "#00afff",
		Text:               "#d0d0d0",
		TextDim:            "#8a8a8a",
		Background:         "#1c1c1c",
		Border:             "#444444",
		BadgeForeground:    "#000000",
		BadgeForegroundAlt: "#ffffff",
		SurfaceAlt:         "#1a1a1a",
		SurfaceHighlight:   "#333333",
		TempHot:            "#ff8c00",
	}

	if theme.Colors != expectedColors {
		t.Errorf("Default theme colors don't match expected values")
	}
}

func TestNordThemeValues(t *testing.T) {
	theme := Themes["nord"]

	if theme.Name != "Nord" {
		t.Errorf("Nord theme Name = %q, want %q", theme.Name, "Nord")
	}

	// Verify Nord color palette
	if theme.Colors.AccentPrimary != "#88c0d0" {
		t.Errorf("Nord AccentPrimary = %q, want %q", theme.Colors.AccentPrimary, "#88c0d0")
	}
	if theme.Colors.Background != "#2e3440" {
		t.Errorf("Nord Background = %q, want %q", theme.Colors.Background, "#2e3440")
	}
}

func TestDraculaThemeValues(t *testing.T) {
	theme := Themes["dracula"]

	if theme.Name != "Dracula" {
		t.Errorf("Dracula theme Name = %q, want %q", theme.Name, "Dracula")
	}

	// Verify Dracula color palette
	if theme.Colors.AccentPrimary != "#8be9fd" {
		t.Errorf("Dracula AccentPrimary = %q, want %q", theme.Colors.AccentPrimary, "#8be9fd")
	}
	if theme.Colors.Background != "#282a36" {
		t.Errorf("Dracula Background = %q, want %q", theme.Colors.Background, "#282a36")
	}
}

// --- User theme loading tests ---

func backupThemes() map[string]Theme {
	backup := make(map[string]Theme, len(Themes))
	for k, v := range Themes {
		backup[k] = v
	}
	return backup
}

func restoreThemes(backup map[string]Theme) {
	// Remove any keys not in backup
	for k := range Themes {
		if _, ok := backup[k]; !ok {
			delete(Themes, k)
		}
	}
	// Restore original values
	for k, v := range backup {
		Themes[k] = v
	}
}

const validThemeYAML = `name: Catppuccin
colors:
  accent_primary: "#89b4fa"
  accent_secondary: "#cba6f7"
  success: "#a6e3a1"
  warning: "#f9e2af"
  danger: "#f38ba8"
  info: "#74c7ec"
  text: "#cdd6f4"
  text_dim: "#6c7086"
  background: "#1e1e2e"
  border: "#313244"
  badge_foreground: "#1e1e2e"
  badge_foreground_alt: "#cdd6f4"
  surface_alt: "#313244"
  surface_highlight: "#45475a"
  temp_hot: "#fab387"
`

func TestLoadUserThemesFromDirectory(t *testing.T) {
	backup := backupThemes()
	defer restoreThemes(backup)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "catppuccin.yaml"), []byte(validThemeYAML), 0644); err != nil {
		t.Fatal(err)
	}

	loadUserThemesFromDir(dir)

	theme, ok := Themes["catppuccin"]
	if !ok {
		t.Fatal("Expected 'catppuccin' theme to be loaded")
	}
	if theme.Name != "Catppuccin" {
		t.Errorf("Theme name = %q, want %q", theme.Name, "Catppuccin")
	}
	if theme.Colors.AccentPrimary != "#89b4fa" {
		t.Errorf("AccentPrimary = %q, want %q", theme.Colors.AccentPrimary, "#89b4fa")
	}
	if theme.Colors.Background != "#1e1e2e" {
		t.Errorf("Background = %q, want %q", theme.Colors.Background, "#1e1e2e")
	}

	// Should appear in ListThemes
	found := false
	for _, name := range ListThemes() {
		if name == "catppuccin" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'catppuccin' to appear in ListThemes()")
	}
}

func TestLoadUserThemesYMLExtension(t *testing.T) {
	backup := backupThemes()
	defer restoreThemes(backup)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mytheme.yml"), []byte(validThemeYAML), 0644); err != nil {
		t.Fatal(err)
	}

	loadUserThemesFromDir(dir)

	if _, ok := Themes["mytheme"]; !ok {
		t.Fatal("Expected 'mytheme' theme to be loaded from .yml file")
	}
}

func TestLoadUserThemesSkipsBuiltin(t *testing.T) {
	backup := backupThemes()
	defer restoreThemes(backup)

	dir := t.TempDir()
	// Try to override the "default" builtin theme
	if err := os.WriteFile(filepath.Join(dir, "default.yaml"), []byte(validThemeYAML), 0644); err != nil {
		t.Fatal(err)
	}

	originalDefault := Themes["default"]
	loadUserThemesFromDir(dir)

	// Default theme should NOT be overridden
	if Themes["default"].Colors.AccentPrimary != originalDefault.Colors.AccentPrimary {
		t.Errorf("Builtin 'default' theme was overridden, AccentPrimary = %q, want %q",
			Themes["default"].Colors.AccentPrimary, originalDefault.Colors.AccentPrimary)
	}
}

func TestLoadUserThemesSkipsInvalidYAML(t *testing.T) {
	backup := backupThemes()
	defer restoreThemes(backup)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.yaml"), []byte("not: [valid: yaml: {{"), 0644); err != nil {
		t.Fatal(err)
	}

	countBefore := len(Themes)
	loadUserThemesFromDir(dir)

	if len(Themes) != countBefore {
		t.Errorf("Invalid YAML should be skipped, theme count changed from %d to %d", countBefore, len(Themes))
	}
}

func TestLoadUserThemesSkipsIncompleteColors(t *testing.T) {
	backup := backupThemes()
	defer restoreThemes(backup)

	dir := t.TempDir()
	incomplete := `name: Incomplete
colors:
  accent_primary: "#89b4fa"
  success: "#a6e3a1"
`
	if err := os.WriteFile(filepath.Join(dir, "incomplete.yaml"), []byte(incomplete), 0644); err != nil {
		t.Fatal(err)
	}

	countBefore := len(Themes)
	loadUserThemesFromDir(dir)

	if len(Themes) != countBefore {
		t.Errorf("Incomplete theme should be skipped, theme count changed from %d to %d", countBefore, len(Themes))
	}
}

func TestLoadUserThemesSkipsDirectories(t *testing.T) {
	backup := backupThemes()
	defer restoreThemes(backup)

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir.yaml"), 0755); err != nil {
		t.Fatal(err)
	}

	countBefore := len(Themes)
	loadUserThemesFromDir(dir)

	if len(Themes) != countBefore {
		t.Errorf("Directories should be skipped, theme count changed from %d to %d", countBefore, len(Themes))
	}
}

func TestLoadUserThemesSkipsNonYAMLFiles(t *testing.T) {
	backup := backupThemes()
	defer restoreThemes(backup)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not a theme"), 0644); err != nil {
		t.Fatal(err)
	}

	countBefore := len(Themes)
	loadUserThemesFromDir(dir)

	if len(Themes) != countBefore {
		t.Errorf("Non-YAML files should be skipped, theme count changed from %d to %d", countBefore, len(Themes))
	}
}

func TestLoadUserThemesFallbackName(t *testing.T) {
	backup := backupThemes()
	defer restoreThemes(backup)

	dir := t.TempDir()
	noName := `colors:
  accent_primary: "#89b4fa"
  accent_secondary: "#cba6f7"
  success: "#a6e3a1"
  warning: "#f9e2af"
  danger: "#f38ba8"
  info: "#74c7ec"
  text: "#cdd6f4"
  text_dim: "#6c7086"
  background: "#1e1e2e"
  border: "#313244"
  badge_foreground: "#1e1e2e"
  badge_foreground_alt: "#cdd6f4"
  surface_alt: "#313244"
  surface_highlight: "#45475a"
  temp_hot: "#fab387"
`
	if err := os.WriteFile(filepath.Join(dir, "noname.yaml"), []byte(noName), 0644); err != nil {
		t.Fatal(err)
	}

	loadUserThemesFromDir(dir)

	theme, ok := Themes["noname"]
	if !ok {
		t.Fatal("Expected 'noname' theme to be loaded")
	}
	if theme.Name != "noname" {
		t.Errorf("Theme without name should use filename as name, got %q", theme.Name)
	}
}

func TestLoadUserThemesNonexistentDir(t *testing.T) {
	backup := backupThemes()
	defer restoreThemes(backup)

	countBefore := len(Themes)
	loadUserThemesFromDir(filepath.Join(t.TempDir(), "nonexistent"))

	if len(Themes) != countBefore {
		t.Errorf("Nonexistent directory should be a no-op, theme count changed from %d to %d", countBefore, len(Themes))
	}
}

func TestIsCompleteTheme(t *testing.T) {
	complete := ColorConfig{
		AccentPrimary:      "#89b4fa",
		AccentSecondary:    "#cba6f7",
		Success:            "#a6e3a1",
		Warning:            "#f9e2af",
		Danger:             "#f38ba8",
		Info:               "#74c7ec",
		Text:               "#cdd6f4",
		TextDim:            "#6c7086",
		Background:         "#1e1e2e",
		Border:             "#313244",
		BadgeForeground:    "#1e1e2e",
		BadgeForegroundAlt: "#cdd6f4",
		SurfaceAlt:         "#313244",
		SurfaceHighlight:   "#45475a",
		TempHot:            "#fab387",
	}
	if !isCompleteTheme(complete) {
		t.Error("Expected complete ColorConfig to be valid")
	}

	// Missing one field
	incomplete := complete
	incomplete.TempHot = ""
	if isCompleteTheme(incomplete) {
		t.Error("Expected incomplete ColorConfig (empty TempHot) to be invalid")
	}

	// Invalid hex
	badHex := complete
	badHex.Success = "not-a-color"
	if isCompleteTheme(badHex) {
		t.Error("Expected ColorConfig with invalid hex to be invalid")
	}
}

func TestGetThemesDir(t *testing.T) {
	dir, err := GetThemesDir()
	if err != nil {
		t.Fatalf("GetThemesDir() error: %v", err)
	}
	if dir == "" {
		t.Fatal("GetThemesDir() returned empty string")
	}
	if filepath.Base(dir) != "themes" {
		t.Errorf("GetThemesDir() should end with 'themes', got %q", filepath.Base(dir))
	}
	if filepath.Base(filepath.Dir(dir)) != "nuther" {
		t.Errorf("GetThemesDir() parent should be 'nuther', got %q", filepath.Base(filepath.Dir(dir)))
	}
}

func TestEmbeddedThemeChecksums(t *testing.T) {
	// Precomputed SHA256 checksums of embedded theme files.
	// Update these values when intentionally modifying a theme.
	expectedChecksums := map[string]string{
		"atomic-tangerine.yaml": "622109598a2764b0c2b95f68832698f77a05e61780e3cb7cb6a415b4a637afef",
		"catppuccin.yaml":       "4bc5211c8b39ecfeafa76749b5e900108e6a1606e3a868be4317f316c2c6c2b0",
		"default.yaml":          "6dfbe21105129f7b8f30ea07f87d55db099379eb8b0fc5e445b2d11f08a04ee0",
		"dracula.yaml":          "ed913700a1d2a53f5fcf4528e2b59c580ed836af06d151eed538def7a5528d71",
		"everforest.yaml":       "dcc874f72bee237540ee677caadd0f16e79ade9fd0242f119ed830ff0afd66ea",
		"gruvbox.yaml":          "378da894b17f0447c3fe9811f97de68fcf1bd5e67b8df60140c7beb727c62301",
		"nord.yaml":             "7129ba1d6cb1b1b8188bcd3a759396b0e660626cd53979c23cabe086bdd9e502",
		"rose-petale.yaml":      "5ea50bede842a8170d2e58febf3d76cf1bd37ea9bec558cda0aa52474f2efdc4",
		"solarized-dark.yaml":   "d39aba4e58239256144d95875b40ac03061e2ebc84bb864bf89f1ebb7a0afa00",
		"sous-bois.yaml":        "100045a880d52d592d0d73b1fbcc296b2bad9f1b39daeafbc2080982e554d22c",
		"tokyo-night.yaml":      "d74a31cac20a99044f68ba656b765ee501e78000dcfc1d01e123e9246ce200a4",
	}

	entries, err := themesFS.ReadDir("themes")
	if err != nil {
		t.Fatalf("failed to read embedded themes: %v", err)
	}

	found := 0
	for _, entry := range entries {
		if entry.IsDir() || !isYAMLFile(entry.Name()) {
			continue
		}
		found++

		data, err := themesFS.ReadFile("themes/" + entry.Name())
		if err != nil {
			t.Fatalf("failed to read %s: %v", entry.Name(), err)
		}
		hash := sha256.Sum256(data)
		checksum := fmt.Sprintf("%x", hash)

		expected, ok := expectedChecksums[entry.Name()]
		if !ok {
			t.Errorf("unknown theme file %q with checksum %q — add to expectedChecksums", entry.Name(), checksum)
			continue
		}
		if checksum != expected {
			t.Errorf("theme %q checksum mismatch:\n  got  %q\n  want %q", entry.Name(), checksum, expected)
		}
	}

	if found != len(expectedChecksums) {
		t.Errorf("expected %d theme files, found %d", len(expectedChecksums), found)
	}
}
