package config

import (
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
	Gatherer GathererConfig `yaml:"gatherer"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port string `yaml:"port"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	InfoLog  string `yaml:"info_log"`
	ErrorLog string `yaml:"error_log"`
}

// GathererConfig represents the gatherer configuration
type GathererConfig struct {
	Interval time.Duration `yaml:"interval"`
}

// ReadConfig reads the configuration from a YAML file
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
	return &config, nil
}
