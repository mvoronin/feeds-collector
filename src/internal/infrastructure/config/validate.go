package config

import (
	"fmt"
	"strconv"
	"time"
)

// ValidateConfig validates the configuration and sets default values
func ValidateConfig(config *Config) error {
	if config.Server.Port == "" {
		config.Server.Port = "8080" // Default port
	}
	port, err := strconv.Atoi(config.Server.Port)
	if err != nil {
		return fmt.Errorf("invalid server.port: %v", err)
	}
	if port <= 1024 || port > 65535 {
		return fmt.Errorf("server.port must be between 1025 and 65535")
	}

	if config.Database.Path == "" {
		config.Database.Path = "feeds.db" // Default database path
	}

	if config.Logging.InfoLog == "" {
		config.Logging.InfoLog = "info.log" // Default info log file
	}

	if config.Logging.ErrorLog == "" {
		config.Logging.ErrorLog = "error.log" // Default error log file
	}

	if config.Gatherer.Interval == 0 {
		config.Gatherer.Interval = 40 * time.Minute
	}

	return nil
}
