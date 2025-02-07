package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CSPF-Founder/iva/scanner/logger"
	"github.com/CSPF-Founder/libs/gmp/protocol"
	"github.com/joho/godotenv"
)

// Config represents the configuration information.
type Config struct {
	LocalTmpDir   string
	DatabaseURI   string `json:"database_uri"`
	DBName        string
	ReporterBin   string
	ZapAPIKey     string
	Proxy         string
	CSVFileName   string
	CSVFilePath   string
	Logging       *logger.Config `json:"logging"`
	OpenvasConfig protocol.OpenvasConfig
}

// Version contains the current project version
var Version = "1"

func loadEnv() {
	// determin bin directory and load .env from there
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

// LoadConfig loads the configuration from the specified filepath
func LoadConfig() (*Config, error) {
	if os.Getenv("USE_DOTENV") != "false" {
		loadEnv()
	}

	// Initialize the OpenVAS configuration
	openVASConfig := protocol.OpenvasConfig{
		Username:     getEnvValueOrError("OPENVAS_USERNAME"),
		Password:     getEnvValueOrError("OPENVAS_PASSWORD"),
		CLIPath:      getEnvValueOrError("OPENVAS_CLI_PATH"),
		ScanConfigID: getEnvValueOrError("OPENVAS_SCAN_CONFIG_ID"),
		PortListID:   getEnvValueOrError("OPENVAS_PORT_LIST_ID"),
		Host:         getEnvValueOrError("OPENVAS_HOST"),
		Port:         9390,
		Timeout:      60,
	}

	config := &Config{
		DatabaseURI:   getEnvValueOrError("DATABASE_URI"),
		DBName:        getEnvValueOrError("DB_NAME"),
		LocalTmpDir:   getEnvValueOrError("LOCAL_TMP_DIR"),
		ReporterBin:   getEnvValueOrError("REPORTER_BIN"),
		ZapAPIKey:     getEnvValueOrError("ZAP_API_KEY"),
		Proxy:         getEnvValueOrError("PROXY"),
		OpenvasConfig: openVASConfig,
		Logging: &logger.Config{
			Level: getEnvValueOrError("LOG_LEVEL"),
			// Log to stdout if no file path is provided
			FilePath: getEnvValueOrDefault("LOG_FILE_PATH", ""),
		},
	}
	return config, nil
}

// Constants for timeouts
const (
	WSTimeout      = 45 * time.Minute //  45 Minutes
	WSPollInterval = 45 * time.Second //  45 Seconds

	NSTimeout      = 45 * time.Minute //  45 Minutes
	NSPollInterval = 45 * time.Second //  45 Seconds

	IPRangeTimeoutPerIP = 30 * time.Minute // 30 minutes
	IPRangePollInterval = 1 * time.Minute  // 1 minute
	IPRangeMaxTimeout   = 24 * time.Hour   // 24 hours
)
