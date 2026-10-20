// Package promtext renders S.M.A.R.T. drive snapshots as Prometheus text
// exposition, for consumption by the node_exporter / Alloy textfile collector.
package promtext

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"nuther/internal/smart"
)

// FileName is the file written inside the configured --textfile directory.
const FileName = "nuther-smart.prom"

type family struct {
	name  string
	help  string
	typ   string
	value func(smart.DriveInfo) (int64, bool)
	// extra appends additional labels to the base label set (e.g. status).
	extra func(smart.DriveInfo) string
	// scanErrorOnly marks families still emitted for a drive whose scan failed.
	scanErrorOnly bool
}

// Render returns the Prometheus text exposition for the given drives.
//
// It is pure and deterministic: no I/O, no global state, and no clock read
// (now is supplied by the caller). Drives are sorted by device name. Every
// series family is emitted with its # HELP and # TYPE lines; the
// wear_leveling_value and written_bytes_total samples are omitted when the
// model reports them as unavailable (-1).
func Render(drives []smart.DriveInfo, now time.Time) string {
	sorted := make([]smart.DriveInfo, len(drives))
	copy(sorted, drives)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Device < sorted[j].Device })

	families := []family{
		{
			name:  "nuther_smart_health_status",
			help:  "Drive health status; 1 for the current status label, other status labels are absent rather than 0.",
			typ:   "gauge",
			value: func(smart.DriveInfo) (int64, bool) { return 1, true },
			extra: func(d smart.DriveInfo) string {
				return fmt.Sprintf(`,status="%s"`, statusLabel(d.HealthStatus))
			},
		},
		{
			name: "nuther_smart_health_percent",
			help: "Remaining drive life as a percentage (0-100); omitted when the drive health is unknown.",
			typ:  "gauge",
			value: func(d smart.DriveInfo) (int64, bool) {
				if d.ScanError != nil || d.HealthStatus == smart.HealthUnknown || d.HealthStatus == "" {
					return 0, false
				}
				return int64(d.HealthPercent()), true
			},
		},
		{
			name:  "nuther_smart_temperature_celsius",
			help:  "Current drive temperature in degrees Celsius.",
			typ:   "gauge",
			value: func(d smart.DriveInfo) (int64, bool) { return int64(d.Temperature), true },
		},
		{
			name:  "nuther_smart_reallocated_sectors",
			help:  "Count of reallocated sectors.",
			typ:   "gauge",
			value: func(d smart.DriveInfo) (int64, bool) { return d.ReallocatedSectors, true },
		},
		{
			name:  "nuther_smart_pending_sectors",
			help:  "Count of sectors pending reallocation.",
			typ:   "gauge",
			value: func(d smart.DriveInfo) (int64, bool) { return d.PendingSectors, true },
		},
		{
			name:  "nuther_smart_uncorrectable_sectors",
			help:  "Count of offline uncorrectable sectors.",
			typ:   "gauge",
			value: func(d smart.DriveInfo) (int64, bool) { return d.UncorrectableSectors, true },
		},
		{
			name:  "nuther_smart_crc_errors",
			help:  "Count of UDMA CRC errors (cable/interface problems).",
			typ:   "gauge",
			value: func(d smart.DriveInfo) (int64, bool) { return d.CRCErrors, true },
		},
		// Power-on hours and cycles are lifetime totals, not monotonic event
		// counters: a replaced drive resets them, so they stay gauges.
		{
			name:  "nuther_smart_power_on_hours",
			help:  "Total power-on hours.",
			typ:   "gauge",
			value: func(d smart.DriveInfo) (int64, bool) { return d.PowerOnHours, true },
		},
		{
			name:  "nuther_smart_power_cycles",
			help:  "Total power cycles.",
			typ:   "gauge",
			value: func(d smart.DriveInfo) (int64, bool) { return d.PowerCycles, true },
		},
		{
			name: "nuther_smart_wear_leveling_value",
			help: "Normalized wear leveling value (100 new, 0 worn out); omitted when unavailable.",
			typ:  "gauge",
			value: func(d smart.DriveInfo) (int64, bool) {
				if d.WearLevelingValue < 0 {
					return 0, false
				}
				return int64(d.WearLevelingValue), true
			},
		},
		{
			name: "nuther_smart_written_bytes_total",
			help: "Total bytes written; omitted when unavailable.",
			typ:  "counter",
			value: func(d smart.DriveInfo) (int64, bool) {
				if d.TotalBytesWritten < 0 {
					return 0, false
				}
				return d.TotalBytesWritten, true
			},
		},
		{
			name:  "nuther_smart_snapshot_timestamp_seconds",
			help:  "Unix timestamp of the snapshot pass that produced these metrics.",
			typ:   "gauge",
			value: func(smart.DriveInfo) (int64, bool) { return now.Unix(), true },
		},
		{
			name: "nuther_smart_scan_success",
			help: "1 when the drive scan succeeded, 0 when this drive's scan errored.",
			typ:  "gauge",
			value: func(d smart.DriveInfo) (int64, bool) {
				if d.ScanError != nil {
					return 0, true
				}
				return 1, true
			},
			scanErrorOnly: true,
		},
	}

	var b strings.Builder
	for _, f := range families {
		fmt.Fprintf(&b, "# HELP %s %s\n", f.name, f.help)
		fmt.Fprintf(&b, "# TYPE %s %s\n", f.name, f.typ)
		for _, d := range sorted {
			// A drive that could not be read gets only scan_success 0; every
			// other series would be fabricated, so it is absent.
			if d.ScanError != nil && !f.scanErrorOnly {
				continue
			}
			value, ok := f.value(d)
			if !ok {
				continue
			}
			labels := baseLabels(d)
			if f.extra != nil {
				labels += f.extra(d)
			}
			fmt.Fprintf(&b, "%s{%s} %d\n", f.name, labels, value)
		}
	}
	return b.String()
}

// baseLabels returns the shared label set for a drive, with label values escaped.
func baseLabels(d smart.DriveInfo) string {
	return fmt.Sprintf(`device="%s",serial="%s",model="%s",type="%s",interface="%s"`,
		escapeLabel(d.Device),
		escapeLabel(d.Serial),
		escapeLabel(d.Model),
		escapeLabel(d.GetDriveType()),
		escapeLabel(d.Interface),
	)
}

// statusLabel maps a drive health status to one of the contract's status
// labels, defaulting to UNKNOWN for an unset or unexpected value.
func statusLabel(status smart.HealthStatus) string {
	switch status {
	case smart.HealthGood, smart.HealthInfo, smart.HealthCaution, smart.HealthBad, smart.HealthUnknown:
		return string(status)
	default:
		return string(smart.HealthUnknown)
	}
}

// escapeLabel escapes a Prometheus label value (backslash, double quote, newline).
func escapeLabel(value string) string {
	if !strings.ContainsAny(value, "\\\"\n") {
		return value
	}
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return value
}
