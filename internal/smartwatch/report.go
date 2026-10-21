package smartwatch

import (
	"encoding/json"
	"errors"
	"os"

	"nuther/internal/smart"
)

// ReportLogMaxBytes caps the report log. Once the active file reaches this
// size it is renamed to "<path>.1" (replacing any previous rotation) and a
// fresh file is started, so the active path is never rewritten in place: a
// tailer such as Alloy follows the renamed inode instead of re-reading bytes
// it already shipped. The active file therefore stays under the cap plus at
// most one scan, and at most two files exist on disk.
const ReportLogMaxBytes = 8 << 20 // 8 MiB

// ReportSnapshot reconstructs the exact Snapshot the store persisted for a
// record, so a report line is the store's own object and not a second shape.
func ReportSnapshot(record SnapshotRecord, drive smart.DriveInfo) Snapshot {
	return Snapshot{
		ID:        record.ID,
		Timestamp: record.Timestamp,
		Reason:    record.Reason,
		Device:    record.Device,
		SMART:     drive,
	}
}

// AppendReport appends one compact JSON line per snapshot to path, each as a
// single O_APPEND write so a reader never observes a partial line. Nothing is
// written when snapshots is empty, and the file is never rewritten in place.
func AppendReport(path string, snapshots []Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	if err := rotateReportIfFull(path); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, snapshot := range snapshots {
		line, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		line = append(line, '\n')
		if _, err := f.Write(line); err != nil {
			return err
		}
	}
	return nil
}

// rotateReportIfFull renames the active file aside once it reaches the cap.
// The renamed file is left for the tailer to finish; it is not deleted while
// it may still be open.
func rotateReportIfFull(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Size() < ReportLogMaxBytes {
		return nil
	}
	return os.Rename(path, path+".1")
}
