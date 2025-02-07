package network

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/CSPF-Founder/iva/scanner/enums"
	"github.com/CSPF-Founder/iva/scanner/internal/repositories"
	"github.com/CSPF-Founder/iva/scanner/logger"
	"github.com/CSPF-Founder/iva/scanner/models"
	"github.com/CSPF-Founder/iva/scanner/utils"
	"github.com/CSPF-Founder/libs/gmp"
)

// ResultsHandler is responsible for handling scan results.
type resultsHandler struct {
	logger      *logger.Logger
	db          *repositories.Repository
	reporterBin string
	target      models.Target
}

// NewResultsHandler initializes and returns a new ResultsHandler.
func NewResultsHandler(
	logger logger.Logger,
	db *repositories.Repository,
	target models.Target,
	reporterBin string,
) *resultsHandler {
	return &resultsHandler{
		logger:      &logger,
		db:          db,
		target:      target,
		reporterBin: reporterBin,
	}
}

func (rh *resultsHandler) run(ctx context.Context, results []gmp.Alert) error {
	if rh.target.IsDS {
		h := NewDSHandler(rh.db, rh.logger, rh.target)
		err := h.handle(ctx, results)
		if err != nil {
			return err
		}
	} else {
		err := rh.AddToDB(ctx, results)
		if err != nil {
			return err
		}
	}

	// call reporter
	rh.logger.Info("Making Report")
	scriptArgs := []string{"-t", rh.target.ID.Hex()}
	cmd := exec.CommandContext(ctx, rh.reporterBin, scriptArgs...)

	// capture the stdout and stderr
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	// run the command
	err := cmd.Run()
	if err != nil {
		errMsg := ""
		errMsg = err.Error()

		if stdOut.String() != "" {
			errMsg = errMsg + "\n" + stdOut.String()
		}

		if stdErr.String() != "" {
			errMsg = errMsg + "\n" + stdErr.String()
		}
		return errors.New(errMsg)
	}

	if stdErr.Len() > 0 {
		return fmt.Errorf("reporter script returned an error: %s", stdErr.String())
	}

	return nil
}

// addToDB adds scan results to the database.
func (rh *resultsHandler) AddToDB(ctx context.Context, results []gmp.Alert) error {
	if len(results) == 0 {
		rh.logger.Info("No results to add to the database")
		return nil
	}

	records := make([]models.ScanResult, 0, len(results))

	dupMap := make(mappedResults)

	for _, entry := range results {
		record, err := parseRecord(entry, rh.target)

		if err != nil {
			rh.logger.Error("Error converting entry to record", err)
			continue
		}

		// Duplicate check
		cmpKey := prepDSCmpFields(
			record.VulnerabilityTitle,
			record.NSData.IP,
			record.NSData.Hostname,
			record.NSData.Port,
		)
		if _, ok := dupMap[cmpKey]; ok {
			continue
		}
		dupMap[cmpKey] = mappedResultData{
			// No need to fill the fields, since we are not going to use it
		}
		records = append(records, *record)
	}
	if len(records) == 0 {
		rh.logger.Warn("No records to add to the database", nil)
		return nil
	}

	_, err := rh.db.ScanResult.AddList(ctx, records)
	if err != nil {
		return err
	}
	return nil
}

// parseRecord converts an input record to an output record
func parseRecord(input gmp.Alert, target models.Target) (*models.ScanResult, error) {

	if input.NVTName == "" {
		return nil, errors.New("error converting entry to record: missing or invalid 'name' field")
	}

	port := input.Port
	portProto := input.PortProto
	if port != "" && portProto != "" {
		port = port + "/" + portProto
	}

	cve_ids, err := UnmarshalCVEList(input.CVEDetails)
	if err != nil {
		cve_ids = []string{}
	}

	return &models.ScanResult{
		VulnerabilityTitle: input.NVTName,
		TargetID:           target.ID,
		CustomerUserName:   target.CustomerUsername,
		Severity:           enums.SeverityFromString(input.Severity),
		Finding:            input.Summary,
		Remediation:        input.Solution,
		Reference:          utils.SplitStrIntoSlice(input.OtherReferences, ","),
		Cause:              input.VulnerabilityInsight,
		Effect:             input.Impact,
		Classification: models.Classification{
			CVSSScore: input.CVSS,
			CVEID:     cve_ids,
		},
		NSData: &models.NSData{
			IP:             input.IP,
			Port:           port,
			Hostname:       input.Hostname,
			Affected:       input.AffectedSoftwareOrOS,
			SpecificResult: input.SpecificResult,
			OID:            input.NVTOID,
		},
	}, nil
}

// UnmarshalCVEList takes a JSON string and unmarshals it into a slice of strings.
func UnmarshalCVEList(data string) ([]string, error) {
	var cves []string

	if data == "NOCVE" {
		return cves, nil
	}

	err := json.Unmarshal([]byte(data), &cves)
	if err != nil {
		return nil, err
	}
	return cves, nil
}
