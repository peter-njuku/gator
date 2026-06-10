package logging

import (
	"log"
	"os"
	"path/filepath"
)

var logFile *os.File

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(home, ".gatorlogs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(logDir, "gator.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logFile = f
	log.SetOutput(f)
	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
