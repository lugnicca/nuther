package smart

import (
	"testing"
)

func TestCriticalAttributesMap(t *testing.T) {
	// Verify critical attributes are properly defined
	expectedCritical := []int{
		AttrReallocatedSectors,
		AttrReportedUncorrect,
		AttrPendingSectors,
		AttrOfflineUncorrectable,
	}

	for _, id := range expectedCritical {
		if !CriticalAttributes[id] {
			t.Errorf("Attribute ID %d should be in CriticalAttributes", id)
		}
	}

	// Verify count
	if len(CriticalAttributes) != len(expectedCritical) {
		t.Errorf("CriticalAttributes has %d entries, want %d", len(CriticalAttributes), len(expectedCritical))
	}

	// Connection issues should NOT be in CriticalAttributes
	connectionIDs := []int{AttrUDMACRCError, AttrCommandTimeout}
	for _, id := range connectionIDs {
		if CriticalAttributes[id] {
			t.Errorf("Attribute ID %d should NOT be in CriticalAttributes (it's a connection issue)", id)
		}
	}
}

func TestConnectionAttributesMap(t *testing.T) {
	expectedConnection := []int{
		AttrCommandTimeout,
		AttrUDMACRCError,
	}

	for _, id := range expectedConnection {
		if !ConnectionAttributes[id] {
			t.Errorf("Attribute ID %d should be in ConnectionAttributes", id)
		}
	}

	if len(ConnectionAttributes) != len(expectedConnection) {
		t.Errorf("ConnectionAttributes has %d entries, want %d", len(ConnectionAttributes), len(expectedConnection))
	}
}


func TestAttributeIDConstants(t *testing.T) {
	// Verify attribute IDs are unique and match common SMART specs
	tests := []struct {
		name string
		id   int
	}{
		{"AttrRawReadErrorRate", AttrRawReadErrorRate},
		{"AttrSpinUpTime", AttrSpinUpTime},
		{"AttrStartStopCount", AttrStartStopCount},
		{"AttrReallocatedSectors", AttrReallocatedSectors},
		{"AttrSeekErrorRate", AttrSeekErrorRate},
		{"AttrPowerOnHours", AttrPowerOnHours},
		{"AttrSpinRetryCount", AttrSpinRetryCount},
		{"AttrPowerCycleCount", AttrPowerCycleCount},
		{"AttrTemperature", AttrTemperature},
		{"AttrPendingSectors", AttrPendingSectors},
		{"AttrOfflineUncorrectable", AttrOfflineUncorrectable},
		{"AttrUDMACRCError", AttrUDMACRCError},
	}

	// Check for common IDs
	expectedIDs := map[int]string{
		1:   "Raw_Read_Error_Rate",
		5:   "Reallocated_Sector_Ct",
		9:   "Power_On_Hours",
		194: "Temperature_Celsius",
		197: "Current_Pending_Sector",
		198: "Offline_Uncorrectable",
		199: "UDMA_CRC_Error_Count",
	}

	for id, name := range expectedIDs {
		found := false
		for _, tt := range tests {
			if tt.id == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected attribute ID %d (%s) not found in constants", id, name)
		}
	}
}

func TestAttributeIDValues(t *testing.T) {
	// Verify specific ID values match SMART specification
	if AttrRawReadErrorRate != 1 {
		t.Errorf("AttrRawReadErrorRate = %d, want 1", AttrRawReadErrorRate)
	}
	if AttrReallocatedSectors != 5 {
		t.Errorf("AttrReallocatedSectors = %d, want 5", AttrReallocatedSectors)
	}
	if AttrPowerOnHours != 9 {
		t.Errorf("AttrPowerOnHours = %d, want 9", AttrPowerOnHours)
	}
	if AttrTemperature != 194 {
		t.Errorf("AttrTemperature = %d, want 194", AttrTemperature)
	}
	if AttrPendingSectors != 197 {
		t.Errorf("AttrPendingSectors = %d, want 197", AttrPendingSectors)
	}
	if AttrOfflineUncorrectable != 198 {
		t.Errorf("AttrOfflineUncorrectable = %d, want 198", AttrOfflineUncorrectable)
	}
	if AttrUDMACRCError != 199 {
		t.Errorf("AttrUDMACRCError = %d, want 199", AttrUDMACRCError)
	}
}
