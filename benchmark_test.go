package log

import (
	"os"
	"testing"
)

// BenchmarkLogPerformance tests the performance of various logging operations
func BenchmarkLogPerformance(b *testing.B) {
	// Create a temporary directory for logs
	tempDir := "/tmp/benchmark_logs"
	defer os.RemoveAll(tempDir)

	opts := NewOptions().
		WithDirectory(tempDir).
		WithPrefix("BENCH_").
		WithLevel("info").
		WithDisableSplitError(true) // Disable error split for simpler benchmarking

	logger := NewLogger(opts)
	defer logger.Sync()

	b.Run("SimpleInfo", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Simple info message")
		}
	})

	b.Run("InfoWithFormat", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Infof("Info message with number: %d", i)
		}
	})

	b.Run("InfoWithStructuredFields", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Infow("Info message with fields", "iteration", i, "type", "benchmark")
		}
	})

	b.Run("ErrorLogging", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Errorw("Error message", "error_code", i, "severity", "high")
		}
	})

	b.Run("ConcurrentLogging", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info("Concurrent log message")
			}
		})
	})

	b.Run("GlobalLoggerAccess", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Test the atomic access to global logger
			_ = DefaultLogger()
		}
	})
}

// BenchmarkLoggerCreation tests logger creation performance
func BenchmarkLoggerCreation(b *testing.B) {
	tempDir := "/tmp/benchmark_logs_creation"
	defer os.RemoveAll(tempDir)

	b.Run("CreateLogger", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			opts := NewOptions().WithDirectory(tempDir)
			logger := NewLogger(opts)
			logger.Sync()
		}
	})

	b.Run("CreateLoggerWithValidation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			opts := NewOptions().
				WithDirectory(tempDir).
				WithLevel("debug").
				WithFormat("json").
				WithMaxSize(50).
				WithMaxBackups(5)
			logger := NewLogger(opts)
			logger.Sync()
		}
	})
}

// BenchmarkConcurrentLoggerReplace tests the performance of logger replacement
func BenchmarkConcurrentLoggerReplace(b *testing.B) {
	tempDir := "/tmp/benchmark_logs_replace"
	defer os.RemoveAll(tempDir)

	b.Run("ReplaceLogger", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Half of the goroutines replace logger, half access it
				if pb.Next() {
					opts := NewOptions().WithDirectory(tempDir)
					newLogger := NewLogger(opts)
					ReplaceLogger(newLogger)
					newLogger.Sync()
				} else {
					_ = DefaultLogger()
				}
			}
		})
	})
}

// BenchmarkMemoryAllocations measures memory allocation patterns
func BenchmarkMemoryAllocations(b *testing.B) {
	tempDir := "/tmp/benchmark_logs_memory"
	defer os.RemoveAll(tempDir)

	opts := NewOptions().
		WithDirectory(tempDir).
		WithPrefix("MEM_TEST_")

	logger := NewLogger(opts)
	defer logger.Sync()

	b.Run("AllocationsPerLog", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Infow("Memory test", "iteration", i, "data", "some_data_here")
		}
	})
}

// BenchmarkOptionsPerformance tests options creation and validation
func BenchmarkOptionsPerformance(b *testing.B) {
	b.Run("CreateOptions", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewOptions()
		}
	})

	b.Run("CreateOptionsWithChaining", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewOptions().
				WithPrefix("TEST_").
				WithLevel("debug").
				WithFormat("json").
				WithMaxSize(100).
				WithMaxBackups(3).
				WithCompress(true)
		}
	})

	b.Run("OptionsValidation", func(b *testing.B) {
		opts := NewOptions().
			WithPrefix("TEST_").
			WithLevel("debug").
			WithFormat("json")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = opts.Validate()
		}
	})
}

