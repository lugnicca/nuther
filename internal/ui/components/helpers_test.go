package components

import (
	"testing"

	"nuther/internal/smart"
)

func TestSortAttributesByID(t *testing.T) {
	original := []smart.SmartAttribute{
		{ID: 5, Name: "Reallocated"},
		{ID: 1, Name: "Read Error"},
		{ID: 9, Name: "Power On"},
		{ID: 3, Name: "Spin Up"},
	}

	sorted := SortAttributesByID(original)

	// Verify sorted order
	expectedIDs := []int{1, 3, 5, 9}
	for i, id := range expectedIDs {
		if sorted[i].ID != id {
			t.Errorf("sorted[%d].ID = %d, want %d", i, sorted[i].ID, id)
		}
	}

	// Verify original is not mutated
	if original[0].ID != 5 {
		t.Errorf("original[0].ID = %d, want 5 (should not be mutated)", original[0].ID)
	}
}

func TestSortAttributesByIDEmpty(t *testing.T) {
	// nil slice
	sorted := SortAttributesByID(nil)
	if len(sorted) != 0 {
		t.Errorf("SortAttributesByID(nil) length = %d, want 0", len(sorted))
	}

	// empty slice
	sorted = SortAttributesByID([]smart.SmartAttribute{})
	if len(sorted) != 0 {
		t.Errorf("SortAttributesByID([]) length = %d, want 0", len(sorted))
	}
}

func TestFormatRawValue(t *testing.T) {
	tests := []struct {
		name     string
		attr     smart.SmartAttribute
		expected string
	}{
		{
			name:     "with RawString",
			attr:     smart.SmartAttribute{RawString: "123 (Min/Max 20/45)", RawValue: 123},
			expected: "123 (Min/Max 20/45)",
		},
		{
			name:     "without RawString",
			attr:     smart.SmartAttribute{RawString: "", RawValue: 456},
			expected: "456",
		},
		{
			name:     "zero value",
			attr:     smart.SmartAttribute{RawString: "", RawValue: 0},
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRawValue(tt.attr)
			if result != tt.expected {
				t.Errorf("FormatRawValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}
