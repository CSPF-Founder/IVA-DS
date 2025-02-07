package models

import (
	"regexp"
	"time"

	"github.com/CSPF-Founder/iva/scanner/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScanInfo struct {
	ScanNumber int       `bson:"scan_number"`
	ScanDate   time.Time `bson:"scan_date"`
}
type Target struct {
	ID               primitive.ObjectID `bson:"_id"`
	CustomerUsername string             `bson:"customer_username"`
	TargetAddress    string             `bson:"target_address"`

	Flag       int                `bson:"flag"`
	ScanStatus enums.TargetStatus `bson:"scan_status"`
	CreatedAt  time.Time          `bson:"createdAt"`
	IsDS       bool               `bson:"is_ds"`
	Scans      []ScanInfo         `bson:"scans"`

	TargetType        enums.TargetType `bson:"target_type"`
	ScannerIP         string           `bson:"scanner_ip"`
	ScannerUsername   string           `bson:"scanner_username"`
	ScanStartedTime   time.Time        `bson:"scan_started_time"`
	ScanCompletedTime time.Time        `bson:"sscan_completed_time"`
	FailureReason     int              `bson:"failure_reason"`

	OverallCVSSScore *float64
	CVSSScoreByHost  map[string]float64
}

func (s *Target) IsIPRange() bool {
	ipRangeRegex := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2}$`)
	return ipRangeRegex.MatchString(s.TargetAddress)
}

// GetNextScanNumber returns the next scan number for the target
func (s *Target) GetNextScanNumber() int {
	if len(s.Scans) == 0 {
		return 1
	}

	bigScanNum := 1
	for _, scan := range s.Scans {
		if scan.ScanNumber > bigScanNum {
			bigScanNum = scan.ScanNumber
		}
	}

	return bigScanNum + 1
}
