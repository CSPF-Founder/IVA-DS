package zapapi

import (
	"context"
	"fmt"
	"time"

	"github.com/CSPF-Founder/libs/zapapi/utils"
	"github.com/CSPF-Founder/libs/zapapi/zap"
)

type Logger interface {
	Info(msg string)
	Infof(format string, a ...any)
	Debug(msg string)
	Error(msg string, err error)
	Errorf(format string, a ...any)
	Fatal(msg string, err error)
	Warn(msg string, err error)
}

type ScanError struct {
	ErrType string
	Msg     string
}

// Implement the error interface
func (e *ScanError) Error() string {
	return e.Msg
}

// Define error types
var (
	ErrUnableToReach    = "unable to reach"
	ErrUnableToInitiate = "unable to initiate"
)

type ScanOptions struct {
	PreURLListFile string
}

type API struct {
	logger  Logger
	workDir string

	Ascan      zap.Ascan
	Spider     zap.Spider
	Core       zap.Core
	ImportURLs zap.ImportURLs
}

func NewZap(
	apiKey string,
	proxy string,
	logger Logger,
	workDir string,
) (*API, error) {

	z := &API{
		logger:  logger,
		workDir: workDir,
	}

	err := z.baseSetup(apiKey, proxy)
	if err != nil {
		return nil, err
	}

	return z, nil
}

func (z *API) baseSetup(apiKey, proxy string) error {

	cfg := &zap.Config{
		Base:   zap.DefaultBase,
		Proxy:  proxy,
		APIKey: apiKey,
	}
	client, err := zap.NewClient(cfg)
	if err != nil {
		return err
	}

	z.Ascan = zap.NewAscan(client)
	z.Spider = zap.NewSpider(client)
	z.Core = zap.NewCore(client)
	z.ImportURLs = zap.NewImportURLs(client)

	return nil
}

// StartScan initiates the scan
func (z *API) StartScan(ctx context.Context, target string, opts ScanOptions) (string, error) {
	err := z.checkConnection(ctx)
	if err != nil {
		return "", &ScanError{
			ErrType: ErrUnableToReach,
			Msg:     err.Error(),
		}
	}

	z.logger.Info("Connected to WS")
	scanID, err := z.start(ctx, target, opts)
	if err != nil {
		if scanErr, ok := err.(*ScanError); ok {
			// return the error as is if it is a ScanError
			return "", scanErr
		}

		return "", &ScanError{
			ErrType: ErrUnableToInitiate,
			Msg:     err.Error(),
		}
	}

	return scanID, nil
}

func (z *API) checkConnection(ctx context.Context) error {
	_, err := z.Core.Version(ctx)
	if err != nil {
		return err
	}
	return nil
}

// ImportURLs imports a list of URLs from a file
func (z *API) importURLs(ctx context.Context, urlListFile string) error {
	if len(urlListFile) > 0 {
		startTime := time.Now()
		err := z.ImportURLs.FromFile(ctx, urlListFile)
		if err != nil {
			return err
		}
		elapsedTime := time.Since(startTime)
		z.logger.Info(fmt.Sprintf("ws - imported url - %v secs", elapsedTime.Seconds()))
	}

	return nil
}

func (z *API) start(ctx context.Context, url string, opts ScanOptions) (string, error) {

	err := z.Core.NewSession(ctx, "", "")
	if err != nil {
		return "", fmt.Errorf("error creating new session: %v", err)
	}

	err = z.Spider.SetOptionParseSitemapXml(ctx, true)
	if err != nil {
		return "", fmt.Errorf("error setting spider option: %v", err)
	}
	err = z.Spider.SetOptionParseRobotsTxt(ctx, true)
	if err != nil {
		return "", fmt.Errorf("error setting spider option: %v", err)
	}

	err = z.Core.AccessURL(ctx, url, "")
	if err != nil {
		return "", &ScanError{
			ErrType: ErrUnableToReach,
			Msg:     fmt.Sprintf("error accessing URL: %v", err),
		}
	}

	if err := z.importURLs(ctx, opts.PreURLListFile); err != nil {
		return "", fmt.Errorf("error importing URLs: %v", err)
	}

	if err := z.spider(ctx, url); err != nil {
		return "", fmt.Errorf("error initiating spider: %v", err)
	}

	// Give the passive scanner a chance to finish
	if err := utils.SleepContext(ctx, 10*time.Second); err != nil {
		return "", fmt.Errorf("error sleeping: %v", err)
	}

	// Initiate the Scan:
	return z.Ascan.Scan(
		ctx,
		url,
		zap.AScanOpts{},
	)
}

// spider initiates the spider scan
func (z *API) spider(ctx context.Context, url string) error {
	if err := utils.SleepContext(ctx, 1*time.Second); err != nil {
		return err
	}

	spiderID, err := z.Spider.Scan(ctx, url, "", "", "", "")
	if err != nil {
		return err
	}

	if err := utils.SleepContext(ctx, 2*time.Second); err != nil {
		return err
	}

	stoppingTime := time.Now().Add(5 * time.Minute)
	if err := z.pollSpiderStatus(ctx, spiderID, stoppingTime); err != nil {
		return err
	}

	return nil
}

// pollSpiderStatus polls the spider status every 4 seconds
// and stops the spider if it takes more than 5 minutes
func (z *API) pollSpiderStatus(
	ctx context.Context,
	spiderID string,
	timeout time.Time,
) error {
	for {
		status, err := z.Spider.Status(ctx, spiderID)
		if err != nil {
			return fmt.Errorf("error getting spider status: %v", err)
		}

		if status >= 100 {
			z.logger.Info("spider scan 100% complete")
			return nil
		}
		// z.logger.Info(fmt.Sprintf("Spider progress :%d", status))

		// Check if stopping time has been exceeded
		if time.Now().After(timeout) {
			z.logger.Info("stopping spider after 5 minutes")
			err := z.Spider.StopAllScans(ctx)
			if err != nil {
				// z.logger.Error("Error stopping spider:", err)
				return fmt.Errorf("error stopping spider: %v", err)
			}
			return nil
		}

		if err := utils.SleepContext(ctx, 4*time.Second); err != nil {
			return err
		}
	}
}

func (z *API) IsScanCompleted(ctx context.Context, scanID string) (bool, error) {
	status, err := z.Ascan.Status(ctx, scanID)
	if err != nil {
		return false, err
	}

	return status == 100, nil
}

func (z *API) GetResults(ctx context.Context, scanID string) ([]zap.Alert, error) {
	_, err := z.Core.Version(ctx)
	if err != nil {
		return nil, fmt.Errorf("scanner failed to respond: %v", err)
	}

	err = z.Ascan.Stop(ctx, scanID)
	if err != nil {
		// ignore error if scan is already stopped
		z.logger.Error("ws Retriever - Error stopping scan", err)
	}

	z.logger.Info("ws Retriever - Getting alert details")

	zapAlerts, err := z.Core.GetAlerts(ctx, zap.OptsToGetAlerts{})
	if err != nil {
		return nil, err
	}
	filteredAlerts := z.filterAlerts(zapAlerts)

	err = z.Ascan.RemoveScan(ctx, scanID)
	if err != nil {
		return nil, err
	}
	output := sortAlerts(filteredAlerts)

	return output.Alerts, nil
}
