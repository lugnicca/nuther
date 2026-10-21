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

// The textfile collector matches files solely by the .prom suffix, so the
// scratch file created by writeAtomic must not carry it or a concurrent
// scrape could read a half-written file.
func TestTempFileCannotBeScraped(t *testing.T) {
	var created string
	orig := createTemp
	createTemp = func(dir string) (*os.File, error) {
		f, err := orig(dir)
		if err == nil {
			created = f.Name()
		}
		return f, err
	}
	t.Cleanup(func() { createTemp = orig })

	dir := t.TempDir()
	if err := writeAtomic(filepath.Join(dir, FileName), func() ([]byte, error) {
		return []byte("x\n"), nil
	}); err != nil {
		t.Fatalf("writeAtomic() error = %v", err)
	}
	if created == "" {
		t.Fatal("writeAtomic() did not create a scratch file through createTemp")
	}
	if strings.HasSuffix(filepath.Base(created), ".prom") {
		t.Fatalf("temp file %q ends in .prom and would be scraped", created)
	}
}

// A real filesystem failure, not a callback error injected before any I/O.
func TestWriteFileFilesystemErrors(t *testing.T) {
	drives := []smart.DriveInfo{ataDrive()}
	now := time.Unix(1700000000, 0).UTC()

	t.Run("missing directory", func(t *testing.T) {
		err := WriteFile(filepath.Join(t.TempDir(), "missing"), drives, now)
		if err == nil {
			t.Fatal("WriteFile() error = nil, want create failure")
		}
	})

	t.Run("target is a directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, FileName), 0o755); err != nil {
			t.Fatalf("Mkdir() error = %v", err)
		}
		err := WriteFile(dir, drives, now)
		if err == nil {
			t.Fatal("WriteFile() error = nil, want rename failure")
		}
	})
}

func TestRenderOmitsHealthPercentForUnknown(t *testing.T) {
	drive := ataDrive()
	drive.HealthStatus = smart.HealthUnknown
	out := Render([]smart.DriveInfo{drive}, time.Unix(1700000000, 0).UTC())
	if strings.Contains(out, "nuther_smart_health_percent{") {
		t.Fatalf("unknown drive still rendered health_percent:\n%s", out)
	}
}

// A drive whose status command failed still yields a usable attribute, so it is
// exported; but with smart_status absent its health is unknown, not failed, so
// the textfile must say UNKNOWN and omit health_percent rather than invent
// BAD/0. This exercises the parse -> render path end to end.
func TestRenderAttributeOnlyDriveStaysUnknown(t *testing.T) {
	raw := []byte(`{
	  "device": {"name": "/dev/probe-attr", "type": "sat", "protocol": "ATA"},
	  "model_name": "ATTRIBUTE ONLY",
	  "serial_number": "ATTRONLY",
	  "rotation_rate": 7200,
	  "power_on_time": {"hours": 42},
	  "ata_smart_attributes": {"table": [{"id": 9, "name": "Power_On_Hours", "value": 99, "worst": 99, "thresh": 0, "raw": {"value": 42, "string": "42"}}]}
	}`)
	drive, err := smart.ParseSmartctlDriveJSON(raw)
	if err != nil {
		t.Fatalf("ParseSmartctlDriveJSON() error = %v", err)
	}
	out := Render([]smart.DriveInfo{drive}, time.Unix(1700000000, 0).UTC())
	if !strings.Contains(out, `nuther_smart_health_status{`) || !strings.Contains(out, `status="UNKNOWN"`) {
		t.Fatalf("attribute-only drive did not export UNKNOWN health:\n%s", out)
	}
	if strings.Contains(out, "nuther_smart_health_percent{") {
		t.Fatalf("unknown drive still rendered health_percent:\n%s", out)
	}
}

// The drive's own smart_status verdict outranks any attribute-derived caution,
// and severity never moves down. An explicit passed:false with sector errors
// stays BAD; an absent verdict with the same errors stays UNKNOWN with
// health_percent omitted; only a genuinely passed:true drive may be raised to
// CAUTION by its counters.
func TestRenderHealthStatusPrecedence(t *testing.T) {
	noDefect := `{"id": 9, "name": "Power_On_Hours", "value": 99, "worst": 99, "thresh": 0, "raw": {"value": 42, "string": "42"}}`
	reallocated := `{"id": 5, "name": "Reallocated_Sector_Ct", "value": 100, "worst": 100, "thresh": 10, "raw": {"value": 1, "string": "1"}}`
	fixture := func(smartStatus, attrs string) []byte {
		return []byte(`{"device": {"name": "/dev/probe", "type": "sat", "protocol": "ATA"},
		  "model_name": "PRECEDENCE", "serial_number": "PREC", "rotation_rate": 7200,
		  ` + smartStatus + `"ata_smart_attributes": {"table": [` + attrs + `]}}`)
	}

	for _, tc := range []struct {
		name        string
		raw         []byte
		wantStatus  string
		wantPercent bool
		wantValue   string
	}{
		{"explicit false with sector errors stays BAD", fixture(`"smart_status": {"passed": false},`, reallocated), "BAD", true, "0"},
		{"absent verdict with sector errors stays UNKNOWN", fixture(``, reallocated), "UNKNOWN", false, ""},
		{"passed true with sector errors is CAUTION", fixture(`"smart_status": {"passed": true},`, reallocated), "CAUTION", true, "50"},
		{"passed true healthy is GOOD", fixture(`"smart_status": {"passed": true},`, noDefect), "GOOD", true, "100"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			drive, err := smart.ParseSmartctlDriveJSON(tc.raw)
			if err != nil {
				t.Fatalf("ParseSmartctlDriveJSON() error = %v", err)
			}
			out := Render([]smart.DriveInfo{drive}, time.Unix(1700000000, 0).UTC())
			if !strings.Contains(out, `status="`+tc.wantStatus+`"`) {
				t.Fatalf("health status = %q, want %q:\n%s", drive.HealthStatus, tc.wantStatus, out)
			}
			hasPercent := strings.Contains(out, "nuther_smart_health_percent{")
			if hasPercent != tc.wantPercent {
				t.Fatalf("health_percent present = %v, want %v:\n%s", hasPercent, tc.wantPercent, out)
			}
			if tc.wantPercent {
				want := `nuther_smart_health_percent{device="/dev/probe",serial="PREC",model="PRECEDENCE",type="HDD",interface="SATA HDD (7200 RPM)"} ` + tc.wantValue
				if !strings.Contains(out, want) {
					t.Fatalf("health_percent sample = want %q:\n%s", want, out)
				}
			}
		})
	}
}

// A measurement smartctl never supplied must not surface as a zero-valued
// sample: availability travels from the JSON parse into Render, so an absent
// field is omitted while a genuinely measured zero is kept. These fixtures
// reproduce the review's cases end to end through ParseSmartctlDriveJSON.
func TestRenderOmitsUnmeasuredMeasurements(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		present []string
		absent  []string
	}{
		{
			name: "devstat only exports only the measured written bytes",
			raw: `{
			  "device": {"name": "/dev/probe", "type": "sat", "protocol": "ATA"},
			  "model_name": "REVIEW ATA", "serial_number": "REVIEW-ATA", "rotation_rate": 7200,
			  "logical_block_size": 512,
			  "ata_device_statistics": {"pages": [{"number": 1, "table": [{"offset": 24, "value": 1000, "flags": {"valid": true}}]}]}
			}`,
			present: []string{
				`nuther_smart_written_bytes_total{device="/dev/probe"`,
			},
			absent: []string{
				"nuther_smart_temperature_celsius{",
				"nuther_smart_reallocated_sectors{",
				"nuther_smart_pending_sectors{",
				"nuther_smart_uncorrectable_sectors{",
				"nuther_smart_crc_errors{",
				"nuther_smart_power_on_hours{",
				"nuther_smart_power_cycles{",
				"nuther_smart_health_percent{",
			},
		},
		{
			name: "healthy nvme omits the ata-specific counters",
			raw: `{
			  "device": {"name": "/dev/nvme0", "type": "nvme", "protocol": "NVMe"},
			  "model_name": "Samsung 980", "serial_number": "NVME456", "rotation_rate": 0,
			  "smart_status": {"passed": true},
			  "nvme_smart_health_information_log": {"temperature": 38, "percentage_used": 2, "data_units_written": 18000000, "power_cycles": 150, "power_on_hours": 2500}
			}`,
			present: []string{
				`nuther_smart_health_percent{device="/dev/nvme0"`,
				`nuther_smart_temperature_celsius{device="/dev/nvme0"`,
				`nuther_smart_power_on_hours{device="/dev/nvme0"`,
				`nuther_smart_power_cycles{device="/dev/nvme0"`,
				`nuther_smart_written_bytes_total{device="/dev/nvme0"`,
			},
			absent: []string{
				"nuther_smart_reallocated_sectors{",
				"nuther_smart_pending_sectors{",
				"nuther_smart_uncorrectable_sectors{",
				"nuther_smart_crc_errors{",
			},
		},
		{
			name: "sparse nvme invents neither life nor written bytes",
			raw: `{
			  "device": {"name": "/dev/nvme0", "type": "nvme", "protocol": "NVMe"},
			  "model_name": "Samsung 980", "serial_number": "NVME456", "rotation_rate": 0,
			  "smart_status": {"passed": true},
			  "nvme_smart_health_information_log": {"temperature": 38}
			}`,
			present: []string{
				`nuther_smart_health_status{device="/dev/nvme0"`,
				`nuther_smart_temperature_celsius{device="/dev/nvme0"`,
			},
			absent: []string{
				"nuther_smart_health_percent{",
				"nuther_smart_reallocated_sectors{",
				"nuther_smart_pending_sectors{",
				"nuther_smart_uncorrectable_sectors{",
				"nuther_smart_crc_errors{",
				"nuther_smart_power_on_hours{",
				"nuther_smart_power_cycles{",
				"nuther_smart_written_bytes_total{",
			},
		},
		{
			name: "attribute-only ata reads power-on hours from the attribute",
			raw: `{
			  "device": {"name": "/dev/probe", "type": "sat", "protocol": "ATA"},
			  "model_name": "ATTRIBUTE ONLY", "serial_number": "ATTRONLY", "rotation_rate": 7200,
			  "ata_smart_attributes": {"table": [{"id": 9, "name": "Power_On_Hours", "value": 99, "worst": 99, "thresh": 0, "raw": {"value": 42, "string": "42"}}]}
			}`,
			present: []string{
				`nuther_smart_power_on_hours{device="/dev/probe",serial="ATTRONLY",model="ATTRIBUTE ONLY",type="HDD",interface="SATA HDD (7200 RPM)"} 42`,
			},
			absent: []string{
				"nuther_smart_health_percent{",
				"nuther_smart_temperature_celsius{",
				"nuther_smart_reallocated_sectors{",
				"nuther_smart_pending_sectors{",
				"nuther_smart_uncorrectable_sectors{",
				"nuther_smart_crc_errors{",
				"nuther_smart_power_cycles{",
				"nuther_smart_written_bytes_total{",
			},
		},
		{
			name: "null attribute raw value is not a measurement",
			raw: `{
			  "device": {"name": "/dev/probe", "type": "sat", "protocol": "ATA"},
			  "model_name": "NULL RAW", "serial_number": "NULLRAW", "rotation_rate": 7200,
			  "logical_block_size": 512,
			  "ata_smart_attributes": {"table": [
			    {"id": 1, "name": "Read_Error_Rate", "value": 100, "worst": 100, "thresh": 0, "raw": {"value": 1, "string": "1"}},
			    {"id": 5, "name": "Reallocated_Sector_Ct", "value": 100, "worst": 100, "thresh": 0, "raw": {"value": null, "string": ""}}
			  ]}
			}`,
			absent: []string{"nuther_smart_reallocated_sectors{"},
		},
		{
			name: "null device statistic value is not a measurement",
			raw: `{
			  "device": {"name": "/dev/probe", "type": "sat", "protocol": "ATA"},
			  "model_name": "NULL STAT", "serial_number": "NULLSTAT", "rotation_rate": 7200,
			  "logical_block_size": 512,
			  "ata_device_statistics": {"pages": [{"number": 1, "table": [{"offset": 24, "value": null, "flags": {"valid": true}}]}]}
			}`,
			absent: []string{"nuther_smart_written_bytes_total{"},
		},
		{
			name: "null attribute 241 falls back to the valid statistics value",
			raw: `{
			  "device": {"name": "/dev/probe", "type": "sat", "protocol": "ATA"},
			  "model_name": "NULL 241", "serial_number": "NULL241", "rotation_rate": 7200,
			  "logical_block_size": 512,
			  "ata_smart_attributes": {"table": [
			    {"id": 241, "name": "Total_LBAs_Written", "value": 100, "worst": 100, "thresh": 0, "raw": {"value": null, "string": ""}}
			  ]},
			  "ata_device_statistics": {"pages": [{"number": 1, "table": [{"offset": 24, "value": 1000, "flags": {"valid": true}}]}]}
			}`,
			present: []string{
				`nuther_smart_written_bytes_total{device="/dev/probe",serial="NULL241",model="NULL 241",type="HDD",interface="SATA HDD (7200 RPM)"} 512000`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drive, err := smart.ParseSmartctlDriveJSON([]byte(tt.raw))
			if err != nil {
				t.Fatalf("ParseSmartctlDriveJSON() error = %v", err)
			}
			out := Render([]smart.DriveInfo{drive}, time.Unix(1700000000, 0).UTC())
			for _, want := range tt.present {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q:\n%s", want, out)
				}
			}
			for _, unwanted := range tt.absent {
				if strings.Contains(out, unwanted) {
					t.Errorf("output unexpectedly contains %q:\n%s", unwanted, out)
				}
			}
		})
	}
}

// A zero that smartctl did measure is a real value and must stay exported, not
// be dropped along with the genuinely absent fields.
func TestRenderKeepsMeasuredZero(t *testing.T) {
	raw := []byte(`{
	  "device": {"name": "/dev/probe", "type": "sat", "protocol": "ATA"},
	  "model_name": "ZERO DRIVE", "serial_number": "ZERO", "rotation_rate": 7200,
	  "smart_status": {"passed": true},
	  "temperature": {"current": 34}, "power_on_time": {"hours": 42}, "power_cycle_count": 7,
	  "ata_smart_attributes": {"table": [
	    {"id": 5, "name": "Reallocated_Sector_Ct", "value": 100, "worst": 100, "thresh": 10, "raw": {"value": 0, "string": "0"}},
	    {"id": 197, "name": "Current_Pending_Sector", "value": 100, "worst": 100, "thresh": 0, "raw": {"value": 0, "string": "0"}},
	    {"id": 198, "name": "Offline_Uncorrectable", "value": 100, "worst": 100, "thresh": 0, "raw": {"value": 0, "string": "0"}},
	    {"id": 199, "name": "UDMA_CRC_Error_Count", "value": 100, "worst": 100, "thresh": 0, "raw": {"value": 0, "string": "0"}}
	  ]}
	}`)
	drive, err := smart.ParseSmartctlDriveJSON(raw)
	if err != nil {
		t.Fatalf("ParseSmartctlDriveJSON() error = %v", err)
	}
	out := Render([]smart.DriveInfo{drive}, time.Unix(1700000000, 0).UTC())
	labels := `device="/dev/probe",serial="ZERO",model="ZERO DRIVE",type="HDD",interface="SATA HDD (7200 RPM)"`
	for _, want := range []string{
		"nuther_smart_reallocated_sectors{" + labels + "} 0",
		"nuther_smart_pending_sectors{" + labels + "} 0",
		"nuther_smart_uncorrectable_sectors{" + labels + "} 0",
		"nuther_smart_crc_errors{" + labels + "} 0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing measured zero %q:\n%s", want, out)
		}
	}
}

// A device statistic measured as zero is a real value and must stay exported,
// even though the same position with a null value is omitted.
func TestRenderKeepsMeasuredZeroStatistic(t *testing.T) {
	raw := []byte(`{
	  "device": {"name": "/dev/probe", "type": "sat", "protocol": "ATA"},
	  "model_name": "ZERO STAT", "serial_number": "ZEROSTAT", "rotation_rate": 7200,
	  "logical_block_size": 512,
	  "ata_device_statistics": {"pages": [{"number": 1, "table": [{"offset": 24, "value": 0, "flags": {"valid": true}}]}]}
	}`)
	drive, err := smart.ParseSmartctlDriveJSON(raw)
	if err != nil {
		t.Fatalf("ParseSmartctlDriveJSON() error = %v", err)
	}
	out := Render([]smart.DriveInfo{drive}, time.Unix(1700000000, 0).UTC())
	want := `nuther_smart_written_bytes_total{device="/dev/probe",serial="ZEROSTAT",model="ZERO STAT",type="HDD",interface="SATA HDD (7200 RPM)"} 0`
	if !strings.Contains(out, want) {
		t.Errorf("output missing measured zero %q:\n%s", want, out)
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
		Measured: smart.MeasuredFields{
			Temperature:   true,
			PowerOnHours:  true,
			PowerCycles:   true,
			Reallocated:   true,
			Pending:       true,
			Uncorrectable: true,
			CRC:           true,
		},
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
		Measured: smart.MeasuredFields{
			Temperature:     true,
			PowerOnHours:    true,
			PowerCycles:     true,
			WrittenBytes:    true,
			NVMePercentUsed: true,
		},
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
		{"nuther_smart_health_status", "Drive health status; 1 for the current status label, other status labels are absent rather than 0."},
		{"nuther_smart_health_percent", "Remaining drive life as a percentage (0-100); omitted when the drive health is unknown."},
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
// (sorted by device: nvme0n1, sda, sdb). The scan-error drive only emits
// scan_success 0; all its other series are absent.
func fullOutput() string {
	nvme := `device="/dev/nvme0n1",serial="NVME456",model="Samsung 980",type="NVMe SSD",interface="NVMe"`
	ata := `device="/dev/sda",serial="ATA123",model="WDC WD10",type="HDD",interface="SATA HDD (7200 RPM)"`
	bad := `device="/dev/sdb",serial="ERR789",model="Broken",type="SATA SSD",interface="SATA SSD"`

	return `# HELP nuther_smart_health_status Drive health status; 1 for the current status label, other status labels are absent rather than 0.
# TYPE nuther_smart_health_status gauge
nuther_smart_health_status{` + nvme + `,status="CAUTION"} 1
nuther_smart_health_status{` + ata + `,status="GOOD"} 1
# HELP nuther_smart_health_percent Remaining drive life as a percentage (0-100); omitted when the drive health is unknown.
# TYPE nuther_smart_health_percent gauge
nuther_smart_health_percent{` + nvme + `} 93
nuther_smart_health_percent{` + ata + `} 100
# HELP nuther_smart_temperature_celsius Current drive temperature in degrees Celsius.
# TYPE nuther_smart_temperature_celsius gauge
nuther_smart_temperature_celsius{` + nvme + `} 44
nuther_smart_temperature_celsius{` + ata + `} 38
# HELP nuther_smart_reallocated_sectors Count of reallocated sectors.
# TYPE nuther_smart_reallocated_sectors gauge
nuther_smart_reallocated_sectors{` + ata + `} 0
# HELP nuther_smart_pending_sectors Count of sectors pending reallocation.
# TYPE nuther_smart_pending_sectors gauge
nuther_smart_pending_sectors{` + ata + `} 0
# HELP nuther_smart_uncorrectable_sectors Count of offline uncorrectable sectors.
# TYPE nuther_smart_uncorrectable_sectors gauge
nuther_smart_uncorrectable_sectors{` + ata + `} 0
# HELP nuther_smart_crc_errors Count of UDMA CRC errors (cable/interface problems).
# TYPE nuther_smart_crc_errors gauge
nuther_smart_crc_errors{` + ata + `} 2
# HELP nuther_smart_power_on_hours Total power-on hours.
# TYPE nuther_smart_power_on_hours gauge
nuther_smart_power_on_hours{` + nvme + `} 999
nuther_smart_power_on_hours{` + ata + `} 1234
# HELP nuther_smart_power_cycles Total power cycles.
# TYPE nuther_smart_power_cycles gauge
nuther_smart_power_cycles{` + nvme + `} 12
nuther_smart_power_cycles{` + ata + `} 56
# HELP nuther_smart_wear_leveling_value Normalized wear leveling value (100 new, 0 worn out); omitted when unavailable.
# TYPE nuther_smart_wear_leveling_value gauge
# HELP nuther_smart_written_bytes_total Total bytes written; omitted when unavailable.
# TYPE nuther_smart_written_bytes_total counter
nuther_smart_written_bytes_total{` + nvme + `} 123456789
# HELP nuther_smart_snapshot_timestamp_seconds Unix timestamp of the snapshot pass that produced these metrics.
# TYPE nuther_smart_snapshot_timestamp_seconds gauge
nuther_smart_snapshot_timestamp_seconds{` + nvme + `} 1700000000
nuther_smart_snapshot_timestamp_seconds{` + ata + `} 1700000000
# HELP nuther_smart_scan_success 1 when the drive scan succeeded, 0 when this drive's scan errored.
# TYPE nuther_smart_scan_success gauge
nuther_smart_scan_success{` + nvme + `} 1
nuther_smart_scan_success{` + ata + `} 1
nuther_smart_scan_success{` + bad + `} 0
`
}
