package smartwatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"nuther/internal/smart"
)

func TestWatcherDetectsStartupAndNewDeviceSnapshots(t *testing.T) {
	store := NewStore(t.TempDir())
	hub := NewEventHub("")
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	scans := [][]string{
		{"SN-A"},
		{"SN-A", "SN-B"},
	}
	scanIndex := 0
	watcher := NewWatcher(WatcherConfig{
		Store: store,
		Hub:   hub,
		Scanner: func() ([]smart.DriveInfo, error) {
			return nil, nil
		},
		Now: func() time.Time {
			now = now.Add(time.Second)
			return now
		},
	})
	watcher.scanner = func() ([]smart.DriveInfo, error) {
		serials := scans[scanIndex]
		if scanIndex < len(scans)-1 {
			scanIndex++
		}
		drives := make([]smart.DriveInfo, 0, len(serials))
		for i, serial := range serials {
			drives = append(drives, testDrive("/dev/sd"+string(rune('a'+i)), serial))
		}
		return drives, nil
	}

	seen := map[string]bool{}
	archived := map[string]bool{}
	if err := watcher.detectNewDevices(t.Context(), seen, archived, true); err != nil {
		t.Fatalf("startup detectNewDevices() error = %v", err)
	}
	if err := watcher.detectNewDevices(t.Context(), seen, archived, false); err != nil {
		t.Fatalf("new-device detectNewDevices() error = %v", err)
	}

	index, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex() error = %v", err)
	}
	if len(index.Snapshots) != 2 {
		t.Fatalf("snapshots = %d, want 2", len(index.Snapshots))
	}
	if index.Snapshots[0].Reason != ReasonStartup {
		t.Fatalf("first reason = %q, want %q", index.Snapshots[0].Reason, ReasonStartup)
	}
	if index.Snapshots[1].Reason != ReasonNewDevice {
		t.Fatalf("second reason = %q, want %q", index.Snapshots[1].Reason, ReasonNewDevice)
	}
}

func TestWatcherSnapshotOnceUsesManualReason(t *testing.T) {
	store := NewStore(t.TempDir())
	watcher := NewWatcher(WatcherConfig{
		Store: store,
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now: func() time.Time {
			return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
		},
	})

	records, err := watcher.SnapshotOnce(t.Context(), ReasonManual)
	if err != nil {
		t.Fatalf("SnapshotOnce() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1", len(records))
	}
	if records[0].Reason != ReasonManual {
		t.Fatalf("reason = %q, want %q", records[0].Reason, ReasonManual)
	}
}

func TestWatcherRunContinuesAfterStartupScanError(t *testing.T) {
	store := NewStore(t.TempDir())
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	scanCalls := 0
	watcher := NewWatcher(WatcherConfig{
		Store:            store,
		Hub:              NewEventHub(""),
		PollInterval:     time.Millisecond,
		SnapshotInterval: 0,
		Scanner: func() ([]smart.DriveInfo, error) {
			scanCalls++
			if scanCalls == 1 {
				return nil, errors.New("no non-demo drives detected")
			}
			cancel()
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now: func() time.Time {
			return time.Date(2026, 6, 28, 10, 0, scanCalls, 0, time.UTC)
		},
	})

	err := watcher.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	if scanCalls < 2 {
		t.Fatalf("scan calls = %d, want at least 2", scanCalls)
	}

	index, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex() error = %v", err)
	}
	if len(index.Snapshots) != 1 {
		t.Fatalf("snapshots = %d, want 1", len(index.Snapshots))
	}
	if index.Snapshots[0].Reason != ReasonNewDevice {
		t.Fatalf("reason = %q, want %q", index.Snapshots[0].Reason, ReasonNewDevice)
	}
}

func TestWatcherStartupSkipsAlreadyArchivedDevice(t *testing.T) {
	store := NewStore(t.TempDir())
	drive := testDrive("/dev/sda", "SN-A")
	if _, err := store.SaveSnapshot(time.Date(2026, 6, 28, 9, 0, 0, 0, time.UTC), ReasonStartup, drive); err != nil {
		t.Fatalf("SaveSnapshot() error = %v", err)
	}

	watcher := NewWatcher(WatcherConfig{
		Store: store,
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{drive}, nil
		},
		Now: func() time.Time {
			return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
		},
	})
	index, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex() error = %v", err)
	}
	archived := map[string]bool{}
	for key := range index.Devices {
		archived[key] = true
	}

	if err := watcher.detectNewDevices(t.Context(), map[string]bool{}, archived, true); err != nil {
		t.Fatalf("detectNewDevices() error = %v", err)
	}
	index, err = store.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex() after detect error = %v", err)
	}
	if len(index.Snapshots) != 1 {
		t.Fatalf("snapshots = %d, want existing snapshot only", len(index.Snapshots))
	}
}
