//go:build windows

package smart

import "fmt"

// GetCommonDevicePaths returns common device paths for Windows systems
func GetCommonDevicePaths() []string {
	paths := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		paths = append(paths, fmt.Sprintf("\\\\.\\PhysicalDrive%d", i))
	}
	return paths
}
