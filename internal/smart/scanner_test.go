package smart

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadTestData(t *testing.T, filename string) SmartctlOutput {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test data %s: %v", filename, err)
	}

	var output SmartctlOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Failed to parse test data %s: %v", filename, err)
	}

	return output
}

func loadTestDataRaw(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test data %s: %v", filename, err)
	}
	return data
}

// withMockedSmartctl replaces runSmartctl with a mock and restores after test
func withMockedSmartctl(t *testing.T, mock func(args ...string) ([]byte, error)) {
	t.Helper()
	orig := runSmartctl
	runSmartctl = mock
	t.Cleanup(func() { runSmartctl = orig })
}

func TestParseDriveInfo_NVMe(t *testing.T) {
	output := loadTestData(t, "smartctl_nvme.json")
	drive := parseDriveInfo("/dev/nvme0", output)

	if drive.Device != "/dev/nvme0" {
		t.Errorf("Device = %q, want %q", drive.Device, "/dev/nvme0")
	}

	if drive.Model != "Samsung SSD 980 PRO 1TB" {
		t.Errorf("Model = %q, want %q", drive.Model, "Samsung SSD 980 PRO 1TB")
	}

	if drive.Serial != "S5GXNF0R123456" {
		t.Errorf("Serial = %q, want %q", drive.Serial, "S5GXNF0R123456")
	}

	if !drive.IsNVMe {
		t.Error("IsNVMe = false, want true")
	}

	if !drive.IsSSD {
		t.Error("IsSSD = false, want true")
	}

	if drive.Interface != "NVMe" {
		t.Errorf("Interface = %q, want %q", drive.Interface, "NVMe")
	}

	if drive.Temperature != 38 {
		t.Errorf("Temperature = %d, want %d", drive.Temperature, 38)
	}

	if drive.PowerOnHours != 2500 {
		t.Errorf("PowerOnHours = %d, want %d", drive.PowerOnHours, 2500)
	}

	if drive.PowerCycles != 150 {
		t.Errorf("PowerCycles = %d, want %d", drive.PowerCycles, 150)
	}

	if !drive.HealthPassed {
		t.Error("HealthPassed = false, want true")
	}

	if drive.HealthStatus != HealthGood {
		t.Errorf("HealthStatus = %q, want %q", drive.HealthStatus, HealthGood)
	}

	// Check NVMe health log
	if drive.NVMeHealthLog == nil {
		t.Fatal("NVMeHealthLog is nil")
	}

	if drive.NVMeHealthLog.AvailableSpare != 100 {
		t.Errorf("AvailableSpare = %d, want %d", drive.NVMeHealthLog.AvailableSpare, 100)
	}

	if drive.NVMeHealthLog.PercentageUsed != 2 {
		t.Errorf("PercentageUsed = %d, want %d", drive.NVMeHealthLog.PercentageUsed, 2)
	}

	if drive.NVMeHealthLog.UnsafeShutdowns != 5 {
		t.Errorf("UnsafeShutdowns = %d, want %d", drive.NVMeHealthLog.UnsafeShutdowns, 5)
	}

	if drive.NVMeHealthLog.MediaErrors != 0 {
		t.Errorf("MediaErrors = %d, want %d", drive.NVMeHealthLog.MediaErrors, 0)
	}

	// Check TotalBytesWritten (DataUnitsWritten=18000000 * 512 * 1000)
	expectedBytes := int64(18000000) * 512 * 1000
	if drive.TotalBytesWritten != expectedBytes {
		t.Errorf("TotalBytesWritten = %d, want %d", drive.TotalBytesWritten, expectedBytes)
	}

	// Check NVMe attributes were created
	if len(drive.NVMeAttributes) == 0 {
		t.Error("NVMeAttributes is empty")
	}
}

func TestParseDriveInfo_SATA_SSD(t *testing.T) {
	output := loadTestData(t, "smartctl_sata_ssd.json")
	drive := parseDriveInfo("/dev/sda", output)

	if drive.Device != "/dev/sda" {
		t.Errorf("Device = %q, want %q", drive.Device, "/dev/sda")
	}

	if drive.Model != "Samsung SSD 870 EVO 500GB" {
		t.Errorf("Model = %q, want %q", drive.Model, "Samsung SSD 870 EVO 500GB")
	}

	if drive.ModelFamily != "Samsung based SSDs" {
		t.Errorf("ModelFamily = %q, want %q", drive.ModelFamily, "Samsung based SSDs")
	}

	if drive.IsNVMe {
		t.Error("IsNVMe = true, want false")
	}

	if !drive.IsSSD {
		t.Error("IsSSD = false, want true")
	}

	if drive.Interface != "SATA SSD" {
		t.Errorf("Interface = %q, want %q", drive.Interface, "SATA SSD")
	}

	if drive.Temperature != 32 {
		t.Errorf("Temperature = %d, want %d", drive.Temperature, 32)
	}

	if drive.RotationRate != 0 {
		t.Errorf("RotationRate = %d, want %d", drive.RotationRate, 0)
	}

	// Check SMART attributes were parsed
	if len(drive.Attributes) == 0 {
		t.Error("Attributes is empty")
	}

	// Verify specific attributes
	foundReallocated := false
	for _, attr := range drive.Attributes {
		if attr.ID == AttrReallocatedSectors {
			foundReallocated = true
			if attr.RawValue != 0 {
				t.Errorf("Reallocated sectors RawValue = %d, want %d", attr.RawValue, 0)
			}
		}
	}

	if !foundReallocated {
		t.Error("Reallocated_Sector_Ct attribute not found")
	}

	// SATA SSD test data has no attr 241, so TotalBytesWritten should remain -1
	if drive.TotalBytesWritten != -1 {
		t.Errorf("TotalBytesWritten = %d, want -1 (no attr 241)", drive.TotalBytesWritten)
	}
}

func TestParseSmartctlDriveJSON_DeviceStatisticsLogicalSectorsWritten(t *testing.T) {
	data := []byte(`{
		"device": {"name": "/dev/sdc", "protocol": "ATA"},
		"model_name": "MK001920GWSSE",
		"rotation_rate": 0,
		"logical_block_size": 512,
		"smart_status": {"passed": true},
		"ata_smart_attributes": {"table": [{"id": 9, "name": "Power_On_Hours", "value": 90, "raw": {"value": 46507}}]},
		"ata_device_statistics": {"pages": [{"number": 1, "table": [
			{"offset": 16, "value": 46507, "flags": {"valid": true}},
			{"offset": 24, "value": 1433271786477, "flags": {"valid": true}}
		]}]}
	}`)

	drive, err := ParseSmartctlDriveJSON(data)
	if err != nil {
		t.Fatalf("ParseSmartctlDriveJSON() error = %v", err)
	}

	const expectedBytes int64 = 1433271786477 * 512
	if drive.TotalBytesWritten != expectedBytes {
		t.Errorf("TotalBytesWritten = %d, want %d", drive.TotalBytesWritten, expectedBytes)
	}
}

func TestParseSmartctlDriveJSON_MicronTotalLBAsWrittenAttribute246(t *testing.T) {
	data := []byte(`{
		"device": {"name": "/dev/sdb", "protocol": "ATA"},
		"model_name": "Micron_5300_MTFDDAK1T9TDS",
		"rotation_rate": 0,
		"logical_block_size": 512,
		"smart_status": {"passed": true},
		"ata_smart_attributes": {"table": [{
			"id": 246,
			"name": "Total_LBAs_Written",
			"value": 100,
			"worst": 100,
			"thresh": 0,
			"raw": {"value": 95807222632, "string": "95807222632"}
		}]}
	}`)

	drive, err := ParseSmartctlDriveJSON(data)
	if err != nil {
		t.Fatalf("ParseSmartctlDriveJSON() error = %v", err)
	}

	const expectedBytes int64 = 95807222632 * 512
	if drive.TotalBytesWritten != expectedBytes {
		t.Errorf("TotalBytesWritten = %d, want %d", drive.TotalBytesWritten, expectedBytes)
	}
}

func TestParseDriveInfo_HDD(t *testing.T) {
	output := loadTestData(t, "smartctl_hdd.json")
	drive := parseDriveInfo("/dev/sdb", output)

	if drive.Device != "/dev/sdb" {
		t.Errorf("Device = %q, want %q", drive.Device, "/dev/sdb")
	}

	if drive.Model != "WDC WD40EFRX-68N32N0" {
		t.Errorf("Model = %q, want %q", drive.Model, "WDC WD40EFRX-68N32N0")
	}

	if drive.IsNVMe {
		t.Error("IsNVMe = true, want false")
	}

	if drive.IsSSD {
		t.Error("IsSSD = true, want false")
	}

	if drive.RotationRate != 5400 {
		t.Errorf("RotationRate = %d, want %d", drive.RotationRate, 5400)
	}

	expectedInterface := "SATA HDD (5400 RPM)"
	if drive.Interface != expectedInterface {
		t.Errorf("Interface = %q, want %q", drive.Interface, expectedInterface)
	}

	if drive.FormFactor != "3.5 inches" {
		t.Errorf("FormFactor = %q, want %q", drive.FormFactor, "3.5 inches")
	}

	if drive.Temperature != 35 {
		t.Errorf("Temperature = %d, want %d", drive.Temperature, 35)
	}

	if drive.PowerOnHours != 15000 {
		t.Errorf("PowerOnHours = %d, want %d", drive.PowerOnHours, 15000)
	}

	// HDD has no attr 241, so TotalBytesWritten should remain -1
	if drive.TotalBytesWritten != -1 {
		t.Errorf("TotalBytesWritten = %d, want -1 (no attr 241)", drive.TotalBytesWritten)
	}
}

func TestBuildNVMeAttributes(t *testing.T) {
	healthLog := SmartctlNvmeHealthLog{
		Temperature:             40,
		AvailableSpare:          95,
		AvailableSpareThreshold: 10,
		PercentageUsed:          5,
		DataUnitsRead:           1000000,
		DataUnitsWritten:        500000,
		HostReads:               100000000,
		HostWrites:              50000000,
		ControllerBusyTime:      100,
		PowerCycles:             50,
		PowerOnHours:            1000,
		UnsafeShutdowns:         2,
		MediaErrors:             0,
		NumErrLogEntries:        0,
	}

	attrs := buildNVMeAttributes(healthLog)

	if len(attrs) != 14 {
		t.Errorf("Expected 14 attributes, got %d", len(attrs))
	}

	// Check temperature attribute
	tempAttr := attrs[0]
	if tempAttr.Name != "Temperature" {
		t.Errorf("First attribute name = %q, want %q", tempAttr.Name, "Temperature")
	}
	if tempAttr.NumericValue != 40 {
		t.Errorf("Temperature NumericValue = %d, want %d", tempAttr.NumericValue, 40)
	}
	if tempAttr.Status != HealthGood {
		t.Errorf("Temperature Status = %q, want %q", tempAttr.Status, HealthGood)
	}

	// Check available spare attribute
	spareAttr := attrs[1]
	if spareAttr.Name != "Available Spare" {
		t.Errorf("Second attribute name = %q, want %q", spareAttr.Name, "Available Spare")
	}
	if spareAttr.Status != HealthGood {
		t.Errorf("Available Spare Status = %q, want %q", spareAttr.Status, HealthGood)
	}
}

func TestUpdateHealthStatus_NVMe_Healthy(t *testing.T) {
	drive := &DriveInfo{
		IsNVMe:       true,
		HealthStatus: HealthGood,
		NVMeHealthLog: &NVMeHealthLog{
			CriticalWarning: 0,
			MediaErrors:     0,
			PercentageUsed:  10,
		},
	}

	updateHealthStatus(drive)

	if drive.HealthStatus != HealthGood {
		t.Errorf("HealthStatus = %q, want %q", drive.HealthStatus, HealthGood)
	}
}

func TestUpdateHealthStatus_NVMe_CriticalWarning(t *testing.T) {
	drive := &DriveInfo{
		IsNVMe:       true,
		HealthStatus: HealthGood,
		NVMeHealthLog: &NVMeHealthLog{
			CriticalWarning: 1,
			MediaErrors:     0,
			PercentageUsed:  10,
		},
	}

	updateHealthStatus(drive)

	if drive.HealthStatus != HealthBad {
		t.Errorf("HealthStatus = %q, want %q", drive.HealthStatus, HealthBad)
	}
}

func TestUpdateHealthStatus_NVMe_MediaErrors(t *testing.T) {
	drive := &DriveInfo{
		IsNVMe:       true,
		HealthStatus: HealthGood,
		NVMeHealthLog: &NVMeHealthLog{
			CriticalWarning: 0,
			MediaErrors:     5,
			PercentageUsed:  10,
		},
	}

	updateHealthStatus(drive)

	if drive.HealthStatus != HealthCaution {
		t.Errorf("HealthStatus = %q, want %q", drive.HealthStatus, HealthCaution)
	}
}

func TestUpdateHealthStatus_NVMe_HighUsage(t *testing.T) {
	drive := &DriveInfo{
		IsNVMe:       true,
		HealthStatus: HealthGood,
		NVMeHealthLog: &NVMeHealthLog{
			CriticalWarning: 0,
			MediaErrors:     0,
			PercentageUsed:  95,
		},
	}

	updateHealthStatus(drive)

	if drive.HealthStatus != HealthCaution {
		t.Errorf("HealthStatus = %q, want %q", drive.HealthStatus, HealthCaution)
	}
}

func TestUpdateHealthStatus_ATA_Healthy(t *testing.T) {
	drive := &DriveInfo{
		IsNVMe:               false,
		HealthStatus:         HealthGood,
		ReallocatedSectors:   0,
		PendingSectors:       0,
		UncorrectableSectors: 0,
		CRCErrors:            0,
	}

	updateHealthStatus(drive)

	if drive.HealthStatus != HealthGood {
		t.Errorf("HealthStatus = %q, want %q", drive.HealthStatus, HealthGood)
	}
}

func TestUpdateHealthStatus_ATA_ReallocatedSectors(t *testing.T) {
	drive := &DriveInfo{
		IsNVMe:               false,
		HealthStatus:         HealthGood,
		ReallocatedSectors:   5,
		PendingSectors:       0,
		UncorrectableSectors: 0,
		CRCErrors:            0,
	}

	updateHealthStatus(drive)

	if drive.HealthStatus != HealthCaution {
		t.Errorf("HealthStatus = %q, want %q", drive.HealthStatus, HealthCaution)
	}
}

func TestUpdateHealthStatus_ATA_HighReallocated(t *testing.T) {
	drive := &DriveInfo{
		IsNVMe:               false,
		HealthStatus:         HealthGood,
		ReallocatedSectors:   150,
		PendingSectors:       0,
		UncorrectableSectors: 0,
		CRCErrors:            0,
	}

	updateHealthStatus(drive)

	if drive.HealthStatus != HealthBad {
		t.Errorf("HealthStatus = %q, want %q", drive.HealthStatus, HealthBad)
	}
}

func TestUpdateHealthStatus_ATA_UncorrectableSectors(t *testing.T) {
	drive := &DriveInfo{
		IsNVMe:               false,
		HealthStatus:         HealthGood,
		ReallocatedSectors:   0,
		PendingSectors:       0,
		UncorrectableSectors: 150,
		CRCErrors:            0,
	}

	updateHealthStatus(drive)

	if drive.HealthStatus != HealthBad {
		t.Errorf("HealthStatus = %q, want %q", drive.HealthStatus, HealthBad)
	}
}

func TestSmartAttributeGetStatus(t *testing.T) {
	tests := []struct {
		name     string
		attr     SmartAttribute
		expected HealthStatus
	}{
		{
			name: "critical attribute with zero value",
			attr: SmartAttribute{
				ID:        AttrReallocatedSectors,
				RawValue:  0,
				Value:     100,
				Threshold: 10,
			},
			expected: HealthGood,
		},
		{
			name: "critical attribute with small value",
			attr: SmartAttribute{
				ID:        AttrReallocatedSectors,
				RawValue:  5,
				Value:     100,
				Threshold: 10,
			},
			expected: HealthCaution,
		},
		{
			name: "critical attribute with high value",
			attr: SmartAttribute{
				ID:        AttrReallocatedSectors,
				RawValue:  200,
				Value:     100,
				Threshold: 10,
			},
			expected: HealthBad,
		},
		{
			name: "non-critical attribute at threshold",
			attr: SmartAttribute{
				ID:        AttrPowerOnHours,
				RawValue:  5000,
				Value:     10,
				Threshold: 10,
			},
			expected: HealthBad,
		},
		{
			name: "non-critical attribute near threshold",
			attr: SmartAttribute{
				ID:        AttrPowerOnHours,
				RawValue:  5000,
				Value:     15,
				Threshold: 10,
			},
			expected: HealthCaution,
		},
		{
			name: "non-critical attribute healthy",
			attr: SmartAttribute{
				ID:        AttrPowerOnHours,
				RawValue:  5000,
				Value:     100,
				Threshold: 10,
			},
			expected: HealthGood,
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

func TestSmartAttributeIsCritical(t *testing.T) {
	tests := []struct {
		name     string
		attrID   int
		expected bool
	}{
		{"reallocated sectors", AttrReallocatedSectors, true},
		{"pending sectors", AttrPendingSectors, true},
		{"offline uncorrectable", AttrOfflineUncorrectable, true},
		{"command timeout", AttrCommandTimeout, false},
		{"udma crc error", AttrUDMACRCError, false},
		{"power on hours", AttrPowerOnHours, false},
		{"temperature", AttrTemperature, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := SmartAttribute{ID: tt.attrID}
			result := attr.IsCritical()
			if result != tt.expected {
				t.Errorf("IsCritical() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDriveInfoGetDriveType(t *testing.T) {
	tests := []struct {
		name     string
		drive    DriveInfo
		expected string
	}{
		{"nvme", DriveInfo{IsNVMe: true}, "NVMe SSD"},
		{"sata ssd", DriveInfo{IsSSD: true}, "SATA SSD"},
		{"hdd", DriveInfo{RotationRate: 7200}, "HDD"},
		{"virtual", DriveInfo{IsVirtual: true}, "Virtual"},
		{"unknown", DriveInfo{}, "Unknown"},
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

func TestDriveInfoHasCriticalIssues(t *testing.T) {
	tests := []struct {
		name     string
		drive    DriveInfo
		expected bool
	}{
		{"healthy", DriveInfo{}, false},
		{"reallocated", DriveInfo{ReallocatedSectors: 1}, true},
		{"pending", DriveInfo{PendingSectors: 1}, true},
		{"uncorrectable", DriveInfo{UncorrectableSectors: 1}, true},
		{"crc errors only", DriveInfo{CRCErrors: 1}, false},
		{"multiple issues", DriveInfo{ReallocatedSectors: 5, CRCErrors: 3}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.drive.HasCriticalIssues()
			if result != tt.expected {
				t.Errorf("HasCriticalIssues() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// --- New tests for exec-dependent functions via runSmartctl mock ---

func TestGetCommonDevicePaths(t *testing.T) {
	paths := GetCommonDevicePaths()
	if len(paths) == 0 {
		t.Error("GetCommonDevicePaths() returned empty slice")
	}
}

func TestGetDriveInfoWithType_Success(t *testing.T) {
	nvmeJSON := loadTestDataRaw(t, "smartctl_nvme.json")
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		return nvmeJSON, nil
	})

	drive, err := GetDriveInfoWithType("/dev/nvme0", "nvme")
	if err != nil {
		t.Fatalf("GetDriveInfoWithType() error = %v", err)
	}
	if drive.Model != "Samsung SSD 980 PRO 1TB" {
		t.Errorf("Model = %q, want %q", drive.Model, "Samsung SSD 980 PRO 1TB")
	}
	if !drive.IsNVMe {
		t.Error("IsNVMe = false, want true")
	}
}

func TestGetDriveInfoWithType_Error(t *testing.T) {
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("smartctl not found")
	})

	_, err := GetDriveInfoWithType("/dev/sda", "ata")
	if err == nil {
		t.Error("GetDriveInfoWithType() error = nil, want error")
	}
}

func TestGetDriveInfoWithType_InvalidJSON(t *testing.T) {
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		return []byte("not json"), nil
	})

	_, err := GetDriveInfoWithType("/dev/sda", "ata")
	if err == nil {
		t.Error("GetDriveInfoWithType() error = nil, want error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("error should mention parse failure, got %q", err.Error())
	}
}

func TestGetDriveInfoWithType_USBFallback(t *testing.T) {
	nvmeJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	callCount := 0
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		callCount++
		// First call (no type / scsi) fails, then "sat" succeeds
		for _, a := range args {
			if a == "sat" {
				return nvmeJSON, nil
			}
		}
		if callCount == 1 {
			return nil, errors.New("scsi failed")
		}
		return nil, errors.New("still failing")
	})

	drive, err := GetDriveInfoWithType("/dev/sda", "scsi")
	if err != nil {
		t.Fatalf("GetDriveInfoWithType(scsi) error = %v", err)
	}
	if drive.Model == "" {
		t.Error("Model should not be empty after USB fallback")
	}
}

func TestGetDriveInfoWithType_NoType(t *testing.T) {
	sataJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		return sataJSON, nil
	})

	drive, err := GetDriveInfoWithType("/dev/sda", "")
	if err != nil {
		t.Fatalf("GetDriveInfoWithType('') error = %v", err)
	}
	if drive.Model != "Samsung SSD 870 EVO 500GB" {
		t.Errorf("Model = %q, want SATA model", drive.Model)
	}
}

func TestGetDriveInfoWithType_Exit64WithDataIsUsable(t *testing.T) {
	sataJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		// smartctl warning exit status 64 with full drive data: common on USB bridges.
		return sataJSON, errors.New("exit status 64")
	})

	drive, err := GetDriveInfoWithType("/dev/sdc", "sat")
	if err != nil {
		t.Fatalf("GetDriveInfoWithType(sat) error = %v, want success with data", err)
	}
	if drive.Model != "Samsung SSD 870 EVO 500GB" {
		t.Errorf("Model = %q, want SATA model from usable output", drive.Model)
	}
}

func TestGetDriveInfoWithType_Exit64FallbackToSat(t *testing.T) {
	sataJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		// First call (scsi, no -d) fails hard; sat attempt returns exit 64 with data.
		for _, a := range args {
			if a == "sat" {
				return sataJSON, errors.New("exit status 64")
			}
		}
		return nil, errors.New("unknown USB bridge")
	})

	drive, err := GetDriveInfoWithType("/dev/sdc", "scsi")
	if err != nil {
		t.Fatalf("GetDriveInfoWithType(scsi) error = %v, want fallback success", err)
	}
	if drive.Model != "Samsung SSD 870 EVO 500GB" {
		t.Errorf("Model = %q, want SATA model from sat fallback", drive.Model)
	}
}

// A failed SMART read that still names the model must not be accepted as a
// successful scan: identification strings (and a bare smart_status boolean)
// are not SMART data. The record must carry ScanError so the textfile reports
// scan_success 0 and omits the health/percent samples.
func TestScanWithSmartctl_FailedReadWithModelIsReportedFailed(t *testing.T) {
	for _, tc := range []struct {
		name       string
		fixture    string
		deviceType string
	}{
		{"sata", "smartctl_partial_sata.json", "sat"},
		{"nvme", "smartctl_partial_nvme.json", "nvme"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := loadTestDataRaw(t, tc.fixture)
			scanJSON, _ := json.Marshal(SmartctlScanResult{
				Devices: []struct {
					Name     string `json:"name"`
					InfoName string `json:"info_name"`
					Type     string `json:"type"`
					Protocol string `json:"protocol"`
				}{
					{Name: "/dev/probe-partial", Type: tc.deviceType},
				},
			})
			withMockedSmartctl(t, func(args ...string) ([]byte, error) {
				if args[0] == "--scan" {
					for _, a := range args {
						if a == "-j" {
							return scanJSON, nil
						}
					}
					return []byte("/dev/probe-partial -d " + tc.deviceType + "\n"), nil
				}
				return raw, errors.New("exit status 4")
			})

			drives, err := ScanDrivesIncludingErrors()
			if err != nil {
				t.Fatalf("ScanDrivesIncludingErrors() error = %v", err)
			}
			if len(drives) != 1 {
				t.Fatalf("ScanDrivesIncludingErrors() = %+v, want exactly one record", drives)
			}
			if drives[0].ScanError == nil {
				t.Fatalf("ScanDrivesIncludingErrors() = %+v, want the failed read reported, not a healthy drive", drives[0])
			}
			if readable := readableDrives(drives); len(readable) != 0 {
				t.Fatalf("readableDrives() = %+v, want none (the TUI path must not expose it)", readable)
			}
		})
	}
}

func TestGetDriveInfo(t *testing.T) {
	sataJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		return sataJSON, nil
	})

	drive, err := GetDriveInfo("/dev/sda")
	if err != nil {
		t.Fatalf("GetDriveInfo() error = %v", err)
	}
	if drive.Model == "" {
		t.Error("GetDriveInfo() should return a drive with model")
	}
}

func TestScanWithSmartctl_JSONScan(t *testing.T) {
	nvmeJSON := loadTestDataRaw(t, "smartctl_nvme.json")
	scanJSON, _ := json.Marshal(SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: "/dev/nvme0", Type: "nvme", Protocol: "NVMe"},
		},
	})

	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		for _, a := range args {
			if a == "--scan" {
				return scanJSON, nil
			}
		}
		// Drive info request
		return nvmeJSON, nil
	})

	drives := scanWithSmartctl()
	if len(drives) == 0 {
		t.Error("scanWithSmartctl() returned no drives")
	}
	if len(drives) > 0 && drives[0].Model != "Samsung SSD 980 PRO 1TB" {
		t.Errorf("Model = %q, want NVMe model", drives[0].Model)
	}
}

func TestScanWithSmartctl_DeduplicatesSamePhysicalDrive(t *testing.T) {
	nvmeJSON := loadTestDataRaw(t, "smartctl_nvme.json")
	scanJSON, _ := json.Marshal(SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: `\\.\PhysicalDrive0`, InfoName: `\\.\PhysicalDrive0`, Type: "nvme", Protocol: "NVMe"},
			{Name: `\\.\PhysicalDrive0`, InfoName: `\\.\PhysicalDrive0`, Type: "auto", Protocol: "ATA"},
		},
	})

	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		for _, a := range args {
			if a == "--scan" {
				return scanJSON, nil
			}
		}
		return nvmeJSON, nil
	})

	drives := scanWithSmartctl()
	if len(drives) != 1 {
		t.Fatalf("scanWithSmartctl() returned %d drives, want 1 after deduplication", len(drives))
	}
	if drives[0].Serial != "S5GXNF0R123456" {
		t.Errorf("Serial = %q, want deduplicated real drive", drives[0].Serial)
	}
}

func TestScanWithSmartctl_TextFallback(t *testing.T) {
	sataJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		for i, a := range args {
			if a == "--scan" {
				// Check if -j flag is also present
				hasJSON := false
				for _, aa := range args {
					if aa == "-j" {
						hasJSON = true
					}
				}
				if hasJSON {
					return nil, errors.New("json scan failed")
				}
				// Text scan output
				return []byte("/dev/sda -d sat # /dev/sda, ATA device\n"), nil
			}
			_ = i
		}
		return sataJSON, nil
	})

	drives := scanWithSmartctl()
	if len(drives) == 0 {
		t.Error("scanWithSmartctl() text fallback returned no drives")
	}
}

// A transient detailed-read failure must not skip the textual discovery
// retry: the fallback decision is based on usable drives, not on records that
// carry an error.
func TestScanWithSmartctl_RetriesTextDiscoveryWhenAllReadsFail(t *testing.T) {
	sataJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	scanJSON, _ := json.Marshal(SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: "/dev/disk99", Type: "sat", Protocol: "ATA"},
		},
	})

	reads := 0
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		if args[0] == "--scan" {
			for _, a := range args {
				if a == "-j" {
					return scanJSON, nil
				}
			}
			return []byte("/dev/disk99 -d sat # /dev/disk99, ATA device\n"), nil
		}
		// Only the transiently-failing disk ever returns data; the common-path
		// fallback keeps failing, so demo data would mean the retry was skipped.
		if !strings.Contains(strings.Join(args, " "), "/dev/disk99") {
			return nil, errors.New("device not found")
		}
		reads++
		if reads == 1 {
			return nil, errors.New("transient detailed-read failure")
		}
		return sataJSON, nil
	})

	drives, err := ScanDrives()
	if err != nil {
		t.Fatalf("ScanDrives() error = %v, want the recovered disk", err)
	}
	if len(drives) != 1 || drives[0].Device != "/dev/disk99" || drives[0].IsDemo {
		t.Fatalf("ScanDrives() = %+v, want readable /dev/disk99", drives)
	}
}

// A device the JSON pass could not read must survive the textual retry when
// the retry only lists some devices: the retry recovers what it can without
// dropping the failures only JSON knew about.
func TestScanWithSmartctl_TextRetryKeepsJSONOnlyFailure(t *testing.T) {
	sataJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	scanJSON, _ := json.Marshal(SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: "/dev/disk98", Type: "sat", Protocol: "ATA"},
			{Name: "/dev/disk99", Type: "sat", Protocol: "ATA"},
		},
	})

	reads := 0
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		if args[0] == "--scan" {
			for _, a := range args {
				if a == "-j" {
					return scanJSON, nil
				}
			}
			// The textual pass only ever learns about /dev/disk99.
			return []byte("/dev/disk99 -d sat # /dev/disk99, ATA device\n"), nil
		}
		switch {
		case strings.Contains(joined, "/dev/disk98"):
			return nil, errors.New("disk98 read failed")
		case strings.Contains(joined, "/dev/disk99"):
			reads++
			if reads == 1 {
				return nil, errors.New("disk99 transient read failure")
			}
			return sataJSON, nil
		}
		return nil, errors.New("device not found")
	})

	drives := scanWithSmartctl()
	byDevice := map[string]DriveInfo{}
	for _, drive := range drives {
		byDevice[drive.Device] = drive
	}
	if _, ok := byDevice["/dev/disk98"]; !ok {
		t.Fatalf("scanWithSmartctl() dropped the JSON-only failure: %+v", drives)
	}
	if byDevice["/dev/disk98"].ScanError == nil {
		t.Fatalf("/dev/disk98 = %+v, want the failed read preserved", byDevice["/dev/disk98"])
	}
	if got := byDevice["/dev/disk99"]; got.ScanError != nil {
		t.Fatalf("/dev/disk99 = %+v, want the retry's successful read", got)
	}
}

// A device the JSON pass listed twice and failed to read twice must still end
// up with exactly one record after the textual retry recovers it: the retry's
// success replaces the base entry rather than leaving a stale duplicate next
// to it (with a serial) or being dropped in favour of the earlier failure
// (without one).
func TestScanWithSmartctl_RetryCoalescesDuplicateBasePaths(t *testing.T) {
	scanJSON, _ := json.Marshal(SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: "/dev/probe-good", Type: "sat", Protocol: "ATA"},
			{Name: "/dev/probe-good", Type: "sat", Protocol: "ATA"},
		},
	})

	for _, tc := range []struct {
		name       string
		raw        []byte
		wantSerial string
	}{
		{"with serial", loadTestDataRaw(t, "smartctl_sata_ssd.json"), "S4HNNS0R123456"},
		{"without serial", withoutSerialNumber(t, loadTestDataRaw(t, "smartctl_sata_ssd.json")), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reads := 0
			withMockedSmartctl(t, func(args ...string) ([]byte, error) {
				if args[0] == "--scan" {
					for _, a := range args {
						if a == "-j" {
							return scanJSON, nil
						}
					}
					return []byte("/dev/probe-good -d sat # /dev/probe-good, ATA device\n"), nil
				}
				reads++
				if reads <= 2 {
					return nil, errors.New("transient detailed-read failure")
				}
				return tc.raw, nil
			})

			drives := scanWithSmartctl()
			if len(drives) != 1 {
				t.Fatalf("scanWithSmartctl() = %d records, want exactly one per device path: %+v", len(drives), drives)
			}
			if drives[0].ScanError != nil {
				t.Fatalf("scanWithSmartctl() kept the stale failure: %+v", drives[0])
			}
			if drives[0].Serial != tc.wantSerial {
				t.Fatalf("Serial = %q, want %q", drives[0].Serial, tc.wantSerial)
			}
		})
	}
}

// The direct JSON return must coalesce duplicate device paths too, not only
// the retry path: when the first detailed read fails and the second succeeds,
// exactly one record per path must survive and the success must win. With a
// serial the two identities used to differ and both were kept; without one the
// earlier failure was kept and the successful read dropped.
func TestScanWithSmartctl_JSONPassCoalescesDuplicatePaths(t *testing.T) {
	scanJSON, _ := json.Marshal(SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: "/dev/probe-good", Type: "sat", Protocol: "ATA"},
			{Name: "/dev/probe-good", Type: "sat", Protocol: "ATA"},
		},
	})

	for _, tc := range []struct {
		name       string
		raw        []byte
		wantSerial string
	}{
		{"with serial", loadTestDataRaw(t, "smartctl_sata_ssd.json"), "S4HNNS0R123456"},
		{"without serial", withoutSerialNumber(t, loadTestDataRaw(t, "smartctl_sata_ssd.json")), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reads := 0
			withMockedSmartctl(t, func(args ...string) ([]byte, error) {
				if args[0] == "--scan" {
					for _, a := range args {
						if a == "-j" {
							return scanJSON, nil
						}
					}
					return nil, errors.New("text scan failed")
				}
				reads++
				if reads == 1 {
					return nil, errors.New("first detailed read failed")
				}
				return tc.raw, nil
			})

			drives := scanWithSmartctl()
			if len(drives) != 1 {
				t.Fatalf("scanWithSmartctl() = %d records, want exactly one per device path: %+v", len(drives), drives)
			}
			if drives[0].ScanError != nil {
				t.Fatalf("scanWithSmartctl() kept the stale failure: %+v", drives[0])
			}
			if drives[0].Serial != tc.wantSerial {
				t.Fatalf("Serial = %q, want %q", drives[0].Serial, tc.wantSerial)
			}
		})
	}
}

// A device-statistics page header is metadata smartctl prints before the
// entries, and an entry whose valid flag is false carries no value: neither
// is a SMART measurement, so neither may satisfy the "usable data" criterion.
func TestHasSmartData_DeviceStatisticsNeedsValidEntry(t *testing.T) {
	for _, tc := range []struct {
		name string
		json string
		want bool
	}{
		{
			name: "page header only",
			json: `{"device":{"protocol":"ATA"},"ata_device_statistics":{"pages":[{"number":1,"name":"General Statistics","revision":1}]}}`,
			want: false,
		},
		{
			name: "invalid entry",
			json: `{"device":{"protocol":"ATA"},"ata_device_statistics":{"pages":[{"number":1,"table":[{"offset":24,"flags":{"valid":false}}]}]}}`,
			want: false,
		},
		{
			name: "valid entry",
			json: `{"device":{"protocol":"ATA"},"ata_device_statistics":{"pages":[{"number":1,"table":[{"offset":24,"value":42,"flags":{"valid":true}}]}]}}`,
			want: true,
		},
		{
			name: "attribute table",
			json: `{"device":{"protocol":"ATA"},"ata_smart_attributes":{"table":[{"id":9,"name":"Power_On_Hours"}]}}`,
			want: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var s SmartctlOutput
			if err := json.Unmarshal([]byte(tc.json), &s); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if got := hasSmartData(s); got != tc.want {
				t.Fatalf("hasSmartData() = %v, want %v", got, tc.want)
			}
		})
	}
}

// A failed read that still prints the device-statistics page header (or an
// invalid entry) must be reported failed: accepting it would export
// health_status BAD/0 and scan_success 1 for a drive whose read failed.
func TestGetDriveInfoWithType_RejectsStatisticsWithoutEntries(t *testing.T) {
	const headerOnly = `{
	  "device": {"name": "/dev/probe-page", "type": "sat", "protocol": "ATA"},
	  "model_name": "PAGE HEADER ONLY",
	  "serial_number": "PAGEONLY",
	  "smart_status": {"passed": true},
	  "ata_device_statistics": {"pages": [{"number": 1, "name": "General Statistics", "revision": 1}]}
	}`
	const invalidEntry = `{
	  "device": {"name": "/dev/probe-page", "type": "sat", "protocol": "ATA"},
	  "model_name": "INVALID ENTRY",
	  "serial_number": "INVALID",
	  "smart_status": {"passed": true},
	  "ata_device_statistics": {"pages": [{"number": 1, "table": [{"offset": 24, "flags": {"valid": false}}]}]}
	}`
	const validEntry = `{
	  "device": {"name": "/dev/probe-page", "type": "sat", "protocol": "ATA"},
	  "model_name": "VALID ENTRY",
	  "serial_number": "VALID",
	  "rotation_rate": 7200,
	  "ata_device_statistics": {"pages": [{"number": 1, "table": [{"offset": 24, "value": 42, "flags": {"valid": true}}]}]}
	}`

	for _, tc := range []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"page header only", headerOnly, true},
		{"invalid entry", invalidEntry, true},
		{"valid entry", validEntry, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withMockedSmartctl(t, func(args ...string) ([]byte, error) {
				return []byte(tc.raw), errors.New("exit status 4")
			})
			drive, err := GetDriveInfoWithType("/dev/probe-page", "sat")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("GetDriveInfoWithType() = %+v, want rejection of a header-only/invalid read", drive)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetDriveInfoWithType() error = %v, want a usable device-statistics entry accepted", err)
			}
		})
	}
}

// Two successful reads of the same path must coalesce to one record even when
// no read failed: the direct JSON return used to bypass path coalescing, so
// both identities survived and the textfile exported two contradictory
// health_status samples for one device path.
func TestScanWithSmartctl_JSONPassCoalescesTwoSuccessfulPaths(t *testing.T) {
	scanJSON, _ := json.Marshal(SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: "/dev/probe-dup", Type: "sat", Protocol: "ATA"},
			{Name: "/dev/probe-dup", Type: "sat", Protocol: "ATA"},
		},
	})

	good := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	bad := withoutSerialAndFailedHealth(t, loadTestDataRaw(t, "smartctl_sata_ssd.json"))

	reads := 0
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		if args[0] == "--scan" {
			for _, a := range args {
				if a == "-j" {
					return scanJSON, nil
				}
			}
			return nil, errors.New("text scan failed")
		}
		reads++
		if reads == 1 {
			return good, nil
		}
		return bad, nil
	})

	drives := scanWithSmartctl()
	if len(drives) != 1 {
		t.Fatalf("scanWithSmartctl() = %d records, want exactly one per device path: %+v", len(drives), drives)
	}
	if drives[0].Device != "/dev/probe-dup" {
		t.Fatalf("Device = %q, want /dev/probe-dup", drives[0].Device)
	}
}

// smart_status is absent when the status command failed, and an absent field is
// not a failed verdict. A drive that still yields a usable attribute is
// accepted, but its health must stay UNKNOWN so the textfile omits
// health_percent instead of inventing BAD/0. An explicit false still means BAD.
func TestGetDriveInfoWithType_AbsentSmartStatusStaysUnknown(t *testing.T) {
	const attrOnly = `{
	  "device": {"name": "/dev/probe-attr", "type": "sat", "protocol": "ATA"},
	  "model_name": "ATTRIBUTE ONLY",
	  "serial_number": "ATTRONLY",
	  "rotation_rate": 7200,
	  "power_on_time": {"hours": 42},
	  "ata_smart_attributes": {"table": [{"id": 9, "name": "Power_On_Hours", "value": 99, "worst": 99, "thresh": 0, "raw": {"value": 42, "string": "42"}}]}
	}`
	const passedFalse = `{
	  "device": {"name": "/dev/probe-attr", "type": "sat", "protocol": "ATA"},
	  "model_name": "EXPLICIT FALSE",
	  "serial_number": "FALSE",
	  "rotation_rate": 7200,
	  "smart_status": {"passed": false},
	  "ata_smart_attributes": {"table": [{"id": 9, "name": "Power_On_Hours", "value": 99, "worst": 99, "thresh": 0, "raw": {"value": 42, "string": "42"}}]}
	}`
	const passedTrue = `{
	  "device": {"name": "/dev/probe-attr", "type": "sat", "protocol": "ATA"},
	  "model_name": "EXPLICIT TRUE",
	  "serial_number": "TRUE",
	  "rotation_rate": 7200,
	  "smart_status": {"passed": true},
	  "ata_smart_attributes": {"table": [{"id": 9, "name": "Power_On_Hours", "value": 99, "worst": 99, "thresh": 0, "raw": {"value": 42, "string": "42"}}]}
	}`

	for _, tc := range []struct {
		name       string
		raw        string
		wantHealth HealthStatus
	}{
		{"absent passed", attrOnly, HealthUnknown},
		{"explicit false", passedFalse, HealthBad},
		{"explicit true", passedTrue, HealthGood},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withMockedSmartctl(t, func(args ...string) ([]byte, error) {
				return []byte(tc.raw), errors.New("exit status 4")
			})
			drive, err := GetDriveInfoWithType("/dev/probe-attr", "sat")
			if err != nil {
				t.Fatalf("GetDriveInfoWithType() error = %v, want the usable attribute accepted", err)
			}
			if drive.HealthStatus != tc.wantHealth {
				t.Fatalf("HealthStatus = %q, want %q", drive.HealthStatus, tc.wantHealth)
			}
		})
	}
}

func withoutSerialAndFailedHealth(t *testing.T, raw []byte) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	delete(doc, "serial_number")
	doc["smart_status"] = map[string]any{"passed": false}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return out
}

func withoutSerialNumber(t *testing.T, raw []byte) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	delete(doc, "serial_number")
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return out
}

func TestScanWithSmartctl_AllFail(t *testing.T) {
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("smartctl not found")
	})

	drives := scanWithSmartctl()
	if drives != nil {
		t.Errorf("scanWithSmartctl() = %v, want nil when all scans fail", drives)
	}
}

func TestScanWithSmartctl_PreservesFailedDevices(t *testing.T) {
	sataJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	scanJSON, _ := json.Marshal(SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: "/dev/sda", Type: "sat", Protocol: "ATA"},
			{Name: "/dev/sdb", Type: "sat", Protocol: "ATA"},
		},
	})

	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		for _, a := range args {
			if a == "--scan" {
				return scanJSON, nil
			}
			if a == "/dev/sdb" {
				return nil, errors.New("smartctl failed")
			}
		}
		return sataJSON, nil
	})

	drives := scanWithSmartctl()
	failed := false
	for _, drive := range drives {
		if drive.Device == "/dev/sdb" && drive.ScanError != nil {
			failed = true
		}
	}
	if !failed {
		t.Fatalf("scanWithSmartctl() dropped the failed device: %+v", drives)
	}

	readable, err := ScanDrives()
	if err != nil {
		t.Fatalf("ScanDrives() error = %v", err)
	}
	for _, drive := range readable {
		if drive.Device == "/dev/sdb" {
			t.Fatal("ScanDrives() exposed the failed device to the TUI")
		}
	}

	withErrors, err := ScanDrivesIncludingErrors()
	if err != nil {
		t.Fatalf("ScanDrivesIncludingErrors() error = %v", err)
	}
	kept := false
	for _, drive := range withErrors {
		if drive.Device == "/dev/sdb" && drive.ScanError != nil {
			kept = true
		}
	}
	if !kept {
		t.Fatalf("ScanDrivesIncludingErrors() = %+v, want the failed device kept", withErrors)
	}
}

func TestScanCommonDevices(t *testing.T) {
	sataJSON := loadTestDataRaw(t, "smartctl_sata_ssd.json")
	callCount := 0
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		callCount++
		// Only succeed for the first device
		if callCount == 1 {
			return sataJSON, nil
		}
		return nil, errors.New("device not found")
	})

	drives := scanCommonDevices()
	if len(drives) == 0 {
		t.Error("scanCommonDevices() should find at least one drive")
	}
}

func TestScanDrives_WithSmartctl(t *testing.T) {
	nvmeJSON := loadTestDataRaw(t, "smartctl_nvme.json")
	scanJSON, _ := json.Marshal(SmartctlScanResult{
		Devices: []struct {
			Name     string `json:"name"`
			InfoName string `json:"info_name"`
			Type     string `json:"type"`
			Protocol string `json:"protocol"`
		}{
			{Name: "/dev/nvme0", Type: "nvme", Protocol: "NVMe"},
		},
	})

	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		for _, a := range args {
			if a == "--scan" {
				return scanJSON, nil
			}
		}
		return nvmeJSON, nil
	})

	drives, err := ScanDrives()
	if err != nil {
		t.Errorf("ScanDrives() error = %v, want nil", err)
	}
	if len(drives) == 0 {
		t.Error("ScanDrives() returned no drives")
	}
}

func TestScanDrives_FallbackDemo(t *testing.T) {
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("no smartctl")
	})

	drives, err := ScanDrives()
	if err == nil {
		t.Error("ScanDrives() should return an error when falling back to demo data")
	}
	if len(drives) == 0 {
		t.Error("ScanDrives() should return demo data as fallback")
	}
	// Demo data should have IsDemo set
	for _, d := range drives {
		if !d.IsDemo {
			t.Error("Fallback drives should be demo data")
		}
	}
}

func TestParseATADrive_Flags(t *testing.T) {
	// Test with WhenFailed, Prefailure=false, UpdatedOnline=false
	s := SmartctlOutput{
		AtaSmartAttributes: SmartctlAtaSmartAttributes{
			Table: []SmartctlAtaAttribute{
				{
					ID:         5,
					Name:       "Reallocated_Sector_Ct",
					Value:      90,
					Worst:      90,
					Thresh:     10,
					WhenFailed: "past",
					Flags: struct {
						Value         int    `json:"value"`
						String        string `json:"string"`
						Prefailure    bool   `json:"prefailure"`
						UpdatedOnline bool   `json:"updated_online"`
						Performance   bool   `json:"performance"`
						ErrorRate     bool   `json:"error_rate"`
						EventCount    bool   `json:"event_count"`
						AutoKeep      bool   `json:"auto_keep"`
					}{
						Prefailure:    false,
						UpdatedOnline: false,
						String:        "------",
					},
					Raw: struct {
						Value  int64  `json:"value"`
						String string `json:"string"`
					}{Value: 10, String: "10"},
				},
				{
					ID:    194,
					Name:  "Temperature_Celsius",
					Value: 100,
					Worst: 100,
					Flags: struct {
						Value         int    `json:"value"`
						String        string `json:"string"`
						Prefailure    bool   `json:"prefailure"`
						UpdatedOnline bool   `json:"updated_online"`
						Performance   bool   `json:"performance"`
						ErrorRate     bool   `json:"error_rate"`
						EventCount    bool   `json:"event_count"`
						AutoKeep      bool   `json:"auto_keep"`
					}{
						Prefailure:    true,
						UpdatedOnline: true,
						String:        "PO----",
					},
					Raw: struct {
						Value  int64  `json:"value"`
						String string `json:"string"`
					}{Value: 35, String: "35"},
				},
			},
		},
	}

	drive := &DriveInfo{}
	parseATADrive(drive, s)

	if len(drive.Attributes) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(drive.Attributes))
	}

	// First attr: Prefailure=false → Old_age, UpdatedOnline=false → Offline
	attr0 := drive.Attributes[0]
	if attr0.Type != "Old_age" {
		t.Errorf("Attr 0 Type = %q, want %q", attr0.Type, "Old_age")
	}
	if attr0.Updated != "Offline" {
		t.Errorf("Attr 0 Updated = %q, want %q", attr0.Updated, "Offline")
	}
	if attr0.WhenFailed != "past" {
		t.Errorf("Attr 0 WhenFailed = %q, want %q", attr0.WhenFailed, "past")
	}

	// Second attr: Prefailure=true → Pre-fail, UpdatedOnline=true → Always
	attr1 := drive.Attributes[1]
	if attr1.Type != "Pre-fail" {
		t.Errorf("Attr 1 Type = %q, want %q", attr1.Type, "Pre-fail")
	}
	if attr1.Updated != "Always" {
		t.Errorf("Attr 1 Updated = %q, want %q", attr1.Updated, "Always")
	}

	// Temperature should be set from attr 194
	if drive.Temperature != 35 {
		t.Errorf("Temperature = %d, want 35", drive.Temperature)
	}

	// ReallocatedSectors should be set from attr 5
	if drive.ReallocatedSectors != 10 {
		t.Errorf("ReallocatedSectors = %d, want 10", drive.ReallocatedSectors)
	}
}

func TestGetDriveInfoWithType_AllUSBFallbacksFail(t *testing.T) {
	withMockedSmartctl(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("always fail")
	})

	_, err := GetDriveInfoWithType("/dev/sda", "")
	if err == nil {
		t.Error("expected error when all USB fallbacks fail")
	}
	if !strings.Contains(err.Error(), "smartctl failed") {
		t.Errorf("error should mention smartctl failure, got %q", err.Error())
	}
}

func TestDriveInfoGetAttributeCount_Extended(t *testing.T) {
	tests := []struct {
		name     string
		drive    DriveInfo
		expected int
	}{
		{
			"nvme with attributes",
			DriveInfo{IsNVMe: true, NVMeAttributes: []NVMeAttribute{{}, {}, {}}},
			3,
		},
		{
			"sata with attributes",
			DriveInfo{Attributes: []SmartAttribute{{}, {}}},
			2,
		},
		{
			"no attributes",
			DriveInfo{},
			0,
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
