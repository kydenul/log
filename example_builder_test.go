package log

import (
	"os"
	"path/filepath"
	"time"
)

// ExampleBuilder demonstrates basic usage of the Builder pattern
func ExampleBuilder() {
	// Create a logger using the builder pattern with method chaining
	logger := NewBuilder().
		Level("debug").
		Format("console").
		Directory("./logs").
		Filename("myapp").
		Prefix("APP_").
		Build()

	logger.Info("Application started")
	logger.Debug("Debug information")
	logger.Sync()
}

// ExampleBuilder_presets demonstrates using presets with the builder
func ExampleBuilder_presets() {
	// Development environment logger
	devLogger := NewBuilder().
		Development().
		Directory("./dev_logs").
		Build()

	devLogger.Debug("Development log message")

	// Production environment logger
	prodLogger := NewBuilder().
		Production().
		Directory("./prod_logs").
		Filename("production").
		Build()

	prodLogger.Info("Production log message")

	// Testing environment logger
	testLogger := NewBuilder().
		Testing().
		Directory("./test_logs").
		Build()

	testLogger.Debug("Test log message")

	// Sync all loggers
	devLogger.Sync()
	prodLogger.Sync()
	testLogger.Sync()
}

// ExampleBuilder_customConfiguration demonstrates advanced configuration
func ExampleBuilder_customConfiguration() {
	// Create a temporary directory for this example
	tempDir := filepath.Join(os.TempDir(), "example_logs")
	defer os.RemoveAll(tempDir)

	// Custom logger with advanced settings
	logger := NewBuilder().
		Level("info").
		Format("json").
		Directory(tempDir).
		Filename("custom").
		Prefix("CUSTOM_").
		MaxSize(50).                  // 50MB max file size
		MaxBackups(3).                // Keep 3 backup files
		Compress(true).               // Compress rotated files
		BufferSize(2048).             // 2KB buffer
		FlushInterval(5*time.Second). // Flush every 5 seconds
		Sampling(true, 100, 1000).    // Enable sampling
		DisableCaller(false).         // Show caller info
		Build()

	logger.Info("Custom configured logger message")
	logger.Warn("Warning message")
	logger.Error("Error message")
	logger.Sync()
}

// ExampleBuilder_presetWithOverrides demonstrates combining presets with custom settings
func ExampleBuilder_presetWithOverrides() {
	// Start with production preset and override specific settings
	logger := NewBuilder().
		Production().          // Start with production defaults
		Level("debug").        // Override to debug level
		Directory("./custom"). // Override directory
		DisableCaller(false).  // Enable caller info (overriding production default)
		Build()

	logger.Debug("Debug message with caller info in production-like setup")
	logger.Info("Info message")
	logger.Sync()
}
