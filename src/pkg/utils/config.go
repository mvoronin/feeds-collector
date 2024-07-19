package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	FeedsUpdateInterval time.Duration `yaml:"feeds_update_interval"`
	Logging             struct {
		InfoLog  string `yaml:"info_log"`
		ErrorLog string `yaml:"error_log"`
	} `yaml:"logging"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
}

// Default config values
const (
	databasePath        = "db/feeds.db"
	feedsUpdateInterval = "40m"
	port                = "8080"
	infoLog             = "info.log"
	errorLog            = "error.log"
)

func ValidateConfig(config *Config) error {
	// check if config.Server.Port is integer between 1 and 65535
	if config.Server.Port == "" {
		config.Server.Port = port
	}
	port, err := strconv.Atoi(config.Server.Port)
	if err != nil {
		return fmt.Errorf("invalid server.port: %v", err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	return nil
}

func ReadConfig(configPath string) (*Config, error) {
	configFile, err := os.Open(filepath.Clean(configPath))
	if err != nil {
		log.Printf("Error opening config file: %v", err)
		return nil, err
	}
	defer func(configFile *os.File) {
		err := configFile.Close()
		if err != nil {
			log.Printf("Error closing config file: %v", err)
			os.Exit(1)
		}
	}(configFile)

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Printf("Error unmarshalling config file: %v", err)
		return nil, err
	}

	err = ValidateConfig(&config)
	if err != nil {
		log.Printf("Error validating config file: %v", err)
		return nil, err
	}

	return &config, nil
}
