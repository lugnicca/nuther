package components

import (
	"fmt"
	"sort"

	"nuther/internal/smart"
)

// SortAttributesByID returns a copy of the attributes slice sorted by ID
func SortAttributesByID(attrs []smart.SmartAttribute) []smart.SmartAttribute {
	sorted := make([]smart.SmartAttribute, len(attrs))
	copy(sorted, attrs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

// FormatRawValue returns the display string for a SMART attribute's raw value
func FormatRawValue(attr smart.SmartAttribute) string {
	if attr.RawString != "" {
		return attr.RawString
	}
	return fmt.Sprintf("%d", attr.RawValue)
}
