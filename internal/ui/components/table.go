package components

import (
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// GetTableRowStyle returns the appropriate row style based on index and selection
func GetTableRowStyle(index, selected int, s *styles.Styles) lipgloss.Style {
	if index == selected {
		return s.TableRowSelected
	}
	if index%2 == 1 {
		return s.TableRowAlt
	}
	return s.TableRow
}
