package datamodels

import (
	"github.com/CSPF-Founder/iva/panel/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Classification struct {
	CVSSScore float64  `json:"cvss_score" bson:"cvss_score"`
	CVEID     []string `json:"cve_id" bson:"cve_id"`

	// CWEID     []string `json:"cwe_id" bson:"cwe_id"`
	CVSSMetrics string `json:"cvss_metrics" bson:"cvss_metrics"`
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
	FoundDate   *primitive.DateTime `json:"found_date,omitempty" bson:"found_date,omitempty"`
	FixedDate   *primitive.DateTime `json:"fixed_date,omitempty" bson:"fixed_date,omitempty"`
	AlertStatus enums.AlertStatus   `json:"alert_status" bson:"alert_status"`

	// Optional sub-structs for specific scan details
	WSData *WSData `json:"web_details,omitempty" bson:"web_details,omitempty"`
	NSData *NSData `json:"network_details,omitempty" bson:"network_details,omitempty"`

	// Non DB fields
	FormatedFoundDate string `gorm:"-"`
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
	CVSS           string `json:"cvss" bson:"cvss"`
	OID            string `json:"oid" bson:"oid"`
}

// GetPOC returns the Proof of Concept for the scan result
//
// nolint:cyclop
func (r ScanResult) GetPOC() string {
	var poc string

	if r.Evidence != "" {
		poc = "Evidence: " + r.Evidence + "\n"
	}
	if r.WSData != nil {
		if r.WSData.URL != "" {
			poc += "URL: " + r.WSData.URL + "\n\n"
		}

		if r.WSData.Param != "" {
			poc += "Parameters: " + r.WSData.Param + "\n\n"
		}

		if r.WSData.Payload != "" {
			poc += "Payload: " + r.WSData.Payload + "\n\n"
		}

		if r.WSData.Other != "" {
			poc += "Other: " + r.WSData.Other
		}

		if r.WSData.Instances != "" {
			poc += "Instances: " + r.WSData.Instances + "\n\n"
		}

	} else if r.NSData != nil {
		if r.NSData.IP != "" {
			poc += "IP: " + r.NSData.IP + "\n\n"
		}

		if r.NSData.Port != "" {
			poc += "Port: " + r.NSData.Port + "\n\n"
		}

		if r.NSData.SpecificResult != "" {
			poc += "Additional Details: \n" + r.NSData.SpecificResult + "\n\n"
		}

	}

	return poc
}

func (r ScanResult) GetDetailsAndImpact() string {
	var info string

	if r.Finding != "" {
		info = "Finding: " + r.Finding + "\n"
	}

	if r.Cause != "" {
		info = "Cause: " + r.Cause + "\n"
	}

	if r.Effect != "" {
		info += "Effect: " + r.Effect + "\n"
	}

	return info

}
