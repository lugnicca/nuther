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
	reportLog        string
}

type WatcherConfig struct {
	Store            *Store
	Hub              *EventHub
	Scanner          Scanner
	PollInterval     time.Duration
	SnapshotInterval time.Duration
	Now              func() time.Time
	TextfileDir      string
	ReportLog        string
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
		reportLog:        cfg.ReportLog,
	}
}

func (w *Watcher) SnapshotOnce(ctx context.Context, reason string) ([]SnapshotRecord, error) {
	records, _, err := w.snapshotOnce(ctx, reason)
	return records, err
}

// snapshotOnce is SnapshotOnce plus whether a textfile was actually published,
// so Run can base its final verdict on publication rather than on the
// aggregate error (which also carries archive failures).
func (w *Watcher) snapshotOnce(ctx context.Context, reason string) ([]SnapshotRecord, bool, error) {
	drives, err := w.scanRealDrives()
	if err != nil {
		return nil, false, err
	}

	var records []SnapshotRecord
	var snapshots []Snapshot
	var errs []error
	for _, drive := range drives {
		if drive.ScanError != nil {
			// Unreadable drives have no SMART data to archive; the textfile
			// reports them with scan_success 0 instead.
			continue
		}
		record, snapshot, err := w.archiveAndPublish(ctx, reason, drive)
		if err != nil {
			errs = append(errs, err)
		}
		// Archiving and reporting are independent: a failed report append
		// still leaves the snapshot archived, and a failed save still gets a
		// report line. Only a successful save yields a record for the caller.
		// The store built the Snapshot even when persistence failed, so the
		// report line carries the same value the archive would hold.
		snapshots = append(snapshots, snapshot)
		if record.ID != "" {
			records = append(records, record)
		}
	}

	// The whole scan is published in one append after every drive was
	// archived, so a batch cannot interleave with another scan and a marshal
	// failure on any drive leaves the report file untouched.
	if w.reportLog != "" {
		if err := appendReport(w.reportLog, snapshots); err != nil {
			errs = append(errs, err)
		}
	}

	// Success requires at least one readable drive whether or not a textfile
	// is configured.
	if !hasSuccessfulDrive(drives) {
		errs = append(errs, errors.New("no readable drive detected"))
	}

	// Archiving and publication are independent in both directions: a save
	// failure must not suppress the export, and an export failure must not
	// suppress the saves above. Both errors are returned, but publication
	// status is reported on its own.
	published, err := w.publish(drives)
	if err != nil {
		errs = append(errs, err)
	}
	return records, published, errors.Join(errs...)
}

func (w *Watcher) Run(ctx context.Context) error {
	index, indexErr := w.store.LoadIndex()
	archived := map[string]bool{}
	for key := range index.Devices {
		archived[key] = true
	}

	seen := map[string]bool{}
	published := false
	if drives, err := w.scanRealDrives(); err != nil {
		slog.Debug("smart watcher startup scan failed", "error", err)
		// A corrupt index still surfaces as the startup failure it was before,
		// but only after the scan was attempted.
		if indexErr != nil {
			return indexErr
		}
	} else {
		// Startup archiving and publication are attempted independently, so an
		// archive failure does not suppress the textfile. A corrupt index is
		// retained as an error but does not gate the scan or the export.
		errs := []error{indexErr}
		// A corrupt index cannot be appended to: SaveSnapshot reloads the same
		// index and fails with the same parse error, which would report it
		// twice. Skip the futile archive and report the corrupt index once.
		if indexErr == nil {
			if err := w.snapshotNewDevices(ctx, seen, archived, true, drives); err != nil {
				errs = append(errs, err)
			}
		}
		ok, err := w.publish(drives)
		if err != nil {
			errs = append(errs, err)
		}
		if ok {
			published = true
		}
		if err := errors.Join(errs...); err != nil {
			return err
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
				slog.Debug("smart watcher scan or archive failed", "error", err)
			}
			if len(drives) == 0 {
				continue
			}
			// Publication is attempted even when the archive above failed.
			ok, err := w.publish(drives)
			if err != nil {
				slog.Debug("smart textfile export failed", "error", err)
			}
			if ok {
				published = true
			}
		case <-snapshotC:
			_, ok, err := w.snapshotOnce(ctx, ReasonInterval)
			if err != nil {
				slog.Debug("smart interval snapshot failed", "error", err)
			}
			if ok {
				published = true
			}
		}
	}
}

func (w *Watcher) detectNewDevices(ctx context.Context, seen map[string]bool, archived map[string]bool, startup bool) ([]smart.DriveInfo, error) {
	drives, err := w.scanRealDrives()
	if err != nil {
		return nil, err
	}
	if err := w.snapshotNewDevices(ctx, seen, archived, startup, drives); err != nil {
		// Return the scanned drives so the caller can still publish the
		// textfile: an archive failure must not suppress the export.
		return drives, err
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
		_, snapshot, saveErr := w.archiveAndPublish(ctx, reason, drive)
		// The daemon reports each newly archived drive as it goes: this path is
		// not batched like --once, and it stops at the first error, so a
		// discovery poll is not one line per measured drive.
		reportErr := w.publishReport(snapshot)
		if saveErr != nil || reportErr != nil {
			return errors.Join(saveErr, reportErr)
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
// freshness marker. It reports whether a textfile was actually written,
// separately from the export error, so callers can distinguish "nothing to
// publish" from "published" without inspecting the error.
func (w *Watcher) publish(drives []smart.DriveInfo) (bool, error) {
	if w.textfileDir == "" {
		return false, nil
	}
	if !hasSuccessfulDrive(drives) {
		err := fmt.Errorf("no successfully scanned drive; textfile left untouched")
		slog.Error("smart textfile export skipped", "error", err)
		return false, err
	}
	if err := promtext.WriteFile(w.textfileDir, drives, w.now()); err != nil {
		slog.Error("smart textfile export failed", "error", err)
		return false, err
	}
	return true, nil
}

func hasSuccessfulDrive(drives []smart.DriveInfo) bool {
	for _, drive := range drives {
		if drive.ScanError == nil {
			return true
		}
	}
	return false
}

// archiveAndPublish archives the drive and publishes its event. It returns the
// store's Snapshot so a scan can publish the whole batch to the report log in
// one append, even when persistence failed.
func (w *Watcher) archiveAndPublish(ctx context.Context, reason string, drive smart.DriveInfo) (SnapshotRecord, Snapshot, error) {
	record, snapshot, saveErr := w.store.saveSnapshot(w.now(), reason, drive)
	if saveErr == nil {
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
	}
	return record, snapshot, saveErr
}

// appendReport is the report-log publication used by the watcher; it is a
// variable so tests can count how many publications a scan makes. Production
// always uses AppendReport.
var appendReport = AppendReport

// publishReport appends one snapshot to the report log when one is configured.
func (w *Watcher) publishReport(snapshot Snapshot) error {
	if w.reportLog == "" {
		return nil
	}
	return appendReport(w.reportLog, []Snapshot{snapshot})
}
