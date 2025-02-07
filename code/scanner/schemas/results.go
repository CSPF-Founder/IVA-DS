package schemas

import (
	"github.com/CSPF-Founder/iva/scanner/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// // schema for ZAP results
// type InputRecord struct {
// 	// Info               InputInfo `json:"info"`
// 	Risk               string   `json:"risk"`
// 	Alert              string   `json:"alert"`
// 	Description        string   `json:"description"`
// 	Remediation        string   `json:"solution"`
// 	WascID             string   `json:"wascid"`
// 	CweID              string   `json:"cweid"`
// 	MessageID          string   `json:"message_id"`
// 	Finding            string   `csv:"summary"`
// 	VulnerabilityTitle string   `json:"vulnerability_title"`
// 	ScanType           string   `json:"scan_type"`
// 	URL                string   `json:"url"`
// 	Reference          []string `json:"reference"`
// 	Evidence           string   `json:"evidence"`
// 	Payload            string   `json:"attack"`
// 	Other              string   `json:"other"`
// 	Instances          string   `json:"instances"`
// 	Param              string   `json:"param"`
// }

// schema for Openvas results
type Alert struct {
	Severity           enums.Severity     `csv:"severity"`
	VulnerabilityTitle string             `csv:"vulnerability_title"`
	Port               string             `csv:"port"`
	IP                 string             `csv:"ip"`
	Hostname           string             `csv:"hostname"`
	CVSs               string             `csv:"cvss"`
	Finding            string             `csv:"summary"`
	Cause              string             `csv:"cause"`
	Effect             string             `csv:"effect"`
	Remediation        string             `csv:"remediation"`
	Reference          string             `csv:"reference"`
	Affected           string             `csv:"affected"`
	CVEDetails         string             `csv:"cve_details"`
	SpecificResult     string             `csv:"specific_result"`
	TargetID           primitive.ObjectID `csv:"target_id"`
}
