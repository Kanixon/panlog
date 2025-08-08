package panlog

import (
	"fmt"
	"log"
	"time"
)

// ExampleUsage demonstrates how to use the log rotation package
func ExampleUsage() {
	// Example 1: Basic daily rotation
	fmt.Println("=== Example 1: Basic Daily Rotation ===")
	basicLogger, err := NewLogger(LoggerConfig{
		LogLevel:      "info",
		LogFile:       "logs/basic.log",
		RotateDaily:   true,
		MaxBackups:    7,
		Compress:      true,
		ConsoleOutput: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer basicLogger.Close()

	basicLogger.Info("This is a basic daily rotation example")
	basicLogger.WithField("user", "john").Info("User logged in")

	// Example 2: Size-based rotation with compression
	fmt.Println("\n=== Example 2: Size-based Rotation ===")
	sizeLogger, err := NewLogger(LoggerConfig{
		LogLevel:      "debug",
		LogFile:       "logs/size.log",
		MaxSize:       1024 * 1024, // 1MB
		MaxBackups:    5,
		Compress:      true,
		RotateDaily:   false,
		ConsoleOutput: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sizeLogger.Close()

	// Write enough data to trigger rotation
	for i := 0; i < 1000; i++ {
		sizeLogger.Debugf("Debug message %d: This is a long message to fill up the log file quickly", i)
	}

	// Example 3: Production logger with JSON format
	fmt.Println("\n=== Example 3: Production Logger ===")
	prodLogger, err := ProductionLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer prodLogger.Close()

	prodLogger.WithFields(map[string]interface{}{
		"request_id": "req-123",
		"user_id":    "user-456",
		"action":     "login",
	}).Info("User authentication successful")

	// Example 4: Development logger
	fmt.Println("\n=== Example 4: Development Logger ===")
	devLogger, err := DevelopmentLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer devLogger.Close()

	devLogger.Debug("Debug information for development")
	devLogger.WithField("module", "auth").Debug("Authentication module debug info")

	// Example 5: Custom configuration
	fmt.Println("\n=== Example 5: Custom Configuration ===")
	customLogger, err := NewLogger(LoggerConfig{
		LogLevel:      "warn",
		LogFile:       "logs/custom.log",
		MaxSize:       5 * 1024 * 1024, // 5MB
		MaxAge:        24 * time.Hour,  // 1 day
		MaxBackups:    3,
		Compress:      false,
		RotateDaily:   true,
		JSONFormat:    false,
		ConsoleOutput: false, // Only file output
	})
	if err != nil {
		log.Fatal(err)
	}
	defer customLogger.Close()

	customLogger.Warn("This is a warning message")
	customLogger.Error("This is an error message")

	// Example 6: Manual rotation
	fmt.Println("\n=== Example 6: Manual Rotation ===")
	manualLogger, err := NewLogger(LoggerConfig{
		LogLevel:      "info",
		LogFile:       "logs/manual.log",
		RotateDaily:   false,
		MaxSize:       10 * 1024 * 1024, // 10MB
		ConsoleOutput: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer manualLogger.Close()

	manualLogger.Info("Before manual rotation")

	// Manually trigger rotation
	if err := manualLogger.Rotate(); err != nil {
		log.Printf("Manual rotation failed: %v", err)
	}

	manualLogger.Info("After manual rotation")

	// Example 7: Get statistics
	fmt.Println("\n=== Example 7: Logger Statistics ===")
	statsLogger, err := DefaultLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer statsLogger.Close()

	statsLogger.Info("Logging some messages to get statistics")
	statsLogger.Warn("Another message")

	stats := statsLogger.GetStats()
	fmt.Printf("Logger Statistics: %+v\n", stats)

	fmt.Println("\nAll examples completed successfully!")
}

// ExampleWithErrorHandling demonstrates proper error handling
func ExampleWithErrorHandling() {
	// Create logger with error handling
	logger, err := NewLogger(LoggerConfig{
		LogLevel:      "error",
		LogFile:       "logs/error.log",
		RotateDaily:   true,
		MaxBackups:    10,
		Compress:      true,
		ConsoleOutput: true,
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("Failed to close logger: %v", err)
		}
	}()

	// Log different types of errors
	logger.WithError(fmt.Errorf("database connection failed")).Error("Database error")
	logger.WithField("status_code", 500).Error("HTTP server error")
	logger.WithFields(map[string]interface{}{
		"user_id": "123",
		"action":  "purchase",
		"amount":  99.99,
	}).Error("Payment processing failed")
}

// ExampleStructuredLogging demonstrates structured logging with fields
func ExampleStructuredLogging() {
	logger, err := NewLogger(LoggerConfig{
		LogLevel:      "info",
		LogFile:       "logs/structured.log",
		RotateDaily:   true,
		JSONFormat:    true,
		ConsoleOutput: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close()

	// Create a logger with common fields
	requestLogger := logger.WithFields(map[string]interface{}{
		"service": "api",
		"version": "1.0",
	})

	// Log different types of requests
	requestLogger.WithFields(map[string]interface{}{
		"method":      "GET",
		"path":        "/users",
		"status":      200,
		"duration_ms": 45,
	}).Info("API request completed")

	requestLogger.WithFields(map[string]interface{}{
		"method":      "POST",
		"path":        "/orders",
		"status":      201,
		"duration_ms": 120,
		"user_id":     "user-123",
	}).Info("Order created successfully")
}
