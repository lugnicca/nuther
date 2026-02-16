package smart

// Critical SMART Attribute IDs
const (
	AttrRawReadErrorRate     = 1
	AttrSpinUpTime           = 3
	AttrStartStopCount       = 4
	AttrReallocatedSectors   = 5
	AttrSeekErrorRate        = 7
	AttrPowerOnHours         = 9
	AttrSpinRetryCount       = 10
	AttrPowerCycleCount      = 12
	AttrSoftReadErrorRate    = 13
	AttrEndToEndError        = 184
	AttrReportedUncorrect    = 187
	AttrCommandTimeout       = 188
	AttrHighFlyWrites        = 189
	AttrAirflowTemp          = 190
	AttrGSenseErrorRate      = 191
	AttrPowerOffRetract      = 192
	AttrLoadCycleCount       = 193
	AttrTemperature          = 194
	AttrHardwareECC          = 195
	AttrReallocEventCount    = 196
	AttrPendingSectors       = 197
	AttrOfflineUncorrectable = 198
	AttrUDMACRCError         = 199
	AttrWriteErrorRate       = 200
	AttrSoftReadErrorRate2   = 201
	AttrTAIncreaseCount      = 202
	AttrWearLevelingCount    = 177
	AttrMediaWearoutInd      = 233
	AttrTotalLBAsWritten     = 241
	AttrTotalLBAsRead        = 242
)

// CriticalAttributes contains the IDs of attributes that indicate drive health issues
var CriticalAttributes = map[int]bool{
	AttrReallocatedSectors:   true,
	AttrReportedUncorrect:    true,
	AttrPendingSectors:       true,
	AttrOfflineUncorrectable: true,
}

// AttributeDescriptions provides human-readable descriptions for known SMART attributes
var AttributeDescriptions = map[int]string{
	AttrRawReadErrorRate:     "Rate of hardware read errors; vendor-specific, high values may be normal",
	AttrSpinUpTime:           "Average time to spin up the drive (milliseconds)",
	AttrStartStopCount:       "Total count of start/stop cycles",
	AttrReallocatedSectors:   "Count of bad sectors remapped to spare area",
	AttrSeekErrorRate:        "Rate of seek errors; vendor-specific, high values may be normal",
	AttrPowerOnHours:         "Total hours the drive has been powered on",
	AttrSpinRetryCount:       "Count of spin-up retry attempts",
	AttrPowerCycleCount:      "Total count of power on/off cycles",
	AttrSoftReadErrorRate:    "Rate of off-track soft read errors",
	AttrEndToEndError:        "End-to-end data path error count",
	AttrReportedUncorrect:    "Count of uncorrectable errors reported to the host",
	AttrCommandTimeout:       "Commands that timed out; may indicate cable or connection issues",
	AttrHighFlyWrites:        "Count of high fly write events (head too far from platter)",
	AttrAirflowTemp:          "Airflow temperature (may differ from drive temperature)",
	AttrGSenseErrorRate:      "Count of shock/vibration events detected by accelerometer",
	AttrPowerOffRetract:      "Count of emergency head retracts during power loss",
	AttrLoadCycleCount:       "Count of head load/unload cycles",
	AttrTemperature:          "Current drive temperature in Celsius",
	AttrHardwareECC:          "Count of errors corrected by hardware ECC",
	AttrReallocEventCount:    "Count of sector reallocation events",
	AttrPendingSectors:       "Sectors waiting to be remapped (unstable sectors)",
	AttrOfflineUncorrectable: "Sectors that could not be corrected during offline scan",
	AttrUDMACRCError:         "CRC errors during data transfer; usually a cable or connector issue",
	AttrWriteErrorRate:       "Rate of errors during write operations",
	AttrSoftReadErrorRate2:   "Rate of off-track soft read errors (vendor alternate)",
	AttrTAIncreaseCount:      "Count of thermal throttling events",
	AttrWearLevelingCount:    "Remaining SSD write endurance (100=new, 0=worn out)",
	AttrMediaWearoutInd:      "SSD media wear indicator (100=new, 0=worn out)",
	AttrTotalLBAsWritten:     "Total logical blocks written to the drive",
	AttrTotalLBAsRead:        "Total logical blocks read from the drive",
}

// ConnectionAttributes contains IDs of attributes that indicate cable/interface issues, not drive defects
var ConnectionAttributes = map[int]bool{
	AttrCommandTimeout: true,
	AttrUDMACRCError:   true,
}

