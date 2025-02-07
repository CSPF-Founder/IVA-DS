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
	"github.com/CSPF-Founder/iva/manager/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

// startNetService starts the network service
// This runs in a loop and checks for network scan jobs
// If a job is found, it starts the scanner process
// if context is cancelled(example: ctrl+c pressed), it stops the service
func (c *MainController) startNetService(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Info(fmt.Sprintf("recovered from panic in network scanner: %v", r))
			if err := utils.SleepContext(ctx, 5*time.Minute); err != nil {
				c.logger.Error("Error sleeping", err)
			}
		}
		defer wg.Done()
	}()

	if err := utils.SleepContext(ctx, 5*time.Minute); err != nil {
		c.logger.Error("Error sleeping", err)
	}

	netScannerReadyCheck(ctx, c.scannerCmd)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Shutting down network scanner")
			return
		default:
			c.netRoutine(ctx)
		}
	}
}

// netRoutine runs the network scan routine
// This function is called in a loop by startNetService
// It checks for network scan jobs and starts the scanner process
// It also updates the network docker and feed
func (c *MainController) netRoutine(ctx context.Context) {
	err := c.handleNetworkScans(ctx)
	if err != nil {
		c.logger.Error("network scan failed", err)
	}

	// err = updater.HandlePanelFeedUpdate(ctx, c.updaterCommand)
	// if err != nil {
	// 	c.logger.Error("panel feed update failed", err)
	// }

	// c.networkUpdateRoutine(ctx)

	if err := utils.SleepContext(ctx, 10*time.Second); err != nil {
		c.logger.Error("Error sleeping", err)
	}
}

// networkUpdateRoutine runs at 00:15 every day to update the network docker and feed
// func (c *MainController) networkUpdateRoutine(ctx context.Context) {
// 	now := time.Now()
// 	if now.Hour() == 0 && now.Minute() == 15 {
// 		if err := updater.UpdateNetworkDocker(ctx, c.updaterCommand); err != nil {
// 			c.logger.Error("Error while updating network docker: ", err)
// 		}
// 		if err := updater.UpdateNetworkFeed(ctx, c.updaterCommand); err != nil {
// 			c.logger.Error("Error while updating network feed: ", err)
// 		}
// 	}
// }

// handleNetworkScans checks for network scan jobs and starts the scanner process
// It also marks all unfinished scans as failed
// If a job is found, it starts the scanner process
func (c *MainController) handleNetworkScans(ctx context.Context) error {
	// mark all unfinished scans as failed
	// ! NOTE: If it is parallel scanning, then this will be a problem
	// ! Or if it is background scanning, then this will be a problem
	// * That time should not use this approach
	// * This is perfect for sequential scanning
	err := c.db.Target.MarkUnfinishedAsFailed(ctx, []string{"ip", "ip_range"})
	if err != nil {
		return err
	}

	target, err := c.db.Target.GetNetworkScanJob(ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		c.logger.Error("Error fetching network scans jobs", err)
		return err
	}

	target.ScanStatus = enums.TargetStatusInitiatingScan
	err = c.db.Target.UpdateScanStatus(ctx, target)
	if err != nil {
		return err
	}

	scanError := c.startScannerProcess(ctx, *target)
	if scanError != nil {
		target.ScanStatus = enums.TargetStatusScanFailed
		_ = c.db.Target.UpdateScanStatus(ctx, target)
		return scanError
	}
	return nil
}

// netScannerReadyCheck checks if the network scanner is ready
// It checks for 5 minutes in a loop
func netScannerReadyCheck(ctx context.Context, scannerCmd string) {
	for i := 0; i < 10; i++ {
		if isNetScannerReady(ctx, scannerCmd) {
			return
		}
		if err := utils.SleepContext(ctx, 30*time.Second); err != nil {
			return
		}
	}
}

func isNetScannerReady(ctx context.Context, scannerCmd string) bool {

	ctx, cancel := context.WithTimeout(ctx, time.Minute*1)
	defer cancel()

	cmd := exec.CommandContext(ctx, scannerCmd, "-m", "check_ns")

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
