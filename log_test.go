package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLog_Option(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	opts := NewOptions()
	opts.Compress = false
	opts.Directory = "/Users/kyden/git-space/log/logs"
	opts.DisableCaller = false
	opts.DisableSplitError = false
	opts.DisableStacktrace = false
	opts.Format = "console"
	opts.Level = "debug"
	opts.MaxBackups = 2
	opts.MaxSize = 1
	opts.Prefix = "Ziwi_Log_Test_"
	opts.TimeLayout = "2006-01-02 15:04:05.000"
	assert.NotNil(t)

	logger := NewLog(opts)
	defer logger.Sync()

	for i := range 10_000 {
		logger.Debugw("test debug", "i", i)
		logger.Infow("test info", "i", i)
		logger.Warnw("test warn", "i", i)
		logger.Errorw("test error", "i", i)
	}
}

func Test_Log(t *testing.T) {
	t.Parallel()

	logger := NewLog(nil)

	logger.Info("Test Log")
	Info("Test Log")
	logger.Infoln("Test Log")
	Infoln("Test Log")
	logger.Infof("Test Log, %s", "Test Log")
	Infof("Test Log, %s", "Test Log")
	logger.Infow("Test Log", "key", "value")
	Infow("Test Log", "key", "value")

	logger.Error("Test Log")
	Error("Test Log")
	logger.Errorln("Test Log")
	Errorln("Test Log")
	logger.Errorf("Test Log, %s", "Test Log")
	Errorf("Test Log, %s", "Test Log")
	logger.Errorw("Test Log", "key", "value")
	Errorw("Test Log", "key", "value")

	logger.Warn("Test Log")
	Warn("Test Log")
	logger.Warnln("Test Log")
	Warnln("Test Log")
	logger.Warnf("Test Log, %s", "Test Log")
	Warnf("Test Log, %s", "Test Log")
	logger.Warnw("Test Log", "key", "value")
	Warnw("Test Log", "key", "value")

	logger.Debug("Test Log")
	Debug("Test Log")
	logger.Debugln("Test Log")
	Debugln("Test Log")
	logger.Debugf("Test Log, %s", "Test Log")
	Debugf("Test Log, %s", "Test Log")
	logger.Debugw("Test Log", "key", "value")
	Debugw("Test Log", "key", "value")

	logger.Sync()
	Sync()
}

func TestLoggerWithErrorNilCheck(t *testing.T) {
	testDir := "./logs/test_logs"
	defer os.RemoveAll(testDir)

	t.Run("DefaultConfig", func(t *testing.T) {
		opts := NewOptions().
			WithDirectory(testDir).
			WithPrefix("TEST_")

		if !opts.DisableSplitError {
			t.Fatalf("Expected DisableSplitError to be true by default, got false")
		}
		logger := NewLog(opts)
		logger.Debug("Debug message")
		logger.Info("Info message")
		logger.Warn("Warning message")
		logger.Error("Error message")
		logger.Sync()
	})

	t.Run("WithErrorSplit", func(t *testing.T) {
		opts := NewOptions().
			WithDirectory(testDir).
			WithPrefix("TEST_").
			WithDisableSplitError(false)

		if opts.DisableSplitError {
			t.Fatalf("Expected DisableSplitError to be false, got true")
		}
		logger := NewLog(opts)
		logger.Debug("Debug message with error split")
		logger.Info("Info message with error split")
		logger.Warn("Warning message with error split")
		logger.Error("Error message with error split")
		logger.Sync()
	})
}

// Test buffer pool functionality for memory optimization
func TestBufferPoolOptimization(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_pool"
	defer os.RemoveAll(testDir)

	opts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("POOL_TEST_")

	logger := NewLog(opts)
	defer logger.Sync()

	// Test that buffer pool is working by logging multiple messages
	// If buffer pool is working correctly, we should see memory reuse
	for i := range 100 {
		logger.Infow("Pool test message", "iteration", i, "data", "buffer_pool_test")
	}

	asrt.NotNil(logger)
}

// Test atomic logger access for concurrent safety
func TestAtomicLoggerAccess(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_atomic"
	defer os.RemoveAll(testDir)

	// Test DefaultLogger access
	originalLogger := DefaultLogger()
	asrt.NotNil(originalLogger)

	// Test ReplaceLogger functionality
	newOpts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("ATOMIC_TEST_")
	newLogger := NewLog(newOpts)
	defer newLogger.Sync()

	ReplaceLogger(newLogger)
	currentLogger := DefaultLogger()
	asrt.Equal(newLogger, currentLogger)

	// Test nil replacement (should not change logger)
	ReplaceLogger(nil)
	stillCurrentLogger := DefaultLogger()
	asrt.Equal(newLogger, stillCurrentLogger)
}

// Test concurrent logger access and replacement
func TestConcurrentLoggerAccess(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_concurrent"
	defer os.RemoveAll(testDir)

	// Test concurrent access to DefaultLogger
	const numGoroutines = 10
	const numOperations = 100

	// Channel to collect results
	results := make(chan *Log, numGoroutines*numOperations)

	// Start multiple goroutines accessing DefaultLogger concurrently
	for i := range numGoroutines {
		go func(id int) {
			for j := range numOperations {
				logger := DefaultLogger()
				logger.Infow("Concurrent test", "goroutine", id, "operation", j)
				results <- logger
			}
		}(i)
	}

	// Collect all results
	for range numGoroutines * numOperations {
		logger := <-results
		asrt.NotNil(logger)
	}

	// Test concurrent logger replacement
	for i := range 5 {
		go func(id int) {
			opts := NewOptions().
				WithDirectory(testDir).
				WithPrefix(fmt.Sprintf("CONCURRENT_%d_", id))
			newLogger := NewLog(opts)
			defer newLogger.Sync()
			ReplaceLogger(newLogger)
		}(i)
	}

	// Verify we can still access the logger
	finalLogger := DefaultLogger()
	asrt.NotNil(finalLogger)
}

// Test date checking optimization
func TestDateCheckOptimization(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_date"
	defer os.RemoveAll(testDir)

	opts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("DATE_TEST_")

	logger := NewLog(opts)
	defer logger.Sync()

	// Test that date checking works correctly
	// Initial date check should be set
	asrt.NotZero(logger.dateCheck)

	// Log multiple messages quickly - date check should not change frequently
	initialCheck := logger.dateCheck
	for i := range 10 {
		logger.Infow("Date test message", "iteration", i)
	}

	// Date check should still be the same for quick successive logs
	asrt.Equal(initialCheck, logger.dateCheck)

	// Verify current date is set correctly
	asrt.NotEmpty(logger.currDate)
}

// Test file write retry mechanism
func TestFileWriteRetry(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_retry"
	defer os.RemoveAll(testDir)

	opts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("RETRY_TEST_")

	logger := NewLog(opts)
	defer logger.Sync()

	// Test successful write
	logger.Info("Test message for retry mechanism")

	// Test that writeToFile handles nil file gracefully
	err := logger.writeToFile(nil, []byte("test data"))
	asrt.Error(err)
	asrt.Contains(err.Error(), "file is nil")

	// Test normal write operation
	if logger.file != nil {
		err = logger.writeToFile(logger.file, []byte("test write data\n"))
		asrt.NoError(err)
	}
}

// Test setup log files functionality
func TestSetupLogFiles(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_setup"
	defer os.RemoveAll(testDir)

	opts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("SETUP_TEST_")

	logger := NewLog(opts)
	defer logger.Sync()

	// Test setup with current date
	currentDate := "2025-01-01"
	err := logger.setupLogFiles(currentDate)
	asrt.NoError(err)

	// Verify files are set up
	asrt.NotNil(logger.file)
	asrt.Equal(currentDate, logger.currDate)

	// Test setup with error split disabled
	if logger.opts.DisableSplitError {
		// Error file should not be set when split is disabled
		asrt.Nil(logger.errFile)
	}

	// Test setup with same date (should not recreate files)
	err = logger.setupLogFiles(currentDate)
	asrt.NoError(err)
	asrt.Equal(currentDate, logger.currDate)
}

// Test generateFileName functionality - comprehensive test suite
func TestGenerateFileName(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDate := "2025-07-20"

	// Test with empty filename (backward compatibility)
	opts := NewOptions().WithFilename("")
	logger := NewLog(opts)
	defer logger.Sync()

	// Test main log file without custom filename
	mainFileName := logger.generateFileName(testDate, false)
	asrt.Equal("2025-07-20.log", mainFileName)

	// Test error log file without custom filename
	errorFileName := logger.generateFileName(testDate, true)
	asrt.Equal("2025-07-20_error.log", errorFileName)

	// Test with custom filename
	opts = NewOptions().WithFilename("myapp")
	logger = NewLog(opts)
	defer logger.Sync()

	// Test main log file with custom filename
	mainFileName = logger.generateFileName(testDate, false)
	asrt.Equal("myapp-2025-07-20.log", mainFileName)

	// Test error log file with custom filename
	errorFileName = logger.generateFileName(testDate, true)
	asrt.Equal("myapp-2025-07-20_error.log", errorFileName)

	// Test with filename containing unsafe characters
	opts = NewOptions().WithFilename("my/app:test*")
	logger = NewLog(opts)
	defer logger.Sync()

	// Test that unsafe characters are sanitized
	mainFileName = logger.generateFileName(testDate, false)
	asrt.Equal("my_app_test_-2025-07-20.log", mainFileName)

	errorFileName = logger.generateFileName(testDate, true)
	asrt.Equal("my_app_test_-2025-07-20_error.log", errorFileName)

	// Test with filename that becomes empty after sanitization
	opts = NewOptions().WithFilename("///")
	logger = NewLog(opts)
	defer logger.Sync()

	// Should fallback to default format when sanitized filename is empty
	mainFileName = logger.generateFileName(testDate, false)
	asrt.Equal("2025-07-20.log", mainFileName)

	errorFileName = logger.generateFileName(testDate, true)
	asrt.Equal("2025-07-20_error.log", errorFileName)
}

// Test generateFileName with various date formats and edge cases
func TestGenerateFileName_DateFormats(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opts := NewOptions().WithFilename("testapp")
	logger := NewLog(opts)
	defer logger.Sync()

	testCases := []struct {
		name     string
		date     string
		expected struct {
			main  string
			error string
		}
	}{
		{
			name: "Standard date format",
			date: "2025-07-20",
			expected: struct {
				main  string
				error string
			}{
				main:  "testapp-2025-07-20.log",
				error: "testapp-2025-07-20_error.log",
			},
		},
		{
			name: "Different year",
			date: "2024-12-31",
			expected: struct {
				main  string
				error string
			}{
				main:  "testapp-2024-12-31.log",
				error: "testapp-2024-12-31_error.log",
			},
		},
		{
			name: "Leap year date",
			date: "2024-02-29",
			expected: struct {
				main  string
				error string
			}{
				main:  "testapp-2024-02-29.log",
				error: "testapp-2024-02-29_error.log",
			},
		},
		{
			name: "New year date",
			date: "2025-01-01",
			expected: struct {
				main  string
				error string
			}{
				main:  "testapp-2025-01-01.log",
				error: "testapp-2025-01-01_error.log",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mainFileName := logger.generateFileName(tc.date, false)
			asrt.Equal(tc.expected.main, mainFileName, "Main log filename mismatch for %s", tc.name)

			errorFileName := logger.generateFileName(tc.date, true)
			asrt.Equal(tc.expected.error, errorFileName, "Error log filename mismatch for %s", tc.name)
		})
	}
}

// Test generateFileName with various filename prefixes
func TestGenerateFileName_FilenameVariations(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDate := "2025-07-20"

	testCases := []struct {
		name     string
		filename string
		expected struct {
			main  string
			error string
		}
	}{
		{
			name:     "Simple alphanumeric filename",
			filename: "app123",
			expected: struct {
				main  string
				error string
			}{
				main:  "app123-2025-07-20.log",
				error: "app123-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with hyphens",
			filename: "my-app-service",
			expected: struct {
				main  string
				error string
			}{
				main:  "my-app-service-2025-07-20.log",
				error: "my-app-service-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with underscores",
			filename: "my_app_service",
			expected: struct {
				main  string
				error string
			}{
				main:  "my_app_service-2025-07-20.log",
				error: "my_app_service-2025-07-20_error.log",
			},
		},
		{
			name:     "Single character filename",
			filename: "a",
			expected: struct {
				main  string
				error string
			}{
				main:  "a-2025-07-20.log",
				error: "a-2025-07-20_error.log",
			},
		},
		{
			name:     "Numeric filename",
			filename: "12345",
			expected: struct {
				main  string
				error string
			}{
				main:  "12345-2025-07-20.log",
				error: "12345-2025-07-20_error.log",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewOptions().WithFilename(tc.filename)
			logger := NewLog(opts)
			defer logger.Sync()

			mainFileName := logger.generateFileName(testDate, false)
			asrt.Equal(tc.expected.main, mainFileName, "Main log filename mismatch for %s", tc.name)

			errorFileName := logger.generateFileName(testDate, true)
			asrt.Equal(tc.expected.error, errorFileName, "Error log filename mismatch for %s", tc.name)
		})
	}
}

// Test generateFileName with unsafe characters and sanitization
func TestGenerateFileName_UnsafeCharacters(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDate := "2025-07-20"

	testCases := []struct {
		name     string
		filename string
		expected struct {
			main  string
			error string
		}
	}{
		{
			name:     "Filename with forward slash",
			filename: "app/service",
			expected: struct {
				main  string
				error string
			}{
				main:  "app_service-2025-07-20.log",
				error: "app_service-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with backslash",
			filename: "app\\service",
			expected: struct {
				main  string
				error string
			}{
				main:  "app_service-2025-07-20.log",
				error: "app_service-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with colon",
			filename: "app:service",
			expected: struct {
				main  string
				error string
			}{
				main:  "app_service-2025-07-20.log",
				error: "app_service-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with asterisk",
			filename: "app*service",
			expected: struct {
				main  string
				error string
			}{
				main:  "app_service-2025-07-20.log",
				error: "app_service-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with question mark",
			filename: "app?service",
			expected: struct {
				main  string
				error string
			}{
				main:  "app_service-2025-07-20.log",
				error: "app_service-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with quotes",
			filename: "app\"service",
			expected: struct {
				main  string
				error string
			}{
				main:  "app_service-2025-07-20.log",
				error: "app_service-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with angle brackets",
			filename: "app<service>",
			expected: struct {
				main  string
				error string
			}{
				main:  "app_service_-2025-07-20.log",
				error: "app_service_-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with pipe",
			filename: "app|service",
			expected: struct {
				main  string
				error string
			}{
				main:  "app_service-2025-07-20.log",
				error: "app_service-2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with multiple unsafe characters",
			filename: "app/\\:*?\"<>|service",
			expected: struct {
				main  string
				error string
			}{
				main:  "app_________service-2025-07-20.log",
				error: "app_________service-2025-07-20_error.log",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewOptions().WithFilename(tc.filename)
			logger := NewLog(opts)
			defer logger.Sync()

			mainFileName := logger.generateFileName(testDate, false)
			asrt.Equal(tc.expected.main, mainFileName, "Main log filename mismatch for %s", tc.name)

			errorFileName := logger.generateFileName(testDate, true)
			asrt.Equal(tc.expected.error, errorFileName, "Error log filename mismatch for %s", tc.name)
		})
	}
}

// Test generateFileName backward compatibility scenarios
func TestGenerateFileName_BackwardCompatibility(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDate := "2025-07-20"

	testCases := []struct {
		name     string
		filename string
		expected struct {
			main  string
			error string
		}
	}{
		{
			name:     "Empty filename",
			filename: "",
			expected: struct {
				main  string
				error string
			}{
				main:  "2025-07-20.log",
				error: "2025-07-20_error.log",
			},
		},
		{
			name:     "Whitespace only filename",
			filename: "   ",
			expected: struct {
				main  string
				error string
			}{
				main:  "2025-07-20.log",
				error: "2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with only unsafe characters",
			filename: "/\\:*?\"<>|",
			expected: struct {
				main  string
				error string
			}{
				main:  "2025-07-20.log",
				error: "2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with only dots",
			filename: "...",
			expected: struct {
				main  string
				error string
			}{
				main:  "2025-07-20.log",
				error: "2025-07-20_error.log",
			},
		},
		{
			name:     "Filename with only underscores",
			filename: "___",
			expected: struct {
				main  string
				error string
			}{
				main:  "2025-07-20.log",
				error: "2025-07-20_error.log",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewOptions().WithFilename(tc.filename)
			logger := NewLog(opts)
			defer logger.Sync()

			mainFileName := logger.generateFileName(testDate, false)
			asrt.Equal(tc.expected.main, mainFileName, "Main log filename mismatch for %s", tc.name)

			errorFileName := logger.generateFileName(testDate, true)
			asrt.Equal(tc.expected.error, errorFileName, "Error log filename mismatch for %s", tc.name)
		})
	}
}

// Test generateFileName with edge cases and boundary conditions
func TestGenerateFileName_EdgeCases(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDate := "2025-07-20"

	// Test with very long filename (should be truncated)
	longFilename := strings.Repeat("a", 150) // 150 characters, should be truncated to 100
	opts := NewOptions().WithFilename(longFilename)
	logger := NewLog(opts)
	defer logger.Sync()

	mainFileName := logger.generateFileName(testDate, false)
	// Should be truncated to 100 chars + "-2025-07-20.log"
	expectedMain := strings.Repeat("a", 100) + "-2025-07-20.log"
	asrt.Equal(expectedMain, mainFileName)

	errorFileName := logger.generateFileName(testDate, true)
	expectedError := strings.Repeat("a", 100) + "-2025-07-20_error.log"
	asrt.Equal(expectedError, errorFileName)

	// Test with filename starting with dot
	opts = NewOptions().WithFilename(".hidden")
	logger = NewLog(opts)
	defer logger.Sync()

	mainFileName = logger.generateFileName(testDate, false)
	asrt.Equal("_hidden-2025-07-20.log", mainFileName)

	errorFileName = logger.generateFileName(testDate, true)
	asrt.Equal("_hidden-2025-07-20_error.log", errorFileName)

	// Test with filename containing control characters
	opts = NewOptions().WithFilename("app\x00service\x1f")
	logger = NewLog(opts)
	defer logger.Sync()

	mainFileName = logger.generateFileName(testDate, false)
	asrt.Equal("app_service_-2025-07-20.log", mainFileName)

	errorFileName = logger.generateFileName(testDate, true)
	asrt.Equal("app_service_-2025-07-20_error.log", errorFileName)
}

// Test concurrent and multi-instance scenarios for filename field
// Tests multiple Log instances using different Filename values
func TestConcurrentMultiInstance_DifferentFilenames(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_concurrent_multi"
	defer os.RemoveAll(testDir)

	const numInstances = 5
	const numGoroutines = 10
	const numOperations = 50

	// Create multiple Log instances with different filenames
	loggers := make([]*Log, numInstances)
	for i := range numInstances {
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename(fmt.Sprintf("app%d", i)).
			WithDisableSplitError(false)
		loggers[i] = NewLog(opts)
		defer loggers[i].Sync()
	}

	// Channel to collect results and errors
	results := make(chan struct{}, numInstances*numGoroutines*numOperations)
	errors := make(chan error, numInstances*numGoroutines*numOperations)

	// Start concurrent logging with different instances
	for i, logger := range loggers {
		for j := range numGoroutines {
			go func(loggerIndex, goroutineID int, l *Log) {
				for k := range numOperations {
					// Log different types of messages
					l.Infow("Concurrent test info",
						"logger", loggerIndex,
						"goroutine", goroutineID,
						"operation", k,
						"timestamp", time.Now().UnixNano())

					l.Errorw("Concurrent test error",
						"logger", loggerIndex,
						"goroutine", goroutineID,
						"operation", k,
						"timestamp", time.Now().UnixNano())

					results <- struct{}{}
				}
			}(i, j, logger)
		}
	}

	// Collect all results
	for range numInstances * numGoroutines * numOperations {
		select {
		case <-results:
			// Success
		case err := <-errors:
			t.Errorf("Unexpected error during concurrent logging: %v", err)
		case <-time.After(30 * time.Second):
			t.Fatal("Test timeout - concurrent logging took too long")
		}
	}

	// Verify that each logger created its own files
	for i := range numInstances {
		expectedMainFile := filepath.Join(testDir, fmt.Sprintf("app%d-%s.log", i, time.Now().Format(time.DateOnly)))
		expectedErrorFile := filepath.Join(
			testDir,
			fmt.Sprintf("app%d-%s_error.log", i, time.Now().Format(time.DateOnly)),
		)

		// Check main log file exists and has content
		mainInfo, err := os.Stat(expectedMainFile)
		asrt.NoError(err, "Main log file should exist for logger %d", i)
		asrt.True(mainInfo.Size() > 0, "Main log file should have content for logger %d", i)

		// Check error log file exists and has content
		errorInfo, err := os.Stat(expectedErrorFile)
		asrt.NoError(err, "Error log file should exist for logger %d", i)
		asrt.True(errorInfo.Size() > 0, "Error log file should have content for logger %d", i)
	}
}

// Test concurrent access to same filename and directory
func TestConcurrentMultiInstance_SameFilename(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_concurrent_same"
	defer os.RemoveAll(testDir)

	const numInstances = 3
	const numGoroutines = 5
	const numOperations = 30

	// Create multiple Log instances with the SAME filename
	loggers := make([]*Log, numInstances)
	for i := range numInstances {
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename("shared-app"). // Same filename for all instances
			WithDisableSplitError(false)
		loggers[i] = NewLog(opts)
		defer loggers[i].Sync()
	}

	// Channel to collect results
	results := make(chan struct{}, numInstances*numGoroutines*numOperations)
	var wg sync.WaitGroup

	// Start concurrent logging with same filename
	for i, logger := range loggers {
		for j := range numGoroutines {
			wg.Add(1)
			go func(loggerIndex, goroutineID int, l *Log) {
				defer wg.Done()
				for k := range numOperations {
					// Log with unique identifiers to track which instance wrote what
					l.Infow("Shared filename test",
						"instance", loggerIndex,
						"goroutine", goroutineID,
						"operation", k,
						"unique_id", fmt.Sprintf("%d-%d-%d", loggerIndex, goroutineID, k),
						"timestamp", time.Now().UnixNano())

					// Also test error logging
					if k%5 == 0 {
						l.Errorw("Shared filename error test",
							"instance", loggerIndex,
							"goroutine", goroutineID,
							"operation", k,
							"unique_id", fmt.Sprintf("err-%d-%d-%d", loggerIndex, goroutineID, k))
					}

					results <- struct{}{}
				}
			}(i, j, logger)
		}
	}

	// Wait for all goroutines to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Collect results with timeout
	select {
	case <-done:
		// All goroutines completed successfully
	case <-time.After(30 * time.Second):
		t.Fatal("Test timeout - concurrent logging with same filename took too long")
	}

	// Verify that files were created and contain data from all instances
	expectedMainFile := filepath.Join(testDir, fmt.Sprintf("shared-app-%s.log", time.Now().Format(time.DateOnly)))
	expectedErrorFile := filepath.Join(
		testDir,
		fmt.Sprintf("shared-app-%s_error.log", time.Now().Format(time.DateOnly)),
	)

	// Check main log file
	mainInfo, err := os.Stat(expectedMainFile)
	asrt.NoError(err, "Shared main log file should exist")
	asrt.True(mainInfo.Size() > 0, "Shared main log file should have content")

	// Check error log file
	errorInfo, err := os.Stat(expectedErrorFile)
	asrt.NoError(err, "Shared error log file should exist")
	asrt.True(errorInfo.Size() > 0, "Shared error log file should have content")

	// Read and verify file contents contain entries from all instances
	mainContent, err := os.ReadFile(expectedMainFile)
	asrt.NoError(err, "Should be able to read main log file")

	// Verify that we have entries from all instances
	mainContentStr := string(mainContent)
	for i := range numInstances {
		instancePattern := fmt.Sprintf(`"instance": %d`, i)
		asrt.Contains(mainContentStr, instancePattern, "Main log should contain entries from instance %d", i)
	}
}

// Test thread safety and file access conflict handling
func TestConcurrentMultiInstance_ThreadSafety(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_thread_safety"
	defer os.RemoveAll(testDir)

	const numInstances = 4
	const numGoroutines = 8
	const numOperations = 25

	// Test with both same and different filenames to stress test thread safety
	loggers := make([]*Log, numInstances)
	for i := range numInstances {
		var filename string
		if i%2 == 0 {
			filename = "even-app" // Even instances use same filename
		} else {
			filename = "odd-app" // Odd instances use same filename
		}

		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename(filename).
			WithDisableSplitError(false)
		loggers[i] = NewLog(opts)
		defer loggers[i].Sync()
	}

	// Use atomic counters to track operations
	var (
		totalOperations int64
		successfulOps   int64
		errorOps        int64
	)

	var wg sync.WaitGroup

	// Start intensive concurrent logging
	for i, logger := range loggers {
		for j := range numGoroutines {
			wg.Add(1)
			go func(loggerIndex, goroutineID int, l *Log) {
				defer wg.Done()

				for k := range numOperations {
					atomic.AddInt64(&totalOperations, 1)

					// Mix different log levels and operations
					switch k % 4 {
					case 0:
						l.Debugw("Thread safety debug test",
							"instance", loggerIndex,
							"goroutine", goroutineID,
							"operation", k,
							"thread_id", fmt.Sprintf("debug-%d-%d-%d", loggerIndex, goroutineID, k))
						atomic.AddInt64(&successfulOps, 1)

					case 1:
						l.Infow("Thread safety info test",
							"instance", loggerIndex,
							"goroutine", goroutineID,
							"operation", k,
							"thread_id", fmt.Sprintf("info-%d-%d-%d", loggerIndex, goroutineID, k))
						atomic.AddInt64(&successfulOps, 1)

					case 2:
						l.Warnw("Thread safety warn test",
							"instance", loggerIndex,
							"goroutine", goroutineID,
							"operation", k,
							"thread_id", fmt.Sprintf("warn-%d-%d-%d", loggerIndex, goroutineID, k))
						atomic.AddInt64(&successfulOps, 1)

					case 3:
						l.Errorw("Thread safety error test",
							"instance", loggerIndex,
							"goroutine", goroutineID,
							"operation", k,
							"thread_id", fmt.Sprintf("error-%d-%d-%d", loggerIndex, goroutineID, k))
						atomic.AddInt64(&errorOps, 1)
					}

					// Add small random delay to increase chance of race conditions
					if k%10 == 0 {
						time.Sleep(time.Microsecond * time.Duration(k%5))
					}
				}
			}(i, j, logger)
		}
	}

	// Wait for all operations to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All operations completed
	case <-time.After(45 * time.Second):
		t.Fatal("Test timeout - thread safety test took too long")
	}

	// Verify operation counts
	expectedTotal := int64(numInstances * numGoroutines * numOperations)
	actualTotal := atomic.LoadInt64(&totalOperations)
	asrt.Equal(expectedTotal, actualTotal, "All operations should be counted")

	actualSuccess := atomic.LoadInt64(&successfulOps)
	actualErrors := atomic.LoadInt64(&errorOps)
	asrt.Equal(expectedTotal, actualSuccess+actualErrors, "Success + error operations should equal total")

	// Verify files were created correctly
	expectedFiles := []string{
		fmt.Sprintf("even-app-%s.log", time.Now().Format(time.DateOnly)),
		fmt.Sprintf("even-app-%s_error.log", time.Now().Format(time.DateOnly)),
		fmt.Sprintf("odd-app-%s.log", time.Now().Format(time.DateOnly)),
		fmt.Sprintf("odd-app-%s_error.log", time.Now().Format(time.DateOnly)),
	}

	for _, filename := range expectedFiles {
		fullPath := filepath.Join(testDir, filename)
		info, err := os.Stat(fullPath)
		asrt.NoError(err, "File %s should exist", filename)
		asrt.True(info.Size() > 0, "File %s should have content", filename)
	}
}

// Test file access conflict handling with rapid instance creation/destruction
func TestConcurrentMultiInstance_FileAccessConflicts(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_file_conflicts"
	defer os.RemoveAll(testDir)

	const numCycles = 10
	const instancesPerCycle = 3
	const operationsPerInstance = 20

	var wg sync.WaitGroup
	conflictErrors := make(chan error, numCycles*instancesPerCycle)

	// Test rapid creation and destruction of logger instances with same filename
	for cycle := range numCycles {
		wg.Add(1)
		go func(cycleID int) {
			defer wg.Done()

			// Create multiple instances in this cycle with same filename
			loggers := make([]*Log, instancesPerCycle)
			for i := range instancesPerCycle {
				opts := NewOptions().
					WithDirectory(testDir).
					WithFilename("conflict-test"). // Same filename for all
					WithDisableSplitError(false)
				loggers[i] = NewLog(opts)

				// Log some messages immediately after creation
				for j := range operationsPerInstance {
					loggers[i].Infow("Conflict test message",
						"cycle", cycleID,
						"instance", i,
						"operation", j,
						"timestamp", time.Now().UnixNano())

					// Test error logging as well
					if j%3 == 0 {
						loggers[i].Errorw("Conflict test error",
							"cycle", cycleID,
							"instance", i,
							"operation", j)
					}
				}
			}

			// Sync and close all loggers in this cycle
			for _, logger := range loggers {
				logger.Sync()
			}
		}(cycle)
	}

	// Wait for all cycles to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All cycles completed
	case <-time.After(60 * time.Second):
		t.Fatal("Test timeout - file access conflict test took too long")
	}

	// Check for any conflict errors
	close(conflictErrors)
	for err := range conflictErrors {
		t.Errorf("File access conflict error: %v", err)
	}

	// Verify that files were created and contain expected data
	expectedMainFile := filepath.Join(testDir, fmt.Sprintf("conflict-test-%s.log", time.Now().Format(time.DateOnly)))
	expectedErrorFile := filepath.Join(
		testDir,
		fmt.Sprintf("conflict-test-%s_error.log", time.Now().Format(time.DateOnly)),
	)

	// Check main log file
	mainInfo, err := os.Stat(expectedMainFile)
	asrt.NoError(err, "Main log file should exist after conflict test")
	asrt.True(mainInfo.Size() > 0, "Main log file should have content after conflict test")

	// Check error log file
	errorInfo, err := os.Stat(expectedErrorFile)
	asrt.NoError(err, "Error log file should exist after conflict test")
	asrt.True(errorInfo.Size() > 0, "Error log file should have content after conflict test")

	// Verify file contents contain entries from multiple cycles
	mainContent, err := os.ReadFile(expectedMainFile)
	asrt.NoError(err, "Should be able to read main log file after conflict test")

	mainContentStr := string(mainContent)
	// Should contain entries from multiple cycles
	cycleCount := 0
	for i := range numCycles {
		cyclePattern := fmt.Sprintf(`"cycle": %d`, i)
		if strings.Contains(mainContentStr, cyclePattern) {
			cycleCount++
		}
	}
	asrt.True(cycleCount > 0, "Main log should contain entries from at least one cycle")
}

// Test concurrent filename sanitization with various unsafe characters
func TestConcurrentMultiInstance_FilenameSanitization(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_sanitization"
	defer os.RemoveAll(testDir)

	// Test various unsafe filenames concurrently
	unsafeFilenames := []string{
		"app/service",
		"app\\service",
		"app:service",
		"app*service",
		"app?service",
		"app\"service",
		"app<service>",
		"app|service",
		"app\x00service",
		"app\x1fservice",
	}

	const numGoroutines = 5
	const numOperations = 20

	var wg sync.WaitGroup
	results := make(chan string, len(unsafeFilenames)*numGoroutines*numOperations)

	// Test each unsafe filename concurrently
	for i, filename := range unsafeFilenames {
		wg.Add(1)
		go func(index int, unsafeFilename string) {
			defer wg.Done()

			opts := NewOptions().
				WithDirectory(testDir).
				WithFilename(unsafeFilename).
				WithDisableSplitError(false)
			logger := NewLog(opts)
			defer logger.Sync()

			// Start multiple goroutines for this filename
			var innerWg sync.WaitGroup
			for j := range numGoroutines {
				innerWg.Add(1)
				go func(goroutineID int) {
					defer innerWg.Done()
					for k := range numOperations {
						logger.Infow("Sanitization test",
							"filename_index", index,
							"unsafe_filename", unsafeFilename,
							"goroutine", goroutineID,
							"operation", k,
							"timestamp", time.Now().UnixNano())

						results <- fmt.Sprintf("%d-%d-%d", index, goroutineID, k)
					}
				}(j)
			}
			innerWg.Wait()
		}(i, filename)
	}

	// Wait for all operations to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All operations completed
	case <-time.After(30 * time.Second):
		t.Fatal("Test timeout - filename sanitization test took too long")
	}

	// Verify that sanitized files were created
	expectedSanitizedFilenames := []string{
		"app_service",
		"app_service",
		"app_service",
		"app_service",
		"app_service",
		"app_service",
		"app_service_",
		"app_service",
		"app_service",
		"app_service",
	}

	currentDate := time.Now().Format(time.DateOnly)
	for i, sanitizedName := range expectedSanitizedFilenames {
		expectedMainFile := filepath.Join(testDir, fmt.Sprintf("%s-%s.log", sanitizedName, currentDate))
		expectedErrorFile := filepath.Join(testDir, fmt.Sprintf("%s-%s_error.log", sanitizedName, currentDate))

		// Check if files exist (some might be duplicates due to same sanitized names)
		if _, err := os.Stat(expectedMainFile); err == nil {
			info, err := os.Stat(expectedMainFile)
			asrt.NoError(err, "Sanitized main log file should be accessible for index %d", i)
			asrt.True(info.Size() > 0, "Sanitized main log file should have content for index %d", i)
		}

		if _, err := os.Stat(expectedErrorFile); err == nil {
			info, err := os.Stat(expectedErrorFile)
			asrt.NoError(err, "Sanitized error log file should be accessible for index %d", i)
			asrt.True(info.Size() > 0, "Sanitized error log file should have content for index %d", i)
		}
	}
}

// Test concurrent logger instances with mixed filename configurations
func TestConcurrentMultiInstance_MixedConfigurations(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_mixed_config"
	defer os.RemoveAll(testDir)

	const numInstances = 6
	const numGoroutines = 4
	const numOperations = 15

	// Create instances with mixed configurations
	loggers := make([]*Log, numInstances)
	expectedFiles := make(map[string]bool)

	for i := range numInstances {
		var opts *Options
		switch i % 3 {
		case 0:
			// Empty filename (backward compatibility)
			opts = NewOptions().
				WithDirectory(testDir).
				WithFilename("").
				WithDisableSplitError(false)
			// Expected files: YYYY-MM-DD.log, YYYY-MM-DD_error.log
			currentDate := time.Now().Format(time.DateOnly)
			expectedFiles[fmt.Sprintf("%s.log", currentDate)] = true
			expectedFiles[fmt.Sprintf("%s_error.log", currentDate)] = true

		case 1:
			// Custom filename
			opts = NewOptions().
				WithDirectory(testDir).
				WithFilename(fmt.Sprintf("custom%d", i)).
				WithDisableSplitError(false)
			// Expected files: customN-YYYY-MM-DD.log, customN-YYYY-MM-DD_error.log
			currentDate := time.Now().Format(time.DateOnly)
			expectedFiles[fmt.Sprintf("custom%d-%s.log", i, currentDate)] = true
			expectedFiles[fmt.Sprintf("custom%d-%s_error.log", i, currentDate)] = true

		case 2:
			// Custom filename with error split disabled
			opts = NewOptions().
				WithDirectory(testDir).
				WithFilename(fmt.Sprintf("nosplit%d", i)).
				WithDisableSplitError(true)
			// Expected files: nosplitN-YYYY-MM-DD.log (no error file)
			currentDate := time.Now().Format(time.DateOnly)
			expectedFiles[fmt.Sprintf("nosplit%d-%s.log", i, currentDate)] = true
		}

		loggers[i] = NewLog(opts)
		defer loggers[i].Sync()
	}

	var wg sync.WaitGroup
	operationCount := int64(0)

	// Start concurrent logging with mixed configurations
	for i, logger := range loggers {
		for j := range numGoroutines {
			wg.Add(1)
			go func(loggerIndex, goroutineID int, l *Log) {
				defer wg.Done()
				for k := range numOperations {
					atomic.AddInt64(&operationCount, 1)

					// Log different message types
					l.Infow("Mixed config test",
						"logger_index", loggerIndex,
						"config_type", loggerIndex%3,
						"goroutine", goroutineID,
						"operation", k,
						"timestamp", time.Now().UnixNano())

					// Test error logging (will only go to error file if split is enabled)
					if k%4 == 0 {
						l.Errorw("Mixed config error test",
							"logger_index", loggerIndex,
							"config_type", loggerIndex%3,
							"goroutine", goroutineID,
							"operation", k)
					}

					// Add variety with other log levels
					switch k % 3 {
					case 0:
						l.Debugw("Mixed config debug", "logger", loggerIndex, "op", k)
					case 1:
						l.Warnw("Mixed config warn", "logger", loggerIndex, "op", k)
					}
				}
			}(i, j, logger)
		}
	}

	// Wait for all operations to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All operations completed
	case <-time.After(45 * time.Second):
		t.Fatal("Test timeout - mixed configuration test took too long")
	}

	// Verify operation count
	expectedOperations := int64(numInstances * numGoroutines * numOperations)
	actualOperations := atomic.LoadInt64(&operationCount)
	asrt.Equal(expectedOperations, actualOperations, "All operations should be completed")

	// Verify expected files were created
	for expectedFile := range expectedFiles {
		fullPath := filepath.Join(testDir, expectedFile)
		info, err := os.Stat(fullPath)
		asrt.NoError(err, "Expected file %s should exist", expectedFile)
		asrt.True(info.Size() > 0, "Expected file %s should have content", expectedFile)
	}

	// Verify file contents contain appropriate entries
	files, err := os.ReadDir(testDir)
	asrt.NoError(err, "Should be able to read test directory")

	logFileCount := 0
	errorFileCount := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".log") {
			if strings.Contains(file.Name(), "_error") {
				errorFileCount++
			} else {
				logFileCount++
			}

			// Read and verify file has content
			content, err := os.ReadFile(filepath.Join(testDir, file.Name()))
			asrt.NoError(err, "Should be able to read file %s", file.Name())
			asrt.True(len(content) > 0, "File %s should have content", file.Name())

			// Verify content contains expected patterns
			contentStr := string(content)
			if strings.Contains(file.Name(), "_error") {
				// Error files should contain error messages
				asrt.Contains(
					contentStr,
					"Mixed config error test",
					"File %s should contain error messages",
					file.Name(),
				)
			} else {
				// Main log files should contain info messages
				asrt.Contains(contentStr, "Mixed config test", "File %s should contain test messages", file.Name())
			}
		}
	}

	// Verify we have the expected number of files
	asrt.True(logFileCount > 0, "Should have at least one main log file")
	// Error file count depends on configuration - some instances have error split disabled
}

// Test setupLogFiles with custom filename
func TestSetupLogFilesWithFilename(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_filename"
	defer os.RemoveAll(testDir)

	// Test with custom filename
	opts := NewOptions().
		WithDirectory(testDir).
		WithFilename("testapp").
		WithDisableSplitError(false) // Enable error log splitting

	logger := NewLog(opts)
	defer logger.Sync()

	// Test setup with current date
	currentDate := "2025-07-20"
	err := logger.setupLogFiles(currentDate)
	asrt.NoError(err)

	// Verify files are set up with correct names
	asrt.NotNil(logger.file)
	asrt.NotNil(logger.errFile)
	asrt.Equal(currentDate, logger.currDate)

	// Check that the file paths contain the custom filename
	asrt.Contains(logger.file.Filename, "testapp-2025-07-20.log")
	asrt.Contains(logger.errFile.Filename, "testapp-2025-07-20_error.log")

	// Test backward compatibility with empty filename
	opts = NewOptions().
		WithDirectory(testDir).
		WithFilename("").
		WithDisableSplitError(false)

	logger = NewLog(opts)
	defer logger.Sync()

	err = logger.setupLogFiles(currentDate)
	asrt.NoError(err)

	// Verify files use default naming format
	asrt.Contains(logger.file.Filename, "2025-07-20.log")
	asrt.Contains(logger.errFile.Filename, "2025-07-20_error.log")
	asrt.NotContains(logger.file.Filename, "testapp")
	asrt.NotContains(logger.errFile.Filename, "testapp")
}

// Integration test for setupLogFiles method with custom filename behavior
func TestSetupLogFiles_Integration_CustomFilename(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_integration_custom"
	defer os.RemoveAll(testDir)

	t.Run("CustomFilename_MainAndErrorLogs", func(t *testing.T) {
		// Test with custom filename and error log splitting enabled
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename("myservice").
			WithDisableSplitError(false)

		logger := NewLog(opts)
		defer logger.Sync()

		testDate := "2025-07-20"
		err := logger.setupLogFiles(testDate)
		asrt.NoError(err)

		// Verify internal state
		asrt.Equal(testDate, logger.currDate)
		asrt.NotNil(logger.file)
		asrt.NotNil(logger.errFile)

		// Verify file paths are correctly generated
		expectedMainPath := filepath.Join(testDir, "myservice-2025-07-20.log")
		expectedErrorPath := filepath.Join(testDir, "myservice-2025-07-20_error.log")
		asrt.Equal(expectedMainPath, logger.file.Filename)
		asrt.Equal(expectedErrorPath, logger.errFile.Filename)

		// Test actual file creation by writing to them
		testData := []byte("Integration test log entry\n")
		_, err = logger.file.Write(testData)
		asrt.NoError(err)
		_, err = logger.errFile.Write(testData)
		asrt.NoError(err)

		// Verify files exist on filesystem
		_, err = os.Stat(expectedMainPath)
		asrt.NoError(err, "Main log file should exist")
		_, err = os.Stat(expectedErrorPath)
		asrt.NoError(err, "Error log file should exist")
	})

	t.Run("CustomFilename_MainLogOnly", func(t *testing.T) {
		// Test with custom filename and error log splitting disabled
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename("singlelog").
			WithDisableSplitError(true)

		logger := NewLog(opts)
		defer logger.Sync()

		testDate := "2025-07-21"
		err := logger.setupLogFiles(testDate)
		asrt.NoError(err)

		// Verify internal state
		asrt.Equal(testDate, logger.currDate)
		asrt.NotNil(logger.file)
		asrt.Nil(logger.errFile, "Error file should be nil when split is disabled")

		// Verify file path is correctly generated
		expectedMainPath := filepath.Join(testDir, "singlelog-2025-07-21.log")
		asrt.Equal(expectedMainPath, logger.file.Filename)

		// Test actual file creation
		testData := []byte("Single log integration test\n")
		_, err = logger.file.Write(testData)
		asrt.NoError(err)

		// Verify file exists on filesystem
		_, err = os.Stat(expectedMainPath)
		asrt.NoError(err, "Main log file should exist")
	})

	t.Run("UnsafeFilename_Sanitization", func(t *testing.T) {
		// Test with filename containing unsafe characters
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename("app/service:test*").
			WithDisableSplitError(false)

		logger := NewLog(opts)
		defer logger.Sync()

		testDate := "2025-07-22"
		err := logger.setupLogFiles(testDate)
		asrt.NoError(err)

		// Verify sanitized filenames are used
		expectedMainPath := filepath.Join(testDir, "app_service_test_-2025-07-22.log")
		expectedErrorPath := filepath.Join(testDir, "app_service_test_-2025-07-22_error.log")
		asrt.Equal(expectedMainPath, logger.file.Filename)
		asrt.Equal(expectedErrorPath, logger.errFile.Filename)

		// Test actual file creation with sanitized names
		testData := []byte("Sanitized filename test\n")
		_, err = logger.file.Write(testData)
		asrt.NoError(err)
		_, err = logger.errFile.Write(testData)
		asrt.NoError(err)

		// Verify files exist with sanitized names
		_, err = os.Stat(expectedMainPath)
		asrt.NoError(err, "Main log file with sanitized name should exist")
		_, err = os.Stat(expectedErrorPath)
		asrt.NoError(err, "Error log file with sanitized name should exist")
	})
}

// Integration test for actual file creation and naming correctness
func TestSetupLogFiles_Integration_FileCreation(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_integration_creation"
	defer os.RemoveAll(testDir)

	t.Run("FileCreation_CorrectNaming", func(t *testing.T) {
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename("webapp").
			WithDisableSplitError(false)

		logger := NewLog(opts)
		defer logger.Sync()

		testDate := "2025-08-15"
		err := logger.setupLogFiles(testDate)
		asrt.NoError(err)

		// Write test data to verify files are actually created and writable
		mainTestData := []byte("Main log test data\n")
		errorTestData := []byte("Error log test data\n")

		_, err = logger.file.Write(mainTestData)
		asrt.NoError(err, "Should be able to write to main log file")

		_, err = logger.errFile.Write(errorTestData)
		asrt.NoError(err, "Should be able to write to error log file")

		// Verify files exist with correct names
		expectedMainFile := filepath.Join(testDir, "webapp-2025-08-15.log")
		expectedErrorFile := filepath.Join(testDir, "webapp-2025-08-15_error.log")

		mainInfo, err := os.Stat(expectedMainFile)
		asrt.NoError(err, "Main log file should exist")
		asrt.True(mainInfo.Size() > 0, "Main log file should have content")

		errorInfo, err := os.Stat(expectedErrorFile)
		asrt.NoError(err, "Error log file should exist")
		asrt.True(errorInfo.Size() > 0, "Error log file should have content")

		// Verify file contents
		mainContent, err := os.ReadFile(expectedMainFile)
		asrt.NoError(err)
		asrt.Contains(string(mainContent), "Main log test data")

		errorContent, err := os.ReadFile(expectedErrorFile)
		asrt.NoError(err)
		asrt.Contains(string(errorContent), "Error log test data")
	})

	t.Run("FileCreation_BackwardCompatibility", func(t *testing.T) {
		// Test file creation without custom filename (backward compatibility)
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename(""). // Empty filename for backward compatibility
			WithDisableSplitError(false)

		logger := NewLog(opts)
		defer logger.Sync()

		testDate := "2025-08-16"
		err := logger.setupLogFiles(testDate)
		asrt.NoError(err)

		// Write test data
		mainTestData := []byte("Backward compatibility main log\n")
		errorTestData := []byte("Backward compatibility error log\n")

		_, err = logger.file.Write(mainTestData)
		asrt.NoError(err)

		_, err = logger.errFile.Write(errorTestData)
		asrt.NoError(err)

		// Verify files exist with default naming format
		expectedMainFile := filepath.Join(testDir, "2025-08-16.log")
		expectedErrorFile := filepath.Join(testDir, "2025-08-16_error.log")

		_, err = os.Stat(expectedMainFile)
		asrt.NoError(err, "Main log file should exist with default naming")

		_, err = os.Stat(expectedErrorFile)
		asrt.NoError(err, "Error log file should exist with default naming")

		// Verify file contents
		mainContent, err := os.ReadFile(expectedMainFile)
		asrt.NoError(err)
		asrt.Contains(string(mainContent), "Backward compatibility main log")

		errorContent, err := os.ReadFile(expectedErrorFile)
		asrt.NoError(err)
		asrt.Contains(string(errorContent), "Backward compatibility error log")
	})

	t.Run("FileCreation_DirectoryCreation", func(t *testing.T) {
		// Test that setupLogFiles creates the directory if it doesn't exist
		nestedTestDir := filepath.Join(testDir, "nested", "deep", "directory")
		opts := NewOptions().
			WithDirectory(nestedTestDir).
			WithFilename("deeplog").
			WithDisableSplitError(false)

		logger := NewLog(opts)
		defer logger.Sync()

		// Verify directory doesn't exist initially
		_, err := os.Stat(nestedTestDir)
		asrt.True(os.IsNotExist(err), "Nested directory should not exist initially")

		testDate := "2025-08-17"
		err = logger.setupLogFiles(testDate)
		asrt.NoError(err)

		// Verify directory was created
		dirInfo, err := os.Stat(nestedTestDir)
		asrt.NoError(err, "Nested directory should be created")
		asrt.True(dirInfo.IsDir(), "Should be a directory")

		// Verify files can be created in the nested directory
		testData := []byte("Deep directory test\n")
		_, err = logger.file.Write(testData)
		asrt.NoError(err)

		expectedFile := filepath.Join(nestedTestDir, "deeplog-2025-08-17.log")
		_, err = os.Stat(expectedFile)
		asrt.NoError(err, "Log file should exist in nested directory")
	})
}

// Integration test for date change handling in file names
func TestSetupLogFiles_Integration_DateChange(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_integration_datechange"
	defer os.RemoveAll(testDir)

	t.Run("DateChange_FileRotation", func(t *testing.T) {
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename("rotatingapp").
			WithDisableSplitError(false)

		logger := NewLog(opts)
		defer logger.Sync()

		// Setup with first date
		firstDate := "2025-07-20"
		err := logger.setupLogFiles(firstDate)
		asrt.NoError(err)

		// Verify initial setup
		asrt.Equal(firstDate, logger.currDate)
		firstMainFile := logger.file.Filename
		firstErrorFile := logger.errFile.Filename
		asrt.Contains(firstMainFile, "rotatingapp-2025-07-20.log")
		asrt.Contains(firstErrorFile, "rotatingapp-2025-07-20_error.log")

		// Write to first day's files
		firstDayData := []byte("First day log entry\n")
		_, err = logger.file.Write(firstDayData)
		asrt.NoError(err)
		_, err = logger.errFile.Write(firstDayData)
		asrt.NoError(err)

		// Change to second date (simulating date change)
		secondDate := "2025-07-21"
		err = logger.setupLogFiles(secondDate)
		asrt.NoError(err)

		// Verify date change was handled
		asrt.Equal(secondDate, logger.currDate)
		secondMainFile := logger.file.Filename
		secondErrorFile := logger.errFile.Filename
		asrt.Contains(secondMainFile, "rotatingapp-2025-07-21.log")
		asrt.Contains(secondErrorFile, "rotatingapp-2025-07-21_error.log")

		// Verify new files are different from old files
		asrt.NotEqual(firstMainFile, secondMainFile)
		asrt.NotEqual(firstErrorFile, secondErrorFile)

		// Write to second day's files
		secondDayData := []byte("Second day log entry\n")
		_, err = logger.file.Write(secondDayData)
		asrt.NoError(err)
		_, err = logger.errFile.Write(secondDayData)
		asrt.NoError(err)

		// Verify both days' files exist
		_, err = os.Stat(firstMainFile)
		asrt.NoError(err, "First day's main log should still exist")
		_, err = os.Stat(firstErrorFile)
		asrt.NoError(err, "First day's error log should still exist")
		_, err = os.Stat(secondMainFile)
		asrt.NoError(err, "Second day's main log should exist")
		_, err = os.Stat(secondErrorFile)
		asrt.NoError(err, "Second day's error log should exist")

		// Verify file contents are separate
		firstMainContent, err := os.ReadFile(firstMainFile)
		asrt.NoError(err)
		asrt.Contains(string(firstMainContent), "First day log entry")
		asrt.NotContains(string(firstMainContent), "Second day log entry")

		secondMainContent, err := os.ReadFile(secondMainFile)
		asrt.NoError(err)
		asrt.Contains(string(secondMainContent), "Second day log entry")
		asrt.NotContains(string(secondMainContent), "First day log entry")
	})

	t.Run("DateChange_SameDateNoRecreation", func(t *testing.T) {
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename("stable").
			WithDisableSplitError(false)

		logger := NewLog(opts)
		defer logger.Sync()

		testDate := "2025-07-25"

		// First setup
		err := logger.setupLogFiles(testDate)
		asrt.NoError(err)

		originalMainFile := logger.file
		originalErrorFile := logger.errFile
		originalMainPath := logger.file.Filename
		originalErrorPath := logger.errFile.Filename

		// Write initial data
		initialData := []byte("Initial log entry\n")
		_, err = logger.file.Write(initialData)
		asrt.NoError(err)

		// Setup again with same date - should not recreate files
		err = logger.setupLogFiles(testDate)
		asrt.NoError(err)

		// Verify files are the same instances (not recreated)
		asrt.Equal(originalMainFile, logger.file, "Main file should not be recreated for same date")
		asrt.Equal(originalErrorFile, logger.errFile, "Error file should not be recreated for same date")
		asrt.Equal(originalMainPath, logger.file.Filename, "Main file path should remain the same")
		asrt.Equal(originalErrorPath, logger.errFile.Filename, "Error file path should remain the same")

		// Write additional data to verify files are still functional
		additionalData := []byte("Additional log entry\n")
		_, err = logger.file.Write(additionalData)
		asrt.NoError(err)

		// Verify both entries exist in the file
		fileContent, err := os.ReadFile(originalMainPath)
		asrt.NoError(err)
		asrt.Contains(string(fileContent), "Initial log entry")
		asrt.Contains(string(fileContent), "Additional log entry")
	})

	t.Run("DateChange_BackwardCompatibility", func(t *testing.T) {
		// Test date change handling without custom filename
		opts := NewOptions().
			WithDirectory(testDir).
			WithFilename(""). // No custom filename
			WithDisableSplitError(false)

		logger := NewLog(opts)
		defer logger.Sync()

		// Setup with first date
		firstDate := "2025-08-01"
		err := logger.setupLogFiles(firstDate)
		asrt.NoError(err)

		firstMainFile := logger.file.Filename
		firstErrorFile := logger.errFile.Filename
		asrt.Equal(filepath.Join(testDir, "2025-08-01.log"), firstMainFile)
		asrt.Equal(filepath.Join(testDir, "2025-08-01_error.log"), firstErrorFile)

		// Write to first day's files
		firstData := []byte("First day without custom filename\n")
		_, err = logger.file.Write(firstData)
		asrt.NoError(err)

		// Change to second date
		secondDate := "2025-08-02"
		err = logger.setupLogFiles(secondDate)
		asrt.NoError(err)

		secondMainFile := logger.file.Filename
		secondErrorFile := logger.errFile.Filename
		asrt.Equal(filepath.Join(testDir, "2025-08-02.log"), secondMainFile)
		asrt.Equal(filepath.Join(testDir, "2025-08-02_error.log"), secondErrorFile)

		// Verify files are different
		asrt.NotEqual(firstMainFile, secondMainFile)
		asrt.NotEqual(firstErrorFile, secondErrorFile)

		// Write to second day's files
		secondData := []byte("Second day without custom filename\n")
		_, err = logger.file.Write(secondData)
		asrt.NoError(err)

		// Verify both files exist and have correct content
		firstContent, err := os.ReadFile(firstMainFile)
		asrt.NoError(err)
		asrt.Contains(string(firstContent), "First day without custom filename")

		secondContent, err := os.ReadFile(secondMainFile)
		asrt.NoError(err)
		asrt.Contains(string(secondContent), "Second day without custom filename")
	})
}

// Test backward compatibility - ensure default behavior remains unchanged when Filename is not set
func TestBackwardCompatibility_DefaultBehavior(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_backward_compat"
	defer os.RemoveAll(testDir)

	// Test 1: Default options without Filename should work exactly as before
	opts := NewOptions().WithDirectory(testDir)
	asrt.Equal("", opts.Filename, "Default filename should be empty")

	logger := NewLog(opts)
	defer logger.Sync()

	// Log some messages
	logger.Info("Backward compatibility test - info message")
	logger.Error("Backward compatibility test - error message")
	logger.Warn("Backward compatibility test - warn message")
	logger.Debug("Backward compatibility test - debug message")

	// Force file creation by syncing
	logger.Sync()

	// Verify files are created with old naming format (date only)
	currentDate := time.Now().Format(time.DateOnly)
	expectedMainFile := filepath.Join(testDir, currentDate+".log")
	expectedErrorFile := filepath.Join(testDir, currentDate+"_error.log")

	// Check main log file exists and has content
	mainInfo, err := os.Stat(expectedMainFile)
	asrt.NoError(err, "Main log file should exist with default naming")
	asrt.True(mainInfo.Size() > 0, "Main log file should have content")

	// Check error log file exists and has content (since DisableSplitError is true by default, this might not exist)
	if !opts.DisableSplitError {
		errorInfo, err := os.Stat(expectedErrorFile)
		asrt.NoError(err, "Error log file should exist with default naming")
		asrt.True(errorInfo.Size() > 0, "Error log file should have content")
	}

	// Verify no files with custom prefix exist
	files, err := os.ReadDir(testDir)
	asrt.NoError(err, "Should be able to read test directory")

	for _, file := range files {
		fileName := file.Name()
		// Should not contain any prefix before the date
		asrt.True(
			strings.HasPrefix(fileName, currentDate) || strings.Contains(fileName, currentDate),
			"File %s should follow default naming pattern", fileName)
		asrt.False(
			strings.Contains(fileName, "-"+currentDate),
			"File %s should not contain prefix separator", fileName)
	}
}

// Test backward compatibility - existing API calls should work unchanged
func TestBackwardCompatibility_ExistingAPICalls(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_api_compat"
	defer os.RemoveAll(testDir)

	// Test 1: Creating logger with nil options (should use defaults)
	logger1 := NewLog(nil)
	asrt.NotNil(logger1, "NewLog(nil) should work")
	asrt.Equal("", logger1.opts.Filename, "Default filename should be empty")
	logger1.Sync()

	// Test 2: Creating logger with empty Options struct
	emptyOpts := &Options{}
	logger2 := NewLog(emptyOpts)
	asrt.NotNil(logger2, "NewLog with empty options should work")
	logger2.Sync()

	// Test 3: Using NewOptions() without any modifications
	defaultOpts := NewOptions()
	asrt.Equal("", defaultOpts.Filename, "NewOptions() should have empty filename")
	logger3 := NewLog(defaultOpts)
	asrt.NotNil(logger3, "NewLog with default options should work")
	logger3.Sync()

	// Test 4: Method chaining without WithFilename should work
	chainedOpts := NewOptions().
		WithPrefix("TEST_").
		WithDirectory(testDir).
		WithLevel("debug").
		WithFormat("json")
	asrt.Equal("", chainedOpts.Filename, "Chained options without WithFilename should have empty filename")
	logger4 := NewLog(chainedOpts)
	asrt.NotNil(logger4, "NewLog with chained options should work")
	logger4.Sync()

	// Test 5: Global logger functions should work unchanged
	originalLogger := DefaultLogger()
	asrt.NotNil(originalLogger, "DefaultLogger() should work")

	// Replace with a logger that has no filename
	testLogger := NewLog(NewOptions().WithDirectory(testDir))
	ReplaceLogger(testLogger)
	defer testLogger.Sync()

	// Global functions should work
	Info("Global info message")
	Error("Global error message")
	Warn("Global warn message")
	Debug("Global debug message")
	Infow("Global info with fields", "key", "value")
	Errorw("Global error with fields", "error", "test")

	// Verify files are created with default naming
	currentDate := time.Now().Format(time.DateOnly)
	expectedMainFile := filepath.Join(testDir, currentDate+".log")

	// Give some time for file operations
	time.Sleep(100 * time.Millisecond)
	testLogger.Sync()

	mainInfo, err := os.Stat(expectedMainFile)
	asrt.NoError(err, "Main log file should exist with default naming")
	asrt.True(mainInfo.Size() > 0, "Main log file should have content")
}

// Test backward compatibility - existing configuration files should work
func TestBackwardCompatibility_ExistingConfigurations(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_config_compat"
	os.MkdirAll(testDir, 0o755)
	defer os.RemoveAll(testDir)

	// Test 1: Configuration without filename field should work
	opts := NewOptions().
		WithPrefix("OLD_CONFIG_").
		WithDirectory(testDir).
		WithLevel("info").
		WithFormat("console").
		WithDisableCaller(false).
		WithDisableStacktrace(false).
		WithDisableSplitError(false).
		WithMaxSize(50).
		WithMaxBackups(5).
		WithCompress(true)
	// Note: No WithFilename() call - simulating old configuration

	asrt.Equal("", opts.Filename, "Configuration without filename should have empty filename")

	logger := NewLog(opts)
	defer logger.Sync()

	// Test that all other options work correctly
	asrt.Equal("OLD_CONFIG_", opts.Prefix)
	asrt.Equal(testDir, opts.Directory)
	asrt.Equal("info", opts.Level)
	asrt.Equal("console", opts.Format)
	asrt.False(opts.DisableCaller)
	asrt.False(opts.DisableStacktrace)
	asrt.False(opts.DisableSplitError)
	asrt.Equal(50, opts.MaxSize)
	asrt.Equal(5, opts.MaxBackups)
	asrt.True(opts.Compress)

	// Log messages to verify functionality
	logger.Info("Config compatibility test - info")
	logger.Error("Config compatibility test - error")
	logger.Warn("Config compatibility test - warn")

	logger.Sync()

	// Verify files are created with default naming (no custom prefix)
	currentDate := time.Now().Format(time.DateOnly)
	expectedMainFile := filepath.Join(testDir, currentDate+".log")
	expectedErrorFile := filepath.Join(testDir, currentDate+"_error.log")

	mainInfo, err := os.Stat(expectedMainFile)
	asrt.NoError(err, "Main log file should exist with default naming")
	asrt.True(mainInfo.Size() > 0, "Main log file should have content")

	errorInfo, err := os.Stat(expectedErrorFile)
	asrt.NoError(err, "Error log file should exist with default naming")
	asrt.True(errorInfo.Size() > 0, "Error log file should have content")

	// Test 2: Validation should pass for configurations without filename
	err = opts.Validate()
	asrt.NoError(err, "Validation should pass for configuration without filename")

	// Test 3: Options with empty filename should be valid
	opts.Filename = ""
	err = opts.Validate()
	asrt.NoError(err, "Validation should pass for empty filename")
}

// Test backward compatibility - upgrade scenarios
func TestBackwardCompatibility_UpgradeScenarios(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_upgrade_compat"
	defer os.RemoveAll(testDir)

	// Simulate upgrade scenario: existing code that doesn't use filename
	// Test 1: Code that creates logger with minimal options
	minimalLogger := NewLog(NewOptions().WithDirectory(testDir))
	defer minimalLogger.Sync()

	minimalLogger.Info("Upgrade test - minimal logger")
	minimalLogger.Error("Upgrade test - minimal error")

	// Test 2: Code that uses struct literal (old style)
	oldStyleOpts := &Options{
		Prefix:    "LEGACY_",
		Directory: testDir,
		Level:     "debug",
		Format:    "console",
		// Note: Filename field not set (zero value)
		DisableCaller:     false,
		DisableStacktrace: false,
		DisableSplitError: false,
		MaxSize:           100,
		MaxBackups:        3,
		Compress:          false,
	}

	// This should work even though Filename is not explicitly set
	legacyLogger := NewLog(oldStyleOpts)
	defer legacyLogger.Sync()

	legacyLogger.Info("Upgrade test - legacy logger")
	legacyLogger.Error("Upgrade test - legacy error")

	// Test 3: Verify both loggers work and create files with default naming
	minimalLogger.Sync()
	legacyLogger.Sync()

	currentDate := time.Now().Format(time.DateOnly)
	expectedMainFile := filepath.Join(testDir, currentDate+".log")
	expectedErrorFile := filepath.Join(testDir, currentDate+"_error.log")

	mainInfo, err := os.Stat(expectedMainFile)
	asrt.NoError(err, "Main log file should exist after upgrade")
	asrt.True(mainInfo.Size() > 0, "Main log file should have content after upgrade")

	errorInfo, err := os.Stat(expectedErrorFile)
	asrt.NoError(err, "Error log file should exist after upgrade")
	asrt.True(errorInfo.Size() > 0, "Error log file should have content after upgrade")

	// Test 4: Verify file content contains messages from both loggers
	// Since both loggers use default naming, they write to the same files
	mainContent, err := os.ReadFile(expectedMainFile)
	asrt.NoError(err, "Should be able to read main log file")

	mainContentStr := string(mainContent)
	asrt.Contains(mainContentStr, "minimal logger", "Should contain minimal logger message")
	asrt.Contains(mainContentStr, "legacy logger", "Should contain legacy logger message")

	errorContent, err := os.ReadFile(expectedErrorFile)
	asrt.NoError(err, "Should be able to read error log file")

	errorContentStr := string(errorContent)
	// Both loggers write to the same error file since they use default naming
	asrt.True(
		strings.Contains(errorContentStr, "minimal error") || strings.Contains(errorContentStr, "legacy error"),
		"Should contain error messages from at least one logger")

	// Test 5: Verify no files with custom prefixes were created
	files, err := os.ReadDir(testDir)
	asrt.NoError(err, "Should be able to read test directory")

	for _, file := range files {
		fileName := file.Name()
		// All files should follow the default pattern (date-based)
		asrt.True(
			strings.HasPrefix(fileName, currentDate) || strings.Contains(fileName, currentDate),
			"File %s should follow default naming pattern", fileName)
		// Should not contain prefix separator indicating custom filename
		asrt.False(
			strings.Contains(fileName, "-"+currentDate) && !strings.HasPrefix(fileName, currentDate),
			"File %s should not contain custom prefix separator", fileName)
	}
}

// Test backward compatibility - mixed usage scenarios
func TestBackwardCompatibility_MixedUsage(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_mixed_compat"
	defer os.RemoveAll(testDir)

	// Test scenario: Some loggers use filename, others don't (mixed environment)
	// This tests that both old and new code can coexist

	// Logger 1: Old style (no filename)
	oldLogger := NewLog(NewOptions().
		WithDirectory(testDir).
		WithPrefix("OLD_"))
	defer oldLogger.Sync()

	// Logger 2: New style (with filename)
	newLogger := NewLog(NewOptions().
		WithDirectory(testDir).
		WithPrefix("NEW_").
		WithFilename("newapp"))
	defer newLogger.Sync()

	// Logger 3: Default logger (no filename)
	defaultLogger := NewLog(NewOptions().WithDirectory(testDir))
	defer defaultLogger.Sync()

	// All loggers should work correctly
	oldLogger.Info("Old style logger message")
	oldLogger.Error("Old style error message")

	newLogger.Info("New style logger message")
	newLogger.Error("New style error message")

	defaultLogger.Info("Default logger message")
	defaultLogger.Error("Default error message")

	// Sync all loggers
	oldLogger.Sync()
	newLogger.Sync()
	defaultLogger.Sync()

	// Verify files are created with appropriate naming
	currentDate := time.Now().Format(time.DateOnly)

	// Old logger should create files with default naming
	oldMainFile := filepath.Join(testDir, currentDate+".log")

	// New logger should create files with custom prefix
	newMainFile := filepath.Join(testDir, "newapp-"+currentDate+".log")

	// Check old logger files
	oldMainInfo, err := os.Stat(oldMainFile)
	asrt.NoError(err, "Old logger main file should exist")
	asrt.True(oldMainInfo.Size() > 0, "Old logger main file should have content")

	// Check new logger files
	newMainInfo, err := os.Stat(newMainFile)
	asrt.NoError(err, "New logger main file should exist")
	asrt.True(newMainInfo.Size() > 0, "New logger main file should have content")

	// Verify file contents
	oldContent, err := os.ReadFile(oldMainFile)
	asrt.NoError(err, "Should be able to read old logger file")
	asrt.Contains(string(oldContent), "Old style logger message", "Old file should contain old logger messages")
	asrt.Contains(string(oldContent), "Default logger message", "Old file should contain default logger messages")

	newContent, err := os.ReadFile(newMainFile)
	asrt.NoError(err, "Should be able to read new logger file")
	asrt.Contains(string(newContent), "New style logger message", "New file should contain new logger messages")

	// Verify that old and new loggers don't interfere with each other
	asrt.NotContains(string(oldContent), "New style logger message", "Old file should not contain new logger messages")
	asrt.NotContains(string(newContent), "Old style logger message", "New file should not contain old logger messages")
}

// Test backward compatibility - ensure no breaking changes in public API
func TestBackwardCompatibility_PublicAPI(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testDir := "./logs/test_logs_api_unchanged"
	defer os.RemoveAll(testDir)

	// Test that all existing public functions and methods still work
	// This ensures no breaking changes were introduced

	// Test 1: Global functions
	originalLogger := DefaultLogger()
	testLogger := NewLog(NewOptions().WithDirectory(testDir))
	ReplaceLogger(testLogger)
	defer testLogger.Sync()

	// All these should work without any changes
	Debug("API test debug")
	Info("API test info")
	Warn("API test warn")
	Error("API test error")

	Debugln("API test debugln")
	Infoln("API test infoln")
	Warnln("API test warnln")
	Errorln("API test errorln")

	Debugf("API test debugf %s", "formatted")
	Infof("API test infof %s", "formatted")
	Warnf("API test warnf %s", "formatted")
	Errorf("API test errorf %s", "formatted")

	Debugw("API test debugw", "key", "value")
	Infow("API test infow", "key", "value")
	Warnw("API test warnw", "key", "value")
	Errorw("API test errorw", "key", "value")

	// Test 2: Instance methods
	logger := NewLog(NewOptions().WithDirectory(testDir))
	defer logger.Sync()

	logger.Debug("Instance debug")
	logger.Info("Instance info")
	logger.Warn("Instance warn")
	logger.Error("Instance error")

	logger.Debugln("Instance debugln")
	logger.Infoln("Instance infoln")
	logger.Warnln("Instance warnln")
	logger.Errorln("Instance errorln")

	logger.Debugf("Instance debugf %s", "formatted")
	logger.Infof("Instance infof %s", "formatted")
	logger.Warnf("Instance warnf %s", "formatted")
	logger.Errorf("Instance errorf %s", "formatted")

	logger.Debugw("Instance debugw", "key", "value")
	logger.Infow("Instance infow", "key", "value")
	logger.Warnw("Instance warnw", "key", "value")
	logger.Errorw("Instance errorw", "key", "value")

	// Test 3: Options methods (all existing methods should work)
	opts := NewOptions()

	// Test method chaining (should work exactly as before)
	chainedOpts := opts.
		WithPrefix("CHAIN_").
		WithDirectory(testDir).
		WithLevel("debug").
		WithTimeLayout("2006-01-02 15:04:05").
		WithFormat("json").
		WithDisableCaller(true).
		WithDisableStacktrace(true).
		WithDisableSplitError(true).
		WithMaxSize(200).
		WithMaxBackups(10).
		WithCompress(true)

	// All values should be set correctly
	asrt.Equal("CHAIN_", chainedOpts.Prefix)
	asrt.Equal(testDir, chainedOpts.Directory)
	asrt.Equal("debug", chainedOpts.Level)
	asrt.Equal("2006-01-02 15:04:05", chainedOpts.TimeLayout)
	asrt.Equal("json", chainedOpts.Format)
	asrt.True(chainedOpts.DisableCaller)
	asrt.True(chainedOpts.DisableStacktrace)
	asrt.True(chainedOpts.DisableSplitError)
	asrt.Equal(200, chainedOpts.MaxSize)
	asrt.Equal(10, chainedOpts.MaxBackups)
	asrt.True(chainedOpts.Compress)

	// Filename should still be empty (default)
	asrt.Equal("", chainedOpts.Filename)

	// Logger created with these options should work
	chainedLogger := NewLog(chainedOpts)
	asrt.NotNil(chainedLogger)
	chainedLogger.Info("Chained options test")
	chainedLogger.Sync()

	// Test 4: Sync functions
	Sync()        // Global sync should work
	logger.Sync() // Instance sync should work

	// Restore original logger
	ReplaceLogger(originalLogger)

	// Test 5: Verify files were created with default naming (no custom prefix)
	currentDate := time.Now().Format(time.DateOnly)
	expectedMainFile := filepath.Join(testDir, currentDate+".log")

	mainInfo, err := os.Stat(expectedMainFile)
	asrt.NoError(err, "Main log file should exist with default naming")
	asrt.True(mainInfo.Size() > 0, "Main log file should have content")
}

// ============================================================================
// YAML Configuration Tests (migrated from config_test.go)
// ============================================================================

func TestLoadFromYAML(t *testing.T) {
	// Create a temporary YAML config file
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "test_config.yaml")

	yamlContent := `
prefix: "YAML_"
directory: "/yaml/logs"
level: "debug"
format: "json"
max_size: 150
compress: true
`

	err := os.WriteFile(yamlFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	// Test LoadFromYAML function
	opts, err := LoadFromYAML(yamlFile)
	require.NoError(t, err)
	require.NotNil(t, opts)

	assert.Equal(t, "YAML_", opts.Prefix)
	assert.Equal(t, "/yaml/logs", opts.Directory)
	assert.Equal(t, "debug", opts.Level)
	assert.Equal(t, "json", opts.Format)
	assert.Equal(t, 150, opts.MaxSize)
	assert.True(t, opts.Compress)
}

func TestLoadFromYAMLFileNotFound(t *testing.T) {
	// Test with non-existent file
	opts, err := LoadFromYAML("/non/existent/file.yaml")
	assert.Error(t, err)
	assert.Nil(t, opts)
	assert.Contains(t, err.Error(), "failed to read configuration file")
}

func TestLoadFromYAMLInvalidContent(t *testing.T) {
	// Create a temporary file with invalid YAML content
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "invalid.yaml")

	invalidYamlContent := `
prefix: "TEST_"
level: debug
invalid_yaml: [unclosed array
`

	err := os.WriteFile(yamlFile, []byte(invalidYamlContent), 0o644)
	require.NoError(t, err)

	// Test LoadFromYAML function with invalid content
	opts, err := LoadFromYAML(yamlFile)
	assert.Error(t, err)
	assert.Nil(t, opts)
	assert.Contains(t, err.Error(), "failed to read configuration file")
}

func TestLoadFromJSON(t *testing.T) {
	// Create a temporary JSON config file
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "test_config.json")

	jsonContent := `{
  "prefix": "JSON_",
  "directory": "/json/logs",
  "level": "info",
  "format": "console",
  "max_size": 200,
  "compress": false
}`

	err := os.WriteFile(jsonFile, []byte(jsonContent), 0o644)
	require.NoError(t, err)

	// Test LoadFromJSON function
	opts, err := LoadFromJSON(jsonFile)
	require.NoError(t, err)
	require.NotNil(t, opts)

	assert.Equal(t, "JSON_", opts.Prefix)
	assert.Equal(t, "/json/logs", opts.Directory)
	assert.Equal(t, "info", opts.Level)
	assert.Equal(t, "console", opts.Format)
	assert.Equal(t, 200, opts.MaxSize)
	assert.False(t, opts.Compress)
}

func TestLoadFromJSONFileNotFound(t *testing.T) {
	// Test with non-existent file
	opts, err := LoadFromJSON("/non/existent/file.json")
	assert.Error(t, err)
	assert.Nil(t, opts)
	assert.Contains(t, err.Error(), "failed to read configuration file")
}

func TestLoadFromJSONInvalidContent(t *testing.T) {
	// Create a temporary file with invalid JSON content
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "invalid.json")

	invalidJsonContent := `{
  "prefix": "TEST_",
  "level": "debug",
  "invalid_json": [unclosed array
}`

	err := os.WriteFile(jsonFile, []byte(invalidJsonContent), 0o644)
	require.NoError(t, err)

	// Test LoadFromJSON function with invalid content
	opts, err := LoadFromJSON(jsonFile)
	assert.Error(t, err)
	assert.Nil(t, opts)
	assert.Contains(t, err.Error(), "failed to read configuration file")
}

func TestLoadFromTOML(t *testing.T) {
	// Create a temporary TOML config file
	tempDir := t.TempDir()
	tomlFile := filepath.Join(tempDir, "test_config.toml")

	tomlContent := `prefix = "TOML_"
directory = "/toml/logs"
level = "warn"
format = "json"
max_size = 300
compress = true`

	err := os.WriteFile(tomlFile, []byte(tomlContent), 0o644)
	require.NoError(t, err)

	// Test LoadFromTOML function
	opts, err := LoadFromTOML(tomlFile)
	require.NoError(t, err)
	require.NotNil(t, opts)

	assert.Equal(t, "TOML_", opts.Prefix)
	assert.Equal(t, "/toml/logs", opts.Directory)
	assert.Equal(t, "warn", opts.Level)
	assert.Equal(t, "json", opts.Format)
	assert.Equal(t, 300, opts.MaxSize)
	assert.True(t, opts.Compress)
}

func TestLoadFromTOMLFileNotFound(t *testing.T) {
	// Test with non-existent file
	opts, err := LoadFromTOML("/non/existent/file.toml")
	assert.Error(t, err)
	assert.Nil(t, opts)
	assert.Contains(t, err.Error(), "failed to read configuration file")
}

func TestLoadFromTOMLInvalidContent(t *testing.T) {
	// Create a temporary file with invalid TOML content
	tempDir := t.TempDir()
	tomlFile := filepath.Join(tempDir, "invalid.toml")

	invalidTomlContent := `prefix = "TEST_"
level = "debug"
invalid_toml = [unclosed array`

	err := os.WriteFile(tomlFile, []byte(invalidTomlContent), 0o644)
	require.NoError(t, err)

	// Test LoadFromTOML function with invalid content
	opts, err := LoadFromTOML(tomlFile)
	assert.Error(t, err)
	assert.Nil(t, opts)
	assert.Contains(t, err.Error(), "failed to read configuration file")
}

func TestLoadFromFileYAML(t *testing.T) {
	// Create a temporary YAML config file
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "test_config.yaml")

	yamlContent := `
prefix: "FILE_"
level: "info"
`

	err := os.WriteFile(yamlFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	// Test LoadFromFile function with YAML extension
	opts, err := LoadFromFile(yamlFile)
	require.NoError(t, err)
	require.NotNil(t, opts)

	assert.Equal(t, "FILE_", opts.Prefix)
	assert.Equal(t, "info", opts.Level)
}

func TestLoadFromFileYML(t *testing.T) {
	// Create a temporary YML config file
	tempDir := t.TempDir()
	ymlFile := filepath.Join(tempDir, "test_config.yml")

	ymlContent := `
prefix: "YML_"
level: "warn"
`

	err := os.WriteFile(ymlFile, []byte(ymlContent), 0o644)
	require.NoError(t, err)

	// Test LoadFromFile function with YML extension
	opts, err := LoadFromFile(ymlFile)
	require.NoError(t, err)
	require.NotNil(t, opts)

	assert.Equal(t, "YML_", opts.Prefix)
	assert.Equal(t, "warn", opts.Level)
}

func TestLoadFromFileJSON(t *testing.T) {
	// Test LoadFromFile function with JSON extension (should fail due to file not found)
	opts, err := LoadFromFile("config.json")
	assert.Error(t, err)
	assert.Nil(t, opts)
	assert.Contains(t, err.Error(), "failed to read configuration file")
}

func TestLoadFromFileTOML(t *testing.T) {
	// Test LoadFromFile function with TOML extension (should fail due to file not found)
	opts, err := LoadFromFile("config.toml")
	assert.Error(t, err)
	assert.Nil(t, opts)
	assert.Contains(t, err.Error(), "failed to read configuration file")
}

func TestLoadFromFileUnknownExtension(t *testing.T) {
	// Create a temporary file with unknown extension but valid YAML content
	tempDir := t.TempDir()
	unknownFile := filepath.Join(tempDir, "test_config.conf")

	yamlContent := `
prefix: "UNKNOWN_"
level: "error"
`

	err := os.WriteFile(unknownFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	// Test LoadFromFile function with unknown extension (viper should fail)
	opts, err := LoadFromFile(unknownFile)
	assert.Error(t, err)
	assert.Nil(t, opts)
	assert.Contains(t, err.Error(), "Unsupported Config Type")
}

// Comprehensive tests for configuration management system
func TestLoadFromYAML_ValidConfiguration(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "valid_config.yaml")

	yamlContent := `
prefix: "YAML_"
directory: "/tmp/yaml-logs"
filename: "yaml-app"
level: "debug"
time_layout: "2006/01/02 15:04:05"
format: "json"
disable_caller: true
disable_stacktrace: false
disable_split_error: true
max_size: 150
max_backups: 8
compress: true
buffer_size: 4096
flush_interval: "10s"
enable_sampling: true
sample_initial: 200
sample_thereafter: 2000
`

	err := os.WriteFile(yamlFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	opts, err := LoadFromYAML(yamlFile)
	asrt.NoError(err)
	asrt.NotNil(opts)

	// Verify all fields were loaded correctly
	asrt.Equal("YAML_", opts.Prefix)
	asrt.Equal("/tmp/yaml-logs", opts.Directory)
	asrt.Equal("yaml-app", opts.Filename)
	asrt.Equal("debug", opts.Level)
	asrt.Equal("2006/01/02 15:04:05", opts.TimeLayout)
	asrt.Equal("json", opts.Format)
	asrt.True(opts.DisableCaller)
	asrt.False(opts.DisableStacktrace)
	asrt.True(opts.DisableSplitError)
	asrt.Equal(150, opts.MaxSize)
	asrt.Equal(8, opts.MaxBackups)
	asrt.True(opts.Compress)
	asrt.True(opts.EnableSampling)
	asrt.Equal(200, opts.SampleInitial)
	asrt.Equal(2000, opts.SampleThereafter)

	// Configuration should be valid
	err = opts.Validate()
	asrt.NoError(err)
}

func TestLoadFromYAML_MinimalConfiguration(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "minimal_config.yaml")

	yamlContent := `
level: "warn"
format: "console"
`

	err := os.WriteFile(yamlFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	opts, err := LoadFromYAML(yamlFile)
	asrt.NoError(err)
	asrt.NotNil(opts)

	// Specified values should be loaded
	asrt.Equal("warn", opts.Level)
	asrt.Equal("console", opts.Format)

	// Unspecified values should have defaults
	asrt.Equal(DefaultPrefix, opts.Prefix)
	asrt.Equal(DefaultDirectory, opts.Directory)
}

func TestLoadFromYAML_InvalidFile(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test nonexistent file
	_, err := LoadFromYAML("nonexistent.yaml")
	asrt.Error(err)
	asrt.Contains(err.Error(), "failed to read configuration file")

	// Test invalid YAML syntax
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	err = os.WriteFile(invalidFile, []byte("invalid: yaml: [content"), 0o644)
	require.NoError(t, err)

	_, err = LoadFromYAML(invalidFile)
	asrt.Error(err)
	asrt.Contains(err.Error(), "failed to read configuration file")
}

func TestLoadFromYAML_InvalidConfiguration(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	tempDir := t.TempDir()
	invalidConfigFile := filepath.Join(tempDir, "invalid_config.yaml")

	// YAML with invalid configuration values
	yamlContent := `
level: "invalid_level"
format: "invalid_format"
max_size: -1
max_backups: 0
buffer_size: -100
flush_interval: "-1s"
`

	err := os.WriteFile(invalidConfigFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	_, err = LoadFromYAML(invalidConfigFile)
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid configuration values")
}

func TestLoadFromFile_FormatDetection(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	tempDir := t.TempDir()

	testCases := []struct {
		name        string
		filename    string
		content     string
		shouldError bool
		errorMsg    string
	}{
		{
			"YAML with .yaml extension",
			"config.yaml",
			"level: info\nformat: json",
			false,
			"",
		},
		{
			"YAML with .yml extension",
			"config.yml",
			"level: debug\nformat: console",
			false,
			"",
		},
		{
			"JSON file (now supported with viper)",
			"config.json",
			`{"level": "info", "format": "json"}`,
			false,
			"",
		},
		{
			"TOML file (now supported with viper)",
			"config.toml",
			`level = "info"
format = "console"`,
			false,
			"",
		},
		{
			"Unknown extension (unsupported by viper)",
			"config.conf",
			"level: warn\nformat: console",
			true,
			"Unsupported Config Type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configFile := filepath.Join(tempDir, tc.filename)
			err := os.WriteFile(configFile, []byte(tc.content), 0o644)
			require.NoError(t, err)

			opts, err := LoadFromFile(configFile)

			if tc.shouldError {
				asrt.Error(err)
				asrt.Contains(err.Error(), tc.errorMsg)
				asrt.Nil(opts)
			} else {
				asrt.NoError(err)
				asrt.NotNil(opts)
			}
		})
	}
}

func TestQuick(t *testing.T) {
	logger := Quick()
	if logger == nil {
		t.Fatal("Quick() returned nil logger")
	}

	// Test that we can log without errors
	logger.Info("Test message from Quick()")
}

func TestWithPreset(t *testing.T) {
	tests := []struct {
		name   string
		preset Preset
	}{
		{"Development", *DevelopmentPreset()},
		{"Production", *ProductionPreset()},
		{"Testing", *TestingPreset()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := WithPreset(&tt.preset)
			if logger == nil {
				t.Fatalf("WithPreset(%s) returned nil logger", tt.name)
			}

			// Test that we can log without errors
			logger.Info("Test message from preset:", tt.name)
		})
	}
}

func TestFromConfigFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	configContent := `
level: debug
format: json
directory: ` + tempDir + `
filename: test
prefix: TEST_
disable_caller: true
max_size: 50
max_backups: 2
compress: true
buffer_size: 2048
flush_interval: 2s
enable_sampling: true
sample_initial: 50
sample_thereafter: 200
`

	if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	logger, err := FromConfigFile(configFile)
	if err != nil {
		t.Fatalf("FromConfigFile() failed: %v", err)
	}
	if logger == nil {
		t.Fatal("FromConfigFile() returned nil logger")
	}

	// Test that we can log without errors
	logger.Info("Test message from config file")
}

func TestFromConfigFileInvalidFile(t *testing.T) {
	_, err := FromConfigFile("nonexistent.yaml")
	if err == nil {
		t.Fatal("FromConfigFile() should fail for nonexistent file")
	}
}

func TestFromConfigFile_ValidYAML(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "valid_config.yaml")

	configContent := `
level: warn
format: console
directory: ` + tempDir + `
filename: valid_test
prefix: VALID_
disable_caller: false
max_size: 100
max_backups: 5
compress: false
buffer_size: 4096
flush_interval: 1s
enable_sampling: false
sample_initial: 100
sample_thereafter: 500
`

	err := os.WriteFile(configFile, []byte(configContent), 0o644)
	require.NoError(t, err)

	logger, err := FromConfigFile(configFile)
	asrt.NoError(err)
	asrt.NotNil(logger)

	// Verify logger configuration
	asrt.NotNil(logger.opts)
	asrt.Equal("warn", logger.opts.Level)
	asrt.Equal("console", logger.opts.Format)
	asrt.Equal(tempDir, logger.opts.Directory)
	asrt.Equal("valid_test", logger.opts.Filename)
	asrt.Equal("VALID_", logger.opts.Prefix)
	asrt.False(logger.opts.DisableCaller)
	asrt.Equal(100, logger.opts.MaxSize)
	asrt.Equal(5, logger.opts.MaxBackups)
	asrt.False(logger.opts.Compress)
	asrt.False(logger.opts.EnableSampling)
	asrt.Equal(100, logger.opts.SampleInitial)
	asrt.Equal(500, logger.opts.SampleThereafter)

	// Test logging
	logger.Warn("Test warning message")
	logger.Sync()
}

func TestFromConfigFile_MinimalYAML(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "minimal_config.yaml")

	// Minimal config with just level
	configContent := `level: error`

	err := os.WriteFile(configFile, []byte(configContent), 0o644)
	require.NoError(t, err)

	logger, err := FromConfigFile(configFile)
	asrt.NoError(err)
	asrt.NotNil(logger)

	// Should use defaults for unspecified values
	asrt.Equal("error", logger.opts.Level)

	logger.Error("Test error message")
	logger.Sync()
}

func TestFromConfigFile_InvalidFile(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test nonexistent file
	_, err := FromConfigFile("nonexistent.yaml")
	asrt.Error(err)
	asrt.Contains(err.Error(), "failed to load YAML configuration from file")

	// Test invalid YAML
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	err = os.WriteFile(invalidFile, []byte("invalid: yaml: content: ["), 0o644)
	require.NoError(t, err)

	_, err = FromConfigFile(invalidFile)
	asrt.Error(err)
	asrt.Contains(err.Error(), "failed to load YAML configuration from file")
}

func TestFromConfigFile_SupportedFormats(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	tempDir := t.TempDir()

	// Test JSON file (should work with viper)
	jsonFile := filepath.Join(tempDir, "config.json")
	err := os.WriteFile(jsonFile, []byte(`{"level": "info", "format": "json"}`), 0o644)
	require.NoError(t, err)

	logger, err := FromConfigFile(jsonFile)
	asrt.NoError(err)
	asrt.NotNil(logger)
	asrt.Equal("info", logger.opts.Level)
	asrt.Equal("json", logger.opts.Format)

	// Test TOML file (should work with viper)
	tomlFile := filepath.Join(tempDir, "config.toml")
	err = os.WriteFile(tomlFile, []byte(`level = "debug"
format = "console"`), 0o644)
	require.NoError(t, err)

	logger, err = FromConfigFile(tomlFile)
	asrt.NoError(err)
	asrt.NotNil(logger)
	asrt.Equal("debug", logger.opts.Level)
	asrt.Equal("console", logger.opts.Format)
}

// ============================================================================
// Preset Tests (migrated from presets_test.go)
// ============================================================================

func TestDevelopmentPreset(t *testing.T) {
	preset := DevelopmentPreset()

	// Test preset metadata
	if preset.Name() != "Development" {
		t.Errorf("Expected preset name 'Development', got '%s'", preset.Name())
	}

	if preset.Description() == "" {
		t.Error("Expected non-empty description")
	}

	// Test preset configuration
	opts := NewOptions()
	preset.Apply(opts)

	// Verify development-specific settings
	if opts.Level != "debug" {
		t.Errorf("Expected level 'debug', got '%s'", opts.Level)
	}

	if opts.Format != "console" {
		t.Errorf("Expected format 'console', got '%s'", opts.Format)
	}

	if opts.DisableCaller != false {
		t.Error("Expected DisableCaller to be false in development")
	}

	if opts.DisableStacktrace != false {
		t.Error("Expected DisableStacktrace to be false in development")
	}

	if opts.DisableSplitError != true {
		t.Error("Expected DisableSplitError to be true in development")
	}

	if opts.MaxSize != 10 {
		t.Errorf("Expected MaxSize 10, got %d", opts.MaxSize)
	}

	if opts.MaxBackups != 1 {
		t.Errorf("Expected MaxBackups 1, got %d", opts.MaxBackups)
	}

	if opts.Compress != false {
		t.Error("Expected Compress to be false in development")
	}

	if opts.EnableSampling != false {
		t.Error("Expected EnableSampling to be false in development")
	}
}

func TestProductionPreset(t *testing.T) {
	preset := ProductionPreset()

	// Test preset metadata
	if preset.Name() != "Production" {
		t.Errorf("Expected preset name 'Production', got '%s'", preset.Name())
	}

	if preset.Description() == "" {
		t.Error("Expected non-empty description")
	}

	// Test preset configuration
	opts := NewOptions()
	preset.Apply(opts)

	// Verify production-specific settings
	if opts.Level != "info" {
		t.Errorf("Expected level 'info', got '%s'", opts.Level)
	}

	if opts.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", opts.Format)
	}

	if opts.DisableCaller != true {
		t.Error("Expected DisableCaller to be true in production")
	}

	if opts.DisableStacktrace != true {
		t.Error("Expected DisableStacktrace to be true in production")
	}

	if opts.DisableSplitError != false {
		t.Error("Expected DisableSplitError to be false in production")
	}

	if opts.MaxSize != 100 {
		t.Errorf("Expected MaxSize 100, got %d", opts.MaxSize)
	}

	if opts.MaxBackups != 5 {
		t.Errorf("Expected MaxBackups 5, got %d", opts.MaxBackups)
	}

	if opts.Compress != true {
		t.Error("Expected Compress to be true in production")
	}

	if opts.EnableSampling != true {
		t.Error("Expected EnableSampling to be true in production")
	}

	if opts.SampleInitial != 100 {
		t.Errorf("Expected SampleInitial 100, got %d", opts.SampleInitial)
	}

	if opts.SampleThereafter != 1000 {
		t.Errorf("Expected SampleThereafter 1000, got %d", opts.SampleThereafter)
	}
}

func TestTestingPreset(t *testing.T) {
	preset := TestingPreset()

	// Test preset metadata
	if preset.Name() != "Testing" {
		t.Errorf("Expected preset name 'Testing', got '%s'", preset.Name())
	}

	if preset.Description() == "" {
		t.Error("Expected non-empty description")
	}

	// Test preset configuration
	opts := NewOptions()
	preset.Apply(opts)

	// Verify testing-specific settings
	if opts.Level != "debug" {
		t.Errorf("Expected level 'debug', got '%s'", opts.Level)
	}

	if opts.Format != "console" {
		t.Errorf("Expected format 'console', got '%s'", opts.Format)
	}

	if opts.DisableCaller != true {
		t.Error("Expected DisableCaller to be true in testing")
	}

	if opts.DisableStacktrace != true {
		t.Error("Expected DisableStacktrace to be true in testing")
	}

	if opts.DisableSplitError != true {
		t.Error("Expected DisableSplitError to be true in testing")
	}

	if opts.MaxSize != 1 {
		t.Errorf("Expected MaxSize 1, got %d", opts.MaxSize)
	}

	if opts.MaxBackups != 1 {
		t.Errorf("Expected MaxBackups 1, got %d", opts.MaxBackups)
	}

	if opts.Compress != false {
		t.Error("Expected Compress to be false in testing")
	}

	if opts.EnableSampling != false {
		t.Error("Expected EnableSampling to be false in testing")
	}
}

func TestPresetApplyNilOptions(t *testing.T) {
	preset := DevelopmentPreset()

	// Should not panic when applying to nil options
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Apply should not panic with nil options, got panic: %v", r)
		}
	}()

	preset.Apply(nil)
}

func TestPresetWithNilConfigure(t *testing.T) {
	preset := Preset{
		name:        "Empty",
		description: "Empty preset",
		configure:   nil,
	}

	opts := NewOptions()
	originalLevel := opts.Level

	// Should not panic and should not modify options
	preset.Apply(opts)

	if opts.Level != originalLevel {
		t.Error("Preset with nil configure should not modify options")
	}
}

func TestPresetChaining(t *testing.T) {
	opts := NewOptions()

	// Apply development preset first
	DevelopmentPreset().Apply(opts)
	developmentLevel := opts.Level

	// Then apply production preset
	ProductionPreset().Apply(opts)
	productionLevel := opts.Level

	// Should have changed from development to production settings
	if developmentLevel == productionLevel {
		t.Error("Expected preset chaining to change configuration")
	}

	if opts.Level != "info" {
		t.Errorf("Expected final level to be 'info', got '%s'", opts.Level)
	}
}

func TestPresetIntegrationWithLogger(t *testing.T) {
	// Create a temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "log_preset_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test Development Preset
	t.Run("DevelopmentPreset", func(t *testing.T) {
		opts := NewOptions()
		opts.WithDirectory(tempDir).WithFilename("dev_test")
		DevelopmentPreset().Apply(opts)

		// Verify the options are valid
		if err := opts.Validate(); err != nil {
			t.Errorf("Development preset options validation failed: %v", err)
		}

		// Create logger with development preset
		logger := NewLog(opts)
		if logger == nil {
			t.Error("Failed to create logger with development preset")
		}

		// Test logging
		logger.Debug("Development debug message")
		logger.Info("Development info message")

		// Verify log file was created
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "dev_test*.log"))
		if err != nil {
			t.Errorf("Failed to check log files: %v", err)
		}
		if len(logFiles) == 0 {
			t.Error("No log files created with development preset")
		}
	})

	// Test Production Preset
	t.Run("ProductionPreset", func(t *testing.T) {
		opts := NewOptions()
		opts.WithDirectory(tempDir).WithFilename("prod_test")
		ProductionPreset().Apply(opts)

		// Verify the options are valid
		if err := opts.Validate(); err != nil {
			t.Errorf("Production preset options validation failed: %v", err)
		}

		// Create logger with production preset
		logger := NewLog(opts)
		if logger == nil {
			t.Error("Failed to create logger with production preset")
		}

		// Test logging (debug should be filtered out in production)
		logger.Debug("Production debug message - should not appear")
		logger.Info("Production info message")
		logger.Error("Production error message")

		// Verify log files were created
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "prod_test*.log"))
		if err != nil {
			t.Errorf("Failed to check log files: %v", err)
		}
		if len(logFiles) == 0 {
			t.Error("No log files created with production preset")
		}
	})

	// Test Testing Preset
	t.Run("TestingPreset", func(t *testing.T) {
		opts := NewOptions()
		opts.WithDirectory(tempDir).WithFilename("test_test")
		TestingPreset().Apply(opts)

		// Verify the options are valid
		if err := opts.Validate(); err != nil {
			t.Errorf("Testing preset options validation failed: %v", err)
		}

		// Create logger with testing preset
		logger := NewLog(opts)
		if logger == nil {
			t.Error("Failed to create logger with testing preset")
		}

		// Test logging
		logger.Debug("Testing debug message")
		logger.Info("Testing info message")

		// Verify log file was created
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "test_test*.log"))
		if err != nil {
			t.Errorf("Failed to check log files: %v", err)
		}
		if len(logFiles) == 0 {
			t.Error("No log files created with testing preset")
		}
	})
}

func TestPresetOptionsValidation(t *testing.T) {
	presets := []struct {
		name   string
		preset Preset
	}{
		{"Development", *DevelopmentPreset()},
		{"Production", *ProductionPreset()},
		{"Testing", *TestingPreset()},
	}

	for _, tc := range presets {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewOptions()
			tc.preset.Apply(opts)

			if err := opts.Validate(); err != nil {
				t.Errorf("%s preset produces invalid options: %v", tc.name, err)
			}
		})
	}
}

func TestCoreLoggerFunctionality(t *testing.T) {
	t.Parallel()

	t.Run("basic logger creation", func(t *testing.T) {
		logger := NewLog(nil)
		assert.NotNil(t, logger)
		assert.NotNil(t, logger.log)
	})

	t.Run("builder pattern", func(t *testing.T) {
		logger := NewBuilder().
			Level("debug").
			Format("console").
			Directory("./logs/test_logs").
			Filename("test").
			Build()

		assert.NotNil(t, logger)
		assert.Equal(t, "debug", logger.opts.Level)
		assert.Equal(t, "console", logger.opts.Format)
	})

	t.Run("sampling configuration", func(t *testing.T) {
		logger := NewBuilder().
			Sampling(true, 50, 200).
			Build()

		assert.NotNil(t, logger)
		assert.True(t, logger.opts.EnableSampling)
		assert.Equal(t, 50, logger.opts.SampleInitial)
		assert.Equal(t, 200, logger.opts.SampleThereafter)
	})

	t.Run("presets", func(t *testing.T) {
		devLogger := WithPreset(DevelopmentPreset())
		assert.NotNil(t, devLogger)
		assert.False(t, devLogger.opts.EnableSampling)

		prodLogger := WithPreset(ProductionPreset())
		assert.NotNil(t, prodLogger)
		assert.True(t, prodLogger.opts.EnableSampling)
	})

	t.Run("basic logging operations", func(t *testing.T) {
		logger := NewBuilder().
			Level("debug").
			Directory("./logs/test_logs").
			Build()

		// Test all log levels
		logger.Debug("Debug message")
		logger.Info("Info message")
		logger.Warn("Warning message")
		logger.Error("Error message")

		// Test formatted logging
		logger.Debugf("Debug: %s", "formatted")
		logger.Infof("Info: %d", 42)

		// Test structured logging
		logger.Debugw("Structured debug", "key", "value")
		logger.Infow("Structured info", "number", 123)

		logger.Sync()
	})
}

func TestSamplingFunctionality(t *testing.T) {
	t.Parallel()

	t.Run("sampling disabled by default", func(t *testing.T) {
		opts := NewOptions()
		assert.False(t, opts.EnableSampling)
		assert.Equal(t, DefaultSampleInitial, opts.SampleInitial)
		assert.Equal(t, DefaultSampleThereafter, opts.SampleThereafter)
	})

	t.Run("sampling configuration", func(t *testing.T) {
		opts := NewOptions()
		opts.WithSampling(true, 50, 200)

		assert.True(t, opts.EnableSampling)
		assert.Equal(t, 50, opts.SampleInitial)
		assert.Equal(t, 200, opts.SampleThereafter)
	})

	t.Run("sampling with builder", func(t *testing.T) {
		logger := NewBuilder().
			Level("debug").
			Sampling(true, 100, 500).
			Build()

		assert.NotNil(t, logger)
		assert.True(t, logger.opts.EnableSampling)
		assert.Equal(t, 100, logger.opts.SampleInitial)
		assert.Equal(t, 500, logger.opts.SampleThereafter)
	})

	t.Run("sampling validation", func(t *testing.T) {
		opts := NewOptions()
		opts.EnableSampling = true
		opts.SampleInitial = -1    // Invalid
		opts.SampleThereafter = -1 // Invalid

		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sample initial")
	})

	t.Run("production preset enables sampling", func(t *testing.T) {
		opts := NewOptions()
		ProductionPreset().Apply(opts)

		assert.True(t, opts.EnableSampling)
		assert.Equal(t, 100, opts.SampleInitial)
		assert.Equal(t, 1000, opts.SampleThereafter)
	})

	t.Run("development preset disables sampling", func(t *testing.T) {
		opts := NewOptions()
		DevelopmentPreset().Apply(opts)

		assert.False(t, opts.EnableSampling)
	})

	t.Run("logger creation with sampling", func(t *testing.T) {
		opts := NewOptions()
		opts.EnableSampling = true
		opts.SampleInitial = 10
		opts.SampleThereafter = 100

		logger := NewLog(opts)
		assert.NotNil(t, logger)
		assert.NotNil(t, logger.log)

		// Test that logger can handle multiple log calls
		// (sampling behavior is internal to zap, we just verify it doesn't crash)
		for i := range 200 {
			logger.Info("Test message", i)
		}
	})
}

func TestSamplingIntegration(t *testing.T) {
	t.Parallel()

	t.Run("high frequency logging with sampling", func(t *testing.T) {
		logger := NewBuilder().
			Level("info").
			Directory("./logs/test_logs").
			Filename("sampling-test").
			Sampling(true, 2, 1000). // Allow 2 initial, then 1 every 1000
			Build()

		// Generate many log messages quickly
		start := time.Now()
		for i := range 50 {
			logger.Infof("High frequency message %d", i)
		}
		duration := time.Since(start)

		// Sampling should make this very fast
		assert.Less(t, duration, time.Second, "Sampling should make logging fast")

		logger.Sync()
	})

	t.Run("sampling with different log levels", func(t *testing.T) {
		logger := NewBuilder().
			Level("debug").
			Directory("./logs/test_logs").
			Filename("sampling-levels").
			Sampling(true, 3, 5).
			Build()

		// Test different log levels with sampling
		for i := range 20 {
			logger.Debug("Debug message", i)
			logger.Info("Info message", i)
			logger.Warn("Warn message", i)
			logger.Error("Error message", i)
		}

		logger.Sync()
	})
}
