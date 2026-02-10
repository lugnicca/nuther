package config

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return "config validation failed: " + strings.Join(msgs, "; ")
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// hexColorRegex matches valid hex color codes
var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// colorFieldMapping maps a display name to a getter/setter for a ColorConfig field.
type colorFieldMapping struct {
	Name string
	Get  func(c *ColorConfig) string
	Set  func(c *ColorConfig, v string)
}

// colorFields returns the mapping for all 15 ColorConfig fields.
// Defined once to avoid duplicating the field list across Validate, ValidateAndFix, and MergeWithThemeColors.
func colorFields() []colorFieldMapping {
	return []colorFieldMapping{
		{"colors.accent_primary", func(c *ColorConfig) string { return c.AccentPrimary }, func(c *ColorConfig, v string) { c.AccentPrimary = v }},
		{"colors.accent_secondary", func(c *ColorConfig) string { return c.AccentSecondary }, func(c *ColorConfig, v string) { c.AccentSecondary = v }},
		{"colors.success", func(c *ColorConfig) string { return c.Success }, func(c *ColorConfig, v string) { c.Success = v }},
		{"colors.warning", func(c *ColorConfig) string { return c.Warning }, func(c *ColorConfig, v string) { c.Warning = v }},
		{"colors.danger", func(c *ColorConfig) string { return c.Danger }, func(c *ColorConfig, v string) { c.Danger = v }},
		{"colors.info", func(c *ColorConfig) string { return c.Info }, func(c *ColorConfig, v string) { c.Info = v }},
		{"colors.text", func(c *ColorConfig) string { return c.Text }, func(c *ColorConfig, v string) { c.Text = v }},
		{"colors.text_dim", func(c *ColorConfig) string { return c.TextDim }, func(c *ColorConfig, v string) { c.TextDim = v }},
		{"colors.background", func(c *ColorConfig) string { return c.Background }, func(c *ColorConfig, v string) { c.Background = v }},
		{"colors.border", func(c *ColorConfig) string { return c.Border }, func(c *ColorConfig, v string) { c.Border = v }},
		{"colors.badge_foreground", func(c *ColorConfig) string { return c.BadgeForeground }, func(c *ColorConfig, v string) { c.BadgeForeground = v }},
		{"colors.badge_foreground_alt", func(c *ColorConfig) string { return c.BadgeForegroundAlt }, func(c *ColorConfig, v string) { c.BadgeForegroundAlt = v }},
		{"colors.surface_alt", func(c *ColorConfig) string { return c.SurfaceAlt }, func(c *ColorConfig, v string) { c.SurfaceAlt = v }},
		{"colors.surface_highlight", func(c *ColorConfig) string { return c.SurfaceHighlight }, func(c *ColorConfig, v string) { c.SurfaceHighlight = v }},
		{"colors.temp_hot", func(c *ColorConfig) string { return c.TempHot }, func(c *ColorConfig, v string) { c.TempHot = v }},
	}
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() ValidationErrors {
	var errs ValidationErrors

	// Validate theme
	if c.Theme != "" {
		if _, ok := Themes[c.Theme]; !ok {
			validThemes := ListThemes()
			errs = append(errs, ValidationError{
				Field:   "theme",
				Message: fmt.Sprintf("unknown theme %q, valid themes are: %s", c.Theme, strings.Join(validThemes, ", ")),
			})
		}
	}

	// Validate colors
	for _, f := range colorFields() {
		value := f.Get(&c.Colors)
		if value != "" && !isValidHexColor(value) {
			errs = append(errs, ValidationError{
				Field:   f.Name,
				Message: fmt.Sprintf("invalid hex color %q, must be in format #RRGGBB", value),
			})
		}
	}

	return errs
}

// isValidHexColor checks if a string is a valid hex color code
func isValidHexColor(color string) bool {
	return hexColorRegex.MatchString(color)
}

// ValidateAndFix validates the configuration and fixes minor issues
// Returns validation errors for all auto-fixed issues
func (c *Config) ValidateAndFix() ValidationErrors {
	var errs ValidationErrors

	// Fix theme if invalid
	if c.Theme != "" {
		if _, ok := Themes[c.Theme]; !ok {
			errs = append(errs, ValidationError{
				Field:   "theme",
				Message: fmt.Sprintf("unknown theme %q, reset to default", c.Theme),
			})
			c.Theme = "default"
		}
	} else {
		c.Theme = "default"
	}

	// Merge theme colors field-by-field (fills empty fields from theme)
	theme := GetTheme(c.Theme)
	MergeWithThemeColors(&c.Colors, theme.Colors)

	// Validate colors that are set and replace invalid ones
	for _, f := range colorFields() {
		value := f.Get(&c.Colors)
		if value != "" && !isValidHexColor(value) {
			themeValue := f.Get(&theme.Colors)
			f.Set(&c.Colors, themeValue)
			errs = append(errs, ValidationError{
				Field:   f.Name,
				Message: fmt.Sprintf("invalid color %q replaced with theme default %q", value, themeValue),
			})
		}
	}

	return errs
}

// MergeWithThemeColors fills empty color fields from the theme defaults
func MergeWithThemeColors(colors *ColorConfig, theme ColorConfig) {
	for _, f := range colorFields() {
		if f.Get(colors) == "" {
			f.Set(colors, f.Get(&theme))
		}
	}
}
