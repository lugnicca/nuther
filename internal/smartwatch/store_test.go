package smartwatch

import (
	"path/filepath"
	"testing"
	"time"

	"nuther/internal/smart"
)

func TestStoreSaveSnapshotMaintainsIndexAndReadableSnapshot(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	drive := testDrive("/dev/sda", "SN123")

	record, err := store.SaveSnapshot(now, ReasonStartup, drive)
	if err != nil {
		t.Fatalf("SaveSnapshot() error = %v", err)
	}
	if record.Reason != ReasonStartup {
		t.Fatalf("record reason = %q, want %q", record.Reason, ReasonStartup)
	}
	if record.Path != filepath.Join("snapshots", record.ID+".json") {
		t.Fatalf("record path = %q, want snapshot path for id", record.Path)
	}

	index, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex() error = %v", err)
	}
	if len(index.Snapshots) != 1 {
		t.Fatalf("index snapshots = %d, want 1", len(index.Snapshots))
	}
	device := index.Devices[record.Device.Key]
	if device.LastSnapshot != record.ID {
		t.Fatalf("device last snapshot = %q, want %q", device.LastSnapshot, record.ID)
	}

	snapshot, err := store.ReadSnapshot(record.ID)
	if err != nil {
		t.Fatalf("ReadSnapshot() error = %v", err)
	}
	if snapshot.SMART.Serial != drive.Serial {
		t.Fatalf("snapshot serial = %q, want %q", snapshot.SMART.Serial, drive.Serial)
	}
}

func TestStoreAvoidsSnapshotIDCollision(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	drive := testDrive("/dev/sda", "SN123")

	first, err := store.SaveSnapshot(now, ReasonManual, drive)
	if err != nil {
		t.Fatalf("first SaveSnapshot() error = %v", err)
	}
	second, err := store.SaveSnapshot(now, ReasonManual, drive)
	if err != nil {
		t.Fatalf("second SaveSnapshot() error = %v", err)
	}
	if first.ID == second.ID {
		t.Fatalf("snapshot ids collided: %q", first.ID)
	}
}

func testDrive(device, serial string) smart.DriveInfo {
	return smart.DriveInfo{
		Device:        device,
		Model:         "Test Drive",
		Serial:        serial,
		WWN:           "wwn-" + serial,
		Capacity:      "1 TB",
		CapacityBytes: 1_000_000_000_000,
		Interface:     "SATA SSD",
		HealthStatus:  smart.HealthGood,
		HealthPassed:  true,
	}
}
