package main

import (
	"github.com/kydenul/log"
)

func main() {
	// Create a logger with aggressive sampling
	logger := log.NewBuilder().
		Level("info").
		Sampling(true, 2, 1000). // Allow 2 initial, then 1 every 1000
		Build()

	// Test 1: Same message (should be sampled)
	println("=== Test 1: Same message (should be sampled) ===")
	for i := 0; i < 20; i++ {
		logger.Info("This is a repeated message")
	}

	// Test 2: Different messages (each unique, no sampling)
	println("\n=== Test 2: Different messages (no sampling) ===")
	for i := 0; i < 5; i++ {
		logger.Infof("This is message number %d", i)
	}

	// Test 3: Same template with different args (should be sampled)
	println("\n=== Test 3: Same template (should be sampled) ===")
	for i := 0; i < 20; i++ {
		logger.Infow("User action", "user_id", i, "action", "login")
	}

	logger.Sync()
}
