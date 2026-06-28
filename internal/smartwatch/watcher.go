package smartwatch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nuther/internal/smart"
)

type Watcher struct {
	store            *Store
	hub              *EventHub
	scanner          Scanner
	pollInterval     time.Duration
	snapshotInterval time.Duration
	now              func() time.Time
}

type WatcherConfig struct {
	Store            *Store
	Hub              *EventHub
	Scanner          Scanner
	PollInterval     time.Duration
	SnapshotInterval time.Duration
	Now              func() time.Time
}

func NewWatcher(cfg WatcherConfig) *Watcher {
	scanner := cfg.Scanner
	if scanner == nil {
		scanner = smart.ScanDrives
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
	}
}

func (w *Watcher) SnapshotOnce(ctx context.Context, reason string) ([]SnapshotRecord, error) {
	drives, err := w.scanRealDrives()
	if err != nil {
		return nil, err
	}

	var records []SnapshotRecord
	for _, drive := range drives {
		record, err := w.saveAndPublish(ctx, reason, drive)
		if err != nil {
			return records, err
		}
		records = append(records, record)
	}
	return records, nil
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
	if drives, err := w.scanRealDrives(); err != nil {
		slog.Debug("smart watcher startup scan failed", "error", err)
	} else if err := w.snapshotNewDevices(ctx, seen, archived, true, drives); err != nil {
		return err
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
			return ctx.Err()
		case <-pollTicker.C:
			if err := w.detectNewDevices(ctx, seen, archived, false); err != nil {
				slog.Debug("smart watcher scan failed", "error", err)
			}
		case <-snapshotC:
			if _, err := w.SnapshotOnce(ctx, ReasonInterval); err != nil {
				slog.Debug("smart interval snapshot failed", "error", err)
			}
		}
	}
}

func (w *Watcher) detectNewDevices(ctx context.Context, seen map[string]bool, archived map[string]bool, startup bool) error {
	drives, err := w.scanRealDrives()
	if err != nil {
		return err
	}
	return w.snapshotNewDevices(ctx, seen, archived, startup, drives)
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
