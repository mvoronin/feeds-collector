package internal

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	InfoLogLevel  = "INFO"
	ErrorLogLevel = "ERROR"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

var basePath string

func init() {
	basePath = os.Args[0]
	basePath = filepath.Dir(basePath)
}

func InitLogging(logFilePath string, logLevel string) (func(), error) {
	if logLevel != InfoLogLevel && logLevel != ErrorLogLevel {
		return nil, fmt.Errorf("invalid log level: %s", logLevel)
	}

	var logWriter io.Writer
	var closeLogging func()

	if logFilePath == "" {
		logWriter = io.Writer(os.Stdout)
	} else {
		logFile, err := os.OpenFile(filepath.Clean(filepath.Join(basePath, logFilePath)),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			log.Fatalf("Failed to open info log file: %v", err)
		}
		closeLogging = func() {
			err := logFile.Close()
			if err != nil {
				log.Fatalf("Failed to close info log file: %v", err)
			}
		}
		logWriter = io.MultiWriter(os.Stdout, logFile)
	}

	if logLevel == InfoLogLevel {
		InfoLogger = log.New(logWriter, InfoLogLevel+": ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		ErrorLogger = log.New(logWriter, ErrorLogLevel+": ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	return closeLogging, nil
}
