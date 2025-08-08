package panlog

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// LogRotator handles automatic log file rotation
type LogRotator struct {
	mu sync.Mutex

	// Configuration
	filePath    string
	maxSize     int64
	maxAge      time.Duration
	maxBackups  int
	compress    bool
	rotateDaily bool
	rotateTime  time.Time

	// Current file handle
	file     *os.File
	fileSize int64
}

// LogRotatorConfig holds configuration for log rotation
type LogRotatorConfig struct {
	FilePath    string        // Path to the log file
	MaxSize     int64         // Maximum size in bytes before rotation
	MaxAge      time.Duration // Maximum age of log files to keep
	MaxBackups  int           // Maximum number of old log files to keep
	Compress    bool          // Whether to compress old log files
	RotateDaily bool          // Whether to rotate daily regardless of size
}

// NewLogRotator creates a new log rotator with the given configuration
func NewLogRotator(config LogRotatorConfig) (*LogRotator, error) {
	if config.FilePath == "" {
		return nil, fmt.Errorf("file path is required")
	}

	// Set defaults
	if config.MaxSize == 0 {
		config.MaxSize = 100 * 1024 * 1024 // 100MB default
	}
	if config.MaxAge == 0 {
		config.MaxAge = 7 * 24 * time.Hour // 7 days default
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 5 // 5 backups default
	}

	lr := &LogRotator{
		filePath:    config.FilePath,
		maxSize:     config.MaxSize,
		maxAge:      config.MaxAge,
		maxBackups:  config.MaxBackups,
		compress:    config.Compress,
		rotateDaily: config.RotateDaily,
		rotateTime:  time.Now().Truncate(24 * time.Hour),
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open the current log file
	if err := lr.openFile(); err != nil {
		return nil, err
	}

	return lr, nil
}

// Write implements io.Writer interface
func (lr *LogRotator) Write(p []byte) (n int, err error) {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	// Check if we need to rotate
	if err := lr.checkRotation(); err != nil {
		return 0, err
	}

	// Write to file
	n, err = lr.file.Write(p)
	if err != nil {
		return n, err
	}

	lr.fileSize += int64(n)
	return n, nil
}

// Close closes the log rotator and the underlying file
func (lr *LogRotator) Close() error {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	if lr.file != nil {
		return lr.file.Close()
	}
	return nil
}

// Rotate manually triggers a log rotation
func (lr *LogRotator) Rotate() error {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	return lr.rotate()
}

// checkRotation checks if rotation is needed and performs it
func (lr *LogRotator) checkRotation() error {
	now := time.Now()

	// Check daily rotation
	if lr.rotateDaily {
		currentDay := now.Truncate(24 * time.Hour)
		if currentDay.After(lr.rotateTime) {
			if err := lr.rotate(); err != nil {
				return err
			}
			lr.rotateTime = currentDay
			return nil
		}
	}

	// Check size-based rotation
	if lr.fileSize >= lr.maxSize {
		return lr.rotate()
	}

	return nil
}

// rotate performs the actual log rotation
func (lr *LogRotator) rotate() error {
	if lr.file == nil {
		return lr.openFile()
	}

	// Close current file
	if err := lr.file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	// Generate rotated filename
	rotatedName := lr.generateRotatedName()
	if rotatedName == "" {
		return fmt.Errorf("failed to generate rotated filename")
	}

	// Rename current file to rotated name
	if err := os.Rename(lr.filePath, rotatedName); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}

	// Compress if enabled
	if lr.compress {
		if err := lr.compressFile(rotatedName); err != nil {
			return fmt.Errorf("failed to compress log file: %w", err)
		}
	}

	// Clean up old files
	if err := lr.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup old files: %w", err)
	}

	// Open new file
	return lr.openFile()
}

// generateRotatedName generates the name for the rotated log file
func (lr *LogRotator) generateRotatedName() string {
	now := time.Now()
	ext := filepath.Ext(lr.filePath)
	base := strings.TrimSuffix(lr.filePath, ext)

	// Format: filename-YYYY-MM-DD-HHMMSS.ext
	timestamp := now.Format("2006-01-02-150405")
	return fmt.Sprintf("%s-%s%s", base, timestamp, ext)
}

// openFile opens the current log file
func (lr *LogRotator) openFile() error {
	file, err := os.OpenFile(lr.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to get file stats: %w", err)
	}

	lr.file = file
	lr.fileSize = stat.Size()
	return nil
}

// compressFile compresses a log file using gzip
func (lr *LogRotator) compressFile(filename string) error {
	// Skip if already compressed
	if strings.HasSuffix(filename, ".gz") {
		return nil
	}

	// Open source file
	source, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer source.Close()

	// Create compressed file
	compressed, err := os.Create(filename + ".gz")
	if err != nil {
		return err
	}
	defer compressed.Close()

	// Create gzip writer
	gw := gzip.NewWriter(compressed)
	defer gw.Close()

	// Copy content
	if _, err := io.Copy(gw, source); err != nil {
		return err
	}

	// Remove original file
	return os.Remove(filename)
}

// cleanup removes old log files based on age and count
func (lr *LogRotator) cleanup() error {
	dir := filepath.Dir(lr.filePath)
	base := filepath.Base(lr.filePath)
	ext := filepath.Ext(base)
	baseWithoutExt := strings.TrimSuffix(base, ext)

	// Find all log files
	pattern := filepath.Join(dir, baseWithoutExt+"-*"+ext+"*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// Sort files by modification time (oldest first)
	var files []struct {
		path    string
		modTime time.Time
	}

	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			continue
		}
		files = append(files, struct {
			path    string
			modTime time.Time
		}{match, stat.ModTime()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// Remove files based on age
	cutoff := time.Now().Add(-lr.maxAge)
	for _, file := range files {
		if file.modTime.Before(cutoff) {
			os.Remove(file.path)
		}
	}

	// Remove files based on count
	if len(files) > lr.maxBackups {
		toRemove := len(files) - lr.maxBackups
		for i := 0; i < toRemove && i < len(files); i++ {
			os.Remove(files[i].path)
		}
	}

	return nil
}

// GetStats returns current statistics about the log rotator
func (lr *LogRotator) GetStats() map[string]interface{} {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	stats := map[string]interface{}{
		"file_path":    lr.filePath,
		"file_size":    lr.fileSize,
		"max_size":     lr.maxSize,
		"max_age":      lr.maxAge,
		"max_backups":  lr.maxBackups,
		"compress":     lr.compress,
		"rotate_daily": lr.rotateDaily,
		"rotate_time":  lr.rotateTime,
	}

	if lr.file != nil {
		stats["file_open"] = true
	} else {
		stats["file_open"] = false
	}

	return stats
}
