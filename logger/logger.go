package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logger     *log.Logger
	logFile    *os.File
	loggerOnce sync.Once
)

// Init initializes the logger with a file path
// Logs are written to ~/.config/port-digger/logs/port-digger.log
func Init() error {
	var err error
	loggerOnce.Do(func() {
		homeDir, e := os.UserHomeDir()
		if e != nil {
			err = fmt.Errorf("failed to get home directory: %w", e)
			return
		}

		logDir := filepath.Join(homeDir, ".config", "port-digger", "logs")
		if e := os.MkdirAll(logDir, 0755); e != nil {
			err = fmt.Errorf("failed to create log directory: %w", e)
			return
		}

		logPath := filepath.Join(logDir, "port-digger.log")
		logFile, e = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if e != nil {
			err = fmt.Errorf("failed to open log file: %w", e)
			return
		}

		logger = log.New(logFile, "", log.LstdFlags)
		Info("Logger initialized, log file: %s", logPath)
	})
	return err
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if logger == nil {
		return
	}
	msg := fmt.Sprintf(format, v...)
	logger.Printf("[INFO] %s", msg)
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if logger == nil {
		return
	}
	msg := fmt.Sprintf(format, v...)
	logger.Printf("[ERROR] %s", msg)
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if logger == nil {
		return
	}
	msg := fmt.Sprintf(format, v...)
	logger.Printf("[DEBUG] %s", msg)
}

// LogLsofQuery logs an lsof query execution
func LogLsofQuery(command []string, portCount int, err error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if err != nil {
		Error("lsof query failed at %s: command=%v, error=%v", timestamp, command, err)
	} else {
		Info("lsof query succeeded at %s: command=%v, found %d ports", timestamp, command, portCount)
	}
}

// LogLLMRequest logs an LLM API request
func LogLLMRequest(command string, serviceName string, err error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if err != nil {
		Error("LLM request failed at %s: command=%s, error=%v", timestamp, command, err)
	} else {
		Info("LLM request succeeded at %s: command=%s, service_name=%s", timestamp, command, serviceName)
	}
}
