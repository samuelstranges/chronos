package util

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/samuelstranges/chronos/types"
)

// ShowIfError displays an error message in the status bar if err is not nil
// seconds: how long to show the error (in seconds)
// err: the potential error to check
// errorMessage: the message to display (err.Error() will be appended)
// weekModel: the model to update with error information
func ShowIfError(weekModel *types.WeekModel, seconds int, err error, errorMessage string) {
	if err != nil {
		fullMessage := errorMessage + ": " + err.Error()
		weekModel.ErrorMessage = fullMessage
		weekModel.ErrorExpiry = time.Now().Add(time.Duration(seconds) * time.Second)

		// Log to file for debugging
		logErrorToFile(fullMessage)
	}
}

// logErrorToFile appends error to chronos error log
func logErrorToFile(message string) {
	// Get user's config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return // Silently fail if can't get config dir
	}

	logPath := filepath.Join(configDir, "chronos", "error.log")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return // Silently fail
	}

	// Open log file in append mode
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Silently fail
	}
	defer logFile.Close()

	// Write timestamped error
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(logFile, "[%s] %s\n", timestamp, message)
}
