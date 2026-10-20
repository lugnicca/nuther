package smart

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

// SmartctlTimeout is the maximum time to wait for smartctl commands
const SmartctlTimeout = 30 * time.Second

// runSmartctl executes smartctl with the given arguments and returns the output.
// This is a package-level variable to allow test mocking.
var runSmartctl = func(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), SmartctlTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "smartctl", args...)
	return cmd.Output()
}

// ScanDrives discovers and retrieves information for all available drives.
// Devices whose detailed read failed are omitted, so the TUI only shows
// readable drives. Returns demo data with a non-nil error when no real drives
// are detected.
func ScanDrives() ([]DriveInfo, error) {
	return scanDrives(false)
}

// ScanDrivesIncludingErrors is ScanDrives but keeps devices whose detailed read
// failed as records carrying ScanError, so monitors can report a failing disk
// instead of letting it silently disappear.
func ScanDrivesIncludingErrors() ([]DriveInfo, error) {
	return scanDrives(true)
}

func scanDrives(includeErrors bool) ([]DriveInfo, error) {
	drives := scanWithSmartctl()
	if !includeErrors {
		drives = readableDrives(drives)
	}
	if len(drives) > 0 {
		return drives, nil
	}

	drives = scanCommonDevices()
	if len(drives) > 0 {
		return drives, nil
	}

	return CreateDemoData(), fmt.Errorf("no drives detected, using demo data")
}

// readableDrives drops records whose detailed read failed.
func readableDrives(drives []DriveInfo) []DriveInfo {
	out := make([]DriveInfo, 0, len(drives))
	for _, drive := range drives {
		if drive.ScanError == nil {
			out = append(out, drive)
		}
	}
	return out
}

func dedupeDrives(drives []DriveInfo) []DriveInfo {
	if len(drives) < 2 {
		return drives
	}

	seen := make(map[string]struct{}, len(drives))
	unique := make([]DriveInfo, 0, len(drives))
	for _, drive := range drives {
		key := driveIdentityKey(drive)
		if _, ok := seen[key]; ok {
			slog.Debug("duplicate drive ignored", "device", drive.Device, "model", drive.Model, "serial", drive.Serial)
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, drive)
	}
	return unique
}

func driveIdentityKey(drive DriveInfo) string {
	if serial := strings.TrimSpace(strings.ToLower(drive.Serial)); serial != "" {
		return "serial:" + serial
	}
	if wwn := strings.TrimSpace(strings.ToLower(drive.WWN)); wwn != "" {
		return "wwn:" + wwn
	}
	return "device:" + strings.TrimSpace(strings.ToLower(drive.Device))
}

// scanWithSmartctl uses smartctl --scan to discover drives
func scanWithSmartctl() []DriveInfo {
	// Devices whose detailed read failed during JSON discovery, kept so a
	// monitor still sees them when the textual retry cannot recover them.
	var failed []DriveInfo

	output, err := runSmartctl("--scan", "-j")
	if err != nil {
		slog.Debug("smartctl --scan -j failed", "error", err)
	} else {
		var scanResult SmartctlScanResult
		if err := json.Unmarshal(output, &scanResult); err != nil {
			slog.Debug("smartctl --scan -j parse failed", "error", err)
		} else if len(scanResult.Devices) > 0 {
			var drives []DriveInfo
			for _, device := range scanResult.Devices {
				drive, err := GetDriveInfoWithType(device.Name, device.Type)
				if err != nil {
					slog.Debug("failed to get drive info", "device", device.Name, "error", err)
					// Keep the device identity so monitors can report scan_success 0.
					drives = append(drives, DriveInfo{Device: device.Name, ScanError: err})
				} else if drive.Model != "" {
					drives = append(drives, drive)
				}
			}
			// The fallback decision counts usable drives only: when every
			// device failed, fall through to the textual retry, which may
			// recover a transient detailed-read failure.
			if len(readableDrives(drives)) > 0 {
				drives = dedupeDrives(coalesceByPath(drives))
				slog.Info("scan complete", "drives", len(drives))
				return drives
			}
			failed = drives
		}
	}

	output, err = runSmartctl("--scan")
	if err != nil {
		slog.Debug("smartctl --scan failed", "error", err)
		return failed
	}

	var drives []DriveInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			devicePath := parts[0]
			deviceType := ""
			for i, p := range parts {
				if p == "-d" && i+1 < len(parts) {
					deviceType = parts[i+1]
					break
				}
			}
			drive, err := GetDriveInfoWithType(devicePath, deviceType)
			if err != nil {
				slog.Debug("failed to get drive info", "device", devicePath, "error", err)
				// Keep the device identity so monitors can report scan_success 0.
				drives = append(drives, DriveInfo{Device: devicePath, ScanError: err})
			} else if drive.Model != "" {
				drives = append(drives, drive)
			}
		}
	}

	if len(drives) == 0 {
		return failed
	}
	return dedupeDrives(mergeScanResults(failed, drives))
}

// coalesceByPath collapses duplicate records for the same device path, letting
// a successful read replace a failed one for that path. It reuses the retry
// path's merge semantics so the direct JSON return and the textual retry agree
// on which record survives for a path.
func coalesceByPath(drives []DriveInfo) []DriveInfo {
	var failed, successful []DriveInfo
	for _, drive := range drives {
		if drive.ScanError != nil {
			failed = append(failed, drive)
		} else {
			successful = append(successful, drive)
		}
	}
	return mergeScanResults(failed, successful)
}

// mergeScanResults combines two discovery passes by device path. Records from
// override replace base records for the same path, so a device the retry
// managed to read is no longer reported as failed; a device known to only one
// pass is preserved. Duplicate base paths (e.g. a device the JSON pass listed
// twice) are coalesced first, so the override always replaces the single entry
// for that path instead of leaving a stale duplicate behind.
func mergeScanResults(base, override []DriveInfo) []DriveInfo {
	if len(base) == 0 {
		return override
	}
	index := make(map[string]int, len(base))
	merged := make([]DriveInfo, 0, len(base)+len(override))
	for _, drive := range base {
		if _, ok := index[drive.Device]; ok {
			slog.Debug("duplicate base device ignored", "device", drive.Device)
			continue
		}
		index[drive.Device] = len(merged)
		merged = append(merged, drive)
	}
	for _, drive := range override {
		if i, ok := index[drive.Device]; ok {
			merged[i] = drive
			continue
		}
		index[drive.Device] = len(merged)
		merged = append(merged, drive)
	}
	return merged
}

// scanCommonDevices checks common device paths for drives
func scanCommonDevices() []DriveInfo {
	commonPaths := GetCommonDevicePaths()

	var drives []DriveInfo
	for _, path := range commonPaths {
		drive, err := GetDriveInfo(path)
		if err != nil {
			slog.Debug("failed to get drive info", "device", path, "error", err)
		} else if drive.Model != "" {
			drives = append(drives, drive)
		}
	}

	drives = dedupeDrives(drives)
	if len(drives) > 0 {
		slog.Info("scan complete", "drives", len(drives))
	}
	return drives
}

// GetDriveInfo retrieves detailed information for a specific device
func GetDriveInfo(device string) (DriveInfo, error) {
	return GetDriveInfoWithType(device, "")
}

// smartctlOutputUsable reports whether smartctl output contains actual SMART
// data even when the command exited non-zero. smartctl returns warning exit
// statuses (e.g. bit 64: some ATA log structure not read) that are common on
// USB bridges but still include a full SMART attribute table. Identification
// strings alone (model/serial) are not evidence that SMART was read.
func smartctlOutputUsable(output []byte) bool {
	var smartOutput SmartctlOutput
	if err := json.Unmarshal(output, &smartOutput); err != nil {
		return false
	}
	return hasSmartData(smartOutput)
}

// hasSmartData reports whether the parsed smartctl output actually carries
// SMART data, as opposed to only the identification block smartctl prints
// before the read fails. ATA/SATA drives must expose at least one parsed
// attribute or device-statistics field; NVMe drives must expose a non-zero
// SMART/Health Information log. smart_status alone is not enough: it is a bare
// boolean the parser cannot distinguish from an absent field, so a read that
// failed and left it as a zero value must not pass as healthy.
func hasSmartData(s SmartctlOutput) bool {
	if s.Device.Protocol == "NVMe" {
		return s.NvmeSmartHealthInformationLog != (SmartctlNvmeHealthLog{})
	}
	return len(s.AtaSmartAttributes.Table) > 0 || len(s.AtaDeviceStatistics.Pages) > 0
}

// GetDriveInfoWithType retrieves detailed information for a specific device with optional type
func GetDriveInfoWithType(device string, deviceType string) (DriveInfo, error) {
	var output []byte
	var err error

	if deviceType != "" && deviceType != "scsi" {
		output, err = runSmartctl("-a", "-l", "devstat", "-j", "-d", deviceType, device)
	} else {
		output, err = runSmartctl("-a", "-l", "devstat", "-j", device)
	}

	if err != nil && !smartctlOutputUsable(output) {
		if deviceType == "scsi" || deviceType == "" {
			usbTypes := []string{"sat", "sat,auto", "usbsunplus", "usbjmicron", "usbcypress"}
			for _, usbType := range usbTypes {
				output, err = runSmartctl("-a", "-l", "devstat", "-j", "-d", usbType, device)
				if err == nil || smartctlOutputUsable(output) {
					break
				}
			}
			if err != nil && !smartctlOutputUsable(output) {
				return DriveInfo{}, fmt.Errorf("smartctl failed for %s: %w", device, err)
			}
		} else {
			return DriveInfo{}, fmt.Errorf("smartctl failed for %s: %w", device, err)
		}
	}

	var smartOutput SmartctlOutput
	if err := json.Unmarshal(output, &smartOutput); err != nil {
		return DriveInfo{}, fmt.Errorf("failed to parse smartctl output: %w", err)
	}
	// A command that succeeded (or warned) but returned no SMART data must not
	// surface as a zero-valued healthy drive: the textfile would then claim
	// scan_success 1 and a fabricated health status.
	if !hasSmartData(smartOutput) {
		return DriveInfo{}, fmt.Errorf("smartctl returned no SMART data for %s", device)
	}

	return parseDriveInfo(device, smartOutput), nil
}

func ParseSmartctlDriveJSON(data []byte) (DriveInfo, error) {
	var smartOutput SmartctlOutput
	if err := json.Unmarshal(data, &smartOutput); err != nil {
		return DriveInfo{}, fmt.Errorf("failed to parse smartctl output: %w", err)
	}

	device := smartOutput.Device.Name
	if device == "" {
		device = smartOutput.Device.InfoName
	}
	return parseDriveInfo(device, smartOutput), nil
}

// parseDriveInfo converts smartctl JSON output to a DriveInfo struct
func parseDriveInfo(device string, s SmartctlOutput) DriveInfo {
	drive := DriveInfo{
		Device:            device,
		Model:             s.ModelName,
		ModelFamily:       s.ModelFamily,
		Serial:            s.SerialNumber,
		Firmware:          s.FirmwareVersion,
		CapacityBytes:     s.UserCapacity.Bytes,
		Capacity:          FormatBytes(s.UserCapacity.Bytes),
		LogicalSector:     s.LogicalBlockSize,
		PhysicalSector:    s.PhysicalBlockSize,
		FormFactor:        s.FormFactor.Name,
		RotationRate:      s.RotationRate,
		Temperature:       s.Temperature.Current,
		PowerOnHours:      s.PowerOnTime.Hours,
		PowerCycles:       s.PowerCycleCount,
		SmartSupported:    true,
		SmartEnabled:      true,
		HealthPassed:      s.SmartStatus.Passed,
		WearLevelingValue: -1,
		TotalBytesWritten: -1,
		LastUpdate:        time.Now(),
	}

	if s.SmartStatus.Passed {
		drive.HealthStatus = HealthGood
	} else {
		drive.HealthStatus = HealthBad
	}

	if s.Device.Protocol == "NVMe" {
		drive.IsNVMe = true
		drive.IsSSD = true
		drive.Interface = "NVMe"
		parseNVMeDrive(&drive, s)
	} else if s.RotationRate == 0 {
		drive.IsSSD = true
		drive.Interface = "SATA SSD"
		parseATADrive(&drive, s)
	} else {
		drive.Interface = fmt.Sprintf("SATA HDD (%d RPM)", s.RotationRate)
		parseATADrive(&drive, s)
	}

	updateHealthStatus(&drive)

	return drive
}

// parseNVMeDrive parses NVMe-specific information
func parseNVMeDrive(drive *DriveInfo, s SmartctlOutput) {
	nvme := s.NvmeSmartHealthInformationLog

	drive.NVMeHealthLog = &NVMeHealthLog{
		CriticalWarning:         nvme.CriticalWarning,
		Temperature:             nvme.Temperature,
		AvailableSpare:          nvme.AvailableSpare,
		AvailableSpareThreshold: nvme.AvailableSpareThreshold,
		PercentageUsed:          nvme.PercentageUsed,
		DataUnitsRead:           nvme.DataUnitsRead,
		DataUnitsWritten:        nvme.DataUnitsWritten,
		HostReadCommands:        nvme.HostReads,
		HostWriteCommands:       nvme.HostWrites,
		ControllerBusyTime:      nvme.ControllerBusyTime,
		PowerCycles:             nvme.PowerCycles,
		PowerOnHours:            nvme.PowerOnHours,
		UnsafeShutdowns:         nvme.UnsafeShutdowns,
		MediaErrors:             nvme.MediaErrors,
		ErrorLogEntries:         nvme.NumErrLogEntries,
		WarningCompTempTime:     nvme.WarningTempTime,
		CriticalCompTempTime:    nvme.CriticalCompTempTime,
	}

	drive.Temperature = nvme.Temperature
	drive.PowerOnHours = nvme.PowerOnHours
	drive.PowerCycles = nvme.PowerCycles
	drive.TotalBytesWritten = nvme.DataUnitsWritten * 512 * 1000

	drive.NVMeAttributes = buildNVMeAttributes(nvme)
}

// buildNVMeAttributes creates the NVMe attributes list from health log
func buildNVMeAttributes(nvme SmartctlNvmeHealthLog) []NVMeAttribute {
	return []NVMeAttribute{
		{
			Name:         "Temperature",
			RawValue:     FormatTemperature(nvme.Temperature),
			NumericValue: int64(nvme.Temperature),
			Status:       GetTemperatureStatus(nvme.Temperature),
			Description:  "Current drive temperature",
		},
		{
			Name:         "Available Spare",
			RawValue:     FormatPercentage(nvme.AvailableSpare),
			NumericValue: int64(nvme.AvailableSpare),
			Status:       GetSpareStatus(nvme.AvailableSpare, nvme.AvailableSpareThreshold),
			Description:  "Remaining spare capacity for wear leveling",
		},
		{
			Name:         "Spare Threshold",
			RawValue:     FormatPercentage(nvme.AvailableSpareThreshold),
			NumericValue: int64(nvme.AvailableSpareThreshold),
			Status:       HealthGood,
			Description:  "Threshold for spare capacity warning",
		},
		{
			Name:         "Percentage Used",
			RawValue:     FormatPercentage(nvme.PercentageUsed),
			NumericValue: int64(nvme.PercentageUsed),
			Status:       GetUsageStatus(nvme.PercentageUsed),
			Description:  "Estimated percentage of NVM subsystem life used",
		},
		{
			Name:         "Data Read",
			RawValue:     FormatDataUnits(nvme.DataUnitsRead),
			NumericValue: nvme.DataUnitsRead,
			Status:       HealthGood,
			Description:  "Total data read from device",
		},
		{
			Name:         "Data Written",
			RawValue:     FormatDataUnits(nvme.DataUnitsWritten),
			NumericValue: nvme.DataUnitsWritten,
			Status:       HealthGood,
			Description:  "Total data written to device",
		},
		{
			Name:         "Host Read Commands",
			RawValue:     FormatNumber(nvme.HostReads),
			NumericValue: nvme.HostReads,
			Status:       HealthGood,
			Description:  "Total host read commands processed",
		},
		{
			Name:         "Host Write Commands",
			RawValue:     FormatNumber(nvme.HostWrites),
			NumericValue: nvme.HostWrites,
			Status:       HealthGood,
			Description:  "Total host write commands processed",
		},
		{
			Name:         "Controller Busy Time",
			RawValue:     fmt.Sprintf("%d minutes", nvme.ControllerBusyTime),
			NumericValue: nvme.ControllerBusyTime,
			Status:       HealthGood,
			Description:  "Time controller was busy with I/O commands",
		},
		{
			Name:         "Power Cycles",
			RawValue:     FormatNumber(nvme.PowerCycles),
			NumericValue: nvme.PowerCycles,
			Status:       HealthGood,
			Description:  "Number of power on/off cycles",
		},
		{
			Name:         "Power On Hours",
			RawValue:     FormatHours(nvme.PowerOnHours),
			NumericValue: nvme.PowerOnHours,
			Status:       HealthGood,
			Description:  "Total time powered on",
		},
		{
			Name:         "Unsafe Shutdowns",
			RawValue:     FormatNumber(nvme.UnsafeShutdowns),
			NumericValue: nvme.UnsafeShutdowns,
			Status:       GetShutdownStatus(nvme.UnsafeShutdowns),
			Description:  "Shutdowns without notification",
		},
		{
			Name:         "Media Errors",
			RawValue:     FormatNumber(nvme.MediaErrors),
			NumericValue: nvme.MediaErrors,
			Status:       GetErrorStatus(nvme.MediaErrors),
			Description:  "Media and data integrity errors",
		},
		{
			Name:         "Error Log Entries",
			RawValue:     FormatNumber(nvme.NumErrLogEntries),
			NumericValue: nvme.NumErrLogEntries,
			Status:       GetErrorStatus(nvme.NumErrLogEntries),
			Description:  "Number of error log entries",
		},
	}
}

// parseATADrive parses ATA/SATA-specific information
func parseATADrive(drive *DriveInfo, s SmartctlOutput) {
	for _, attr := range s.AtaSmartAttributes.Table {
		smartAttr := SmartAttribute{
			ID:         attr.ID,
			Name:       attr.Name,
			Value:      attr.Value,
			Worst:      attr.Worst,
			Threshold:  attr.Thresh,
			RawValue:   attr.Raw.Value,
			RawString:  attr.Raw.String,
			Flags:      attr.Flags.String,
			WhenFailed: attr.WhenFailed,
		}

		if attr.Flags.Prefailure {
			smartAttr.Type = "Pre-fail"
		} else {
			smartAttr.Type = "Old_age"
		}

		if attr.Flags.UpdatedOnline {
			smartAttr.Updated = "Always"
		} else {
			smartAttr.Updated = "Offline"
		}

		drive.Attributes = append(drive.Attributes, smartAttr)

		switch attr.ID {
		case AttrReallocatedSectors:
			drive.ReallocatedSectors = attr.Raw.Value
		case AttrPendingSectors:
			drive.PendingSectors = attr.Raw.Value
		case AttrOfflineUncorrectable:
			drive.UncorrectableSectors = attr.Raw.Value
		case AttrUDMACRCError:
			drive.CRCErrors = attr.Raw.Value
		case AttrTemperature:
			if drive.Temperature == 0 {
				drive.Temperature = int(attr.Raw.Value)
			}
		case AttrWearLevelingCount, AttrMediaWearoutInd:
			if drive.WearLevelingValue < 0 {
				drive.WearLevelingValue = attr.Value
			}
		}

		if isTotalLBAsWrittenAttribute(attr.ID, attr.Name) {
			drive.TotalBytesWritten = attr.Raw.Value * sectorSize(drive)
		}
	}

	// Some firmwares (e.g. HPE-branded Samsung SSDs) omit attribute 241 but
	// still expose the counter in the ATA Device Statistics log (0x04).
	if drive.TotalBytesWritten < 0 {
		if sectors, ok := deviceStatistic(s, devstatGeneralPage, devstatLogicalSectorsWritten); ok {
			drive.TotalBytesWritten = sectors * sectorSize(drive)
		}
	}
}

// ATA Device Statistics: General Statistics page, Logical Sectors Written offset.
const (
	devstatGeneralPage           = 0x01
	devstatLogicalSectorsWritten = 0x18
)

func deviceStatistic(s SmartctlOutput, page, offset int) (int64, bool) {
	for _, p := range s.AtaDeviceStatistics.Pages {
		if p.Number != page {
			continue
		}
		for _, entry := range p.Table {
			if entry.Offset == offset && entry.Flags.Valid {
				return entry.Value, true
			}
		}
	}
	return 0, false
}

func sectorSize(drive *DriveInfo) int64 {
	if drive.LogicalSector == 0 {
		return 512
	}
	return int64(drive.LogicalSector)
}

func isTotalLBAsWrittenAttribute(id int, name string) bool {
	if id == AttrTotalLBAsWritten {
		return true
	}

	normalizedName := strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(name))
	return normalizedName == "totallbaswritten"
}

// updateHealthStatus updates the drive health status based on critical attributes.
func updateHealthStatus(drive *DriveInfo) {
	if drive.IsNVMe {
		if drive.NVMeHealthLog != nil {
			if drive.NVMeHealthLog.CriticalWarning > 0 {
				drive.HealthStatus = HealthBad
				return
			}
			if drive.NVMeHealthLog.MediaErrors > 0 || drive.NVMeHealthLog.PercentageUsed > NVMeHealthPercentageUsedCaution {
				drive.HealthStatus = HealthCaution
				return
			}
		}
	} else {
		if drive.ReallocatedSectors > ReallocatedSectorsBadThreshold || drive.UncorrectableSectors > UncorrectableSectorsBadThreshold {
			drive.HealthStatus = HealthBad
			return
		}
		if drive.ReallocatedSectors > 0 || drive.PendingSectors > 0 ||
			drive.UncorrectableSectors > 0 {
			drive.HealthStatus = HealthCaution
			return
		}
	}
}
