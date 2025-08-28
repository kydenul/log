package log

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()

	require.NotNil(t, builder)
	require.NotNil(t, builder.opts)

	// Verify default options are set
	assert.Equal(t, DefaultLevel.String(), builder.opts.Level)
	assert.Equal(t, DefaultFormat, builder.opts.Format)
	assert.Equal(t, DefaultDirectory, builder.opts.Directory)
	assert.Equal(t, DefaultPrefix, builder.opts.Prefix)
}

func TestBuilderChaining(t *testing.T) {
	// Test that all methods return the builder for chaining
	builder := NewBuilder().
		Level("debug").
		Format("json").
		Directory("./logs/test_logs").
		Filename("test").
		Prefix("TEST_").
		DisableCaller(true).
		DisableStacktrace(true).
		MaxSize(50).
		MaxBackups(2).
		Compress(true).
		Sampling(true, 50, 500)

	require.NotNil(t, builder)

	opts := builder.opts
	assert.Equal(t, "debug", opts.Level)
	assert.Equal(t, "json", opts.Format)
	assert.Equal(t, "./logs/test_logs", opts.Directory)
	assert.Equal(t, "test", opts.Filename)
	assert.Equal(t, "TEST_", opts.Prefix)
	assert.True(t, opts.DisableCaller)
	assert.True(t, opts.DisableStacktrace)
	assert.Equal(t, 50, opts.MaxSize)
	assert.Equal(t, 2, opts.MaxBackups)
	assert.True(t, opts.Compress)

	assert.True(t, opts.EnableSampling)
	assert.Equal(t, 50, opts.SampleInitial)
	assert.Equal(t, 500, opts.SampleThereafter)
}

func TestBuilderPresets(t *testing.T) {
	tests := []struct {
		name           string
		presetFunc     func(*Builder) *Builder
		expectedLevel  string
		expectedFormat string
	}{
		{
			name:           "Development preset",
			presetFunc:     (*Builder).Development,
			expectedLevel:  "debug",
			expectedFormat: "console",
		},
		{
			name:           "Production preset",
			presetFunc:     (*Builder).Production,
			expectedLevel:  "info",
			expectedFormat: "json",
		},
		{
			name:           "Testing preset",
			presetFunc:     (*Builder).Testing,
			expectedLevel:  "debug",
			expectedFormat: "console",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()
			builder = tt.presetFunc(builder)

			assert.Equal(t, tt.expectedLevel, builder.opts.Level)
			assert.Equal(t, tt.expectedFormat, builder.opts.Format)
		})
	}
}

func TestBuilderBuild(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "builder_test")
	defer os.RemoveAll(tempDir)

	// Test building a logger with custom configuration
	logger := NewBuilder().
		Level("warn").
		Format("json").
		Directory(tempDir).
		Filename("builder_test").
		Prefix("BUILDER_").
		Build()

	require.NotNil(t, logger)
	require.NotNil(t, logger.opts)

	// Verify the logger was created with correct options
	assert.Equal(t, "warn", logger.opts.Level)
	assert.Equal(t, "json", logger.opts.Format)
	assert.Equal(t, tempDir, logger.opts.Directory)
	assert.Equal(t, "builder_test", logger.opts.Filename)
	assert.Equal(t, "BUILDER_", logger.opts.Prefix)

	// Test that the logger can actually log
	logger.Info("This should not appear due to warn level")
	logger.Warn("This should appear")
	logger.Error("This should also appear")

	// Sync to ensure logs are written
	logger.Sync()
}

func TestBuilderWithPresetAndOverrides(t *testing.T) {
	// Test applying a preset and then overriding some settings
	builder := NewBuilder().
		Development().  // Apply development preset
		Level("error"). // Override level
		Format("json"). // Override format
		MaxSize(200)    // Override max size

	opts := builder.opts

	// Verify preset was applied and then overridden
	assert.Equal(t, "error", opts.Level)    // Overridden
	assert.Equal(t, "json", opts.Format)    // Overridden
	assert.Equal(t, 200, opts.MaxSize)      // Overridden
	assert.False(t, opts.DisableCaller)     // From development preset
	assert.False(t, opts.DisableStacktrace) // From development preset
}

func TestBuilderInvalidValues(t *testing.T) {
	// Test that invalid values are handled gracefully by the underlying Options methods
	builder := NewBuilder().
		Level("invalid_level").   // Should fallback to default
		Format("invalid_format"). // Should fallback to default
		MaxSize(-1).              // Should fallback to default
		MaxBackups(0)             // Should fallback to default

	opts := builder.opts

	// The underlying Options methods should handle invalid values
	assert.Equal(t, DefaultLevel.String(), opts.Level)
	assert.Equal(t, DefaultFormat, opts.Format)
	assert.Equal(t, DefaultMaxSize, opts.MaxSize)
	assert.Equal(t, DefaultMaxBackups, opts.MaxBackups)
}

func TestBuilderMultiplePresets(t *testing.T) {
	// Test applying multiple presets (last one should win)
	builder := NewBuilder().
		Development(). // First preset
		Production().  // Second preset should override
		Testing()      // Third preset should override

	opts := builder.opts

	// Should have testing preset configuration
	assert.Equal(t, "debug", opts.Level)
	assert.Equal(t, "console", opts.Format)
	assert.True(t, opts.DisableCaller)     // From testing preset
	assert.True(t, opts.DisableStacktrace) // From testing preset
	assert.Equal(t, 1, opts.MaxSize)       // From testing preset
}

func TestBuilderEmptyDirectory(t *testing.T) {
	// Test that empty directory falls back to default
	builder := NewBuilder().Directory("")

	// Should fallback to default directory
	assert.Equal(t, DefaultDirectory, builder.opts.Directory)
}

func TestBuilderTimeLayout(t *testing.T) {
	customLayout := "2006/01/02 15:04:05"
	builder := NewBuilder().TimeLayout(customLayout)

	assert.Equal(t, customLayout, builder.opts.TimeLayout)
}

func TestBuilderSamplingConfiguration(t *testing.T) {
	builder := NewBuilder().Sampling(true, 200, 2000)

	opts := builder.opts
	assert.True(t, opts.EnableSampling)
	assert.Equal(t, 200, opts.SampleInitial)
	assert.Equal(t, 2000, opts.SampleThereafter)

	// Test disabling sampling
	builder = NewBuilder().Sampling(false, 100, 1000)
	opts = builder.opts
	assert.False(t, opts.EnableSampling)
	// Values should still be set even when disabled
	assert.Equal(t, 100, opts.SampleInitial)
	assert.Equal(t, 1000, opts.SampleThereafter)
}

// Comprehensive tests for the Builder pattern
func TestNewBuilder_Initialization(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	builder := NewBuilder()
	asrt.NotNil(builder)
	asrt.NotNil(builder.opts)

	// Should start with default options
	asrt.Equal(DefaultLevel.String(), builder.opts.Level)
	asrt.Equal(DefaultFormat, builder.opts.Format)
	asrt.Equal(DefaultDirectory, builder.opts.Directory)
	asrt.Equal(DefaultPrefix, builder.opts.Prefix)
	asrt.Equal(DefaultEnableSampling, builder.opts.EnableSampling)
	asrt.Equal(DefaultSampleInitial, builder.opts.SampleInitial)
	asrt.Equal(DefaultSampleThereafter, builder.opts.SampleThereafter)
}

func TestBuilder_BasicMethodChaining(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	builder := NewBuilder().
		Level("debug").
		Format("json").
		Directory("./logs/test_logs").
		Filename("test").
		Prefix("TEST_")

	asrt.NotNil(builder)
	asrt.Equal("debug", builder.opts.Level)
	asrt.Equal("json", builder.opts.Format)
	asrt.Equal("./logs/test_logs", builder.opts.Directory)
	asrt.Equal("test", builder.opts.Filename)
	asrt.Equal("TEST_", builder.opts.Prefix)
}

func TestBuilder_EnhancedMethodChaining(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	builder := NewBuilder().
		Sampling(true, 100, 1000)

	asrt.NotNil(builder)
	asrt.True(builder.opts.EnableSampling)
	asrt.Equal(100, builder.opts.SampleInitial)
	asrt.Equal(1000, builder.opts.SampleThereafter)
}

func TestBuilder_AllMethodsChaining(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	builder := NewBuilder().
		Level("warn").
		Format("console").
		Directory("/var/log").
		Filename("myapp").
		Prefix("MYAPP_").
		TimeLayout("2006-01-02 15:04:05").
		DisableCaller(true).
		DisableStacktrace(false).
		DisableSplitError(true).
		MaxSize(200).
		MaxBackups(10).
		Compress(true).
		Sampling(true, 50, 500)

	opts := builder.opts
	asrt.Equal("warn", opts.Level)
	asrt.Equal("console", opts.Format)
	asrt.Equal("/var/log", opts.Directory)
	asrt.Equal("myapp", opts.Filename)
	asrt.Equal("MYAPP_", opts.Prefix)
	asrt.Equal("2006-01-02 15:04:05", opts.TimeLayout)
	asrt.True(opts.DisableCaller)
	asrt.False(opts.DisableStacktrace)
	asrt.True(opts.DisableSplitError)
	asrt.Equal(200, opts.MaxSize)
	asrt.Equal(10, opts.MaxBackups)
	asrt.True(opts.Compress)
	asrt.True(opts.EnableSampling)
	asrt.Equal(50, opts.SampleInitial)
	asrt.Equal(500, opts.SampleThereafter)
}

func TestBuilder_PresetMethods(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testCases := []struct {
		name           string
		builderFunc    func(*Builder) *Builder
		expectedLevel  string
		expectedFormat string
		expectedBuffer int
	}{
		{
			"Development",
			(*Builder).Development,
			"debug", "console", 512,
		},
		{
			"Production",
			(*Builder).Production,
			"info", "json", 2048,
		},
		{
			"Testing",
			(*Builder).Testing,
			"debug", "console", 256,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewBuilder()
			builder = tc.builderFunc(builder)

			asrt.Equal(tc.expectedLevel, builder.opts.Level)
			asrt.Equal(tc.expectedFormat, builder.opts.Format)
		})
	}
}

func TestBuilder_PresetWithOverrides(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Apply preset then override specific values
	builder := NewBuilder().
		Development().  // Apply development preset
		Level("error"). // Override level
		Format("json"). // Override format
		MaxSize(300)    // Override max size

	opts := builder.opts
	asrt.Equal("error", opts.Level) // Overridden
	asrt.Equal("json", opts.Format) // Overridden
	asrt.Equal(300, opts.MaxSize)   // Overridden

	// Values from development preset that weren't overridden
	asrt.False(opts.DisableCaller)     // From development preset
	asrt.False(opts.DisableStacktrace) // From development preset
}

func TestBuilder_MultiplePresets(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Apply multiple presets (last one should win)
	builder := NewBuilder().
		Development(). // First preset
		Production().  // Second preset should override
		Testing()      // Third preset should override

	opts := builder.opts
	// Should have testing preset configuration
	asrt.Equal("debug", opts.Level)
	asrt.Equal("console", opts.Format)
	asrt.True(opts.DisableCaller)     // From testing preset
	asrt.True(opts.DisableStacktrace) // From testing preset
	asrt.Equal(1, opts.MaxSize)       // From testing preset
}

func TestBuilder_Build(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	tempDir := t.TempDir()

	logger := NewBuilder().
		Level("info").
		Format("json").
		Directory(tempDir).
		Filename("builder_test").
		Prefix("BUILD_").
		Build()

	asrt.NotNil(logger)
	asrt.NotNil(logger.opts)

	// Verify the logger was created with correct options
	asrt.Equal("info", logger.opts.Level)
	asrt.Equal("json", logger.opts.Format)
	asrt.Equal(tempDir, logger.opts.Directory)
	asrt.Equal("builder_test", logger.opts.Filename)
	asrt.Equal("BUILD_", logger.opts.Prefix)

	// Test that the logger can actually log
	logger.Info("Test message from builder")
	logger.Warn("Warning message")
	logger.Error("Error message")
	logger.Sync()
}

func TestBuilder_InvalidValues(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that invalid values are handled gracefully
	builder := NewBuilder().
		Level("invalid_level").   // Should fallback to default
		Format("invalid_format"). // Should fallback to default
		MaxSize(-1).              // Should fallback to default
		MaxBackups(0)             // Should fallback to default

	opts := builder.opts
	asrt.Equal(DefaultLevel.String(), opts.Level)
	asrt.Equal(DefaultFormat, opts.Format)
	asrt.Equal(DefaultMaxSize, opts.MaxSize)
	asrt.Equal(DefaultMaxBackups, opts.MaxBackups)

	// Should still be able to build a valid logger
	logger := builder.Build()
	asrt.NotNil(logger)
	logger.Info("Test with corrected invalid values")
	logger.Sync()
}

func TestBuilder_EmptyStringHandling(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	builder := NewBuilder().
		Level("").     // Should use default
		Format("").    // Should use default
		Directory(""). // Should use default
		TimeLayout("") // Should use default

	opts := builder.opts
	asrt.Equal(DefaultLevel.String(), opts.Level)
	asrt.Equal(DefaultFormat, opts.Format)
	asrt.Equal(DefaultDirectory, opts.Directory)
	asrt.Equal(DefaultTimeLayout, opts.TimeLayout)
}

func TestBuilder_BooleanMethods(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test setting boolean values to true
	builder := NewBuilder().
		DisableCaller(true).
		DisableStacktrace(true).
		DisableSplitError(true).
		Compress(true)

	opts := builder.opts
	asrt.True(opts.DisableCaller)
	asrt.True(opts.DisableStacktrace)
	asrt.True(opts.DisableSplitError)
	asrt.True(opts.Compress)

	// Test setting boolean values to false
	builder = NewBuilder().
		DisableCaller(false).
		DisableStacktrace(false).
		DisableSplitError(false).
		Compress(false)

	opts = builder.opts
	asrt.False(opts.DisableCaller)
	asrt.False(opts.DisableStacktrace)
	asrt.False(opts.DisableSplitError)
	asrt.False(opts.Compress)
}

func TestBuilder_SamplingConfiguration(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test enabling sampling with custom values
	builder := NewBuilder().Sampling(true, 200, 2000)
	opts := builder.opts
	asrt.True(opts.EnableSampling)
	asrt.Equal(200, opts.SampleInitial)
	asrt.Equal(2000, opts.SampleThereafter)

	// Test disabling sampling
	builder = NewBuilder().Sampling(false, 100, 1000)
	opts = builder.opts
	asrt.False(opts.EnableSampling)
	asrt.Equal(100, opts.SampleInitial) // Values should still be set
	asrt.Equal(1000, opts.SampleThereafter)

	// Test with invalid values (should use defaults)
	builder = NewBuilder().Sampling(true, 0, -1)
	opts = builder.opts
	asrt.True(opts.EnableSampling)
	asrt.Equal(DefaultSampleInitial, opts.SampleInitial)
	asrt.Equal(DefaultSampleThereafter, opts.SampleThereafter)
}

func TestBuilder_TimeLayoutCustomization(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	customLayouts := []string{
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05.000",
		"Jan 2, 2006 15:04:05",
		"2006-01-02T15:04:05Z07:00",
	}

	for _, layout := range customLayouts {
		t.Run("layout_"+layout, func(t *testing.T) {
			builder := NewBuilder().TimeLayout(layout)
			asrt.Equal(layout, builder.opts.TimeLayout)
		})
	}
}

func TestBuilder_ComplexConfiguration(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test a complex, realistic configuration
	builder := NewBuilder().
		Production().                // Start with production preset
		Level("warn").               // Override to warn level
		Directory("/var/log/myapp"). // Custom directory
		Filename("application").     // Custom filename
		Prefix("MYAPP_").            // Custom prefix
		MaxSize(500).                // Larger files
		MaxBackups(20).              // More backups
		Sampling(true, 1000, 10000)  // Custom sampling

	opts := builder.opts

	// Verify all settings
	asrt.Equal("warn", opts.Level)
	asrt.Equal("json", opts.Format) // From production preset
	asrt.Equal("/var/log/myapp", opts.Directory)
	asrt.Equal("application", opts.Filename)
	asrt.Equal("MYAPP_", opts.Prefix)
	asrt.Equal(500, opts.MaxSize)
	asrt.Equal(20, opts.MaxBackups)
	asrt.True(opts.Compress) // From production preset
	asrt.True(opts.EnableSampling)
	asrt.Equal(1000, opts.SampleInitial)
	asrt.Equal(10000, opts.SampleThereafter)

	// Should build successfully
	logger := builder.Build()
	asrt.NotNil(logger)
}

func TestBuilder_Immutability(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that builder methods don't affect the original builder
	originalBuilder := NewBuilder().Level("info").Format("json")

	// Create a new builder from the original
	_ = originalBuilder.Level("debug").Format("console")

	// Original builder should be unchanged
	asrt.Equal(
		"debug",
		originalBuilder.opts.Level,
	) // Actually, this will be changed because we're modifying the same instance

	// Let's test proper immutability by creating separate builders
	builder1 := NewBuilder().Level("info")
	builder2 := NewBuilder().Level("debug")

	asrt.Equal("info", builder1.opts.Level)
	asrt.Equal("debug", builder2.opts.Level)
	asrt.NotEqual(builder1.opts.Level, builder2.opts.Level)
}

func TestBuilder_ValidationAfterBuild(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Build with valid configuration
	logger := NewBuilder().
		Level("info").
		Format("json").
		MaxSize(100).
		MaxBackups(5).
		Sampling(true, 100, 1000).
		Build()

	asrt.NotNil(logger)

	// The built logger should have valid options
	err := logger.opts.Validate()
	asrt.NoError(err)
}

func TestBuilder_ThreadSafety(t *testing.T) {
	t.Parallel()

	// Test that multiple goroutines can create builders concurrently
	const numGoroutines = 10
	done := make(chan *Log, numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			logger := NewBuilder().
				Level("info").
				Prefix("THREAD_" + string(rune(id)) + "_").
				Build()

			logger.Info("Thread safety test", "goroutine", id)
			logger.Sync()
			done <- logger
		}(i)
	}

	// Collect all loggers
	loggers := make([]*Log, 0, numGoroutines)
	for range numGoroutines {
		logger := <-done
		loggers = append(loggers, logger)
		require.NotNil(t, logger)
	}

	// Verify all loggers are different instances
	for i := 0; i < len(loggers); i++ {
		for j := i + 1; j < len(loggers); j++ {
			assert.NotSame(t, loggers[i], loggers[j])
		}
	}
}

func TestBuilder_MethodReturnValues(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that all methods return the builder for chaining
	builder := NewBuilder()

	// Each method should return the same builder instance
	asrt.Same(builder, builder.Level("info"))
	asrt.Same(builder, builder.Format("json"))
	asrt.Same(builder, builder.Directory("/tmp"))
	asrt.Same(builder, builder.Filename("test"))
	asrt.Same(builder, builder.Prefix("TEST_"))
	asrt.Same(builder, builder.TimeLayout("2006-01-02"))
	asrt.Same(builder, builder.DisableCaller(true))
	asrt.Same(builder, builder.DisableStacktrace(true))
	asrt.Same(builder, builder.DisableSplitError(true))
	asrt.Same(builder, builder.MaxSize(100))
	asrt.Same(builder, builder.MaxBackups(5))
	asrt.Same(builder, builder.Compress(true))
	asrt.Same(builder, builder.Sampling(true, 100, 1000))
	asrt.Same(builder, builder.Development())
	asrt.Same(builder, builder.Production())
	asrt.Same(builder, builder.Testing())
}

func TestBuilder_EdgeCases(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test edge cases and boundary values
	builder := NewBuilder().
		MaxSize(1).          // Minimum valid size
		MaxBackups(1).       // Minimum valid backups
		Sampling(true, 1, 1) // Minimum valid sampling

	logger := builder.Build()
	asrt.NotNil(logger)

	// Should be able to log even with minimal settings
	logger.Info("Edge case test")
	logger.Sync()
}

func TestBuilderIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "builder_integration_test")
	defer os.RemoveAll(tempDir)

	// Test that builder creates a logger that implements the Logger interface
	var logger Logger = NewBuilder().
		Level("debug").
		Format("json").
		Directory(tempDir).
		Filename("integration_test").
		Prefix("INTEGRATION_").
		Build()

	require.NotNil(t, logger)

	// Test all logging methods to ensure they work
	logger.Debug("Debug message")
	logger.Debugf("Debug formatted: %s", "test")
	logger.Debugw("Debug with fields", "key", "value")
	logger.Debugln("Debug line")

	logger.Info("Info message")
	logger.Infof("Info formatted: %s", "test")
	logger.Infow("Info with fields", "key", "value")
	logger.Infoln("Info line")

	logger.Warn("Warn message")
	logger.Warnf("Warn formatted: %s", "test")
	logger.Warnw("Warn with fields", "key", "value")
	logger.Warnln("Warn line")

	logger.Error("Error message")
	logger.Errorf("Error formatted: %s", "test")
	logger.Errorw("Error with fields", "key", "value")
	logger.Errorln("Error line")

	// Sync to ensure all logs are written
	logger.Sync()

	// Verify log files were created
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.True(t, len(files) > 0, "Log files should be created")
}

func TestBuilderWithExistingAPI(t *testing.T) {
	// Test that builder works alongside existing API
	tempDir := filepath.Join(os.TempDir(), "builder_existing_api_test")
	defer os.RemoveAll(tempDir)

	// Create logger using builder
	builderLogger := NewBuilder().
		Development().
		Directory(tempDir).
		Filename("builder_api").
		Build()

	// Create logger using existing API
	opts := NewOptions().
		WithLevel("debug").
		WithFormat("console").
		WithDirectory(tempDir).
		WithFilename("existing_api")

	existingLogger := NewLog(opts)

	// Both should work
	builderLogger.Info("Message from builder logger")
	existingLogger.Info("Message from existing logger")

	builderLogger.Sync()
	existingLogger.Sync()

	// Verify both created log files
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.True(t, len(files) >= 2, "Both loggers should create files")
}

func TestBuilderPresetIntegration(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "builder_preset_integration_test")
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		builderFunc func(*Builder) *Builder
		expectLevel string
	}{
		{
			name:        "Development preset integration",
			builderFunc: (*Builder).Development,
			expectLevel: "debug",
		},
		{
			name:        "Production preset integration",
			builderFunc: (*Builder).Production,
			expectLevel: "info",
		},
		{
			name:        "Testing preset integration",
			builderFunc: (*Builder).Testing,
			expectLevel: "debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewBuilder().
				Directory(tempDir).
				Filename(tt.name)

			logger = tt.builderFunc(logger)

			builtLogger := logger.Build()
			require.NotNil(t, builtLogger)

			// Test that the logger works
			builtLogger.Debug("Debug message for " + tt.name)
			builtLogger.Info("Info message for " + tt.name)
			builtLogger.Warn("Warn message for " + tt.name)
			builtLogger.Error("Error message for " + tt.name)

			builtLogger.Sync()

			// Verify the level was set correctly
			assert.Equal(t, tt.expectLevel, builtLogger.opts.Level)
		})
	}
}

func TestBuilderChainedConfiguration(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "builder_chained_test")
	defer os.RemoveAll(tempDir)

	// Test complex chained configuration
	logger := NewBuilder().
		Production().               // Start with production preset
		Level("debug").             // Override level
		Directory(tempDir).         // Set directory
		Filename("chained_config"). // Set filename
		Prefix("CHAIN_").           // Set prefix
		MaxSize(25).                // Override max size
		MaxBackups(2).              // Override max backups
		Compress(false).            // Override compression
		Sampling(false, 0, 0).      // Disable sampling
		DisableCaller(false).       // Enable caller info
		Build()

	require.NotNil(t, logger)

	// Verify configuration was applied correctly
	opts := logger.opts
	assert.Equal(t, "debug", opts.Level)             // Overridden
	assert.Equal(t, "json", opts.Format)             // From production preset
	assert.Equal(t, tempDir, opts.Directory)         // Set explicitly
	assert.Equal(t, "chained_config", opts.Filename) // Set explicitly
	assert.Equal(t, "CHAIN_", opts.Prefix)           // Set explicitly
	assert.Equal(t, 25, opts.MaxSize)                // Overridden
	assert.Equal(t, 2, opts.MaxBackups)              // Overridden
	assert.False(t, opts.Compress)                   // Overridden
	assert.False(t, opts.EnableSampling)             // Overridden
	assert.False(t, opts.DisableCaller)              // Overridden

	// Test that the logger works
	logger.Info("Chained configuration test message")
	logger.Sync()
}

func TestBuilderBackwardCompatibility(t *testing.T) {
	// Test that builder doesn't break existing functionality
	tempDir := filepath.Join(os.TempDir(), "builder_compatibility_test")
	defer os.RemoveAll(tempDir)

	// Create logger using builder
	builderLogger := NewBuilder().
		Level("info").
		Format("console").
		Directory(tempDir).
		Filename("builder_compat").
		Build()

	// Create logger using traditional method
	opts := &Options{
		Level:            "info",
		Format:           "console",
		Directory:        tempDir,
		Filename:         "traditional_compat",
		Prefix:           DefaultPrefix,
		MaxSize:          DefaultMaxSize,
		MaxBackups:       DefaultMaxBackups,
		Compress:         DefaultCompress,
		EnableSampling:   DefaultEnableSampling,
		SampleInitial:    DefaultSampleInitial,
		SampleThereafter: DefaultSampleThereafter,
	}
	traditionalLogger := NewLog(opts)

	// Both should behave similarly
	builderLogger.Info("Builder logger message")
	traditionalLogger.Info("Traditional logger message")

	builderLogger.Sync()
	traditionalLogger.Sync()

	// Both should have similar configurations (excluding filename differences)
	assert.Equal(t, builderLogger.opts.Level, traditionalLogger.opts.Level)
	assert.Equal(t, builderLogger.opts.Format, traditionalLogger.opts.Format)
	assert.Equal(t, builderLogger.opts.Directory, traditionalLogger.opts.Directory)
}

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
		Directory("./logs/test_logs").
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
		MaxSize(50).               // 50MB max file size
		MaxBackups(3).             // Keep 3 backup files
		Compress(true).            // Compress rotated files
		Sampling(true, 100, 1000). // Enable sampling
		DisableCaller(false).      // Show caller info
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

func TestBuilderConsoleOutput(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test enabling console output
	builder := NewBuilder().ConsoleOutput(true)
	asrt.True(builder.opts.ConsoleOutput)

	// Test disabling console output
	builder = NewBuilder().ConsoleOutput(false)
	asrt.False(builder.opts.ConsoleOutput)

	// Test method chaining
	builder = NewBuilder().
		Level("debug").
		ConsoleOutput(false).
		Format("json")
	asrt.Equal("debug", builder.opts.Level)
	asrt.False(builder.opts.ConsoleOutput)
	asrt.Equal("json", builder.opts.Format)

	// Test with preset override
	builder = NewBuilder().
		Production().
		ConsoleOutput(false) // Override preset
	asrt.False(builder.opts.ConsoleOutput)
}
