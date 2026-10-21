package smartwatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"nuther/internal/smart"
)

type Store struct {
	dir string
	mu  sync.Mutex
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) Dir() string {
	return s.dir
}

func (s *Store) LoadIndex() (Index, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadIndexLocked()
}

func (s *Store) SaveSnapshot(now time.Time, reason string, drive smart.DriveInfo) (SnapshotRecord, error) {
	record, _, err := s.saveSnapshot(now, reason, drive)
	return record, err
}

// saveSnapshot builds the Snapshot once, then persists it, returning the
// object that was archived (or would have been) so a caller can publish the
// same value to the report log even when persistence fails. The store's
// on-disk shape is unchanged.
func (s *Store) saveSnapshot(now time.Time, reason string, drive smart.DriveInfo) (SnapshotRecord, Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	summary := SummaryForDrive(drive)
	snapshot := Snapshot{
		Timestamp: now.UTC(),
		Reason:    reason,
		Device:    summary,
		SMART:     drive,
	}

	// MkdirAll first: uniqueSnapshotIDLocked stats into this directory, and a
	// path that is not a directory (or is unreadable) must fail here rather
	// than spin the uniqueness loop.
	if err := os.MkdirAll(filepath.Join(s.dir, "snapshots"), 0755); err != nil {
		snapshot.ID = snapshotID(now, reason, summary.Key)
		return SnapshotRecord{}, snapshot, err
	}

	id := snapshotID(now, reason, summary.Key)
	id = s.uniqueSnapshotIDLocked(id)
	snapshot.ID = id

	index, err := s.loadIndexLocked()
	if err != nil {
		return SnapshotRecord{}, snapshot, err
	}

	relPath := filepath.Join("snapshots", id+".json")
	absPath := filepath.Join(s.dir, relPath)
	if err := writeJSONAtomic(absPath, snapshot); err != nil {
		return SnapshotRecord{}, snapshot, err
	}

	record := SnapshotRecord{
		ID:        id,
		Timestamp: snapshot.Timestamp,
		Reason:    reason,
		Path:      relPath,
		Device:    summary,
	}

	device := index.Devices[summary.Key]
	if device.Key == "" {
		device = DeviceRecord{
			Key:       summary.Key,
			FirstSeen: snapshot.Timestamp,
		}
	}
	device.LastSeen = snapshot.Timestamp
	device.LastSnapshot = id
	device.SnapshotIDs = append(device.SnapshotIDs, id)
	device.Summary = summary
	index.Devices[summary.Key] = device
	index.Snapshots = append(index.Snapshots, record)
	index.UpdatedAt = snapshot.Timestamp

	sort.Slice(index.Snapshots, func(i, j int) bool {
		return index.Snapshots[i].Timestamp.Before(index.Snapshots[j].Timestamp)
	})

	if err := writeJSONAtomic(filepath.Join(s.dir, "index.json"), index); err != nil {
		return SnapshotRecord{}, snapshot, err
	}
	return record, snapshot, nil
}

func (s *Store) ReadSnapshot(id string) (Snapshot, error) {
	if id == "" || strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return Snapshot{}, fmt.Errorf("invalid snapshot id")
	}
	var snapshot Snapshot
	path := filepath.Join(s.dir, "snapshots", id+".json")
	if err := readJSON(path, &snapshot); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func (s *Store) loadIndexLocked() (Index, error) {
	index := Index{
		Version: 1,
		Devices: map[string]DeviceRecord{},
	}
	path := filepath.Join(s.dir, "index.json")
	if err := readJSON(path, &index); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return index, nil
		}
		return Index{}, err
	}
	if index.Devices == nil {
		index.Devices = map[string]DeviceRecord{}
	}
	return index, nil
}

func snapshotID(t time.Time, reason, key string) string {
	return t.UTC().Format("20060102T150405.000000000Z") + "-" + safeToken(reason, 32) + "-" + safeToken(key, 72)
}

func (s *Store) uniqueSnapshotIDLocked(base string) string {
	id := base
	for i := 2; ; i++ {
		if _, err := os.Stat(filepath.Join(s.dir, "snapshots", id+".json")); errors.Is(err, os.ErrNotExist) {
			return id
		}
		id = fmt.Sprintf("%s-%02d", base, i)
	}
}

func readJSON(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
