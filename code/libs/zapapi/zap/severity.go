package zap

import "strings"

type Severity int

const (
	SeverityCritical Severity = 1
	SeverityHigh     Severity = 2
	SeverityMedium   Severity = 3
	SeverityLow      Severity = 4
	SeverityInfo     Severity = 5
)

func SeverityFromString(severity string) Severity {
	switch strings.ToLower(severity) {
	case "critical":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	case "low":
		return SeverityLow
	default:
		return SeverityInfo
	}
}
