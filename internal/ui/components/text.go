package components

import "strings"

// Truncate truncates a string to the specified maximum rune length, appending "..." if truncated
func Truncate(str string, maxLen int) string {
	runes := []rune(str)
	if len(runes) <= maxLen {
		return str
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// PadCenter centers a string within the given length, padding with spaces
func PadCenter(str string, length int) string {
	if len(str) >= length {
		return str
	}
	padding := length - len(str)
	leftPad := padding / 2
	rightPad := padding - leftPad
	return strings.Repeat(" ", leftPad) + str + strings.Repeat(" ", rightPad)
}
