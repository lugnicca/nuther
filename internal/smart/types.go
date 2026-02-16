package smart

import "time"

// HealthStatus represents the health state of a drive or attribute
type HealthStatus string

// Health status constants
const (
	HealthGood    HealthStatus = "GOOD"
	HealthInfo    HealthStatus = "INFO"
	HealthCaution HealthStatus = "CAUTION"
	HealthBad     HealthStatus = "BAD"
	HealthUnknown HealthStatus = "UNKNOWN"
)

// SmartAttribute represents a single S.M.A.R.T. attribute for SATA drives
type SmartAttribute struct {
	ID         int
	Name       string
	Value      int
	Worst      int
	Threshold  int
	RawValue   int64
	RawString  string
	Flags      string
	Type       string
	Updated    string
	WhenFailed string
}

// GetStatus returns the health status of this attribute
func (a *SmartAttribute) GetStatus() HealthStatus {
	if CriticalAttributes[a.ID] && a.RawValue > 0 {
		if a.RawValue > CriticalAttrBadThreshold {
			return HealthBad
		}
		return HealthCaution
	}

	if ConnectionAttributes[a.ID] && a.RawValue > 0 {
		return HealthInfo
	}

	if a.Threshold > 0 {
		if a.Value <= a.Threshold {
			return HealthBad
		}
		if a.Value < a.Threshold+AttrValueMargin {
			return HealthCaution
		}
	}

	return HealthGood
}

// IsCritical returns true if this is a critical attribute
func (a *SmartAttribute) IsCritical() bool {
	return CriticalAttributes[a.ID]
}

// NVMeAttribute represents a health attribute for NVMe drives
type NVMeAttribute struct {
	Name         string
	RawValue     string
	NumericValue int64
	Status       HealthStatus
	Description  string
}

// NVMeHealthLog represents the NVMe SMART/Health Information Log
type NVMeHealthLog struct {
	CriticalWarning         int
	Temperature             int
	AvailableSpare          int
	AvailableSpareThreshold int
	PercentageUsed          int
	DataUnitsRead           int64
	DataUnitsWritten        int64
	HostReadCommands        int64
	HostWriteCommands       int64
	ControllerBusyTime      int64
	PowerCycles             int64
	PowerOnHours            int64
	UnsafeShutdowns         int64
	MediaErrors             int64
	ErrorLogEntries         int64
	WarningCompTempTime     int64
	CriticalCompTempTime    int64
	TempSensor1             int
	TempSensor2             int
	TempSensor3             int
	TempSensor4             int
}

// DriveInfo contains all information about a storage drive
type DriveInfo struct {
	Device      string
	Model       string
	ModelFamily string
	Serial      string
	Firmware    string
	WWN         string

	Capacity       string
	CapacityBytes  int64
	LogicalSector  int
	PhysicalSector int

	FormFactor   string
	RotationRate int
	Interface    string

	SmartSupported bool
	SmartEnabled   bool
	HealthStatus   HealthStatus
	HealthPassed   bool

	Temperature  int
	PowerOnHours int64
	PowerCycles  int64

	ReallocatedSectors   int64
	PendingSectors       int64
	UncorrectableSectors int64
	CRCErrors            int64
	WearLevelingValue    int   // Normalized wear value (100=new, 0=worn out), -1 if unavailable
	TotalBytesWritten   int64 // Total bytes written, -1 if unavailable

	Attributes     []SmartAttribute
	NVMeAttributes []NVMeAttribute
	NVMeHealthLog  *NVMeHealthLog

	IsNVMe    bool
	IsSSD     bool
	IsUSB     bool
	IsVirtual bool
	IsDemo    bool // True if this is demo/sample data

	LastUpdate time.Time
	ScanError  error
}

// HealthPercent returns the health/remaining-life percentage for the drive.
// NVMe: 100 - PercentageUsed. SATA SSD: normalized wear leveling value.
// HDD/other: mapped from HealthStatus (GOOD=100, CAUTION=50, BAD=0).
func (d *DriveInfo) HealthPercent() int {
	if d.IsNVMe && d.NVMeHealthLog != nil {
		p := 100 - d.NVMeHealthLog.PercentageUsed
		if p < 0 {
			return 0
		}
		return p
	}
	if d.WearLevelingValue >= 0 {
		return d.WearLevelingValue
	}
	// Fallback: derive from overall health status
	switch d.HealthStatus {
	case HealthBad:
		return 0
	case HealthCaution:
		return 50
	default:
		return 100
	}
}

// GetDriveType returns a human-readable drive type string
func (d *DriveInfo) GetDriveType() string {
	if d.IsNVMe {
		return "NVMe SSD"
	}
	if d.IsSSD {
		return "SATA SSD"
	}
	if d.RotationRate > 0 {
		return "HDD"
	}
	if d.IsVirtual {
		return "Virtual"
	}
	return "Unknown"
}

// HasCriticalIssues returns true if any critical attributes have problems
func (d *DriveInfo) HasCriticalIssues() bool {
	return d.ReallocatedSectors > 0 ||
		d.PendingSectors > 0 ||
		d.UncorrectableSectors > 0
}

// HasConnectionIssues returns true if cable/interface errors are detected (not a drive defect)
func (d *DriveInfo) HasConnectionIssues() bool {
	return d.CRCErrors > 0
}

// GetAttributeCount returns the number of attributes
func (d *DriveInfo) GetAttributeCount() int {
	if d.IsNVMe {
		return len(d.NVMeAttributes)
	}
	return len(d.Attributes)
}

// Smartctl JSON response types

// SmartctlDevice represents the device section of smartctl JSON output
type SmartctlDevice struct {
	Name     string `json:"name"`
	InfoName string `json:"info_name"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
}

// SmartctlCapacity represents user capacity information
type SmartctlCapacity struct {
	Blocks int64 `json:"blocks"`
	Bytes  int64 `json:"bytes"`
}

// SmartctlFormFactor represents form factor information
type SmartctlFormFactor struct {
	ATAValue int    `json:"ata_value"`
	Name     string `json:"name"`
}

// SmartctlSmartStatus represents SMART overall health status
type SmartctlSmartStatus struct {
	Passed bool `json:"passed"`
}

// SmartctlTemperature represents temperature information
type SmartctlTemperature struct {
	Current int `json:"current"`
}

// SmartctlPowerOnTime represents power-on time
type SmartctlPowerOnTime struct {
	Hours int64 `json:"hours"`
}

// SmartctlAtaAttribute represents a single ATA SMART attribute
type SmartctlAtaAttribute struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Value      int    `json:"value"`
	Worst      int    `json:"worst"`
	Thresh     int    `json:"thresh"`
	WhenFailed string `json:"when_failed"`
	Flags      struct {
		Value         int    `json:"value"`
		String        string `json:"string"`
		Prefailure    bool   `json:"prefailure"`
		UpdatedOnline bool   `json:"updated_online"`
		Performance   bool   `json:"performance"`
		ErrorRate     bool   `json:"error_rate"`
		EventCount    bool   `json:"event_count"`
		AutoKeep      bool   `json:"auto_keep"`
	} `json:"flags"`
	Raw struct {
		Value  int64  `json:"value"`
		String string `json:"string"`
	} `json:"raw"`
}

// SmartctlAtaSmartAttributes represents the ATA SMART attributes table
type SmartctlAtaSmartAttributes struct {
	Revision int                    `json:"revision"`
	Table    []SmartctlAtaAttribute `json:"table"`
}

// SmartctlNvmeHealthLog represents NVMe SMART/Health Information
type SmartctlNvmeHealthLog struct {
	CriticalWarning         int   `json:"critical_warning"`
	Temperature             int   `json:"temperature"`
	AvailableSpare          int   `json:"available_spare"`
	AvailableSpareThreshold int   `json:"available_spare_threshold"`
	PercentageUsed          int   `json:"percentage_used"`
	DataUnitsRead           int64 `json:"data_units_read"`
	DataUnitsWritten        int64 `json:"data_units_written"`
	HostReads               int64 `json:"host_reads"`
	HostWrites              int64 `json:"host_writes"`
	ControllerBusyTime      int64 `json:"controller_busy_time"`
	PowerCycles             int64 `json:"power_cycles"`
	PowerOnHours            int64 `json:"power_on_hours"`
	UnsafeShutdowns         int64 `json:"unsafe_shutdowns"`
	MediaErrors             int64 `json:"media_errors"`
	NumErrLogEntries        int64 `json:"num_err_log_entries"`
	WarningTempTime         int64 `json:"warning_temp_time"`
	CriticalCompTempTime    int64 `json:"critical_comp_temp_time"`
}

// SmartctlOutput represents the complete smartctl JSON output
type SmartctlOutput struct {
	JSONFormatVersion []int `json:"json_format_version"`
	Smartctl          struct {
		Version      []int    `json:"version"`
		SVNRevision  string   `json:"svn_revision"`
		PlatformInfo string   `json:"platform_info"`
		BuildInfo    string   `json:"build_info"`
		Argv         []string `json:"argv"`
		ExitStatus   int      `json:"exit_status"`
	} `json:"smartctl"`
	Device       SmartctlDevice `json:"device"`
	ModelName    string         `json:"model_name"`
	ModelFamily  string         `json:"model_family"`
	SerialNumber string         `json:"serial_number"`
	WWN          struct {
		NAA int64 `json:"naa"`
		OUI int64 `json:"oui"`
		ID  int64 `json:"id"`
	} `json:"wwn"`
	FirmwareVersion               string                     `json:"firmware_version"`
	UserCapacity                  SmartctlCapacity           `json:"user_capacity"`
	LogicalBlockSize              int                        `json:"logical_block_size"`
	PhysicalBlockSize             int                        `json:"physical_block_size"`
	RotationRate                  int                        `json:"rotation_rate"`
	FormFactor                    SmartctlFormFactor         `json:"form_factor"`
	InSmartctlDatabase            bool                       `json:"in_smartctl_database"`
	SmartStatus                   SmartctlSmartStatus        `json:"smart_status"`
	Temperature                   SmartctlTemperature        `json:"temperature"`
	PowerOnTime                   SmartctlPowerOnTime        `json:"power_on_time"`
	PowerCycleCount               int64                      `json:"power_cycle_count"`
	AtaSmartAttributes            SmartctlAtaSmartAttributes `json:"ata_smart_attributes"`
	NvmeSmartHealthInformationLog SmartctlNvmeHealthLog      `json:"nvme_smart_health_information_log"`
}

// SmartctlScanResult represents the result of smartctl --scan
type SmartctlScanResult struct {
	JSONFormatVersion []int `json:"json_format_version"`
	Smartctl          struct {
		ExitStatus int `json:"exit_status"`
	} `json:"smartctl"`
	Devices []struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"devices"`
}
