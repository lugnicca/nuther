package smart

import (
	"testing"
)

func TestTemperatureThresholds(t *testing.T) {
	// Verify temperature thresholds are in ascending order
	temps := []struct {
		name  string
		value int
	}{
		{"TempCool", TempCool},
		{"TempGoodMax", TempGoodMax},
		{"TempCautionMax", TempCautionMax},
		{"TempCritical", TempCritical},
	}

	for i := 1; i < len(temps); i++ {
		if temps[i].value <= temps[i-1].value {
			t.Errorf("%s (%d) should be greater than %s (%d)",
				temps[i].name, temps[i].value, temps[i-1].name, temps[i-1].value)
		}
	}
}

func TestNVMeThresholds(t *testing.T) {
	// Verify NVMe threshold values are sensible
	if NVMePercentageUsedCaution >= NVMePercentageUsedBad {
		t.Errorf("NVMePercentageUsedCaution (%d) should be less than NVMePercentageUsedBad (%d)",
			NVMePercentageUsedCaution, NVMePercentageUsedBad)
	}

	if NVMeSpareMargin <= 0 {
		t.Errorf("NVMeSpareMargin (%d) should be positive", NVMeSpareMargin)
	}
}

func TestShutdownThresholds(t *testing.T) {
	if UnsafeShutdownGoodMax >= UnsafeShutdownCautionMax {
		t.Errorf("UnsafeShutdownGoodMax (%d) should be less than UnsafeShutdownCautionMax (%d)",
			UnsafeShutdownGoodMax, UnsafeShutdownCautionMax)
	}
}

func TestErrorCountThresholds(t *testing.T) {
	if ErrorCountGoodMax > ErrorCountCautionMax {
		t.Errorf("ErrorCountGoodMax (%d) should be <= ErrorCountCautionMax (%d)",
			ErrorCountGoodMax, ErrorCountCautionMax)
	}
}

func TestATAThresholds(t *testing.T) {
	if CriticalAttrBadThreshold <= 0 {
		t.Errorf("CriticalAttrBadThreshold (%d) should be positive", CriticalAttrBadThreshold)
	}

	if ReallocatedSectorsBadThreshold <= 0 {
		t.Errorf("ReallocatedSectorsBadThreshold (%d) should be positive", ReallocatedSectorsBadThreshold)
	}

	if UncorrectableSectorsBadThreshold <= 0 {
		t.Errorf("UncorrectableSectorsBadThreshold (%d) should be positive", UncorrectableSectorsBadThreshold)
	}

	if AttrValueMargin <= 0 {
		t.Errorf("AttrValueMargin (%d) should be positive", AttrValueMargin)
	}
}

func TestThresholdValues(t *testing.T) {
	// Verify exact threshold values as documented
	tests := []struct {
		name     string
		actual   int
		expected int
	}{
		{"TempCool", TempCool, 35},
		{"TempGoodMax", TempGoodMax, 45},
		{"TempCautionMax", TempCautionMax, 55},
		{"TempCritical", TempCritical, 65},
		{"NVMePercentageUsedCaution", NVMePercentageUsedCaution, 80},
		{"NVMePercentageUsedBad", NVMePercentageUsedBad, 95},
		{"NVMeHealthPercentageUsedCaution", NVMeHealthPercentageUsedCaution, 90},
		{"NVMeSpareMargin", NVMeSpareMargin, 10},
		{"UnsafeShutdownGoodMax", int(UnsafeShutdownGoodMax), 50},
		{"UnsafeShutdownCautionMax", int(UnsafeShutdownCautionMax), 200},
		{"ErrorCountGoodMax", int(ErrorCountGoodMax), 0},
		{"ErrorCountCautionMax", int(ErrorCountCautionMax), 10},
		{"CriticalAttrBadThreshold", int(CriticalAttrBadThreshold), 100},
		{"ReallocatedSectorsBadThreshold", int(ReallocatedSectorsBadThreshold), 100},
		{"UncorrectableSectorsBadThreshold", int(UncorrectableSectorsBadThreshold), 100},
		{"AttrValueMargin", AttrValueMargin, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.actual, tt.expected)
			}
		})
	}
}
