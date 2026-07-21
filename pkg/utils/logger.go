package utils

import (
	"io"
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

	multiWriter := io.MultiWriter(file, os.Stdout)
	logger = log.New(multiWriter, "", log.LstdFlags)

	log.Printf("Log directory: %s", logDir)
}

func LogInfo(message string) {
	if logger != nil {
		logger.Println("[INFO] " + message)
	}
}

func LogErrorf(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, args...)
	}
}

func LogInfof(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+format, args...)
	}
}
