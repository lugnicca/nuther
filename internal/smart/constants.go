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

// ConnectionAttributes contains IDs of attributes that indicate cable/interface issues, not drive defects
var ConnectionAttributes = map[int]bool{
	AttrCommandTimeout: true,
	AttrUDMACRCError:   true,
}

