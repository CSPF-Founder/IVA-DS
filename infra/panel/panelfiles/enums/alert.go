package enums

// Type definition for AlertStatus
type AlertStatus int

const (
	AlertUnfixed       AlertStatus = 0
	AlertFixed         AlertStatus = 1
	AlertIgnored       AlertStatus = 2
	AlertFalsePositive AlertStatus = 3
	AlertInvalid       AlertStatus = 4
)

// Type alias with underlying type of IntEnumMap[AlertStatus]
type AlertStatusMapType = IntEnumMap[AlertStatus]

// AlertStatusMap is the map of AlertStatus to string
var AlertStatusMap = AlertStatusMapType{
	AlertUnfixed:       "Unfixed",
	AlertFixed:         "Fixed",
	AlertIgnored:       "Ignored",
	AlertFalsePositive: "FP",
	AlertInvalid:       "Invalid",
}

func AlertStatusToString(alert AlertStatus) string {
	return AlertStatusMap[alert]
}
