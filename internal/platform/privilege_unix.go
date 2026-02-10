//go:build !windows

package platform

import "os"

// IsElevated returns true if running as root
func IsElevated() bool {
	return os.Geteuid() == 0
}
