package internal

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// autoSyncSetup ensures that SetupAutoSync is only called once.
var autoSyncSetup sync.Once

// NewBaseEncoder creates a new encoder.
func NewBaseEncoder(format, timeLayout string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(timeLayout)

	if strings.ToLower(format) == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// SetupAutoSync sets up automatic synchronization of logs.
// It registers signal handlers to flush logs before the application terminates.
func SetupAutoSync(syncFunc func() error) {
	autoSyncSetup.Do(func() {
		// Create signal channel with buffer size 1
		signalChan := make(chan os.Signal, 1)

		// Register signals to capture
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		// Start a goroutine to handle signals
		go func() {
			sig := <-signalChan

			// Call Sync() function when signal received
			fmt.Fprintf(os.Stderr, "Received termination signal (%v), flushing logs...\n", sig)
			syncFunc()

			// Stop receiving signals on this channel
			signal.Stop(signalChan)
			close(signalChan)

			// Exit cleanly after flushing logs
			// Use exit code 0 for SIGHUP (reload), 130 for SIGINT (Ctrl+C), 143 for SIGTERM
			switch sig {
			case syscall.SIGHUP:
				os.Exit(0)
			case syscall.SIGINT:
				os.Exit(130) // 128 + 2 (SIGINT)
			case syscall.SIGTERM:
				os.Exit(143) // 128 + 15 (SIGTERM)
			default:
				os.Exit(1)
			}
		}()
	})
}

// ValidateTimeLayout validates the time layout string.
// It returns an error if the layout string is invalid.
func ValidateTimeLayout(layout string) error {
	if layout == "" {
		return errors.New("time layout is empty")
	}

	referenceTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	formattedTime := referenceTime.Format(layout)

	_, err := time.Parse(layout, formattedTime)
	if err != nil {
		return fmt.Errorf("invalid time layout: %w", err)
	}

	return nil
}
