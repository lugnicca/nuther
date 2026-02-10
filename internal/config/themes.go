package config

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed themes/*.yaml
var themesFS embed.FS

// Theme represents a complete color theme
type Theme struct {
	Name   string
	Colors ColorConfig
}

// Themes contains all available themes
var Themes map[string]Theme

// builtinThemeKeys tracks which theme keys are builtin and cannot be overridden
var builtinThemeKeys map[string]struct{}

func init() {
	Themes = make(map[string]Theme)
	builtinThemeKeys = make(map[string]struct{})

	entries, err := themesFS.ReadDir("themes")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded themes: %v", err))
	}

	for _, entry := range entries {
		if entry.IsDir() || !isYAMLFile(entry.Name()) {
			continue
		}

		data, err := themesFS.ReadFile("themes/" + entry.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to read embedded theme %s: %v\n", entry.Name(), err)
			continue
		}

		var ut userTheme
		if err := yaml.Unmarshal(data, &ut); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to parse embedded theme %s: %v\n", entry.Name(), err)
			continue
		}

		key := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		Themes[key] = Theme{Name: ut.Name, Colors: ut.Colors}
		builtinThemeKeys[key] = struct{}{}
		slog.Debug("loaded embedded theme", "key", key, "name", ut.Name)
	}

	// Ensure "default" theme always exists (fallback if embedded YAML was corrupted)
	if _, ok := Themes["default"]; !ok {
		Themes["default"] = Theme{
			Name: "Default",
			Colors: ColorConfig{
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
			},
		}
		builtinThemeKeys["default"] = struct{}{}
	}
}

// GetTheme returns a theme by name, falling back to default
func GetTheme(name string) Theme {
	if theme, ok := Themes[name]; ok {
		return theme
	}
	return Themes["default"]
}

// ListThemes returns all available theme names in sorted order
func ListThemes() []string {
	names := make([]string, 0, len(Themes))
	for name := range Themes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetThemesDir returns the user themes directory path
func GetThemesDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "nuther", "themes"), nil
}

// userTheme is the YAML structure for a user-defined theme file
type userTheme struct {
	Name   string      `yaml:"name"`
	Colors ColorConfig `yaml:"colors"`
}

// LoadUserThemes loads theme YAML files from the user themes directory.
// User themes cannot override builtin themes.
func LoadUserThemes() {
	dir, err := GetThemesDir()
	if err != nil {
		return
	}
	loadUserThemesFromDir(dir)
}

// loadUserThemesFromDir loads theme YAML files from the given directory.
// User themes cannot override builtin themes.
func loadUserThemesFromDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !isYAMLFile(entry.Name()) {
			continue
		}

		key := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		// Don't allow overriding builtin themes
		if _, builtin := builtinThemeKeys[key]; builtin {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		var ut userTheme
		if err := yaml.Unmarshal(data, &ut); err != nil {
			continue
		}

		if !isCompleteTheme(ut.Colors) {
			continue
		}

		name := ut.Name
		if name == "" {
			name = key
		}

		Themes[key] = Theme{Name: name, Colors: ut.Colors}
		slog.Debug("loaded user theme", "key", key, "name", name)
	}
}

func isYAMLFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}

func isCompleteTheme(c ColorConfig) bool {
	fields := []string{
		c.AccentPrimary, c.AccentSecondary, c.Success, c.Warning,
		c.Danger, c.Info, c.Text, c.TextDim, c.Background, c.Border,
		c.BadgeForeground, c.BadgeForegroundAlt, c.SurfaceAlt,
		c.SurfaceHighlight, c.TempHot,
	}
	for _, f := range fields {
		if !isValidHexColor(f) {
			return false
		}
	}
	return true
}
