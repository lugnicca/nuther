package smartwatch

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"unicode"

	"nuther/internal/smart"
)

func StableDeviceKey(d smart.DriveInfo) string {
	if strings.TrimSpace(d.Serial) != "" {
		return "serial-" + safeToken(d.Serial, 80)
	}

	parts := []string{
		d.Model,
		d.WWN,
		d.Capacity,
		strconv.FormatInt(d.CapacityBytes, 10),
		d.Device,
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "drive-" + hex.EncodeToString(sum[:12])
}

func SummaryForDrive(d smart.DriveInfo) DeviceSummary {
	return DeviceSummary{
		Key:           StableDeviceKey(d),
		Device:        d.Device,
		Serial:        d.Serial,
		Model:         d.Model,
		WWN:           d.WWN,
		Capacity:      d.Capacity,
		CapacityBytes: d.CapacityBytes,
		Interface:     d.Interface,
		HealthStatus:  d.HealthStatus,
	}
}

func safeToken(value string, maxLen int) string {
	value = strings.TrimSpace(strings.ToLower(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case r == '.', r == '_':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
		if maxLen > 0 && b.Len() >= maxLen {
			break
		}
	}
	out := strings.Trim(b.String(), "-._")
	if out == "" {
		return "unknown"
	}
	return out
}
