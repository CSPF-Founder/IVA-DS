package models

import (
	"github.com/CSPF-Founder/iva/scanner/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Classification struct {
	CVSSScore float64  `json:"cvss_score" bson:"cvss_score"`
	CVEID     []string `json:"cve_id" bson:"cve_id"`

	// CWEID     []string `json:"cwe_id" bson:"cwe_id"`
	// CVSSMetrics string   `json:"cvss_metrics" bson:"cvss_metrics"`
}

type ScanResult struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// Relationship Fields
	CustomerUserName string             `bson:"customer_username"`
	TargetID         primitive.ObjectID `json:"target_id" bson:"target_id"`

	// Parsed fields
	VulnerabilityTitle string         `json:"vulnerability_title" bson:"vulnerability_title"`
	Finding            string         `json:"finding" bson:"finding"`
	Severity           enums.Severity `json:"severity" bson:"severity"`
	Remediation        string         `json:"remediation" bson:"remediation"`
	Reference          []string       `json:"reference" bson:"reference"`
	Evidence           string         `json:"evidence" bson:"evidence"`
	Classification     Classification `json:"classification" bson:"classification"`

	// Other alert fields (optional)
	Cause  string `json:"cause" bson:"cause"`
	Effect string `json:"effect" bson:"effect"`

	// Fields for DS Scans
	ScanNumbers []int               `json:"scan_numbers" bson:"scan_numbers"`
	FoundDate   *primitive.DateTime `json:"found_date" bson:"found_date"`
	FixedDate   *primitive.DateTime `json:"fixed_date" bson:"fixed_date"`
	AlertStatus enums.AlertStatus   `json:"alert_status" bson:"alert_status"`

	// Optional sub-structs for specific scan details
	WSData *WSData `json:"web_details,omitempty" bson:"web_details,omitempty"`
	NSData *NSData `json:"network_details,omitempty" bson:"network_details,omitempty"`
}

type WSData struct {
	// Web alert fields:
	URL       string `json:"url" bson:"url"`
	Param     string `json:"param" bson:"param"`
	Instances string `json:"instances" bson:"instances"`
	Payload   string `json:"payload" bson:"payload"`
	Other     string `json:"other" bson:"other"`
}

type NSData struct {
	// Network alert fields:
	Port           string `json:"port" bson:"port"`
	IP             string `json:"ip" bson:"ip"`
	Hostname       string `json:"hostname" bson:"hostname"`
	Affected       string `json:"affected" bson:"affected"`
	SpecificResult string `json:"specific_result" bson:"specific_result"`
	OID            string `json:"oid" bson:"oid"`
}
