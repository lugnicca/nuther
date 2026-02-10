package smart

// Thresholds for health status evaluation
// These constants define the boundaries between GOOD, CAUTION, and BAD states

// Temperature thresholds (Celsius)
//
// Single scale used for both UI color coding and health evaluation:
//
//	< TempCool     → Info (cool/blue)
//	< TempGoodMax  → Success (green)          ← health = GOOD
//	< TempCautionMax → Warning (yellow/orange) ← health = CAUTION
//	< TempCritical → Hot (deep orange)         ← health = BAD
//	≥ TempCritical → Danger (red)              ← health = BAD
const (
	// TempCool is the upper bound for "cool" display color
	TempCool = 35
	// TempGoodMax is the maximum temperature considered "good"
	TempGoodMax = 45
	// TempCautionMax is the maximum temperature before critical
	TempCautionMax = 55
	// TempCritical is the temperature above which the drive is in danger
	TempCritical = 65
)

// NVMe-specific thresholds
//
// There are two sets of percentage-used thresholds:
//   - NVMePercentageUsedCaution/Bad (80/95): used for the per-attribute display
//     status in GetUsageStatus(). Warns early so the user notices wear.
//   - NVMeHealthPercentageUsedCaution (90): used for the overall drive health
//     evaluation in updateHealthStatus(). Triggers later because the global
//     health badge should only turn yellow when wear is genuinely significant.
const (
	// NVMePercentageUsedCaution is the display threshold for percentage used warning (attribute level)
	NVMePercentageUsedCaution = 80
	// NVMePercentageUsedBad is the display threshold for percentage used critical (attribute level)
	NVMePercentageUsedBad = 95
	// NVMeHealthPercentageUsedCaution is the threshold for overall drive health status
	NVMeHealthPercentageUsedCaution = 90

	// NVMeSpareMargin is the buffer above threshold before showing CAUTION
	// If spare > threshold + margin = GOOD, else CAUTION
	NVMeSpareMargin = 10
)

// Unsafe shutdown thresholds
const (
	// UnsafeShutdownGoodMax is the maximum unsafe shutdowns considered "good"
	UnsafeShutdownGoodMax = 50
	// UnsafeShutdownCautionMax is the maximum unsafe shutdowns before critical
	UnsafeShutdownCautionMax = 200
)

// Error count thresholds
const (
	// ErrorCountGoodMax is the maximum error count considered "good" (0 = good)
	ErrorCountGoodMax = 0
	// ErrorCountCautionMax is the maximum error count before critical
	ErrorCountCautionMax = 10
)

// ATA/SATA critical attribute thresholds
const (
	// CriticalAttrBadThreshold is the raw value above which a critical attribute is BAD
	// At or below this, a non-zero value shows CAUTION
	CriticalAttrBadThreshold = 100

	// ReallocatedSectorsBadThreshold is the count above which drive health is BAD
	ReallocatedSectorsBadThreshold = 100
	// UncorrectableSectorsBadThreshold is the count above which drive health is BAD
	UncorrectableSectorsBadThreshold = 100
)

// Attribute value thresholds (normalized values)
const (
	// AttrValueMargin is the buffer above threshold before showing CAUTION
	// If value > threshold + margin = GOOD, else CAUTION
	AttrValueMargin = 10
)
