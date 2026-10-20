package smartwatch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"nuther/internal/promtext"
	"nuther/internal/smart"
)

type Watcher struct {
	store            *Store
	hub              *EventHub
	scanner          Scanner
	pollInterval     time.Duration
	snapshotInterval time.Duration
	now              func() time.Time
	textfileDir      string
}

type WatcherConfig struct {
	Store            *Store
	Hub              *EventHub
	Scanner          Scanner
	PollInterval     time.Duration
	SnapshotInterval time.Duration
	Now              func() time.Time
	TextfileDir      string
}

func NewWatcher(cfg WatcherConfig) *Watcher {
	scanner := cfg.Scanner
	if scanner == nil {
		scanner = smart.ScanDrivesIncludingErrors
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &Watcher{
		store:            cfg.Store,
		hub:              cfg.Hub,
		scanner:          scanner,
		pollInterval:     cfg.PollInterval,
		snapshotInterval: cfg.SnapshotInterval,
		now:              now,
		textfileDir:      cfg.TextfileDir,
	}
}

func (w *Watcher) SnapshotOnce(ctx context.Context, reason string) ([]SnapshotRecord, error) {
	drives, err := w.scanRealDrives()
	if err != nil {
		return nil, err
	}

	var records []SnapshotRecord
	for _, drive := range drives {
		if drive.ScanError != nil {
			// Unreadable drives have no SMART data to archive; the textfile
			// reports them with scan_success 0 instead.
			continue
		}
		record, err := w.saveAndPublish(ctx, reason, drive)
		if err != nil {
			return records, err
		}
		records = append(records, record)
	}
	// Publication is separate from archiving: an export failure still returns
	// an error (so --once exits non-zero) without losing the snapshots above.
	return records, w.publish(drives)
}

func (w *Watcher) Run(ctx context.Context) error {
	index, err := w.store.LoadIndex()
	if err != nil {
		return err
	}
	archived := map[string]bool{}
	for key := range index.Devices {
		archived[key] = true
	}

	seen := map[string]bool{}
	published := false
	if drives, err := w.scanRealDrives(); err != nil {
		slog.Debug("smart watcher startup scan failed", "error", err)
	} else {
		if err := w.snapshotNewDevices(ctx, seen, archived, true, drives); err != nil {
			return err
		}
		if err := w.publish(drives); err == nil {
			published = true
		}
	}

	pollTicker := time.NewTicker(w.pollInterval)
	defer pollTicker.Stop()

	var snapshotTicker *time.Ticker
	var snapshotC <-chan time.Time
	if w.snapshotInterval > 0 {
		snapshotTicker = time.NewTicker(w.snapshotInterval)
		defer snapshotTicker.Stop()
		snapshotC = snapshotTicker.C
	}

	for {
		select {
		case <-ctx.Done():
			// A textfile that was never published is a failure the fleet's
			// freshness alerting must not read as success.
			if w.textfileDir != "" && !published {
				return errors.New("smart textfile was never published")
			}
			return ctx.Err()
		case <-pollTicker.C:
			drives, err := w.detectNewDevices(ctx, seen, archived, false)
			if err != nil {
				slog.Debug("smart watcher scan failed", "error", err)
				continue
			}
			if err := w.publish(drives); err == nil {
				published = true
			}
		case <-snapshotC:
			if _, err := w.SnapshotOnce(ctx, ReasonInterval); err != nil {
				slog.Debug("smart interval snapshot failed", "error", err)
				continue
			}
			published = true
		}
	}
}

func (w *Watcher) detectNewDevices(ctx context.Context, seen map[string]bool, archived map[string]bool, startup bool) ([]smart.DriveInfo, error) {
	drives, err := w.scanRealDrives()
	if err != nil {
		return nil, err
	}
	if err := w.snapshotNewDevices(ctx, seen, archived, startup, drives); err != nil {
		return nil, err
	}
	return drives, nil
}

func (w *Watcher) snapshotNewDevices(ctx context.Context, seen map[string]bool, archived map[string]bool, startup bool, drives []smart.DriveInfo) error {
	current := map[string]bool{}
	for _, drive := range drives {
		key := StableDeviceKey(drive)
		current[key] = true
		if seen[key] {
			continue
		}
		if startup && archived[key] {
			continue
		}
		if drive.ScanError != nil {
			continue
		}
		reason := ReasonNewDevice
		if startup && !archived[key] {
			reason = ReasonStartup
		}
		if _, err := w.saveAndPublish(ctx, reason, drive); err != nil {
			return err
		}
		archived[key] = true
	}

	for key := range seen {
		delete(seen, key)
	}
	for key := range current {
		seen[key] = true
	}
	return nil
}

func (w *Watcher) scanRealDrives() ([]smart.DriveInfo, error) {
	drives, err := w.scanner()
	realDrives := drives[:0]
	for _, drive := range drives {
		if !drive.IsDemo {
			realDrives = append(realDrives, drive)
		}
	}
	if len(realDrives) == 0 && err != nil {
		return nil, err
	}
	if len(realDrives) == 0 {
		return nil, fmt.Errorf("no non-demo drives detected")
	}
	return realDrives, nil
}

// publish atomically replaces the textfile. A run with no successfully scanned
// drive leaves the previous file untouched, so its mtime stays a truthful
// freshness marker. It logs and returns the export error so callers can report
// it without suppressing snapshot saving.
func (w *Watcher) publish(drives []smart.DriveInfo) error {
	if w.textfileDir == "" {
		return nil
	}
	if !hasSuccessfulDrive(drives) {
		err := fmt.Errorf("no successfully scanned drive; textfile left untouched")
		slog.Error("smart textfile export skipped", "error", err)
		return err
	}
	if err := promtext.WriteFile(w.textfileDir, drives, w.now()); err != nil {
		slog.Error("smart textfile export failed", "error", err)
		return err
	}
	return nil
}

func hasSuccessfulDrive(drives []smart.DriveInfo) bool {
	for _, drive := range drives {
		if drive.ScanError == nil {
			return true
		}
	}
	return false
}

func (w *Watcher) saveAndPublish(ctx context.Context, reason string, drive smart.DriveInfo) (SnapshotRecord, error) {
	record, err := w.store.SaveSnapshot(w.now(), reason, drive)
	if err != nil {
		return SnapshotRecord{}, err
	}
	event := Event{
		Type:       "smart_snapshot",
		Timestamp:  record.Timestamp,
		DeviceKey:  record.Device.Key,
		SnapshotID: record.ID,
		Path:       record.Path,
		Reason:     record.Reason,
		Summary:    record.Device,
	}
	w.hub.Publish(ctx, event)
	return record, nil
}
