package smartwatch

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"nuther/internal/promtext"
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
	if _, err := watcher.detectNewDevices(t.Context(), seen, archived, true); err != nil {
		t.Fatalf("startup detectNewDevices() error = %v", err)
	}
	if _, err := watcher.detectNewDevices(t.Context(), seen, archived, false); err != nil {
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

	if _, err := watcher.detectNewDevices(t.Context(), map[string]bool{}, archived, true); err != nil {
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

func failedDrive(device string) smart.DriveInfo {
	return smart.DriveInfo{Device: device, ScanError: errors.New("smartctl failed")}
}

func readTextfile(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, promtext.FileName))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	return string(data)
}

func TestWatcherSnapshotOncePublishesTextfile(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(t.TempDir())
	watcher := NewWatcher(WatcherConfig{
		Store: store,
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now:         func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		TextfileDir: dir,
	})

	records, err := watcher.SnapshotOnce(t.Context(), ReasonManual)
	if err != nil {
		t.Fatalf("SnapshotOnce() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1", len(records))
	}
	out := readTextfile(t, dir)
	for _, want := range []string{
		`nuther_smart_scan_success{device="/dev/sda"`,
		`nuther_smart_snapshot_timestamp_seconds{device="/dev/sda"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("textfile missing %q:\n%s", want, out)
		}
	}
}

func TestWatcherSnapshotOnceExportFailureStillSavesSnapshot(t *testing.T) {
	store := NewStore(t.TempDir())
	watcher := NewWatcher(WatcherConfig{
		Store: store,
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now:         func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		TextfileDir: filepath.Join(t.TempDir(), "does-not-exist"),
	})

	records, err := watcher.SnapshotOnce(t.Context(), ReasonManual)
	if err == nil {
		t.Fatal("SnapshotOnce() error = nil, want export failure")
	}
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1", len(records))
	}
	index, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex() error = %v", err)
	}
	if len(index.Snapshots) != 1 {
		t.Fatalf("snapshots = %d, want 1 (export failure must not suppress the save)", len(index.Snapshots))
	}
}

func TestWatcherSnapshotOnceMixedRecords(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(t.TempDir())
	watcher := NewWatcher(WatcherConfig{
		Store: store,
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A"), failedDrive("/dev/sdb")}, nil
		},
		Now:         func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		TextfileDir: dir,
	})

	records, err := watcher.SnapshotOnce(t.Context(), ReasonManual)
	if err != nil {
		t.Fatalf("SnapshotOnce() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1 (only readable drives archived)", len(records))
	}
	out := readTextfile(t, dir)
	if !strings.Contains(out, `nuther_smart_scan_success{device="/dev/sda"`) {
		t.Fatalf("textfile missing successful scan_success:\n%s", out)
	}
	if !strings.Contains(out, `nuther_smart_scan_success{device="/dev/sdb"`) || !strings.Contains(out, `device="/dev/sdb"`) {
		t.Fatalf("textfile missing failed scan_success:\n%s", out)
	}
	if got := strings.Count(out, `device="/dev/sdb"`); got != 1 {
		t.Fatalf("failed drive emitted %d series, want only scan_success:\n%s", got, out)
	}
}

func TestWatcherSnapshotOnceAllFailedLeavesFileUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, promtext.FileName)
	if err := os.WriteFile(path, []byte("previous\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	store := NewStore(t.TempDir())
	watcher := NewWatcher(WatcherConfig{
		Store: store,
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{failedDrive("/dev/sdb")}, nil
		},
		Now:         func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		TextfileDir: dir,
	})

	records, err := watcher.SnapshotOnce(t.Context(), ReasonManual)
	if err == nil {
		t.Fatal("SnapshotOnce() error = nil, want skipped-publication error")
	}
	if len(records) != 0 {
		t.Fatalf("records = %d, want 0", len(records))
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() after error = %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("file mtime changed: %v -> %v", before.ModTime(), after.ModTime())
	}
	if data, _ := os.ReadFile(path); string(data) != "previous\n" {
		t.Fatalf("file content = %q, want unchanged", data)
	}
}

func TestWatcherNoDrivesLeavesFileUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, promtext.FileName)
	if err := os.WriteFile(path, []byte("previous\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	watcher := NewWatcher(WatcherConfig{
		Store: NewStore(t.TempDir()),
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return nil, errors.New("no non-demo drives detected")
		},
		TextfileDir: dir,
	})
	if _, err := watcher.SnapshotOnce(t.Context(), ReasonManual); err == nil {
		t.Fatal("SnapshotOnce() error = nil, want scan failure")
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() after error = %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("file mtime changed: %v -> %v", before.ModTime(), after.ModTime())
	}
}

// An entirely unreadable fleet must fail SnapshotOnce even when no textfile is
// configured: success requires at least one readable drive.
func TestWatcherSnapshotOnceAllUnreadableWithoutTextfileFails(t *testing.T) {
	watcher := NewWatcher(WatcherConfig{
		Store: NewStore(t.TempDir()),
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{failedDrive("/dev/sdb")}, nil
		},
	})

	records, err := watcher.SnapshotOnce(t.Context(), ReasonManual)
	if err == nil {
		t.Fatal("SnapshotOnce() error = nil, want failure when no drive is readable")
	}
	if len(records) != 0 {
		t.Fatalf("records = %d, want 0", len(records))
	}
}

// Archiving and publication are independent: a snapshot-store failure must not
// stop the textfile from being published, and the error must still be returned.
func TestWatcherSnapshotOnceArchiveFailureStillPublishes(t *testing.T) {
	dir := t.TempDir()
	storeFile := filepath.Join(t.TempDir(), "store-is-a-file")
	if err := os.WriteFile(storeFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	watcher := NewWatcher(WatcherConfig{
		Store: NewStore(storeFile),
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now:         func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		TextfileDir: dir,
	})

	_, err := watcher.SnapshotOnce(t.Context(), ReasonManual)
	if err == nil {
		t.Fatal("SnapshotOnce() error = nil, want archive failure")
	}
	out := readTextfile(t, dir)
	if !strings.Contains(out, `nuther_smart_scan_success{device="/dev/sda"`) {
		t.Fatalf("textfile not published despite archive failure:\n%s", out)
	}
}

func TestWatcherTextfileDisabled(t *testing.T) {
	dir := t.TempDir()
	// With no TextfileDir the watcher must not write anything, not even into
	// the process working directory where an empty dir would resolve.
	t.Chdir(dir)
	watcher := NewWatcher(WatcherConfig{
		Store: NewStore(t.TempDir()),
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
	})
	if _, err := watcher.SnapshotOnce(t.Context(), ReasonManual); err != nil {
		t.Fatalf("SnapshotOnce() error = %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("textfile written with output disabled: %v", entries)
	}
}

func TestWatcherRunRefreshesTextfileOnPoll(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(t.TempDir())
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var mu sync.Mutex
	scans := 0
	watcher := NewWatcher(WatcherConfig{
		Store:        store,
		Hub:          NewEventHub(""),
		PollInterval: 20 * time.Millisecond,
		Scanner: func() ([]smart.DriveInfo, error) {
			mu.Lock()
			defer mu.Unlock()
			scans++
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now: func() time.Time {
			mu.Lock()
			defer mu.Unlock()
			return time.Unix(1700000000+int64(scans), 0).UTC()
		},
		TextfileDir: dir,
	})

	done := make(chan error, 1)
	go func() { done <- watcher.Run(ctx) }()

	// The startup publish stamps 1700000001; a poll must replace the file
	// with a later stamp.
	deadline := time.Now().Add(3 * time.Second)
	refreshed := false
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(filepath.Join(dir, promtext.FileName))
		if err == nil && strings.Contains(string(data), " 1700000002\n") {
			refreshed = true
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !refreshed {
		t.Fatal("textfile was not refreshed on a poll tick")
	}

	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
}

func TestWatcherRunTextfileNeverPublishedFails(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	scanCalls := 0
	watcher := NewWatcher(WatcherConfig{
		Store:        NewStore(t.TempDir()),
		Hub:          NewEventHub(""),
		PollInterval: time.Millisecond,
		Scanner: func() ([]smart.DriveInfo, error) {
			scanCalls++
			if scanCalls >= 2 {
				cancel()
			}
			return nil, errors.New("no non-demo drives detected")
		},
		TextfileDir: t.TempDir(),
	})

	err := watcher.Run(ctx)
	if err == nil || errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want non-zero publication failure", err)
	}
}

func TestWatcherRunWithoutTextfileCancelsCleanly(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	scanCalls := 0
	watcher := NewWatcher(WatcherConfig{
		Store:        NewStore(t.TempDir()),
		Hub:          NewEventHub(""),
		PollInterval: time.Millisecond,
		Scanner: func() ([]smart.DriveInfo, error) {
			scanCalls++
			if scanCalls >= 2 {
				cancel()
			}
			return nil, errors.New("no non-demo drives detected")
		},
	})

	err := watcher.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
}

// blockedSnapshotStore returns a store whose "snapshots" path is a regular
// file, so every SaveSnapshot fails while LoadIndex keeps working.
func blockedSnapshotStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "snapshots"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return NewStore(dir)
}

func waitForTextfile(t *testing.T, dir string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(filepath.Join(dir, promtext.FileName))
		if err == nil && strings.Contains(string(data), `nuther_smart_scan_success{device="/dev/sda"`) {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("textfile was never published")
}

// An interval snapshot can publish the textfile and still return an archive
// error through errors.Join. The final "never published" verdict must follow
// publication, not the aggregate error.
func TestWatcherRunPublishedDespiteArchiveError(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	scans := 0
	watcher := NewWatcher(WatcherConfig{
		Store:            blockedSnapshotStore(t),
		Hub:              NewEventHub(""),
		PollInterval:     time.Hour,
		SnapshotInterval: 5 * time.Millisecond,
		Scanner: func() ([]smart.DriveInfo, error) {
			scans++
			if scans == 1 {
				return nil, errors.New("startup discovery failed")
			}
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now:         func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		TextfileDir: dir,
	})

	done := make(chan error, 1)
	go func() { done <- watcher.Run(ctx) }()

	waitForTextfile(t, dir)
	cancel()

	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled (the textfile was published)", err)
	}
}

// The startup archive failure must not suppress the startup export.
func TestWatcherRunStartupArchiveFailureStillPublishes(t *testing.T) {
	dir := t.TempDir()
	watcher := NewWatcher(WatcherConfig{
		Store: blockedSnapshotStore(t),
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now:         func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		TextfileDir: dir,
	})

	if err := watcher.Run(t.Context()); err == nil {
		t.Fatal("Run() error = nil, want the archive failure")
	}
	readTextfile(t, dir)
}

// The polling archive failure must not suppress the periodic export.
func TestWatcherRunPollArchiveFailureStillPublishes(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	scans := 0
	watcher := NewWatcher(WatcherConfig{
		Store:        blockedSnapshotStore(t),
		Hub:          NewEventHub(""),
		PollInterval: 5 * time.Millisecond,
		Scanner: func() ([]smart.DriveInfo, error) {
			scans++
			if scans == 1 {
				return nil, errors.New("startup discovery failed")
			}
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now:         func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		TextfileDir: dir,
	})

	done := make(chan error, 1)
	go func() { done <- watcher.Run(ctx) }()

	waitForTextfile(t, dir)
	cancel()

	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled (the textfile was published)", err)
	}
}

// An aggregate of two independent failures must expose both causes, not just
// the first one.
func TestWatcherSnapshotOnceJoinsArchiveAndExportErrors(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	watcher := NewWatcher(WatcherConfig{
		Store: blockedSnapshotStore(t),
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{testDrive("/dev/sda", "SN-A")}, nil
		},
		Now:         func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		TextfileDir: dir,
	})

	_, err := watcher.SnapshotOnce(t.Context(), ReasonManual)
	if err == nil {
		t.Fatal("SnapshotOnce() error = nil, want both failures")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("SnapshotOnce() error = %v, want the export cause visible", err)
	}
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) || len(joined.Unwrap()) < 2 {
		t.Fatalf("SnapshotOnce() error = %v, want both causes joined", err)
	}
}
