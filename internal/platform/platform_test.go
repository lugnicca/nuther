package platform

import (
	"testing"
)

func TestIsElevated(t *testing.T) {
	// IsElevated should return a boolean without panicking
	result := IsElevated()

	// On all platforms, we just verify it returns a valid boolean
	// and does not panic. We cannot guarantee root/admin access in tests.
	_ = result
}

func TestIsElevatedType(t *testing.T) {
	// Ensure the function returns a bool
	var result bool = IsElevated()
	_ = result
}
