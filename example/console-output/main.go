package main

import (
	"fmt"

	"github.com/kydenul/log"
)

func main() {
	logger, err := log.FromConfigFile("log.yaml")
	if err != nil {
		fmt.Printf("Error loading YAML config: %v\n", err)
		// Fallback to quick logger if config file not found
		logger = log.Quick()
	} else {
		logger.Info("Logger created from YAML configuration")
	}

	for range 1 {
		logger.Info("This is an info message")
		logger.Warn("This is a warning message")
		logger.Error("This is an error message")
	}
}
