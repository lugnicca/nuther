package smart

import (
	"fmt"
	"strings"
)

// FormatBytes formats a byte count into a human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1000
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatNumber formats a number with thousands separators
func FormatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	s := fmt.Sprintf("%d", n)
	var result strings.Builder
	result.Grow(len(s) + len(s)/3)

	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}

	return result.String()
}

// FormatHours formats hours into a human-readable duration string
func FormatHours(hours int64) string {
	if hours < 24 {
		return fmt.Sprintf("%d hours", hours)
	}

	days := hours / 24
	remainingHours := hours % 24

	if days < 365 {
		if remainingHours == 0 {
			return fmt.Sprintf("%d days", days)
		}
		return fmt.Sprintf("%d days, %d hrs", days, remainingHours)
	}

	years := days / 365
	remainingDays := days % 365

	if remainingDays == 0 {
		return fmt.Sprintf("%d years", years)
	}
	return fmt.Sprintf("%d years, %d days", years, remainingDays)
}

// FormatDataUnits formats NVMe data units (512KB blocks) into a readable string
func FormatDataUnits(units int64) string {
	bytes := units * 512 * 1000
	return FormatBytes(bytes)
}

// FormatPercentage formats a percentage value
func FormatPercentage(value int) string {
	return fmt.Sprintf("%d%%", value)
}

// FormatTemperature formats a temperature value with unit
func FormatTemperature(temp int) string {
	return fmt.Sprintf("%d°C", temp)
}

// GetTemperatureStatus returns the health status for a temperature value
func GetTemperatureStatus(temp int) HealthStatus {
	switch {
	case temp <= TempGoodMax:
		return HealthGood
	case temp <= TempCautionMax:
		return HealthCaution
	default:
		return HealthBad
	}
}

// GetSpareStatus returns the health status for available spare percentage
func GetSpareStatus(spare, threshold int) HealthStatus {
	if spare > threshold+NVMeSpareMargin {
		return HealthGood
	}
	if spare > threshold {
		return HealthCaution
	}
	return HealthBad
}

// GetUsageStatus returns the health status for percentage used
func GetUsageStatus(used int) HealthStatus {
	if used < NVMePercentageUsedCaution {
		return HealthGood
	}
	if used < NVMePercentageUsedBad {
		return HealthCaution
	}
	return HealthBad
}

// GetShutdownStatus returns the health status for unsafe shutdown count
func GetShutdownStatus(count int64) HealthStatus {
	if count < UnsafeShutdownGoodMax {
		return HealthGood
	}
	if count < UnsafeShutdownCautionMax {
		return HealthCaution
	}
	return HealthBad
}

// GetErrorStatus returns the health status for error counts
func GetErrorStatus(count int64) HealthStatus {
	if count <= ErrorCountGoodMax {
		return HealthGood
	}
	if count <= ErrorCountCautionMax {
		return HealthCaution
	}
	return HealthBad
}

// BoolToYesNo converts a boolean to "Yes" or "No"
func BoolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// BoolToPassFail converts a boolean to "PASSED" or "FAILED"
func BoolToPassFail(b bool) string {
	if b {
		return "PASSED"
	}
	return "FAILED"
}
