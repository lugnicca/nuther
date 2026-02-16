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
// Returns demo data with a non-nil error when no real drives are detected.
func ScanDrives() ([]DriveInfo, error) {
	drives := scanWithSmartctl()
	if len(drives) > 0 {
		return drives, nil
	}

	drives = scanCommonDevices()
	if len(drives) > 0 {
		return drives, nil
	}

	return CreateDemoData(), fmt.Errorf("no drives detected, using demo data")
}

// scanWithSmartctl uses smartctl --scan to discover drives
func scanWithSmartctl() []DriveInfo {
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
				} else if drive.Model != "" {
					drives = append(drives, drive)
				}
			}
			if len(drives) > 0 {
				slog.Info("scan complete", "drives", len(drives))
				return drives
			}
		}
	}

	output, err = runSmartctl("--scan")
	if err != nil {
		slog.Debug("smartctl --scan failed", "error", err)
		return nil
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
			} else if drive.Model != "" {
				drives = append(drives, drive)
			}
		}
	}

	return drives
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

	if len(drives) > 0 {
		slog.Info("scan complete", "drives", len(drives))
	}
	return drives
}

// GetDriveInfo retrieves detailed information for a specific device
func GetDriveInfo(device string) (DriveInfo, error) {
	return GetDriveInfoWithType(device, "")
}

// GetDriveInfoWithType retrieves detailed information for a specific device with optional type
func GetDriveInfoWithType(device string, deviceType string) (DriveInfo, error) {
	var output []byte
	var err error

	if deviceType != "" && deviceType != "scsi" {
		output, err = runSmartctl("-a", "-j", "-d", deviceType, device)
	} else {
		output, err = runSmartctl("-a", "-j", device)
	}

	if err != nil {
		if deviceType == "scsi" || deviceType == "" {
			usbTypes := []string{"sat", "sat,auto", "usbsunplus", "usbjmicron", "usbcypress"}
			for _, usbType := range usbTypes {
				output, err = runSmartctl("-a", "-j", "-d", usbType, device)
				if err == nil {
					break
				}
			}
			if err != nil {
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

	return parseDriveInfo(device, smartOutput), nil
}

// parseDriveInfo converts smartctl JSON output to a DriveInfo struct
func parseDriveInfo(device string, s SmartctlOutput) DriveInfo {
	drive := DriveInfo{
		Device:         device,
		Model:          s.ModelName,
		ModelFamily:    s.ModelFamily,
		Serial:         s.SerialNumber,
		Firmware:       s.FirmwareVersion,
		CapacityBytes:  s.UserCapacity.Bytes,
		Capacity:       FormatBytes(s.UserCapacity.Bytes),
		LogicalSector:  s.LogicalBlockSize,
		PhysicalSector: s.PhysicalBlockSize,
		FormFactor:     s.FormFactor.Name,
		RotationRate:   s.RotationRate,
		Temperature:    s.Temperature.Current,
		PowerOnHours:   s.PowerOnTime.Hours,
		PowerCycles:    s.PowerCycleCount,
		SmartSupported: true,
		SmartEnabled:   true,
		HealthPassed:   s.SmartStatus.Passed,
		LastUpdate:     time.Now(),
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
		}
	}
}

// updateHealthStatus updates the drive health status based on critical attributes
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
