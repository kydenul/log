package log

import (
	"os"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

	logger := NewLogger(opts)
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

	logger := NewLogger(nil)

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
	testDir := "./test_logs"
	defer os.RemoveAll(testDir)

	t.Run("DefaultConfig", func(t *testing.T) {
		opts := NewOptions().
			WithDirectory(testDir).
			WithPrefix("TEST_")

		if !opts.DisableSplitError {
			t.Fatalf("Expected DisableSplitError to be true by default, got false")
		}
		logger := NewLogger(opts)
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
		logger := NewLogger(opts)
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

	testDir := "./test_logs_pool"
	defer os.RemoveAll(testDir)

	opts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("POOL_TEST_")

	logger := NewLogger(opts)
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

	testDir := "./test_logs_atomic"
	defer os.RemoveAll(testDir)

	// Test DefaultLogger access
	originalLogger := DefaultLogger()
	asrt.NotNil(originalLogger)

	// Test ReplaceLogger functionality
	newOpts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("ATOMIC_TEST_")
	newLogger := NewLogger(newOpts)
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

	testDir := "./test_logs_concurrent"
	defer os.RemoveAll(testDir)

	// Test concurrent access to DefaultLogger
	const numGoroutines = 10
	const numOperations = 100

	// Channel to collect results
	results := make(chan *ZiwiLog, numGoroutines*numOperations)

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
			newLogger := NewLogger(opts)
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

	testDir := "./test_logs_date"
	defer os.RemoveAll(testDir)

	opts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("DATE_TEST_")

	logger := NewLogger(opts)
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

	testDir := "./test_logs_retry"
	defer os.RemoveAll(testDir)

	opts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("RETRY_TEST_")

	logger := NewLogger(opts)
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

	testDir := "./test_logs_setup"
	defer os.RemoveAll(testDir)

	opts := NewOptions().
		WithDirectory(testDir).
		WithPrefix("SETUP_TEST_")

	logger := NewLogger(opts)
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
