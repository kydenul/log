// Package log provides a structured logging facility for the Ziwi framework.
//
// This package is built on top of the zap logging library (go.uber.org/zap) and provides
// a simplified interface for common logging patterns. It supports various log levels,
// structured logging with key-value pairs, and different output formats.
//
// Features:
//   - Multiple log levels: Debug, Info, Warn, Error, Panic, Fatal
//   - Structured logging with key-value pairs
//   - Printf-style logging with format strings
//   - Println-style logging
//   - JSON and console output formats
//   - Configurable time layout
//   - Log file rotation by date
//   - Separate error log files
//   - Optional caller information and stack traces
//
// Basic Usage:
//
//	// Initialize with custom options
//	opts := log.NewOptions()
//	opts.Level = "debug"
//	opts.Format = "json"
//	log.Init(opts)
//
//	// Simple logging
//	log.Debug("Debug message")
//	log.Info("Info message")
//	log.Warn("Warning message")
//	log.Error("Error message")
//
//	// Structured logging with key-value pairs
//	log.Infow("User logged in", "user_id", 123, "ip", "192.168.1.1")
//	log.Errorw("Database connection failed", "error", err, "retry", true)
//
//	// Format string logging
//	log.Debugf("Processing item %d of %d", i, total)
//	log.Errorf("Failed to connect to %s: %v", host, err)
//
// Configuration:
// The logger can be configured through the Options struct:
//
//	type Options struct {
//	    Prefix    string // Log prefix, e.g., "ZIWI"
//	    Directory string // Log file directory, e.g., "logs"
//
//	    Level      string // "debug", "info", "warn", "error", "dpanic", "panic", "fatal"
//	    TimeLayout string // Time format, default: "2006-01-02 15:04:05.000"
//	    Format     string // "console" or "json"
//
//	    DisableCaller     bool // Disable caller information
//	    DisableStacktrace bool // Disable stack traces
//	    DisableSplitError bool // Disable separate error log files
//	}
//
// Example configuration in ziwi.yaml:
//
//	log:
//	  disable-caller: false
//	  disable-stacktrace: false
//	  level: debug
//	  format: console
//	  output-paths: [/tmp/ziwi.log, stdout]
//
// Custom Logger:
// You can create a custom logger instance for specific components:
//
//	customOpts := log.NewOptions()
//	customOpts.Prefix = "MYCOMPONENT"
//	customLogger := log.NewLogger(customOpts)
//	customLogger.Info("Component initialized")
package log
