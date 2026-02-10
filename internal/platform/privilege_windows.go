//go:build windows

package platform

import "golang.org/x/sys/windows"

// IsElevated checks whether the current process is running with administrator privileges
func IsElevated() bool {
	token := windows.GetCurrentProcessToken()
	return token.IsElevated()
}
