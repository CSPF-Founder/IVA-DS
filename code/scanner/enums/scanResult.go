package enums

// Type definition for AlertStatus
type AlertStatus int

const (
	AlertStatusUnfixed AlertStatus = 0
	AlertStatusFixed   AlertStatus = 1
	AlertStatusIgnored AlertStatus = 2
	AlertStatusFP      AlertStatus = 3
	AlertStatusInvalid AlertStatus = 4
)

// Type alias with underlying type of IntEnumMap[AlertStatus]
type AlertStatusMapType = IntEnumMap[AlertStatus]

// AlertStatusMap is the map of AlertStatus to string
var AlertStatusMap = AlertStatusMapType{
	AlertStatusUnfixed: "Unfixed",
	AlertStatusFixed:   "Fixed",
	AlertStatusIgnored: "Ignored",
	AlertStatusFP:      "FP",
	AlertStatusInvalid: "Invalid",
}
