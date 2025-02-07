package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/CSPF-Founder/iva/manager/enums"
	"github.com/CSPF-Founder/iva/manager/internal/updater"
	"github.com/CSPF-Founder/iva/manager/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

// startWebService starts the web service
// This runs in a loop and checks for web scan jobs
// If a job is found, it starts the scanner process
// if context is cancelled(example: ctrl+c pressed), it stops the service
func (c *MainController) startWebService(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Info(fmt.Sprintf("recovered from panic in web scanner: %v", r))
			if err := utils.SleepContext(ctx, 5*time.Minute); err != nil {
				c.logger.Error("Error sleeping", err)
			}
		}
		defer wg.Done()
	}()

	// First time if scanner is not ready, wait for 3 minutes
	webScannerReadyCheck(ctx, c.scannerCmd)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Shutting down web scanner")
			return
		default:
			c.webRoutine(ctx)
		}
	}
}

// webRoutine runs the web scan routine
// This function is called in a loop by startWebService
// It checks for web scan jobs and starts the scanner process
// It also updates the web docker
func (c *MainController) webRoutine(ctx context.Context) {
	err := c.handleWebScans(ctx)
	if err != nil {
		c.logger.Error("web scan failed", err)
	}

	c.webUpdateRoutine(ctx)

	if err := utils.SleepContext(ctx, 10*time.Second); err != nil {
		c.logger.Error("Error sleeping", err)
	}
}

// updateWebFeed runs at 00:15 every day to update the web docker
func (c *MainController) webUpdateRoutine(ctx context.Context) {
	now := time.Now()
	if now.Hour() == 0 && now.Minute() == 15 {
		if err := updater.UpdateWebDocker(ctx, c.updaterCommand); err != nil {
			c.logger.Error("Error while updating web docker: ", err)
		}
	}
}

// handleWebScans checks for web scan jobs and starts the scanner process
// It also marks all unfinished scans as failed
// If a job is found, it starts the scanner process
func (c *MainController) handleWebScans(ctx context.Context) error {
	// mark all unfinished scans as failed
	// ! NOTE: If it is parallel scanning, then this will be a problem
	// ! Or if it is background scanning, then this will be a problem
	// * That time should not use this approach
	// * This is perfect for sequential scanning
	err := c.db.Target.MarkUnfinishedAsFailed(ctx, []string{"url"})
	if err != nil {
		return err
	}

	target, err := c.db.Target.GetWebScanJob(ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}

		c.logger.Error("Error fetching web scans jobs", err)
		return err
	}

	target.ScanStatus = enums.TargetStatusInitiatingScan
	err = c.db.Target.UpdateScanStatus(ctx, target)
	if err != nil {
		return err
	}

	err = c.startScannerProcess(ctx, *target)
	if err != nil {
		target.ScanStatus = enums.TargetStatusScanFailed
		_ = c.db.Target.UpdateScanStatus(ctx, target)
		return err
	}
	return nil
}

// webScannerReadyCheck checks if the web scanner is ready
// It waits for 3 minutes if the scanner is not ready
func webScannerReadyCheck(ctx context.Context, scannerCmd string) {
	for i := 0; i < 6; i++ {
		if isWebScannerReady(ctx, scannerCmd) {
			return
		}
		if err := utils.SleepContext(ctx, 30*time.Second); err != nil {
			return
		}
	}
}

func isWebScannerReady(ctx context.Context, scannerCmd string) bool {

	ctx, cancel := context.WithTimeout(ctx, time.Minute*1)
	defer cancel()

	cmd := exec.CommandContext(ctx, scannerCmd, "-m", "check_ws")

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		return false
	} else if stdErr.Len() > 0 {
		return false
	}

	output := strings.TrimSpace(stdOut.String())
	statusJSON := struct {
		IsReady bool `json:"is_ready"`
	}{}

	err = json.Unmarshal([]byte(output), &statusJSON)
	if err != nil {
		return false
	}

	return statusJSON.IsReady
}
