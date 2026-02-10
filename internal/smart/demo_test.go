package smart

import (
	"testing"
)

func TestCreateDemoData(t *testing.T) {
	drives := CreateDemoData()

	if len(drives) != 3 {
		t.Errorf("CreateDemoData() returned %d drives, want 3", len(drives))
	}

	// All demo drives should be marked as demo
	for i, drive := range drives {
		if !drive.IsDemo {
			t.Errorf("Drive %d should have IsDemo=true", i)
		}
	}

	// Check NVMe demo drive
	nvmeDrive := drives[0]
	if !nvmeDrive.IsNVMe {
		t.Error("First demo drive should be NVMe")
	}
	if nvmeDrive.Model != "Samsung SSD 970 EVO Plus 1TB" {
		t.Errorf("NVMe model = %q, want %q", nvmeDrive.Model, "Samsung SSD 970 EVO Plus 1TB")
	}
	if nvmeDrive.HealthStatus != HealthGood {
		t.Errorf("NVMe HealthStatus = %q, want %q", nvmeDrive.HealthStatus, HealthGood)
	}
	if len(nvmeDrive.NVMeAttributes) == 0 {
		t.Error("NVMe drive should have NVMeAttributes")
	}

	// Check good HDD demo drive
	hddGood := drives[1]
	if hddGood.IsNVMe || hddGood.IsSSD {
		t.Error("Second demo drive should be HDD")
	}
	if hddGood.RotationRate != 5400 {
		t.Errorf("HDD RotationRate = %d, want %d", hddGood.RotationRate, 5400)
	}
	if hddGood.HealthStatus != HealthGood {
		t.Errorf("HDD HealthStatus = %q, want %q", hddGood.HealthStatus, HealthGood)
	}
	if len(hddGood.Attributes) == 0 {
		t.Error("HDD should have SMART Attributes")
	}

	// Check caution HDD demo drive
	hddCaution := drives[2]
	if hddCaution.HealthStatus != HealthCaution {
		t.Errorf("Caution HDD HealthStatus = %q, want %q", hddCaution.HealthStatus, HealthCaution)
	}
	if hddCaution.ReallocatedSectors == 0 {
		t.Error("Caution HDD should have reallocated sectors")
	}
	if hddCaution.PendingSectors == 0 {
		t.Error("Caution HDD should have pending sectors")
	}
}

func TestDemoDataIntegrity(t *testing.T) {
	drives := CreateDemoData()

	for i, drive := range drives {
		// All demo drives should have basic info
		if drive.Device == "" {
			t.Errorf("Drive %d has empty Device", i)
		}
		if drive.Model == "" {
			t.Errorf("Drive %d has empty Model", i)
		}
		if drive.Serial == "" {
			t.Errorf("Drive %d has empty Serial", i)
		}
		if drive.Capacity == "" {
			t.Errorf("Drive %d has empty Capacity", i)
		}
		if drive.CapacityBytes == 0 {
			t.Errorf("Drive %d has zero CapacityBytes", i)
		}

		// All should be SMART enabled
		if !drive.SmartSupported {
			t.Errorf("Drive %d SmartSupported = false", i)
		}
		if !drive.SmartEnabled {
			t.Errorf("Drive %d SmartEnabled = false", i)
		}

		// All should have valid health status
		validStatus := map[HealthStatus]bool{
			HealthGood:    true,
			HealthCaution: true,
			HealthBad:     true,
			HealthUnknown: true,
		}
		if !validStatus[drive.HealthStatus] {
			t.Errorf("Drive %d has invalid HealthStatus: %q", i, drive.HealthStatus)
		}

		// Temperature should be reasonable
		if drive.Temperature < 0 || drive.Temperature > 100 {
			t.Errorf("Drive %d has unreasonable Temperature: %d", i, drive.Temperature)
		}

		// PowerOnHours should be positive
		if drive.PowerOnHours < 0 {
			t.Errorf("Drive %d has negative PowerOnHours: %d", i, drive.PowerOnHours)
		}

		// LastUpdate should not be zero
		if drive.LastUpdate.IsZero() {
			t.Errorf("Drive %d has zero LastUpdate", i)
		}
	}
}

func TestDemoNVMeAttributes(t *testing.T) {
	drives := CreateDemoData()
	nvmeDrive := drives[0]

	if len(nvmeDrive.NVMeAttributes) != 14 {
		t.Errorf("NVMe drive has %d attributes, want 14", len(nvmeDrive.NVMeAttributes))
	}

	// Check specific attributes exist
	attrNames := make(map[string]bool)
	for _, attr := range nvmeDrive.NVMeAttributes {
		attrNames[attr.Name] = true

		// All attributes should have valid status
		validStatus := map[HealthStatus]bool{
			HealthGood:    true,
			HealthCaution: true,
			HealthBad:     true,
		}
		if !validStatus[attr.Status] {
			t.Errorf("Attribute %q has invalid Status: %q", attr.Name, attr.Status)
		}

		// All attributes should have a description
		if attr.Description == "" {
			t.Errorf("Attribute %q has empty Description", attr.Name)
		}
	}

	expectedAttrs := []string{
		"Temperature",
		"Available Spare",
		"Percentage Used",
		"Data Read",
		"Data Written",
		"Power Cycles",
		"Power On Hours",
		"Media Errors",
	}

	for _, name := range expectedAttrs {
		if !attrNames[name] {
			t.Errorf("Missing expected attribute: %q", name)
		}
	}
}

func TestDemoHDDAttributes(t *testing.T) {
	drives := CreateDemoData()
	hddDrive := drives[1]

	if len(hddDrive.Attributes) == 0 {
		t.Fatal("HDD drive has no SMART attributes")
	}

	// Check for critical attributes
	attrIDs := make(map[int]bool)
	for _, attr := range hddDrive.Attributes {
		attrIDs[attr.ID] = true

		// Validate attribute structure
		if attr.Name == "" {
			t.Errorf("Attribute ID %d has empty Name", attr.ID)
		}
		if attr.Value < 0 || attr.Value > 255 {
			t.Errorf("Attribute %q has invalid Value: %d", attr.Name, attr.Value)
		}
		if attr.Worst < 0 || attr.Worst > 255 {
			t.Errorf("Attribute %q has invalid Worst: %d", attr.Name, attr.Worst)
		}
	}

	// Check for common HDD attributes
	commonAttrs := []int{
		AttrReallocatedSectors,
		AttrPowerOnHours,
		AttrTemperature,
		AttrPendingSectors,
	}

	for _, id := range commonAttrs {
		if !attrIDs[id] {
			t.Errorf("Missing common HDD attribute ID: %d", id)
		}
	}
}
