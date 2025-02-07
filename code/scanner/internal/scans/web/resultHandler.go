package web

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/CSPF-Founder/iva/scanner/enums"
	"github.com/CSPF-Founder/iva/scanner/internal/repositories"
	"github.com/CSPF-Founder/iva/scanner/logger"
	"github.com/CSPF-Founder/iva/scanner/models"
	"github.com/CSPF-Founder/iva/scanner/utils"
	"github.com/CSPF-Founder/libs/zapapi/zap"
)

// ResultsHandler is responsible for handling scan results.
type resultsHandler struct {
	logger      *logger.Logger
	reporterBin string
	target      models.Target
	db          *repositories.Repository
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
		reporterBin: reporterBin,
		target:      target,
	}
}

// Run processes the results and makes a report.
func (rh *resultsHandler) run(ctx context.Context, results []zap.Alert) error {

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

	// call Reporter
	rh.logger.Info("Making Report")
	scriptArgs := []string{"-t", rh.target.ID.Hex()}
	cmd := exec.CommandContext(ctx, rh.reporterBin, scriptArgs...)

	// capture the stdout and stderr
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	// Run the command
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
func (rh *resultsHandler) AddToDB(ctx context.Context, results []zap.Alert) error {
	if len(results) == 0 {
		rh.logger.Info("No records to add to the database")
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
		cmpKey := prepDSCmpFields(record.VulnerabilityTitle, record.WSData.URL, record.WSData.Param)
		if _, ok := dupMap[cmpKey]; ok {
			continue
		}
		dupMap[cmpKey] = mappedResultData{
			// No need to fill the fields, since we are not going to use it
		}
		records = append(records, *record)
	}

	if len(records) == 0 {
		rh.logger.Error("No records to add to the database", nil)
		return nil
	}

	_, err := rh.db.ScanResult.AddList(ctx, records)
	if err != nil {
		return err
	}
	return nil
}

// parseRecord converts an input record to an output record
func parseRecord(input zap.Alert, target models.Target) (*models.ScanResult, error) {

	if input.Title == "" {
		return nil, errors.New("error converting entry to record: missing or invalid 'name' field")
	}

	other := input.Other
	if other != "" {
		other = strings.ReplaceAll(other, "\r\n\r\n", "\n\n")
		other = strings.ReplaceAll(other, "Att ", "\nAttribute ")
	}
	severity := enums.SeverityFromString(input.Risk)
	cvssScore := utils.CalculateCVSSBySeverity(severity)

	references := utils.SplitStrIntoSlice(input.Reference, "\n")

	return &models.ScanResult{
		CustomerUserName: target.CustomerUsername,
		TargetID:         target.ID,

		VulnerabilityTitle: input.Title,
		Finding:            input.Description,
		Severity:           severity,
		Remediation:        input.Solution,
		Reference:          references,
		Classification:     models.Classification{CVSSScore: cvssScore},
		WSData: &models.WSData{
			URL:       input.URL,
			Param:     input.Param,
			Payload:   input.Payload,
			Other:     other,
			Instances: input.Instances,
		},
	}, nil
}
