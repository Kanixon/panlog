package panlog

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLogRotator(t *testing.T) {
	// Test with valid configuration
	config := LogRotatorConfig{
		FilePath:    "testdata/test.log",
		MaxSize:     1024,
		MaxAge:      time.Hour,
		MaxBackups:  3,
		Compress:    true,
		RotateDaily: true,
	}

	lr, err := NewLogRotator(config)
	if err != nil {
		t.Fatalf("Failed to create log rotator: %v", err)
	}
	defer lr.Close()

	// Test with empty file path
	_, err = NewLogRotator(LogRotatorConfig{})
	if err == nil {
		t.Error("Expected error for empty file path")
	}
}

func TestLogRotatorWrite(t *testing.T) {
	config := LogRotatorConfig{
		FilePath:    "testdata/write_test.log",
		MaxSize:     100,
		MaxAge:      time.Hour,
		MaxBackups:  3,
		Compress:    false,
		RotateDaily: false,
	}

	lr, err := NewLogRotator(config)
	if err != nil {
		t.Fatalf("Failed to create log rotator: %v", err)
	}
	defer lr.Close()

	// Write data
	data := []byte("test log message\n")
	n, err := lr.Write(data)
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, got %d", len(data), n)
	}

	// Check file size
	stat, err := os.Stat(config.FilePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if stat.Size() != int64(len(data)) {
		t.Errorf("Expected file size %d, got %d", len(data), stat.Size())
	}
}

func TestLogRotatorRotation(t *testing.T) {
	config := LogRotatorConfig{
		FilePath:    "testdata/rotation_test.log",
		MaxSize:     50,
		MaxAge:      time.Hour,
		MaxBackups:  2,
		Compress:    false,
		RotateDaily: false,
	}

	lr, err := NewLogRotator(config)
	if err != nil {
		t.Fatalf("Failed to create log rotator: %v", err)
	}
	defer lr.Close()

	// Write enough data to trigger rotation
	longMessage := "This is a very long message that should trigger rotation when written multiple times\n"
	for i := 0; i < 10; i++ {
		_, err := lr.Write([]byte(longMessage))
		if err != nil {
			t.Fatalf("Failed to write: %v", err)
		}
	}

	// Check if rotation occurred
	dir := filepath.Dir(config.FilePath)
	base := filepath.Base(config.FilePath)
	ext := filepath.Ext(base)
	baseWithoutExt := base[:len(base)-len(ext)]

	pattern := filepath.Join(dir, baseWithoutExt+"-*"+ext)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Failed to glob pattern: %v", err)
	}

	if len(matches) == 0 {
		t.Error("Expected rotated files, but none found")
	}
}

func TestLogRotatorCompression(t *testing.T) {
	config := LogRotatorConfig{
		FilePath:    "testdata/compression_test.log",
		MaxSize:     50,
		MaxAge:      time.Hour,
		MaxBackups:  2,
		Compress:    true,
		RotateDaily: false,
	}

	lr, err := NewLogRotator(config)
	if err != nil {
		t.Fatalf("Failed to create log rotator: %v", err)
	}
	defer lr.Close()

	// Write data to trigger rotation
	longMessage := "This is a very long message that should trigger rotation and compression\n"
	for i := 0; i < 10; i++ {
		_, err := lr.Write([]byte(longMessage))
		if err != nil {
			t.Fatalf("Failed to write: %v", err)
		}
	}

	// Check if compressed files exist
	dir := filepath.Dir(config.FilePath)
	base := filepath.Base(config.FilePath)
	ext := filepath.Ext(base)
	baseWithoutExt := base[:len(base)-len(ext)]

	pattern := filepath.Join(dir, baseWithoutExt+"-*"+ext+".gz")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Failed to glob pattern: %v", err)
	}

	if len(matches) == 0 {
		t.Error("Expected compressed files, but none found")
	}
}

func TestNewLogger(t *testing.T) {
	// Test with valid configuration
	config := LoggerConfig{
		LogLevel:      "info",
		LogFile:       "testdata/logger_test.log",
		MaxSize:       1024,
		MaxAge:        time.Hour,
		MaxBackups:    3,
		Compress:      true,
		RotateDaily:   true,
		JSONFormat:    false,
		ConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test logging
	logger.Info("Test log message")
	logger.WithField("test", "value").Info("Structured log message")
}

func TestDefaultLogger(t *testing.T) {
	logger, err := DefaultLogger()
	if err != nil {
		t.Fatalf("Failed to create default logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Default logger test")
}

func TestDevelopmentLogger(t *testing.T) {
	logger, err := DevelopmentLogger()
	if err != nil {
		t.Fatalf("Failed to create development logger: %v", err)
	}
	defer logger.Close()

	logger.Debug("Development logger debug message")
	logger.Info("Development logger info message")
}

func TestProductionLogger(t *testing.T) {
	logger, err := ProductionLogger()
	if err != nil {
		t.Fatalf("Failed to create production logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Production logger test")
	logger.WithFields(map[string]interface{}{
		"request_id": "test-123",
		"user_id":    "user-456",
	}).Info("Production structured log")
}

func TestLoggerRotation(t *testing.T) {
	config := LoggerConfig{
		LogLevel:      "debug",
		LogFile:       "testdata/logger_rotation_test.log",
		MaxSize:       100,
		MaxAge:        time.Hour,
		MaxBackups:    2,
		Compress:      false,
		RotateDaily:   false,
		JSONFormat:    false,
		ConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Write enough data to trigger rotation
	for i := 0; i < 20; i++ {
		logger.Debugf("Debug message %d: This is a long message to trigger rotation", i)
	}

	// Check if rotation occurred
	dir := filepath.Dir(config.LogFile)
	base := filepath.Base(config.LogFile)
	ext := filepath.Ext(base)
	baseWithoutExt := base[:len(base)-len(ext)]

	pattern := filepath.Join(dir, baseWithoutExt+"-*"+ext)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Failed to glob pattern: %v", err)
	}

	if len(matches) == 0 {
		t.Error("Expected rotated files, but none found")
	}
}

func TestLoggerStats(t *testing.T) {
	config := LoggerConfig{
		LogLevel:      "info",
		LogFile:       "testdata/stats_test.log",
		MaxSize:       1024,
		MaxAge:        time.Hour,
		MaxBackups:    3,
		Compress:      false,
		RotateDaily:   false,
		JSONFormat:    false,
		ConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log some messages
	logger.Info("Test message 1")
	logger.Warn("Test message 2")

	// Get stats
	stats := logger.GetStats()
	if stats == nil {
		t.Error("Expected stats, got nil")
	}

	// Check if config is in stats
	if configStats, ok := stats["config"]; !ok {
		t.Error("Expected config in stats")
	} else {
		if configStats == nil {
			t.Error("Expected config stats, got nil")
		}
	}
}

func TestManualRotation(t *testing.T) {
	config := LoggerConfig{
		LogLevel:      "info",
		LogFile:       "testdata/manual_rotation_test.log",
		MaxSize:       1024 * 1024, // 1MB
		MaxAge:        time.Hour,
		MaxBackups:    3,
		Compress:      false,
		RotateDaily:   false,
		JSONFormat:    false,
		ConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log initial message
	logger.Info("Before manual rotation")

	// Manually trigger rotation
	err = logger.Rotate()
	if err != nil {
		t.Fatalf("Manual rotation failed: %v", err)
	}

	// Log message after rotation
	logger.Info("After manual rotation")

	// Check if rotation occurred
	dir := filepath.Dir(config.LogFile)
	base := filepath.Base(config.LogFile)
	ext := filepath.Ext(base)
	baseWithoutExt := base[:len(base)-len(ext)]

	pattern := filepath.Join(dir, baseWithoutExt+"-*"+ext)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Failed to glob pattern: %v", err)
	}

	if len(matches) == 0 {
		t.Error("Expected rotated files after manual rotation, but none found")
	}
}

// Cleanup function to remove test files
func TestMain(m *testing.M) {
	// Create testdata directory
	os.MkdirAll("testdata", 0755)

	// Run tests
	code := m.Run()

	// Cleanup test files
	os.RemoveAll("testdata")
	os.RemoveAll("logs")

	os.Exit(code)
}
