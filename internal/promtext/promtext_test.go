package promtext

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nuther/internal/smart"
)

func TestRender(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()

	tests := []struct {
		name   string
		drives []smart.DriveInfo
		want   string
	}{
		{
			name:   "no drives",
			drives: nil,
			want:   headerOnly(),
		},
		{
			name:   "nvme ata and scan error",
			drives: []smart.DriveInfo{nvmeDrive(), ataDrive(), scanErrorDrive()},
			want:   fullOutput(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.drives, now)
			if got != tt.want {
				t.Fatalf("Render() mismatch\n--- got ---\n%s\n--- want ---\n%s", got, tt.want)
			}
		})
	}
}

func TestRenderOmitsUnavailableSeries(t *testing.T) {
	out := Render([]smart.DriveInfo{ataDrive()}, time.Unix(1700000000, 0).UTC())
	for _, omitted := range []string{
		"nuther_smart_wear_leveling_value{",
		"nuther_smart_written_bytes_total{",
	} {
		if strings.Contains(out, omitted) {
			t.Fatalf("output unexpectedly contains %q", omitted)
		}
	}
	// HELP/TYPE blocks are still present even when no sample is emitted.
	for _, header := range []string{
		"# HELP nuther_smart_wear_leveling_value",
		"# TYPE nuther_smart_written_bytes_total counter",
	} {
		if !strings.Contains(out, header) {
			t.Fatalf("output missing %q", header)
		}
	}
}

func TestWriteAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	if err := writeAtomic(path, func() ([]byte, error) { return []byte("first\n"), nil }); err != nil {
		t.Fatalf("writeAtomic() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o644 {
		t.Fatalf("mode = %o, want 644", got)
	}

	previous, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := writeAtomic(path, func() ([]byte, error) { return nil, errors.New("render failed") }); err == nil {
		t.Fatal("writeAtomic() error = nil, want render error")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() after failure error = %v", err)
	}
	if !bytes.Equal(got, previous) {
		t.Fatalf("file changed on render failure: got %q, want %q", got, previous)
	}
}

func ataDrive() smart.DriveInfo {
	return smart.DriveInfo{
		Device:            "/dev/sda",
		Serial:            "ATA123",
		Model:             "WDC WD10",
		Interface:         "SATA HDD (7200 RPM)",
		RotationRate:      7200,
		HealthStatus:      smart.HealthGood,
		Temperature:       38,
		PowerOnHours:      1234,
		PowerCycles:       56,
		CRCErrors:         2,
		WearLevelingValue: -1,
		TotalBytesWritten: -1,
	}
}

func nvmeDrive() smart.DriveInfo {
	return smart.DriveInfo{
		Device:            "/dev/nvme0n1",
		Serial:            "NVME456",
		Model:             "Samsung 980",
		Interface:         "NVMe",
		IsNVMe:            true,
		IsSSD:             true,
		HealthStatus:      smart.HealthCaution,
		Temperature:       44,
		PowerOnHours:      999,
		PowerCycles:       12,
		WearLevelingValue: -1,
		TotalBytesWritten: 123456789,
		NVMeHealthLog:     &smart.NVMeHealthLog{PercentageUsed: 7},
	}
}

func scanErrorDrive() smart.DriveInfo {
	return smart.DriveInfo{
		Device:            "/dev/sdb",
		Serial:            "ERR789",
		Model:             "Broken",
		Interface:         "SATA SSD",
		IsSSD:             true,
		HealthStatus:      smart.HealthUnknown,
		WearLevelingValue: -1,
		TotalBytesWritten: -1,
		ScanError:         errors.New("smartctl failed"),
	}
}

// headerOnly is the Render output for an empty drive slice: every family's
// HELP/TYPE block, with no samples.
func headerOnly() string {
	var b strings.Builder
	for _, h := range [][2]string{
		{"nuther_smart_health_status", "Drive health status, one-hot: 1 for the current status label."},
		{"nuther_smart_health_percent", "Remaining drive life as a percentage (0-100)."},
		{"nuther_smart_temperature_celsius", "Current drive temperature in degrees Celsius."},
		{"nuther_smart_reallocated_sectors", "Count of reallocated sectors."},
		{"nuther_smart_pending_sectors", "Count of sectors pending reallocation."},
		{"nuther_smart_uncorrectable_sectors", "Count of offline uncorrectable sectors."},
		{"nuther_smart_crc_errors", "Count of UDMA CRC errors (cable/interface problems)."},
		{"nuther_smart_power_on_hours", "Total power-on hours."},
		{"nuther_smart_power_cycles", "Total power cycles."},
		{"nuther_smart_wear_leveling_value", "Normalized wear leveling value (100 new, 0 worn out); omitted when unavailable."},
		{"nuther_smart_written_bytes_total", "Total bytes written; omitted when unavailable."},
		{"nuther_smart_snapshot_timestamp_seconds", "Unix timestamp of the snapshot pass that produced these metrics."},
		{"nuther_smart_scan_success", "1 when the drive scan succeeded, 0 when this drive's scan errored."},
	} {
		typ := "gauge"
		if h[0] == "nuther_smart_written_bytes_total" {
			typ = "counter"
		}
		b.WriteString("# HELP " + h[0] + " " + h[1] + "\n")
		b.WriteString("# TYPE " + h[0] + " " + typ + "\n")
	}
	return b.String()
}

// fullOutput is the Render output for nvmeDrive, ataDrive and scanErrorDrive
// (sorted by device: nvme0n1, sda, sdb).
func fullOutput() string {
	nvme := `device="/dev/nvme0n1",serial="NVME456",model="Samsung 980",type="NVMe SSD",interface="NVMe"`
	ata := `device="/dev/sda",serial="ATA123",model="WDC WD10",type="HDD",interface="SATA HDD (7200 RPM)"`
	bad := `device="/dev/sdb",serial="ERR789",model="Broken",type="SATA SSD",interface="SATA SSD"`

	return `# HELP nuther_smart_health_status Drive health status, one-hot: 1 for the current status label.
# TYPE nuther_smart_health_status gauge
nuther_smart_health_status{` + nvme + `,status="CAUTION"} 1
nuther_smart_health_status{` + ata + `,status="GOOD"} 1
nuther_smart_health_status{` + bad + `,status="UNKNOWN"} 1
# HELP nuther_smart_health_percent Remaining drive life as a percentage (0-100).
# TYPE nuther_smart_health_percent gauge
nuther_smart_health_percent{` + nvme + `} 93
nuther_smart_health_percent{` + ata + `} 100
nuther_smart_health_percent{` + bad + `} 100
# HELP nuther_smart_temperature_celsius Current drive temperature in degrees Celsius.
# TYPE nuther_smart_temperature_celsius gauge
nuther_smart_temperature_celsius{` + nvme + `} 44
nuther_smart_temperature_celsius{` + ata + `} 38
nuther_smart_temperature_celsius{` + bad + `} 0
# HELP nuther_smart_reallocated_sectors Count of reallocated sectors.
# TYPE nuther_smart_reallocated_sectors gauge
nuther_smart_reallocated_sectors{` + nvme + `} 0
nuther_smart_reallocated_sectors{` + ata + `} 0
nuther_smart_reallocated_sectors{` + bad + `} 0
# HELP nuther_smart_pending_sectors Count of sectors pending reallocation.
# TYPE nuther_smart_pending_sectors gauge
nuther_smart_pending_sectors{` + nvme + `} 0
nuther_smart_pending_sectors{` + ata + `} 0
nuther_smart_pending_sectors{` + bad + `} 0
# HELP nuther_smart_uncorrectable_sectors Count of offline uncorrectable sectors.
# TYPE nuther_smart_uncorrectable_sectors gauge
nuther_smart_uncorrectable_sectors{` + nvme + `} 0
nuther_smart_uncorrectable_sectors{` + ata + `} 0
nuther_smart_uncorrectable_sectors{` + bad + `} 0
# HELP nuther_smart_crc_errors Count of UDMA CRC errors (cable/interface problems).
# TYPE nuther_smart_crc_errors gauge
nuther_smart_crc_errors{` + nvme + `} 0
nuther_smart_crc_errors{` + ata + `} 2
nuther_smart_crc_errors{` + bad + `} 0
# HELP nuther_smart_power_on_hours Total power-on hours.
# TYPE nuther_smart_power_on_hours gauge
nuther_smart_power_on_hours{` + nvme + `} 999
nuther_smart_power_on_hours{` + ata + `} 1234
nuther_smart_power_on_hours{` + bad + `} 0
# HELP nuther_smart_power_cycles Total power cycles.
# TYPE nuther_smart_power_cycles gauge
nuther_smart_power_cycles{` + nvme + `} 12
nuther_smart_power_cycles{` + ata + `} 56
nuther_smart_power_cycles{` + bad + `} 0
# HELP nuther_smart_wear_leveling_value Normalized wear leveling value (100 new, 0 worn out); omitted when unavailable.
# TYPE nuther_smart_wear_leveling_value gauge
# HELP nuther_smart_written_bytes_total Total bytes written; omitted when unavailable.
# TYPE nuther_smart_written_bytes_total counter
nuther_smart_written_bytes_total{` + nvme + `} 123456789
# HELP nuther_smart_snapshot_timestamp_seconds Unix timestamp of the snapshot pass that produced these metrics.
# TYPE nuther_smart_snapshot_timestamp_seconds gauge
nuther_smart_snapshot_timestamp_seconds{` + nvme + `} 1700000000
nuther_smart_snapshot_timestamp_seconds{` + ata + `} 1700000000
nuther_smart_snapshot_timestamp_seconds{` + bad + `} 1700000000
# HELP nuther_smart_scan_success 1 when the drive scan succeeded, 0 when this drive's scan errored.
# TYPE nuther_smart_scan_success gauge
nuther_smart_scan_success{` + nvme + `} 1
nuther_smart_scan_success{` + ata + `} 1
nuther_smart_scan_success{` + bad + `} 0
`
}
