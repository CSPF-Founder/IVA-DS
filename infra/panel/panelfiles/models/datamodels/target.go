package datamodels

import (
	"errors"
	"strings"
	"time"

	"github.com/CSPF-Founder/iva/panel/config"
	"github.com/CSPF-Founder/iva/panel/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Scan represents an individual scan with a ScanID and Timestamp
type ScanInfo struct {
	ScanNumber int                 `bson:"scan_number,omitempty" json:"scan_number,omitempty"`
	ScanDate   *primitive.DateTime `bson:"scan_date,omitempty" json:"scan_date,omitempty"`
}

type Target struct {
	ID                primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	CustomerName      string              `bson:"customer_username"`
	TargetAddress     string              `bson:"target_address"`
	Flag              enums.ScanFlag      `bson:"flag"`
	ScanStatus        enums.TargetStatus  `bson:"scan_status"`
	ScanStartedTime   *primitive.DateTime `bson:"scan_started_time,omitempty"`
	ScanCompletedTime *primitive.DateTime `bson:"scan_completed_time,omitempty"`
	TargetType        enums.TargetType    `bson:"target_type"`
	OverallCVSSScore  float64             `bson:"overall_cvss_score,omitempty"`
	CVSSScoreByHost   map[string]float64  `bson:"cvss_score_by_host,omitempty"`
	IsDS              bool                `bson:"is_ds,omitempty"`
	Scans             []ScanInfo          `bson:"scans,omitempty" json:"scans,omitempty"`
	CreatedAt         primitive.DateTime  `bson:"created_at"`
}

func (t Target) ScanStartedTimeStr() string {
	if t.ScanStartedTime == nil {
		return "-"
	}

	sst := t.ScanStartedTime.Time()
	istLocation, _ := time.LoadLocation("Asia/Kolkata")
	return sst.In(istLocation).Format("02 Jan 06 03:04:05 PM")
}

func (t Target) ScanCompletedTimeStr() string {
	if t.ScanCompletedTime == nil {
		return "-"
	}

	sct := t.ScanCompletedTime.Time()
	istLocation, _ := time.LoadLocation("Asia/Kolkata")
	return sct.In(istLocation).Format("02 Jan 06 03:04:05 PM")
}

func (t Target) GetReportDir() (string, error) {
	// Load the config
	conf, _ := config.LoadConfig()
	if strings.Contains(t.CustomerName, "/") {
		return "", errors.New("Invalid username or target ID")
	}
	return conf.ReportDir + "/" + t.CustomerName + "/" + t.ID.Hex() + "/", nil
}

func (t Target) GetReportPath() (string, error) {
	reportDir, err := t.GetReportDir()
	if err != nil {
		return "", err
	}
	return reportDir + "report.docx", nil
}

func (t Target) GetScanStatusText() string {
	if t.TargetType != enums.TargetTypeURL && t.ScanStatus == enums.TargetStatusUnreachable {
		return "No Open Ports"
	}

	scanStatusText, err := enums.TargetStatusMap.GetText(t.ScanStatus)
	if err != nil {
		return "Unknown"
	}
	return scanStatusText
}

// CanDelete checks if the target can be deleted
// Target can be deleted if it is in one of the following states:
// 1. Scan status is YetToStart, ReportGenerated, ScanFailed or Unreachable
// 2. Scan started time is more than 24 hours ago
func (t Target) CanDelete() bool {
	allowedStatus := map[enums.TargetStatus]bool{
		enums.TargetStatusYetToStart:      true,
		enums.TargetStatusReportGenerated: true,
		enums.TargetStatusScanFailed:      true,
		enums.TargetStatusUnreachable:     true,
	}
	if allowedStatus[t.ScanStatus] {
		return true
	}

	// If scan started time is more than 24 hours ago, then target can be deleted
	if t.ScanStartedTime != nil && time.Since(t.ScanStartedTime.Time()) > 24*time.Hour {
		return true
	}

	return false
}

// CanRescan checks if the target can be rescanned
// Target can be rescanned if it is a DS and in one of the following states:
// 1. if there is only one scans(i.e for 2nd time)
// 2. ScanFailed
// 3. Unreachable
func (t Target) CanRescan() bool {
	if !t.IsDS {
		return false
	}

	if len(t.Scans) == 1 {
		return true
	}

	return t.ScanStatus == enums.TargetStatusScanFailed || t.ScanStatus == enums.TargetStatusUnreachable
}

// Mark as main button enable
// if the status is ReportGenerated, ScanFailed or Unreachable
func (t Target) CanMarkAsMain() bool {
	if !t.IsDS {
		return false
	}

	return t.ScanStatus == enums.TargetStatusReportGenerated || t.ScanStatus == enums.TargetStatusScanFailed || t.ScanStatus == enums.TargetStatusUnreachable
}
