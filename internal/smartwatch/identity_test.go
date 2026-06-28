package smartwatch

import (
	"strings"
	"testing"

	"nuther/internal/smart"
)

func TestStableDeviceKeyPrefersSerial(t *testing.T) {
	drive := smart.DriveInfo{
		Device:        "/dev/sda",
		Model:         "Model A",
		Serial:        " SN 123/ABC ",
		WWN:           "wwn-1",
		CapacityBytes: 100,
	}

	got := StableDeviceKey(drive)
	if got != "serial-sn-123-abc" {
		t.Fatalf("StableDeviceKey() = %q, want serial-sn-123-abc", got)
	}
}

func TestStableDeviceKeyFallsBackToHashedDriveIdentity(t *testing.T) {
	drive := smart.DriveInfo{
		Device:   "/dev/sdb",
		Model:    "Model B",
		WWN:      "wwn-2",
		Capacity: "1 TB",
	}

	got := StableDeviceKey(drive)
	if !strings.HasPrefix(got, "drive-") {
		t.Fatalf("StableDeviceKey() = %q, want drive- prefix", got)
	}
	if got != StableDeviceKey(drive) {
		t.Fatalf("StableDeviceKey() must be deterministic")
	}

	drive.Device = "/dev/sdc"
	changed := StableDeviceKey(drive)
	if changed == got {
		t.Fatalf("fallback key should include device path when no serial is available")
	}
}
