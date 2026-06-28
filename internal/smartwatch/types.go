package smartwatch

import (
	"time"

	"nuther/internal/smart"
)

const (
	ReasonNewDevice = "new_device"
	ReasonInterval  = "interval"
	ReasonManual    = "manual"
	ReasonStartup   = "startup"
)

type DeviceSummary struct {
	Key           string             `json:"key"`
	Device        string             `json:"device"`
	Serial        string             `json:"serial,omitempty"`
	Model         string             `json:"model,omitempty"`
	WWN           string             `json:"wwn,omitempty"`
	Capacity      string             `json:"capacity,omitempty"`
	CapacityBytes int64              `json:"capacity_bytes,omitempty"`
	Interface     string             `json:"interface,omitempty"`
	HealthStatus  smart.HealthStatus `json:"health_status"`
}

type Snapshot struct {
	ID        string          `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Reason    string          `json:"reason"`
	Device    DeviceSummary   `json:"device"`
	SMART     smart.DriveInfo `json:"smart"`
}

type SnapshotRecord struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Reason    string        `json:"reason"`
	Path      string        `json:"path"`
	Device    DeviceSummary `json:"device"`
}

type DeviceRecord struct {
	Key          string        `json:"key"`
	FirstSeen    time.Time     `json:"first_seen"`
	LastSeen     time.Time     `json:"last_seen"`
	SnapshotIDs  []string      `json:"snapshot_ids"`
	LastSnapshot string        `json:"last_snapshot,omitempty"`
	Summary      DeviceSummary `json:"summary"`
}

type Index struct {
	Version   int                     `json:"version"`
	UpdatedAt time.Time               `json:"updated_at"`
	Devices   map[string]DeviceRecord `json:"devices"`
	Snapshots []SnapshotRecord        `json:"snapshots"`
}

type Event struct {
	Type       string        `json:"type"`
	Timestamp  time.Time     `json:"timestamp"`
	DeviceKey  string        `json:"device_key"`
	SnapshotID string        `json:"snapshot_id"`
	Path       string        `json:"path"`
	Reason     string        `json:"reason"`
	Summary    DeviceSummary `json:"summary"`
}

type Scanner func() ([]smart.DriveInfo, error)
