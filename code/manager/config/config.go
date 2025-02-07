package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CSPF-Founder/iva/manager/logger"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Env          string
	DatabaseURI  string
	DatabaseName string

	LogLevel       string
	ScannerCmd     string
	UpdaterCommand string

	ScanLogsDir string
	Logging     *logger.Config
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func getEnvValueOrError(key string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.GetFallBackLogger().Error(fmt.Sprintf("Environment variable %s not set", key), nil)
		os.Exit(1)
	}
	return value
}

func getEnvValueOrDefault(key string, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value
}

func loadEnv() {
	//determin bin directory and load .env from there
	exe, err := os.Executable()
	if err != nil {
		logger.GetFallBackLogger().Fatal("Error loading .env file", err)
	}
	binDir := filepath.Dir(exe)
	envPath := filepath.Join(binDir, ".env")
	if err := godotenv.Load(envPath); err == nil {
		return
	}

	// try to load .env from current directory
	envPath = ".env"
	if err := godotenv.Load(envPath); err == nil {
		return
	}
	logger.GetFallBackLogger().Error("Error loading .env file", err)
	os.Exit(1)

}

// LoadConfig loads the configuration from the specified filepath
func LoadConfig() AppConfig {
	if os.Getenv("USE_DOTENV") != "false" {
		loadEnv()
	}

	logLevel := getEnv("LOG_LEVEL", "debug")

	scansLogDir := getEnvValueOrError("SCAN_LOGS_DIR")
	if scansLogDir[len(scansLogDir)-1] != '/' {
		scansLogDir += "/"
	}

	return AppConfig{
		DatabaseURI:    getEnvValueOrError("DATABASE_URI"),
		DatabaseName:   getEnvValueOrError("DATABASE_NAME"),
		LogLevel:       logLevel,
		ScannerCmd:     getEnvValueOrError("SCANNER_CMD"),
		UpdaterCommand: getEnvValueOrError("UPDATER_COMMAND"),
		ScanLogsDir:    scansLogDir,
		Logging: &logger.Config{
			Level: logLevel,
			// Log to stdout if no file path is provided
			FilePath: getEnvValueOrDefault("LOG_FILE_PATH", ""),
		},
	}
}

// Constants for timeouts
const (
	WebScanTimeout      = 1 * time.Hour    //  1 hour
	NetworkScanTimeout  = 1 * time.Hour    //  1 hour
	IPRangeTimeoutPerIP = 40 * time.Minute // 30 minutes + 10 minutes per IP
	IPRangeMaxTimeout   = 25 * time.Hour   // 24 hours + 1 hour
)
