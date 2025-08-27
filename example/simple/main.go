package main

import (
	"fmt"
	"time"

	"github.com/kydenul/log"
	"github.com/kydenul/log/logutil"
)

func main() {
	fmt.Println("=== Simple Logging Examples ===")

	// Example 1: Zero-configuration quick start
	fmt.Println("\n1. Quick start (zero configuration):")
	quickLogger := log.Quick()
	quickLogger.Info("Hello from quick logger!")
	quickLogger.Debug("This debug message won't show (default level is info)")

	// Example 2: Using environment presets
	fmt.Println("\n2. Using development preset:")
	devLogger := log.WithPreset(log.DevelopmentPreset())
	devLogger.Debug("Debug messages are visible in development")
	devLogger.Info("Development preset includes caller information")
	devLogger.Warn("This is a warning with stack trace info")

	// Example 3: Using builder pattern
	fmt.Println("\n3. Using builder pattern:")
	builderLogger := log.NewBuilder().
		Level("debug").
		Format("console").
		Directory("./logs").
		Filename("simple-example").
		Sampling(true, 100, 1000).
		Build()

	builderLogger.Info("Logger created with builder pattern")
	builderLogger.Debug("Custom configuration applied")

	// Example 4: Using YAML configuration
	fmt.Println("\n4. Using YAML configuration:")
	yamlLogger, err := log.FromConfigFile("log.yaml")
	if err != nil {
		fmt.Printf("Error loading YAML config: %v\n", err)
		// Fallback to quick logger if config file not found
		yamlLogger = log.Quick()
	} else {
		yamlLogger.Info("Logger created from YAML configuration")
	}

	// Example 5: Using utility functions
	fmt.Println("\n5. Using utility functions:")
	utilLogger := log.Quick()

	// Timer utility
	defer logutil.Timer(utilLogger, "main_function")()

	// Error handling utility
	err = simulateOperation()
	logutil.LogError(utilLogger, err, "Simulated operation completed")

	// Conditional logging (using configuration instead of environment variables)
	debugMode := true // This could be read from YAML config
	logutil.InfoIf(utilLogger, debugMode, "Debug mode is enabled")
	logutil.InfoIf(utilLogger, !debugMode, "Debug mode is disabled")

	// Application lifecycle logging
	logutil.LogStartup(utilLogger, "simple-example", "v1.0.0", 0)

	// Example 6: Structured logging
	fmt.Println("\n6. Structured logging:")
	structLogger := log.Quick()

	structLogger.Infow("User action",
		"user_id", 12345,
		"action", "login",
		"ip_address", "192.168.1.100",
		"user_agent", "Mozilla/5.0...",
		"timestamp", time.Now(),
	)

	structLogger.Errorw("Database operation failed",
		"operation", "user_lookup",
		"table", "users",
		"query_time_ms", 1500,
		"error", "connection timeout",
	)

	// Example 7: Traditional API (still works)
	fmt.Println("\n7. Traditional API (backward compatibility):")
	log.Info("Traditional global logger still works")
	log.Infof("Formatted message: %s", "Hello World")
	log.Infoln("Println style message")

	fmt.Println("\n=== Examples completed ===")
	fmt.Println("Check the ./logs directory for log files")
}

func simulateOperation() error {
	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	// Return nil (no error) for this example
	return nil
}
