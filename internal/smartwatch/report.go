package smartwatch

import (
	"encoding/json"
	"errors"
	"os"
)

// ReportLogMaxBytes caps the report log. Once the active file reaches this
// size it is renamed to "<path>.1" (replacing any previous rotation) and a
// fresh file is started, so the active path is never rewritten in place: a
// tailer such as Alloy follows the renamed inode instead of re-reading bytes
// it already shipped. The active file therefore stays under the cap plus at
// most one scan, and at most two report files exist on disk (plus the
// "<path>.lock" sidecar used to serialize rotation).
const ReportLogMaxBytes = 8 << 20 // 8 MiB

// AppendReport appends the scan's snapshots to path as one compact JSON object
// per line. The whole batch is encoded in memory first, so a marshal failure
// leaves nothing behind, and it is written in one O_APPEND Write call.
// Concurrent AppendReport calls cannot interleave because each holds the
// cross-process lock below for the whole rotation-and-append.
//
// What that does not guarantee: os.File.Write may issue more than one write
// syscall after a short write, and a hard failure part-way through a write can
// leave a truncated final line. A tailer must therefore ignore any line that
// does not parse as JSON. An append that fails outright (before any byte is
// written) leaves the file untouched.
//
// The size check, rotation, open and append are serialized by a blocking
// cross-process lock on path+".lock" (a stable sidecar, because rotation
// replaces the active file's inode), so overlapping scans cannot both rotate
// the same active file away. If that lock cannot be taken at all, the report
// is still written unlocked rather than dropped: waiting for a held lock is
// preferred, but a lock error must not lose a line.
func AppendReport(path string, snapshots []Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	batch, err := encodeReportBatch(snapshots)
	if err != nil {
		return err
	}

	release, lockErr := lockReportFile(path)
	if lockErr == nil {
		defer release()
	}
	// lockErr != nil: proceed unlocked so the report is not silently lost.

	if err := rotateReportIfFull(path); err != nil {
		return err
	}
	return appendReportBatch(path, batch)
}

// encodeReportBatch marshals every snapshot before anything touches the file,
// so a failure leaves no partial batch behind.
func encodeReportBatch(snapshots []Snapshot) ([]byte, error) {
	var batch []byte
	for _, snapshot := range snapshots {
		line, err := json.Marshal(snapshot)
		if err != nil {
			return nil, err
		}
		batch = append(batch, line...)
		batch = append(batch, '\n')
	}
	return batch, nil
}

// appendReportBatch makes the whole batch one O_APPEND write. The mode passed
// to OpenFile is masked by umask and rotation replaces the file, so 0644 is
// set explicitly afterwards: a non-root tailer such as Alloy must not lose
// read access because the daemon ran under umask 0077. A chmod failure is
// returned, not swallowed.
func appendReportBatch(path string, batch []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Chmod(0o644); err != nil {
		return err
	}
	_, err = f.Write(batch)
	return err
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
