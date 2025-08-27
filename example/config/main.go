package main

import (
	"fmt"

	"github.com/kydenul/log"
)

func main() {
	fmt.Println("=== YAML Configuration Example ===")

	// Example 1: Load from YAML file
	fmt.Println("\n1. Loading from YAML file:")
	logger1, err := log.FromConfigFile("config.yaml")
	if err != nil {
		fmt.Printf("Error loading from file: %v\n", err)
	} else {
		logger1.Info("Logger created from YAML config file")
	}

	// Example 2: Quick logger with defaults
	fmt.Println("\n2. Quick logger with defaults:")
	logger2 := log.Quick()
	logger2.Info("Quick logger with default configuration")

	// Example 3: Using presets
	fmt.Println("\n3. Using development preset:")
	logger3 := log.WithPreset(log.DevelopmentPreset())
	logger3.Debug("Development logger with preset configuration")

	// Example 4: Manual configuration loading
	fmt.Println("\n4. Manual configuration loading:")
	opts, err := log.LoadFromYAML("config.yaml")
	if err != nil {
		fmt.Printf("Error loading YAML config: %v\n", err)
	} else {
		fmt.Printf("YAML configuration - Level: %s, Format: %s, Directory: %s\n",
			opts.Level, opts.Format, opts.Directory)

		logger4 := log.NewLog(opts)
		logger4.Info("Logger created with manually loaded YAML configuration")
	}
}
