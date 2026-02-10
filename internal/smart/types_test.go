package smart

import (
	"testing"
)

func TestDriveInfoGetAttributeCount(t *testing.T) {
	tests := []struct {
		name     string
		drive    DriveInfo
		expected int
	}{
		{
			name: "NVMe drive with attributes",
			drive: DriveInfo{
				IsNVMe: true,
				NVMeAttributes: []NVMeAttribute{
					{Name: "Temperature"},
					{Name: "Available Spare"},
					{Name: "Percentage Used"},
				},
			},
			expected: 3,
		},
		{
			name: "NVMe drive with no attributes",
			drive: DriveInfo{
				IsNVMe: true,
			},
			expected: 0,
		},
		{
			name: "SATA drive with attributes",
			drive: DriveInfo{
				IsNVMe: false,
				Attributes: []SmartAttribute{
					{ID: 1, Name: "Raw_Read_Error_Rate"},
					{ID: 5, Name: "Reallocated_Sector_Ct"},
				},
			},
			expected: 2,
		},
		{
			name: "SATA drive with no attributes",
			drive: DriveInfo{
				IsNVMe: false,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.drive.GetAttributeCount()
			if result != tt.expected {
				t.Errorf("GetAttributeCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestSmartAttributeGetStatusEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		attr     SmartAttribute
		expected HealthStatus
	}{
		{
			name: "zero threshold with high value",
			attr: SmartAttribute{
				ID:        AttrPowerOnHours,
				RawValue:  50000,
				Value:     100,
				Threshold: 0,
			},
			expected: HealthGood,
		},
		{
			name: "critical attribute at exact bad threshold",
			attr: SmartAttribute{
				ID:        AttrReallocatedSectors,
				RawValue:  100,
				Value:     100,
				Threshold: 10,
			},
			expected: HealthCaution,
		},
		{
			name: "critical attribute just above bad threshold",
			attr: SmartAttribute{
				ID:        AttrReallocatedSectors,
				RawValue:  101,
				Value:     100,
				Threshold: 10,
			},
			expected: HealthBad,
		},
		{
			name: "pending sectors with value",
			attr: SmartAttribute{
				ID:        AttrPendingSectors,
				RawValue:  10,
				Value:     100,
				Threshold: 0,
			},
			expected: HealthCaution,
		},
		{
			name: "offline uncorrectable with high count",
			attr: SmartAttribute{
				ID:        AttrOfflineUncorrectable,
				RawValue:  150,
				Value:     100,
				Threshold: 0,
			},
			expected: HealthBad,
		},
		{
			name: "UDMA CRC error",
			attr: SmartAttribute{
				ID:        AttrUDMACRCError,
				RawValue:  5,
				Value:     200,
				Threshold: 0,
			},
			expected: HealthCaution,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attr.GetStatus()
			if result != tt.expected {
				t.Errorf("GetStatus() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSmartAttributeIsCriticalAllCriticalIDs(t *testing.T) {
	criticalIDs := []int{
		AttrReallocatedSectors,
		AttrReportedUncorrect,
		AttrCommandTimeout,
		AttrPendingSectors,
		AttrOfflineUncorrectable,
		AttrUDMACRCError,
	}

	for _, id := range criticalIDs {
		attr := SmartAttribute{ID: id}
		if !attr.IsCritical() {
			t.Errorf("Attribute ID %d should be critical", id)
		}
	}
}

func TestSmartAttributeIsCriticalNonCriticalIDs(t *testing.T) {
	nonCriticalIDs := []int{
		AttrRawReadErrorRate,
		AttrSpinUpTime,
		AttrStartStopCount,
		AttrSeekErrorRate,
		AttrPowerOnHours,
		AttrSpinRetryCount,
		AttrPowerCycleCount,
		AttrTemperature,
		AttrLoadCycleCount,
	}

	for _, id := range nonCriticalIDs {
		attr := SmartAttribute{ID: id}
		if attr.IsCritical() {
			t.Errorf("Attribute ID %d should NOT be critical", id)
		}
	}
}

func TestDriveInfoGetDriveTypeAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		drive    DriveInfo
		expected string
	}{
		{
			name:     "NVMe takes priority over SSD",
			drive:    DriveInfo{IsNVMe: true, IsSSD: true},
			expected: "NVMe SSD",
		},
		{
			name:     "SSD takes priority over rotation rate",
			drive:    DriveInfo{IsSSD: true, RotationRate: 0},
			expected: "SATA SSD",
		},
		{
			name:     "HDD with 7200 RPM",
			drive:    DriveInfo{RotationRate: 7200},
			expected: "HDD",
		},
		{
			name:     "HDD with 5400 RPM",
			drive:    DriveInfo{RotationRate: 5400},
			expected: "HDD",
		},
		{
			name:     "Virtual drive",
			drive:    DriveInfo{IsVirtual: true},
			expected: "Virtual",
		},
		{
			name:     "Unknown drive",
			drive:    DriveInfo{},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.drive.GetDriveType()
			if result != tt.expected {
				t.Errorf("GetDriveType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDriveInfoHasCriticalIssuesAllCombinations(t *testing.T) {
	tests := []struct {
		name                 string
		reallocatedSectors   int64
		pendingSectors       int64
		uncorrectableSectors int64
		crcErrors            int64
		expected             bool
	}{
		{"all zero", 0, 0, 0, 0, false},
		{"only reallocated", 1, 0, 0, 0, true},
		{"only pending", 0, 1, 0, 0, true},
		{"only uncorrectable", 0, 0, 1, 0, true},
		{"only crc", 0, 0, 0, 1, true},
		{"all issues", 5, 3, 2, 1, true},
		{"high values", 1000, 500, 250, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drive := DriveInfo{
				ReallocatedSectors:   tt.reallocatedSectors,
				PendingSectors:       tt.pendingSectors,
				UncorrectableSectors: tt.uncorrectableSectors,
				CRCErrors:            tt.crcErrors,
			}
			result := drive.HasCriticalIssues()
			if result != tt.expected {
				t.Errorf("HasCriticalIssues() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNVMeHealthLogFields(t *testing.T) {
	log := NVMeHealthLog{
		CriticalWarning:         1,
		Temperature:             45,
		AvailableSpare:          100,
		AvailableSpareThreshold: 10,
		PercentageUsed:          5,
		DataUnitsRead:           1000000,
		DataUnitsWritten:        500000,
		HostReadCommands:        1000000000,
		HostWriteCommands:       500000000,
		ControllerBusyTime:      1000,
		PowerCycles:             100,
		PowerOnHours:            5000,
		UnsafeShutdowns:         10,
		MediaErrors:             0,
		ErrorLogEntries:         5,
		WarningCompTempTime:     0,
		CriticalCompTempTime:    0,
		TempSensor1:             45,
		TempSensor2:             43,
		TempSensor3:             0,
		TempSensor4:             0,
	}

	if log.CriticalWarning != 1 {
		t.Errorf("CriticalWarning = %d, want 1", log.CriticalWarning)
	}
	if log.Temperature != 45 {
		t.Errorf("Temperature = %d, want 45", log.Temperature)
	}
	if log.AvailableSpare != 100 {
		t.Errorf("AvailableSpare = %d, want 100", log.AvailableSpare)
	}
}

func TestSmartctlOutputParsing(t *testing.T) {
	// Test that struct fields are correctly mapped
	output := SmartctlOutput{
		ModelName:       "Test Drive",
		SerialNumber:    "ABC123",
		FirmwareVersion: "1.0",
	}

	if output.ModelName != "Test Drive" {
		t.Errorf("ModelName = %q, want %q", output.ModelName, "Test Drive")
	}
	if output.SerialNumber != "ABC123" {
		t.Errorf("SerialNumber = %q, want %q", output.SerialNumber, "ABC123")
	}
	if output.FirmwareVersion != "1.0" {
		t.Errorf("FirmwareVersion = %q, want %q", output.FirmwareVersion, "1.0")
	}
}

func TestSmartctlScanResultDevices(t *testing.T) {
	result := SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: "/dev/sda", Type: "ata", Protocol: "ATA"},
			{Name: "/dev/nvme0", Type: "nvme", Protocol: "NVMe"},
		},
	}

	if len(result.Devices) != 2 {
		t.Errorf("Devices count = %d, want 2", len(result.Devices))
	}

	if result.Devices[0].Name != "/dev/sda" {
		t.Errorf("First device name = %q, want %q", result.Devices[0].Name, "/dev/sda")
	}
	if result.Devices[1].Protocol != "NVMe" {
		t.Errorf("Second device protocol = %q, want %q", result.Devices[1].Protocol, "NVMe")
	}
}
