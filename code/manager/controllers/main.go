package controllers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/CSPF-Founder/iva/manager/config"
	"github.com/CSPF-Founder/iva/manager/enums"
	"github.com/CSPF-Founder/iva/manager/internal/repositories"
	"github.com/CSPF-Founder/iva/manager/logger"
	"github.com/CSPF-Founder/iva/manager/models"
	"github.com/CSPF-Founder/iva/manager/utils/iputils"
)

type MainController struct {
	db             repositories.Repository
	scannerCmd     string
	updaterCommand string
	logger         *logger.Logger
	scanLogsDir    string
}

func NewMainController(
	db repositories.Repository,
	lgr *logger.Logger,
	scannerCmd string,
	updaterCommand string,
	scanLogsDir string,
) *MainController {
	return &MainController{
		db:             db,
		logger:         lgr,
		scannerCmd:     scannerCmd,
		updaterCommand: updaterCommand,
		scanLogsDir:    scanLogsDir,
	}
}

// Run starts both services and waits for them to complete.
func (c *MainController) Run(ctx context.Context) {
	// if the service is restarted, mark all unfinished scans as failed
	c.markAllUnfinishedAsFailed(ctx)

	var wg sync.WaitGroup

	wg.Add(2)

	// start web service in a new goroutine
	go c.startWebService(ctx, &wg)

	// start network service in a new goroutine
	go c.startNetService(ctx, &wg)

	// wait for both goroutines to finish
	wg.Wait()
}

// calcScanTimeout calculates the scan timeout based on the target type
// If the target is a URL, it uses the web scan timeout
// if the target is an IP, it uses the network scan timeout
// If the target is an IP range, it calculates the timeout based on the number of IPs
//   - It uses the base timeout and adds the timeout per IP
//   - If the total timeout exceeds the max timeout, it uses the max timeout
func calcScanTimeout(target models.Target) time.Duration {
	// if the target is a URL, use the web scan timeout
	if strings.HasPrefix(target.TargetAddress, "http") {
		return config.WebScanTimeout
	}

	scanTimeout := config.NetworkScanTimeout

	// get the count of IP addresses if the target is a range
	ipCount, _ := iputils.GetIPCountIfRange(target.TargetAddress)

	if ipCount > 1 {
		scanTimeout += config.IPRangeTimeoutPerIP * time.Duration(ipCount)
	}

	if scanTimeout > config.IPRangeMaxTimeout {
		scanTimeout = config.IPRangeMaxTimeout
	}

	return scanTimeout
}

// startScannerProcess starts the scanner process
// It runs the scanner command with the target ID as an argument
// It sets the LOG_FILE_PATH environment variable if the scanLogsDir is set
func (c *MainController) startScannerProcess(ctx context.Context, target models.Target) error {

	c.logger.Infof("Starting scan process for target %s", target.ID.Hex())
	scanTimeout := calcScanTimeout(target)

	scanCtx, cancel := context.WithTimeout(ctx, scanTimeout)
	defer cancel()

	cmd := exec.CommandContext(scanCtx, c.scannerCmd, "-t", target.ID.Hex())

	// set environment variables
	if c.scanLogsDir != "" {
		scanLogPath := filepath.Join(c.scanLogsDir, target.ID.Hex()+".log")
		cmd.Env = append(cmd.Env, "LOG_FILE_PATH="+scanLogPath)
	}

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		errMsg := err.Error()

		if stdOut.String() != "" {
			errMsg = errMsg + "\n" + stdOut.String()
		}

		if stdErr.String() != "" {
			errMsg = errMsg + "\n" + stdErr.String()
		}
		return errors.New(errMsg)
	}

	if stdErr.Len() > 0 {
		return fmt.Errorf("scanner process error: %s", stdErr.String())
	}

	c.logger.Info(fmt.Sprintf("Scan process done for target %s", target.ID.Hex()))

	updatedTarget, err := c.db.Target.FindById(ctx, target.ID)
	if err != nil {
		return err
	}

	// if it is not in allowed states, then mark it as scan failed
	allowedStates := map[enums.TargetStatus]bool{
		enums.TargetStatusReportGenerated: true,
		enums.TargetStatusScanFailed:      true,
		enums.TargetStatusUnreachable:     true,
	}

	if _, ok := allowedStates[updatedTarget.ScanStatus]; !ok {
		c.logger.Error(fmt.Sprintf("scan failed for target %s", target.ID.Hex()), nil)
		updatedTarget.ScanStatus = enums.TargetStatusScanFailed
		err = c.db.Target.UpdateScanStatus(ctx, updatedTarget)
		if err != nil {
			return err
		}
	}

	return nil
}

// Any unfinished scans are marked as failed at starting the service
func (c *MainController) markAllUnfinishedAsFailed(ctx context.Context) {
	err := c.db.Target.MarkUnfinishedAsFailed(ctx, []string{"ip", "ip_range", "web"})
	if err != nil {
		c.logger.Error("Error marking unfinished scans as failed", err)
	}
}
