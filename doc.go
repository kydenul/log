// Package log provides a high-performance, structured logging facility for Go applications.
//
// This package is built on top of the zap logging library (go.uber.org/zap) and provides
// enhanced usability features with a simplified interface. It supports various log levels,
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
//   - Environment presets for development, production, and testing
//   - Builder pattern for fluent configuration
//   - YAML configuration support
//   - HTTP middleware for request/response logging
//   - Performance optimizations (buffering, sampling)
//
// Quick Start:
//
//	// Zero-configuration quick start
//	logger := log.Quick()
//	logger.Info("Hello, World!")
//
//	// Use environment presets
//	devLogger := log.WithPreset(log.DevelopmentPreset())
//	prodLogger := log.WithPreset(log.ProductionPreset())
//
//	// Builder pattern for custom configuration
//	logger := log.NewBuilder().
//	    Level("debug").
//	    Format("json").
//	    Directory("./logs").
//	    Build()
//
// Structured Logging:
//
//	// Structured logging with key-value pairs
//	logger.Infow("User logged in", "user_id", 123, "ip", "192.168.1.1")
//	logger.Errorw("Database connection failed", "error", err, "retry", true)
//
//	// Format string logging
//	logger.Debugf("Processing item %d of %d", i, total)
//	logger.Errorf("Failed to connect to %s: %v", host, err)
//
// Configuration:
// The logger can be configured through multiple methods:
//
//	// YAML configuration
//	logger, err := log.FromConfigFile("config.yaml")
//
//	// Traditional options
//	logger := log.NewLog(log.NewOptions().
//	    WithLevel("debug").
//	    WithFormat("json").
//	    WithDirectory("./logs"))
//
// HTTP Middleware:
//
//	middleware := log.HTTPMiddleware(logger)
//	http.Handle("/api", middleware(handler))
//
// Utility Functions:
//
//	import "github.com/kydenul/log/logutil"
//
//	// Error handling
//	logutil.LogError(logger, err, "Operation failed")
//	logutil.FatalOnError(logger, err, "Critical error")
//
//	// Performance timing
//	defer logutil.Timer(logger, "operation_name")()
//
//	// Conditional logging
//	logutil.InfoIf(logger, condition, "Message", "key", "value")
package log
