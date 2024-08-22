package logging

import (
	"log"
	"os"
)

// InitializeLogger initializes and returns a logger instance
func InitializeLogger(prefix string) *log.Logger {
	logger := log.New(os.Stdout, prefix, log.LstdFlags)
	return logger
}

// PublishLog publishes a log message
func PublishLog(logger *log.Logger, message string) {
	logger.Println(message)
}
