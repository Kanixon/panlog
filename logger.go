package panlog

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	LogLevel      string        // Log level (debug, info, warn, error, fatal, panic)
	LogFile       string        // Path to log file
	MaxSize       int64         // Maximum size in bytes before rotation
	MaxAge        time.Duration // Maximum age of log files to keep
	MaxBackups    int           // Maximum number of old log files to keep
	Compress      bool          // Whether to compress old log files
	RotateDaily   bool          // Whether to rotate daily regardless of size
	JSONFormat    bool          // Whether to use JSON format
	ConsoleOutput bool          // Whether to output to console as well
}

// Logger wraps logrus with log rotation capabilities
type Logger struct {
	logrus.Logger
	rotator *LogRotator
	config  LoggerConfig
}

// NewLogger creates a new logger with log rotation
func NewLogger(config LoggerConfig) (*Logger, error) {
	// Parse log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}

	// Set defaults
	if config.MaxSize == 0 {
		config.MaxSize = 100 * 1024 * 1024 // 100MB
	}
	if config.MaxAge == 0 {
		config.MaxAge = 7 * 24 * time.Hour // 7 days
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 5
	}

	// Create log rotator if file logging is enabled
	var rotator *LogRotator
	var writers []io.Writer

	if config.LogFile != "" {
		rotator, err = NewLogRotator(LogRotatorConfig{
			FilePath:    config.LogFile,
			MaxSize:     config.MaxSize,
			MaxAge:      config.MaxAge,
			MaxBackups:  config.MaxBackups,
			Compress:    config.Compress,
			RotateDaily: config.RotateDaily,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create log rotator: %w", err)
		}
		writers = append(writers, rotator)
	}

	// Add console output if requested
	if config.ConsoleOutput {
		writers = append(writers, os.Stdout)
	}

	// Create multi-writer
	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else if len(writers) > 1 {
		output = io.MultiWriter(writers...)
	} else {
		output = os.Stdout
	}

	// Create logger
	logger := &Logger{
		Logger: logrus.Logger{
			Out:       output,
			Formatter: getFormatter(config.JSONFormat),
			Level:     level,
		},
		rotator: rotator,
		config:  config,
	}

	return logger, nil
}

// Close closes the logger and its underlying rotator
func (l *Logger) Close() error {
	if l.rotator != nil {
		return l.rotator.Close()
	}
	return nil
}

// Rotate manually triggers log rotation
func (l *Logger) Rotate() error {
	if l.rotator != nil {
		return l.rotator.Rotate()
	}
	return nil
}

// GetStats returns statistics about the logger and rotator
func (l *Logger) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"config": l.config,
	}

	if l.rotator != nil {
		stats["rotator"] = l.rotator.GetStats()
	}

	return stats
}

// getFormatter returns the appropriate formatter based on configuration
func getFormatter(jsonFormat bool) logrus.Formatter {
	if jsonFormat {
		return &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		}
	}

	return &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		DisableColors:   false,
	}
}

// DefaultLogger creates a logger with sensible defaults
func DefaultLogger() (*Logger, error) {
	return NewLogger(LoggerConfig{
		LogLevel:      "info",
		LogFile:       "logs/app.log",
		MaxSize:       100 * 1024 * 1024,  // 100MB
		MaxAge:        7 * 24 * time.Hour, // 7 days
		MaxBackups:    5,
		Compress:      true,
		RotateDaily:   true,
		JSONFormat:    false,
		ConsoleOutput: true,
	})
}

// DevelopmentLogger creates a logger suitable for development
func DevelopmentLogger() (*Logger, error) {
	return NewLogger(LoggerConfig{
		LogLevel:      "debug",
		LogFile:       "logs/dev.log",
		MaxSize:       50 * 1024 * 1024,   // 50MB
		MaxAge:        3 * 24 * time.Hour, // 3 days
		MaxBackups:    3,
		Compress:      false,
		RotateDaily:   true,
		JSONFormat:    false,
		ConsoleOutput: true,
	})
}

// ProductionLogger creates a logger suitable for production
func ProductionLogger() (*Logger, error) {
	return NewLogger(LoggerConfig{
		LogLevel:      "info",
		LogFile:       "logs/production.log",
		MaxSize:       200 * 1024 * 1024,   // 200MB
		MaxAge:        30 * 24 * time.Hour, // 30 days
		MaxBackups:    10,
		Compress:      true,
		RotateDaily:   true,
		JSONFormat:    true,
		ConsoleOutput: false,
	})
}
