package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CSPF-Founder/iva/scanner/config"
	"github.com/CSPF-Founder/iva/scanner/db"
	"github.com/CSPF-Founder/iva/scanner/internal/repositories"
	"github.com/CSPF-Founder/iva/scanner/internal/scans/network"
	"github.com/CSPF-Founder/iva/scanner/internal/scans/web"
	"github.com/CSPF-Founder/iva/scanner/logger"
)

type Application struct {
	Config *config.Config
	DB     *repositories.Repository
	Logger *logger.Logger
}

func NewApplication(conf *config.Config, appLogger *logger.Logger) *Application {
	return &Application{
		Config: conf,
		Logger: appLogger,
	}
}

func main() {
	// Load the config
	conf, err := config.LoadConfig()
	// Just warn if a contact address hasn't been configured
	if err != nil {
		log.Fatal("Error loading config", err)
	}

	appLogger, err := logger.NewLogger(conf.Logging)
	if err != nil {
		log.Fatal("Error setting up logging", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Set up a signal channel to capture interrupt and termination signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Handle signals in a goroutine
	go func() {
		// Wait for the interrupt signal
		<-interrupt

		// Perform cleanup operations before exiting (if needed)
		appLogger.Info("Scanner is stopping...")

		// Cancel the context to signal a graceful shutdown
		cancel()
	}()

	cliInput, err := parseCLI()
	if err != nil {
		appLogger.Fatal("Error handling CLI", err)
		return
	}

	handleModule(ctx, conf, appLogger, cliInput)
}

// Handle the module based on the CLI input
func handleModule(
	ctx context.Context,
	conf *config.Config,
	appLogger *logger.Logger,
	cliInput CLIInput,
) {

	switch cliInput.Module {
	case ModCheckNS:
		runNSReadyCheck(ctx, conf, appLogger)
	case ModCheckWS:
		runWSReadyCheck(ctx, conf, appLogger)
	case ModScan:
		app, err := baseSetup(ctx, conf, appLogger)
		if err != nil {
			appLogger.Fatal("Error setting up app", err)
			return
		}

		select {
		case <-ctx.Done():
			// Context has timed out
			app.Logger.Info("Scanner has stopped")
		default:
			err = app.RunScan(ctx, cliInput.TargetID)
			if err != nil {
				appLogger.Fatal("Error running scan", err)
				return
			}

		}
	default:
		appLogger.Fatal("Invalid module", nil)
		return
	}
}

func baseSetup(ctx context.Context, conf *config.Config, appLogger *logger.Logger) (*Application, error) {
	app := NewApplication(conf, appLogger)

	dbRepos, err := db.SetupDatabase(ctx, conf)
	if err != nil {
		return nil, err
	}
	app.DB = dbRepos

	return app, nil
}

func (app *Application) RunScan(ctx context.Context, targetID string) error {
	target, err := app.DB.Target.FindByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("error fetching target: %w", err)
	}

	switch target.TargetType {
	case "url":
		app.Logger.Infof("Scanning Web: %s | Customer: %s", target.TargetAddress, target.CustomerUsername)
		webScanner := web.NewWebScanner(target, app.Logger, app.Config, app.DB)
		webScanner.Run(ctx)
	case "ip", "ip_range":
		app.Logger.Infof("Scanning Network: %s | Customer: %s", target.TargetAddress, target.CustomerUsername)
		networkScanner := network.NewNetworkScanner(target, app.Logger, app.Config, app.DB)
		networkScanner.Run(ctx)
	default:
		return errors.New("invalid target category")
	}

	return nil
}

func runNSReadyCheck(ctx context.Context, conf *config.Config, appLogger *logger.Logger) {
	statusJson := struct {
		IsReady bool `json:"is_ready"`
	}{}

	statusJson.IsReady = network.IsScannerReady(ctx, conf, appLogger)

	msgStr, err := json.Marshal(statusJson)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(string(msgStr))
}

func runWSReadyCheck(ctx context.Context, conf *config.Config, appLogger *logger.Logger) {
	statusJson := struct {
		IsReady bool `json:"is_ready"`
	}{}

	statusJson.IsReady = web.IsScannerReady(ctx, conf, appLogger)

	msgStr, err := json.Marshal(statusJson)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(string(msgStr))
}
