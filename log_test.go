package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog_Option(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	opts := NewOptions()
	opts.Compress = false
	opts.Directory = "/Users/kyden/git-space/ziwi/logs"
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
