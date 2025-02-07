package gmp

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/CSPF-Founder/libs/gmp/protocol"
	"github.com/CSPF-Founder/libs/gmp/utils"
)

func Connect(config protocol.OpenvasConfig) (*protocol.OpenvasAPI, error) {
	return protocol.NewOpenvasAPI(config)
}

type Logger interface {
	Info(msg string)
	Infof(format string, a ...any)
	Debug(msg string)
	Error(msg string, err error)
	Errorf(format string, a ...any)
	Fatal(msg string, err error)
	Warn(msg string, err error)
}

type API struct {
	logger Logger
	// workDir string

	OV     *protocol.OpenvasAPI
	Config protocol.OpenvasConfig
}

func NewGMP(
	logger Logger,
	ovConfig protocol.OpenvasConfig,
) (*API, error) {
	ov, err := Connect(ovConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to openvas: %w", err)
	}
	return &API{
		logger: logger,
		OV:     ov,
		Config: ovConfig,
	}, nil
}

type ScanData struct {
	OVTargetID string
	TaskID     string
	ReportID   string
}

func (a *API) createTarget(
	ctx context.Context,
	targetName string,
	target string,
) (string, error) {
	oTargetIDS, err := a.OV.GetTargetIDs(ctx, targetName)
	if err != nil {
		return "", fmt.Errorf("failed to get target ids: %w", err)
	}

	if len(oTargetIDS) > 0 && oTargetIDS[0] != "" {
		// target already exists
		return oTargetIDS[0], nil
	}

	// create target
	tID, err := a.OV.CreateTarget(
		ctx,
		targetName,
		target,
		&protocol.CreateTargetPortList{ID: a.Config.PortListID},
	)

	if err != nil {
		return "", err
	}

	return tID, nil
}

func (a *API) StartScan(ctx context.Context, target string, uniqID string) (*ScanData, error) {
	targetName := target + "_" + uniqID
	targetName = strings.ReplaceAll(targetName, "/", "_")
	targetName = strings.ReplaceAll(targetName, ".", "_")

	ovTargetID, err := a.createTarget(ctx, targetName, target)
	if err != nil {
		return nil, fmt.Errorf("failed to create target: %w", err)
	}

	// creating task
	taskName := "Task_" + targetName
	taskID, err := a.OV.CreateTask(ctx, taskName, ovTargetID, a.Config.ScanConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	reportID, err := a.OV.StartTask(ctx, taskID)
	if err != nil {
		a.logger.Warn("failed to start task:", err)

		if err := utils.SleepContext(ctx, 120*time.Second); err != nil {
			return nil, fmt.Errorf("failed to sleep: %w", err)
		}

		// retry starting the task after sleep
		_, err := a.OV.StartTask(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("retry failed to start task: %w", err)
		}
	}

	return &ScanData{
		OVTargetID: ovTargetID,
		TaskID:     taskID,
		ReportID:   reportID,
	}, nil
}

// IsScanCompleted checks if the scan is completed
// returns true if completed, false if not completed
// returns error if failed to get the progress
func (a *API) IsScanCompleted(ctx context.Context, taskID string) (bool, error) {
	progress, err := a.OV.GetProgress(ctx, taskID)
	if err != nil {
		return false, err
	}

	// if progress 100 or -1 means scan is completed
	return progress == 100 || progress == -1, nil
}

// GetResults returns the scan results
func (a *API) GetResults(ctx context.Context, taskID, reportID string) ([]Alert, error) {
	err := a.OV.StopTask(ctx, taskID)
	if err != nil {
		a.logger.Error("unable to stop the openvas task", err)
		// return nil, err
	}

	// Get Reports in CSV Format
	formatID := "c1645568-627a-11e3-a660-406186ea4fc5"
	reportFilter := "sort-reverse=severity"
	result, err := a.OV.GetReports(ctx, formatID, reportID, true, true, reportFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)

	}

	content, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode report content: %w", err)
	}

	alerts, err := parseCSVData(content, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse alerts csv: %w", err)
	}
	return alerts, nil
}

func (a *API) CleanUpScan(ctx context.Context, sd ScanData) error {

	// Clear Scan Tasks
	if err := a.OV.DeleteTask(ctx, sd.TaskID); err != nil {
		return fmt.Errorf("failed to delete task %s: %w", sd.TaskID, err)
	}

	// Delete Target
	if err := a.OV.DeleteTarget(ctx, sd.OVTargetID); err != nil {
		return fmt.Errorf("failed to delete target %s: %w", sd.OVTargetID, err)
	}

	// Delete Credential
	if err := a.OV.DeleteCredential(ctx, ""); err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	// Delete Report
	if err := a.OV.DeleteReport(ctx, sd.ReportID); err != nil {
		return fmt.Errorf("failed to delete report %s: %w", sd.ReportID, err)
	}

	return nil
}
