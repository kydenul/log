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
func SetupAutoSync(syncFunc func()) {
	autoSyncSetup.Do(func() {
		// Create signal channel
		signalChan := make(chan os.Signal, 1)

		// Register signals to capture
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		// Start a goroutine to handle signals
		go func() {
			<-signalChan

			// Call Sync() function when signal received
			fmt.Println("Received termination signal, flushing logs...")
			syncFunc()

			// Stop signal channel
			signal.Stop(signalChan)

			// Send signal to default signal handler
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(syscall.SIGTERM)
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
