package smart

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"small bytes", 500, "500 B"},
		{"kilobytes", 1500, "1.50 KB"},
		{"megabytes", 1500000, "1.50 MB"},
		{"gigabytes", 1500000000, "1.50 GB"},
		{"terabytes", 1500000000000, "1.50 TB"},
		{"petabytes", 1500000000000000, "1.50 PB"},
		{"exact kilobyte", 1000, "1.00 KB"},
		{"exact gigabyte", 1000000000, "1.00 GB"},
		{"500GB", 500000000000, "500.00 GB"},
		{"1TB", 1000000000000, "1.00 TB"},
		{"4TB", 4000000000000, "4.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   int64
		expected string
	}{
		{"zero", 0, "0"},
		{"single digit", 5, "5"},
		{"two digits", 42, "42"},
		{"three digits", 999, "999"},
		{"thousands", 1000, "1,000"},
		{"ten thousands", 12345, "12,345"},
		{"hundred thousands", 123456, "123,456"},
		{"millions", 1234567, "1,234,567"},
		{"billions", 1234567890, "1,234,567,890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNumber(tt.number)
			if result != tt.expected {
				t.Errorf("FormatNumber(%d) = %q, want %q", tt.number, result, tt.expected)
			}
		})
	}
}

func TestFormatHours(t *testing.T) {
	tests := []struct {
		name     string
		hours    int64
		expected string
	}{
		{"zero hours", 0, "0 hours"},
		{"single hour", 1, "1 hours"},
		{"several hours", 12, "12 hours"},
		{"23 hours", 23, "23 hours"},
		{"exactly one day", 24, "1 days"},
		{"one day with hours", 30, "1 days, 6 hrs"},
		{"several days", 72, "3 days"},
		{"several days with hours", 75, "3 days, 3 hrs"},
		{"just under a year", 8759, "364 days, 23 hrs"},
		{"exactly one year", 8760, "1 years"},
		{"one year plus one day", 8784, "1 years, 1 days"},
		{"two years", 17520, "2 years"},
		{"two years and one day", 17544, "2 years, 1 days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatHours(tt.hours)
			if result != tt.expected {
				t.Errorf("FormatHours(%d) = %q, want %q", tt.hours, result, tt.expected)
			}
		})
	}
}

func TestFormatDataUnits(t *testing.T) {
	tests := []struct {
		name     string
		units    int64
		expected string
	}{
		{"zero units", 0, "0 B"},
		{"small units", 1, "512.00 KB"},
		{"larger units", 1000, "512.00 MB"},
		{"gigabyte range", 2000000, "1.02 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDataUnits(tt.units)
			if result != tt.expected {
				t.Errorf("FormatDataUnits(%d) = %q, want %q", tt.units, result, tt.expected)
			}
		})
	}
}

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected string
	}{
		{"zero", 0, "0%"},
		{"fifty", 50, "50%"},
		{"hundred", 100, "100%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPercentage(tt.value)
			if result != tt.expected {
				t.Errorf("FormatPercentage(%d) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestFormatTemperature(t *testing.T) {
	tests := []struct {
		name     string
		temp     int
		expected string
	}{
		{"zero", 0, "0°C"},
		{"room temp", 25, "25°C"},
		{"hot", 55, "55°C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTemperature(tt.temp)
			if result != tt.expected {
				t.Errorf("FormatTemperature(%d) = %q, want %q", tt.temp, result, tt.expected)
			}
		})
	}
}

func TestGetTemperatureStatus(t *testing.T) {
	tests := []struct {
		name     string
		temp     int
		expected HealthStatus
	}{
		{"cold", 20, HealthGood},
		{"normal", 35, HealthGood},
		{"at good threshold", 45, HealthGood},
		{"warm", 50, HealthCaution},
		{"at caution threshold", 55, HealthCaution},
		{"hot", 60, HealthBad},
		{"very hot", 70, HealthBad},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTemperatureStatus(tt.temp)
			if result != tt.expected {
				t.Errorf("GetTemperatureStatus(%d) = %q, want %q", tt.temp, result, tt.expected)
			}
		})
	}
}

func TestGetSpareStatus(t *testing.T) {
	tests := []struct {
		name      string
		spare     int
		threshold int
		expected  HealthStatus
	}{
		{"plenty of spare", 100, 10, HealthGood},
		{"good spare", 50, 10, HealthGood},
		{"at margin threshold", 21, 10, HealthGood},
		{"within margin", 20, 10, HealthCaution},
		{"at threshold", 10, 10, HealthBad},
		{"below threshold", 5, 10, HealthBad},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSpareStatus(tt.spare, tt.threshold)
			if result != tt.expected {
				t.Errorf("GetSpareStatus(%d, %d) = %q, want %q", tt.spare, tt.threshold, result, tt.expected)
			}
		})
	}
}

func TestGetUsageStatus(t *testing.T) {
	tests := []struct {
		name     string
		used     int
		expected HealthStatus
	}{
		{"new drive", 0, HealthGood},
		{"low usage", 50, HealthGood},
		{"at good threshold", 79, HealthGood},
		{"moderate usage", 80, HealthCaution},
		{"high usage", 90, HealthCaution},
		{"at caution threshold", 94, HealthCaution},
		{"critical usage", 95, HealthBad},
		{"very high usage", 100, HealthBad},
		{"over 100", 120, HealthBad},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUsageStatus(tt.used)
			if result != tt.expected {
				t.Errorf("GetUsageStatus(%d) = %q, want %q", tt.used, result, tt.expected)
			}
		})
	}
}

func TestGetShutdownStatus(t *testing.T) {
	tests := []struct {
		name     string
		count    int64
		expected HealthStatus
	}{
		{"no shutdowns", 0, HealthGood},
		{"few shutdowns", 10, HealthGood},
		{"just below threshold", 49, HealthGood},
		{"at good threshold", 50, HealthCaution},
		{"many shutdowns", 100, HealthCaution},
		{"just below caution threshold", 199, HealthCaution},
		{"at caution threshold", 200, HealthBad},
		{"very many shutdowns", 500, HealthBad},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetShutdownStatus(tt.count)
			if result != tt.expected {
				t.Errorf("GetShutdownStatus(%d) = %q, want %q", tt.count, result, tt.expected)
			}
		})
	}
}

func TestGetErrorStatus(t *testing.T) {
	tests := []struct {
		name     string
		count    int64
		expected HealthStatus
	}{
		{"no errors", 0, HealthGood},
		{"one error", 1, HealthCaution},
		{"few errors", 5, HealthCaution},
		{"at caution threshold", 10, HealthCaution},
		{"many errors", 11, HealthBad},
		{"lots of errors", 100, HealthBad},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorStatus(tt.count)
			if result != tt.expected {
				t.Errorf("GetErrorStatus(%d) = %q, want %q", tt.count, result, tt.expected)
			}
		})
	}
}

func TestBoolToYesNo(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{"true", true, "Yes"},
		{"false", false, "No"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolToYesNo(tt.value)
			if result != tt.expected {
				t.Errorf("BoolToYesNo(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestBoolToPassFail(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{"true", true, "PASSED"},
		{"false", false, "FAILED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolToPassFail(tt.value)
			if result != tt.expected {
				t.Errorf("BoolToPassFail(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}
