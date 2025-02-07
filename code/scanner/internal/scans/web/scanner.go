package web

import (
	"context"
	"time"

	"github.com/CSPF-Founder/iva/scanner/config"
	"github.com/CSPF-Founder/iva/scanner/enums"
	"github.com/CSPF-Founder/iva/scanner/internal/repositories"
	"github.com/CSPF-Founder/iva/scanner/logger"
	"github.com/CSPF-Founder/iva/scanner/models"
	"github.com/CSPF-Founder/iva/scanner/utils"
	"github.com/CSPF-Founder/libs/zapapi"
	"github.com/CSPF-Founder/libs/zapapi/zap"
)

type scanner struct {
	logger           *logger.Logger
	db               *repositories.Repository
	conf             *config.Config
	target           models.Target
	scanFailedStatus enums.TargetStatus
}

func NewWebScanner(
	target models.Target,
	logger *logger.Logger,
	conf *config.Config,
	db *repositories.Repository,
) *scanner {
	return &scanner{
		logger:           logger,
		target:           target,
		db:               db,
		conf:             conf,
		scanFailedStatus: enums.TargetStatusScanFailed,
	}
}

func (s *scanner) UpdateScanStatus(ctx context.Context, status enums.TargetStatus) bool {
	s.target.ScanStatus = status
	isUpdated, err := s.db.Target.UpdateScanStatus(ctx, &s.target)
	if err != nil {
		s.logger.Error("Failed to update Scan Status", err)
	}
	return isUpdated
}

func (s *scanner) Run(ctx context.Context) {
	if !s.UpdateScanStatus(ctx, enums.TargetStatusScanStarted) {
		return
	}

	results, err := s.runScan(ctx)
	if err != nil {
		s.logger.Error("Failed to run scan", err)
		_ = s.UpdateScanStatus(ctx, s.scanFailedStatus)
		return
	}

	if !s.UpdateScanStatus(ctx, enums.TargetStatusScanRetrieved) {
		s.logger.Error("Failed to update Scan Status", nil)
		return
	}

	handler := NewResultsHandler(*s.logger, s.db, s.target, s.conf.ReporterBin)
	if err := handler.run(ctx, results); err != nil {
		s.logger.Error("Failed to handle scan results", err)
		_ = s.UpdateScanStatus(ctx, s.scanFailedStatus)
		return
	}
}

func (s *scanner) runScan(ctx context.Context) ([]zap.Alert, error) {
	s.logger.Infof("Starting scan for target : %s", s.target.TargetAddress)

	zapAPI, err := zapapi.NewZap(
		s.conf.ZapAPIKey,
		s.conf.Proxy,
		s.logger,
		s.conf.LocalTmpDir,
	)
	if err != nil {
		return nil, err
	}

	scanID, err := zapAPI.StartScan(ctx, s.target.TargetAddress, zapapi.ScanOptions{})
	if err != nil {
		if scanErr, ok := err.(*zapapi.ScanError); ok && scanErr.ErrType == zapapi.ErrUnableToReach {
			// If the target is unreachable, set the status to unreachable
			s.scanFailedStatus = enums.TargetStatusUnreachable
		}
		return nil, err
	}

	s.logger.Info("Scan started")

	return s.pollScan(ctx, zapAPI, scanID)
}

// pollScan polls the scan status and returns the results
func (s *scanner) pollScan(ctx context.Context, zapAPI *zapapi.API, scanID string) ([]zap.Alert, error) {
	// Wait for the scan to complete
	endTime := time.Now().Add(config.WSTimeout)
	for time.Now().Before(endTime) {
		// checking if the scan is completed
		completed, err := zapAPI.IsScanCompleted(ctx, scanID)
		if err != nil {
			s.logger.Error("Error while checking scan status", err)
			return nil, err
		}

		if completed {
			s.logger.Info("Scan completed")
			break
		}

		if err := utils.SleepContext(ctx, config.WSPollInterval); err != nil {
			return nil, err
		}
	}

	// Fetch and return the scan results
	results, err := zapAPI.GetResults(ctx, scanID)
	if err != nil {
		s.logger.Error("Failed to get scan results", err)
		return nil, err
	}

	s.logger.Info("Returning scan results")
	return results, nil
}

// IsScannerReady checks if the scanner is alive
func IsScannerReady(ctx context.Context, conf *config.Config, lgr *logger.Logger) bool {
	zapAPI, err := zapapi.NewZap(
		conf.ZapAPIKey,
		conf.Proxy,
		lgr,
		conf.LocalTmpDir,
	)
	if err != nil {
		return false
	}

	zVer, err := zapAPI.Core.Version(ctx)
	if err != nil {
		return false
	}

	if zVer == "" {
		return false
	}

	return true
}
