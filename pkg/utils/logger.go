package utils

import (
	"log"
	"os"
	"path/filepath"
)

var logger *log.Logger

func InitLogger() {
	logDir, err := os.UserConfigDir()
	if err != nil {
		logDir = "."
	}
	logDir = filepath.Join(logDir, "ClassManager", "logs")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
	}

	logFile := filepath.Join(logDir, "app.log")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		logger = log.New(os.Stdout, "", log.LstdFlags)
		return
	}

	logger = log.New(file, "", log.LstdFlags)
}

func LogInfo(message string) {
	if logger != nil {
		logger.Println("[INFO] " + message)
	}
}

func LogError(err error) {
	if logger != nil {
		logger.Println("[ERROR] " + err.Error())
	}
}

func LogErrorf(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, args...)
	}
}

func GetLogDirectory() string {
	logDir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	return filepath.Join(logDir, "ClassManager", "logs")
}
