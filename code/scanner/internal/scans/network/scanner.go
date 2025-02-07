package network

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CSPF-Founder/iva/scanner/config"
	"github.com/CSPF-Founder/iva/scanner/enums"
	"github.com/CSPF-Founder/iva/scanner/internal/repositories"
	"github.com/CSPF-Founder/iva/scanner/logger"
	"github.com/CSPF-Founder/iva/scanner/models"
	"github.com/CSPF-Founder/iva/scanner/utils"
	"github.com/CSPF-Founder/libs/gmp"
	"github.com/CSPF-Founder/libs/gmp/protocol"
)

type scanner struct {
	logger              *logger.Logger
	db                  *repositories.Repository
	conf                *config.Config
	openvasConfig       protocol.OpenvasConfig
	target              models.Target
	totalWaitTime       time.Duration
	resultCheckInterval time.Duration
}

func NewNetworkScanner(
	target models.Target,
	logger *logger.Logger,
	conf *config.Config,
	db *repositories.Repository,
) *scanner {
	return &scanner{
		logger:              logger,
		target:              target,
		db:                  db,
		openvasConfig:       conf.OpenvasConfig,
		conf:                conf,
		totalWaitTime:       config.NSTimeout,
		resultCheckInterval: config.NSPollInterval,
	}
}

func (s *scanner) ScanFailed(ctx context.Context, scanStatus enums.TargetStatus) bool {
	s.target.ScanStatus = scanStatus
	isUpdated, err := s.db.Target.UpdateScanStatus(ctx, &s.target)
	if err != nil {
		s.logger.Error("Failed to update Scan Status in Scan Failed", err)
	}
	return isUpdated
}

func (s *scanner) UpdateScanStatus(ctx context.Context, status enums.TargetStatus) bool {
	s.target.ScanStatus = status
	isUpdated, err := s.db.Target.UpdateScanStatus(ctx, &s.target)
	if err != nil {
		s.logger.Error("Failed to update Scan Status", err)
	}
	return isUpdated
}

func (s *scanner) CalculateScanTimeout() error {
	ipCount := utils.GetIPCountIfRange(s.target.TargetAddress)

	if ipCount > 256 {
		return errors.New("ip count is more than 256")
	}
	if ipCount > 1 {
		// x minute per IP
		s.totalWaitTime += config.IPRangeTimeoutPerIP * time.Duration(ipCount)
		// Result poll interval is 1 minute
		s.resultCheckInterval = config.IPRangePollInterval
	}

	// Ensure the total wait time does not exceed the maximum allowed timeout
	if s.totalWaitTime > config.IPRangeMaxTimeout {
		s.totalWaitTime = config.IPRangeMaxTimeout
	}
	return nil
}

func (s *scanner) Run(ctx context.Context) {
	if !s.UpdateScanStatus(ctx, enums.TargetStatusScanStarted) {
		return
	}

	err := s.CalculateScanTimeout()
	if err != nil {
		s.logger.Error("Unable to calculate scan timeout", err)
		return
	}

	results, err := s.runScan(ctx)
	if err != nil {
		s.logger.Error("Failed to run scan", err)
		_ = s.UpdateScanStatus(ctx, enums.TargetStatusScanFailed)
		return
	}

	if !s.UpdateScanStatus(ctx, enums.TargetStatusScanRetrieved) {
		return
	}

	handler := NewResultsHandler(*s.logger, s.db, s.target, s.conf.ReporterBin)

	if err := handler.run(ctx, results); err != nil {
		s.logger.Error("Failed to run handler", err)
		_ = s.UpdateScanStatus(ctx, enums.TargetStatusScanFailed)
		return
	}
}

func (s *scanner) runScan(ctx context.Context) ([]gmp.Alert, error) {
	s.logger.Info(fmt.Sprintf("Starting scan for target : %s", string(s.target.TargetAddress)))

	gmp, err := gmp.NewGMP(
		s.logger,
		s.openvasConfig,
	)
	if err != nil {
		return nil, err
	}

	scanData, err := gmp.StartScan(ctx, s.target.TargetAddress, s.target.ID.Hex())
	if err != nil {
		s.logger.Error("Error adding. Doing cleanup", err)
		return nil, err
	}

	defer func() {
		if err := gmp.CleanUpScan(ctx, *scanData); err != nil {
			s.logger.Error("error cleaning up scan", err)
		}
	}()

	return s.pollScan(ctx, gmp, scanData)
}

// pollScan polls the scan status and returns the results
func (s *scanner) pollScan(
	ctx context.Context,
	gmp *gmp.API,
	scanData *gmp.ScanData,
) ([]gmp.Alert, error) {
	endTime := time.Now().Add(s.totalWaitTime)

	err := s.waitTillScanCompleted(ctx, gmp, scanData.TaskID, endTime)
	if err != nil {
		return nil, err
	}

	// Get the scan results
	results, err := gmp.GetResults(ctx, scanData.TaskID, scanData.ReportID)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// waitTillScanCompleted waits for the scan to complete
func (s *scanner) waitTillScanCompleted(ctx context.Context, gmp *gmp.API, taskID string, endTime time.Time) error {
	statusAttempt := 0

	for time.Now().Before(endTime) {
		isCompleted, err := gmp.IsScanCompleted(ctx, taskID)
		if err != nil {
			if statusAttempt >= 3 {
				return err
			}

			// Retry getting the scan status
			statusAttempt++
			s.logger.Error("Error getting scan status. Retrying", err)
			if err := utils.SleepContext(ctx, 30*time.Second); err != nil {
				return err
			}
			continue
		}

		// If the scan is completed, break the loop
		if isCompleted {
			s.logger.Info(fmt.Sprintf("Scan Completed for target: %s\n", s.target.TargetAddress))
			return nil
		}

		// if the scan is not completed, wait for the result check interval
		if err := utils.SleepContext(ctx, s.resultCheckInterval); err != nil {
			return err
		}
	}

	s.logger.Info(fmt.Sprintf("Partial retrieval for %s\n", s.target.TargetAddress))
	return nil
}

// IsScannerReady checks if the scanner is alive
func IsScannerReady(ctx context.Context, conf *config.Config, lgr *logger.Logger) bool {
	gmp, err := gmp.NewGMP(lgr, conf.OpenvasConfig)
	if err != nil {
		return false
	}

	oVer, err := gmp.OV.GetVersion(ctx)
	if err != nil {
		return false
	}

	if oVer == "" {
		return false
	}

	return true
}
