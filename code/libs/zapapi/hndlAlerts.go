package zapapi

import (
	"regexp"
	"strings"

	"github.com/CSPF-Founder/libs/zapapi/zap"
)

var GENERIC_ALERTS_TO_GROUP = []string{
	"Content Security Policy (CSP) Header Not Set",
	"Missing Anti-clickjacking Header",
	"X-Content-Type-Options Header Missing",
	"Absence of Anti-CSRF Tokens",
	`Server Leaks Version Information via "Server" HTTP Response Header Field`,
	"Cookie No HttpOnly Flag",
	"Cookie without SameSite Attribute",
	"Loosely Scoped Cookie",
	"X-Frame-Options Setting Malformed",
	"X-Frame-Options Defined via META (Non-compliant with Spec)",
	"Multiple X-Frame-Options Header Entries",
	"Cookie Without Secure Flag",
	"Re-examine Cache-control Directives",
	"Cross-Domain JavaScript Source File Inclusion",
	"Content-Type Header Missing",
	"Anti-clickjacking Header",
	"Cookie Poisoning",
	"Emails Found in the Viewstate",
	"Old Asp.Net Version in Use",
	"Viewstate without MAC Signature (Unsure)",
	"Viewstate without MAC Signature (Sure)",
	"Split Viewstate in Use",
	"CSP",
	"X-Debug-Token Information Leak",
	"Session ID in URL Rewrite",
	"Charset Mismatch (Header Versus Meta Content-Type Charset)",
	"Session Management Response Identified",
}

var GENERIC_ALERTS_TO_GROUP_REGEX = []*regexp.Regexp{
	regexp.MustCompile(`CSP: .*`),
}

var ALERTS_TO_IGNORE = []string{
	"Modern Web Application",
	"Buffer Overflow",
}

func isAlertIgnored(alertTitle string, alertsToIgnore []string) bool {
	for _, alert := range alertsToIgnore {
		if alert == alertTitle {
			return true
		}
	}
	return false
}

// helper function to check if alert is in slice
func isInSlice(alertTitle string, alerts []string) bool {
	for _, alert := range alerts {
		if alert == alertTitle {
			return true
		}
	}
	return false
}

// function to check if alert is in group (using exact and regex match)
func isAlertInGroup(alertTitle string) bool {
	// checking for exact match in GENERIC_ALERTS_TO_GROUP
	if isInSlice(alertTitle, GENERIC_ALERTS_TO_GROUP) {
		return true
	}

	// checking for regex match in GENERIC_ALERTS_TO_GROUP_REGEX
	for _, alertPattern := range GENERIC_ALERTS_TO_GROUP_REGEX {
		if alertPattern.MatchString(alertTitle) {
			return true
		}
	}
	return false
}

func (z *API) filterAlerts(alerts []zap.Alert) []zap.Alert {
	filteredAlerts := []zap.Alert{}
	groupedAlerts := make(map[string]any)

	for _, alert := range alerts {
		alertTitle := alert.Title
		if alertTitle == "" || isAlertIgnored(alertTitle, ALERTS_TO_IGNORE) {
			continue
		}

		severity := zap.SeverityFromString(alert.Risk)
		if (severity == zap.SeverityInfo) && strings.Contains(alertTitle, "Fuzzer") {
			// Skip fuzzer and scanner alerts
			continue
		}

		// Update grouped alerts
		if isAlertInGroup(alertTitle) {
			groupedAlerts[alertTitle] = alert
			continue
		}

		alert.ScanType = zap.ScanTypeActive

		filteredAlerts = append(filteredAlerts, alert)
	}

	return filteredAlerts
}

type SortedAlerts struct {
	Alerts       []zap.Alert
	Distribution map[string]int
}

var severityOrder = []string{"Critical", "High", "Medium", "Low", "Information"}
var severityIndex = map[string]int{
	"Critical":    0,
	"High":        1,
	"Medium":      2,
	"Low":         3,
	"Information": 4,
}

// sortAlerts sorts the alerts by severity and returns the distribution of alerts
func sortAlerts(alerts []zap.Alert) *SortedAlerts {
	groupedAlerts := make(map[string][]zap.Alert)
	for _, alert := range alerts {
		severity := alert.Risk
		//check if severity is valid
		if _, ok := severityIndex[severity]; !ok {
			// Skip invalid severity
			continue
		}

		groupedAlerts[severity] = append(groupedAlerts[severity], alert)
	}

	// Create the sorted alerts list and distribution map
	sortedAlertsList := []zap.Alert{}
	distributionOfAlerts := map[string]int{}

	for _, severity := range severityOrder {
		if alerts, ok := groupedAlerts[severity]; ok {
			sortedAlertsList = append(sortedAlertsList, alerts...)
			distributionOfAlerts[severity] = len(alerts)
		}
	}

	return &SortedAlerts{
		Alerts:       sortedAlertsList,
		Distribution: distributionOfAlerts,
	}
}
