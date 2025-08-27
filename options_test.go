package log

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func Test_NewOptions(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions()
	asrt.Equal(DefaultPrefix, opt.Prefix)
	asrt.Equal(DefaultDirectory, opt.Directory)
	asrt.Equal(DefaultFilename, opt.Filename)
	asrt.Equal(DefaultLevel.String(), opt.Level)
	asrt.Equal(DefaultTimeLayout, opt.TimeLayout)
	asrt.Equal(DefaultFormat, opt.Format)
	asrt.Equal(DefaultDisableCaller, opt.DisableCaller)
	asrt.Equal(DefaultDisableStacktrace, opt.DisableStacktrace)
	asrt.Equal(DefaultDisableSplitError, opt.DisableSplitError)
	asrt.Equal(DefaultMaxSize, opt.MaxSize)
	asrt.Equal(DefaultMaxBackups, opt.MaxBackups)
	asrt.Equal(DefaultCompress, opt.Compress)

	// Test new enhanced functionality fields
	asrt.Equal(DefaultEnableSampling, opt.EnableSampling)
	asrt.Equal(DefaultSampleInitial, opt.SampleInitial)
	asrt.Equal(DefaultSampleThereafter, opt.SampleThereafter)
}

func Test_Options_WithPrefix(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithPrefix("test_")
	asrt.Equal("test_", opt.Prefix)

	opt = NewOptions().WithPrefix("")
	asrt.Equal("", opt.Prefix)
}

func Test_Options_WithDirectory(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithDirectory("test_dir")
	asrt.Equal("test_dir", opt.Directory)

	opt = NewOptions().WithDirectory("")
	asrt.Equal(DefaultDirectory, opt.Directory)
}

func Test_Options_WithFilename(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test setting a valid filename
	opt := NewOptions().WithFilename("myapp")
	asrt.Equal("myapp", opt.Filename)

	// Test setting an empty filename
	opt = NewOptions().WithFilename("")
	asrt.Equal("", opt.Filename)

	// Test setting a filename with special characters
	opt = NewOptions().WithFilename("my-app_v1.0")
	asrt.Equal("my-app_v1.0", opt.Filename)

	// Test chaining with other methods
	opt = NewOptions().WithPrefix("TEST_").WithFilename("service").WithDirectory("/tmp/logs")
	asrt.Equal("TEST_", opt.Prefix)
	asrt.Equal("service", opt.Filename)
	asrt.Equal("/tmp/logs", opt.Directory)
}

func Test_Options_WithLevel(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithLevel(zapcore.DebugLevel.String())
	asrt.Equal(zapcore.DebugLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.InfoLevel.String())
	asrt.Equal(zapcore.InfoLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.WarnLevel.String())
	asrt.Equal(zapcore.WarnLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.ErrorLevel.String())
	asrt.Equal(zapcore.ErrorLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.DPanicLevel.String())
	asrt.Equal(zapcore.DPanicLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.PanicLevel.String())
	asrt.Equal(zapcore.PanicLevel.String(), opt.Level)

	opt = NewOptions().WithLevel(zapcore.FatalLevel.String())
	asrt.Equal(zapcore.FatalLevel.String(), opt.Level)

	opt = NewOptions().WithLevel("")
	asrt.Equal(DefaultLevel.String(), opt.Level)
}

func Test_Options_WithTimeLayout(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithTimeLayout("2006-01-02 15:04:05.000")
	asrt.Equal("2006-01-02 15:04:05.000", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006-01-02 15:04:05")
	asrt.Equal("2006-01-02 15:04:05", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006-01-02 15:04")
	asrt.Equal("2006-01-02 15:04", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006-01-02")
	asrt.Equal("2006-01-02", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006-01")
	asrt.Equal("2006-01", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006")
	asrt.Equal("2006", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01/02 15:04:05.000")
	asrt.Equal("2006/01/02 15:04:05.000", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01/02 15:04:05")
	asrt.Equal("2006/01/02 15:04:05", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01/02 15:04")
	asrt.Equal("2006/01/02 15:04", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01/02")
	asrt.Equal("2006/01/02", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006/01")
	asrt.Equal("2006/01", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("2006")
	asrt.Equal("2006", opt.TimeLayout)

	opt = NewOptions().WithTimeLayout("")
	asrt.Equal(DefaultTimeLayout, opt.TimeLayout)
}

func Test_Options_WithFormat(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithFormat("console")
	asrt.Equal("console", opt.Format)

	opt = NewOptions().WithFormat("json")
	asrt.Equal("json", opt.Format)

	opt = NewOptions().WithFormat("")
	asrt.Equal(DefaultFormat, opt.Format)
}

func Test_Options_WithDisableCaller(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithDisableCaller(true)
	asrt.True(opt.DisableCaller)

	opt = NewOptions().WithDisableCaller(false)
	asrt.False(opt.DisableCaller)
}

func Test_Options_WithDisableStacktrace(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithDisableStacktrace(true)
	asrt.True(opt.DisableStacktrace)

	opt = NewOptions().WithDisableStacktrace(false)
	asrt.False(opt.DisableStacktrace)
}

func Test_Options_WithDisableSplitError(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithDisableSplitError(true)
	asrt.True(opt.DisableSplitError)

	opt = NewOptions().WithDisableSplitError(false)
	asrt.False(opt.DisableSplitError)
}

func Test_Options_WithMaxSize(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithMaxSize(100)
	asrt.Equal(100, opt.MaxSize)

	opt = NewOptions().WithMaxSize(1000)
	asrt.Equal(1000, opt.MaxSize)

	opt = NewOptions().WithMaxSize(10000)
	asrt.Equal(10000, opt.MaxSize)

	opt = NewOptions().WithMaxSize(-1)
	asrt.Equal(DefaultMaxSize, opt.MaxSize)

	opt = NewOptions().WithMaxSize(0)
	asrt.Equal(DefaultMaxSize, opt.MaxSize)
}

func Test_Options_WithMaxBackups(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithMaxBackups(100)
	asrt.Equal(100, opt.MaxBackups)

	opt = NewOptions().WithMaxBackups(1000)
	asrt.Equal(1000, opt.MaxBackups)

	opt = NewOptions().WithMaxBackups(10000)
	asrt.Equal(10000, opt.MaxBackups)

	opt = NewOptions().WithMaxBackups(-1)
	asrt.Equal(DefaultMaxBackups, opt.MaxBackups)

	opt = NewOptions().WithMaxBackups(0)
	asrt.Equal(DefaultMaxBackups, opt.MaxBackups)
}

func Test_Options_WithCompress(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opt := NewOptions().WithCompress(true)
	asrt.True(opt.Compress)

	opt = NewOptions().WithCompress(false)
	asrt.False(opt.Compress)
}

// Test the optimized level validation using slices.Contains
func Test_isValidLevelString(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test valid levels
	validLevels := []string{
		zapcore.DebugLevel.String(),
		zapcore.InfoLevel.String(),
		zapcore.WarnLevel.String(),
		zapcore.ErrorLevel.String(),
		zapcore.DPanicLevel.String(),
		zapcore.PanicLevel.String(),
		zapcore.FatalLevel.String(),
	}

	for _, level := range validLevels {
		asrt.True(isValidLevelString(level), "Level %s should be valid", level)
	}

	// Test invalid levels
	invalidLevels := []string{
		"",
		"invalid",
		"INFO",  // uppercase
		"Debug", // mixed case
		" info", // with space
		"info ", // with trailing space
	}

	for _, level := range invalidLevels {
		asrt.False(isValidLevelString(level), "Level %s should be invalid", level)
	}
}

// Test enhanced options validation
func Test_Options_Validate_Enhanced(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test valid options
	validOpts := NewOptions()
	err := validOpts.Validate()
	asrt.NoError(err)

	// Test valid options with filename
	validOptsWithFilename := NewOptions().WithFilename("myapp")
	err = validOptsWithFilename.Validate()
	asrt.NoError(err)

	// Test invalid directory
	invalidDirOpts := NewOptions()
	invalidDirOpts.Directory = ""
	err = invalidDirOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid directory")

	// Test invalid filename that becomes empty after sanitization
	invalidFilenameOpts := NewOptions()
	invalidFilenameOpts.Filename = "/\\:*?\"<>|"
	err = invalidFilenameOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid filename")

	// Test invalid level
	invalidLevelOpts := NewOptions()
	invalidLevelOpts.Level = "invalid_level"
	err = invalidLevelOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid level")

	// Test invalid format
	invalidFormatOpts := NewOptions()
	invalidFormatOpts.Format = "xml"
	err = invalidFormatOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid format")

	// Test invalid max size
	invalidSizeOpts := NewOptions()
	invalidSizeOpts.MaxSize = -1
	err = invalidSizeOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid max size")

	// Test invalid max backups
	invalidBackupsOpts := NewOptions()
	invalidBackupsOpts.MaxBackups = 0
	err = invalidBackupsOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid max backups")
}

// Test options method chaining with validation
func Test_Options_ChainedValidation(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that invalid values get corrected by With* methods
	opts := NewOptions().
		WithLevel("").     // Should use default
		WithFormat("xml"). // Should use default
		WithMaxSize(-5).   // Should use default
		WithMaxBackups(0). // Should use default
		WithDirectory("")  // Should use default

	// All values should be corrected to defaults
	asrt.Equal(DefaultLevel.String(), opts.Level)
	asrt.Equal(DefaultFormat, opts.Format)
	asrt.Equal(DefaultMaxSize, opts.MaxSize)
	asrt.Equal(DefaultMaxBackups, opts.MaxBackups)
	asrt.Equal(DefaultDirectory, opts.Directory)

	// Validation should pass
	err := opts.Validate()
	asrt.NoError(err)
}

// Test level validation consistency between functions
func Test_LevelValidation_Consistency(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that isValidLevelString and isValidLevel are consistent
	validLevels := []string{
		zapcore.DebugLevel.String(),
		zapcore.InfoLevel.String(),
		zapcore.WarnLevel.String(),
		zapcore.ErrorLevel.String(),
		zapcore.DPanicLevel.String(),
		zapcore.PanicLevel.String(),
		zapcore.FatalLevel.String(),
	}

	for _, level := range validLevels {
		asrt.True(isValidLevelString(level), "isValidLevelString failed for %s", level)
		asrt.True(isValidLevel(level), "isValidLevel failed for %s", level)
	}

	invalidLevels := []string{
		"",
		"invalid",
		"INFO",
	}

	for _, level := range invalidLevels {
		asrt.False(isValidLevelString(level), "isValidLevelString should reject %s", level)
		asrt.False(isValidLevel(level), "isValidLevel should reject %s", level)
	}
}

// Test sanitizeFilename function with various edge cases and invalid inputs
func Test_sanitizeFilename(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test valid filename - should remain unchanged
	result := sanitizeFilename("valid_filename")
	asrt.Equal("valid_filename", result)

	// Test empty string - should return empty
	result = sanitizeFilename("")
	asrt.Equal("", result)

	// Test filename with unsafe characters - should be replaced with underscores
	result = sanitizeFilename("file/name\\with:unsafe*chars?\"<>|")
	asrt.Equal("file_name_with_unsafe_chars_____", result)

	// Test filename with whitespace - should be trimmed
	result = sanitizeFilename("  filename_with_spaces  ")
	asrt.Equal("filename_with_spaces", result)

	// Test filename starting with dot - should be prefixed with underscore
	result = sanitizeFilename(".hidden_file")
	asrt.Equal("_hidden_file", result)

	// Test very long filename - should be truncated to 100 characters
	longFilename := strings.Repeat("a", 150)
	result = sanitizeFilename(longFilename)
	asrt.Equal(100, len(result))
	asrt.Equal(strings.Repeat("a", 100), result)

	// Test filename with control characters - should be replaced
	result = sanitizeFilename("file\x00name\x1f")
	asrt.Equal("file_name_", result)

	// Test filename with DEL character (ASCII 127) - should be replaced
	result = sanitizeFilename("file\x7fname")
	asrt.Equal("file_name", result)

	// Test filename that becomes empty after cleaning - should return empty
	result = sanitizeFilename("   ")
	asrt.Equal("", result)

	// Test filename with only unsafe characters - should return empty after cleaning
	result = sanitizeFilename("/\\:*?\"<>|")
	asrt.Equal("", result)

	// Test filename with mixed valid and invalid characters
	result = sanitizeFilename("my-app_v1.2.3/config\\file:backup*")
	asrt.Equal("my-app_v1.2.3_config_file_backup_", result)

	// Test filename with Unicode characters - should be preserved
	result = sanitizeFilename("文件名_测试")
	asrt.Equal("文件名_测试", result)

	// Test filename with numbers and special allowed characters
	result = sanitizeFilename("file-123_test.backup")
	asrt.Equal("file-123_test.backup", result)

	// Test edge case: exactly 100 characters
	exactLength := strings.Repeat("a", 100)
	result = sanitizeFilename(exactLength)
	asrt.Equal(100, len(result))
	asrt.Equal(exactLength, result)

	// Test edge case: 101 characters should be truncated
	overLength := strings.Repeat("a", 101)
	result = sanitizeFilename(overLength)
	asrt.Equal(100, len(result))

	// Test combination of issues: long filename with unsafe chars and leading dot
	complexFilename := "." + strings.Repeat("a/b\\c", 30) // Creates a long filename with unsafe chars
	result = sanitizeFilename(complexFilename)
	asrt.Equal(100, len(result))
	asrt.True(strings.HasPrefix(result, "_"))
	asrt.NotContains(result, "/")
	asrt.NotContains(result, "\\")
}

// Test sanitizeFilename boundary conditions
func Test_sanitizeFilename_BoundaryConditions(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test all unsafe characters individually
	unsafeChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range unsafeChars {
		result := sanitizeFilename("file" + char + "name")
		asrt.Equal("file_name", result, "Failed to sanitize character: %s", char)
	}

	// Test all control characters (0-31 and 127)
	for i := range 32 {
		filename := "file" + string(rune(i)) + "name"
		result := sanitizeFilename(filename)
		asrt.Equal("file_name", result, "Failed to sanitize control character: %d", i)
	}

	// Test DEL character (127)
	result := sanitizeFilename("file" + string(rune(127)) + "name")
	asrt.Equal("file_name", result)

	// Test multiple consecutive unsafe characters
	result = sanitizeFilename("file///name")
	asrt.Equal("file___name", result)

	// Test filename that becomes empty after removing only dots and spaces
	result = sanitizeFilename("  ...  ")
	asrt.Equal("", result)

	// Test filename with only control characters
	result = sanitizeFilename("\x00\x01\x02")
	asrt.Equal("", result)
}

// Test sanitizeFilename with realistic scenarios
func Test_sanitizeFilename_RealisticScenarios(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test typical application names
	testCases := []struct {
		input    string
		expected string
	}{
		{"my-web-app", "my-web-app"},
		{"MyApp_v2.1", "MyApp_v2.1"},
		{"service:8080", "service_8080"},
		{"app/logs", "app_logs"},
		{"backup-2024-01-01", "backup-2024-01-01"},
		{"temp*file", "temp_file"},
		{"config.json", "config.json"},
		{"log_file_2024", "log_file_2024"},
		{"user@domain", "user@domain"}, // @ is allowed
		{"file#1", "file#1"},           // # is allowed
		{"test$var", "test$var"},       // $ is allowed
		{"file%20name", "file%20name"}, // % is allowed
		{"app&service", "app&service"}, // & is allowed
		{"file(1)", "file(1)"},         // () are allowed
		{"file[0]", "file[0]"},         // [] are allowed
		{"file{id}", "file{id}"},       // {} are allowed
		{"file+backup", "file+backup"}, // + is allowed
		{"file=value", "file=value"},   // = is allowed
		{"file,list", "file,list"},     // , is allowed
		{"file;end", "file;end"},       // ; is allowed
		{"file'quote", "file'quote"},   // ' is allowed
		{"file~temp", "file~temp"},     // ~ is allowed
		{"file`cmd", "file`cmd"},       // ` is allowed
	}

	for _, tc := range testCases {
		result := sanitizeFilename(tc.input)
		asrt.Equal(tc.expected, result, "Failed for input: %s", tc.input)
	}
}

// Test filename validation in Options.Validate()
func Test_Options_Validate_Filename(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test valid filenames
	validFilenames := []string{
		"",           // Empty filename should be valid (uses default)
		"myapp",      // Simple filename
		"my-app",     // With dash
		"my_app",     // With underscore
		"app123",     // With numbers
		"MyApp",      // With uppercase
		"app.log",    // With dot
		"service-v1", // Complex valid name
	}

	for _, filename := range validFilenames {
		opts := NewOptions().WithFilename(filename)
		err := opts.Validate()
		asrt.NoError(err, "Filename '%s' should be valid", filename)
	}

	// Test invalid filenames that become empty after sanitization
	invalidFilenames := []string{
		"/\\:*?\"<>|", // All unsafe characters
		"   ",         // Only spaces
		"...",         // Only dots
		"  ...  ",     // Spaces and dots
		"\x00\x01",    // Control characters
	}

	for _, filename := range invalidFilenames {
		opts := NewOptions().WithFilename(filename)
		err := opts.Validate()
		asrt.Error(err, "Filename '%s' should be invalid", filename)
		asrt.Contains(err.Error(), "invalid filename")
	}

	// Test filenames that get sanitized but remain valid
	sanitizedFilenames := []struct {
		input    string
		expected bool // whether validation should pass
	}{
		{"my/app", true},    // Gets sanitized to "my_app"
		{"app:log", true},   // Gets sanitized to "app_log"
		{"file*name", true}, // Gets sanitized to "file_name"
		{".hidden", true},   // Gets sanitized to "_hidden"
		{"app\\log", true},  // Gets sanitized to "app_log"
	}

	for _, tc := range sanitizedFilenames {
		opts := NewOptions().WithFilename(tc.input)
		err := opts.Validate()
		if tc.expected {
			asrt.NoError(err, "Filename '%s' should be valid after sanitization", tc.input)
		} else {
			asrt.Error(err, "Filename '%s' should be invalid even after sanitization", tc.input)
		}
	}
}

// Test filename integration with other options
func Test_Options_Filename_Integration(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that filename works with method chaining
	opts := NewOptions().
		WithPrefix("TEST_").
		WithFilename("myservice").
		WithDirectory("/var/log").
		WithLevel("debug").
		WithFormat("json")

	asrt.Equal("TEST_", opts.Prefix)
	asrt.Equal("myservice", opts.Filename)
	asrt.Equal("/var/log", opts.Directory)
	asrt.Equal("debug", opts.Level)
	asrt.Equal("json", opts.Format)

	// Validation should pass
	err := opts.Validate()
	asrt.NoError(err)

	// Test that filename can be changed multiple times
	opts = NewOptions().
		WithFilename("first").
		WithFilename("second").
		WithFilename("final")

	asrt.Equal("final", opts.Filename)
	err = opts.Validate()
	asrt.NoError(err)
}

// Test DefaultFilename constant
func Test_DefaultFilename(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that DefaultFilename is empty string
	asrt.Equal("", DefaultFilename)

	// Test that NewOptions uses DefaultFilename
	opts := NewOptions()
	asrt.Equal(DefaultFilename, opts.Filename)
}

func Test_Options_WithSampling(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test enabling sampling with valid values
	opt := NewOptions().WithSampling(true, 50, 200)
	asrt.True(opt.EnableSampling)
	asrt.Equal(50, opt.SampleInitial)
	asrt.Equal(200, opt.SampleThereafter)

	// Test disabling sampling
	opt = NewOptions().WithSampling(false, 50, 200)
	asrt.False(opt.EnableSampling)
	asrt.Equal(50, opt.SampleInitial)
	asrt.Equal(200, opt.SampleThereafter)

	// Test enabling sampling with zero initial - should use default
	opt = NewOptions().WithSampling(true, 0, 200)
	asrt.True(opt.EnableSampling)
	asrt.Equal(DefaultSampleInitial, opt.SampleInitial)
	asrt.Equal(200, opt.SampleThereafter)

	// Test enabling sampling with negative initial - should use default
	opt = NewOptions().WithSampling(true, -10, 200)
	asrt.True(opt.EnableSampling)
	asrt.Equal(DefaultSampleInitial, opt.SampleInitial)
	asrt.Equal(200, opt.SampleThereafter)

	// Test enabling sampling with zero thereafter - should use default
	opt = NewOptions().WithSampling(true, 50, 0)
	asrt.True(opt.EnableSampling)
	asrt.Equal(50, opt.SampleInitial)
	asrt.Equal(DefaultSampleThereafter, opt.SampleThereafter)

	// Test enabling sampling with negative thereafter - should use default
	opt = NewOptions().WithSampling(true, 50, -10)
	asrt.True(opt.EnableSampling)
	asrt.Equal(50, opt.SampleInitial)
	asrt.Equal(DefaultSampleThereafter, opt.SampleThereafter)

	// Test enabling sampling with both values zero - should use defaults
	opt = NewOptions().WithSampling(true, 0, 0)
	asrt.True(opt.EnableSampling)
	asrt.Equal(DefaultSampleInitial, opt.SampleInitial)
	asrt.Equal(DefaultSampleThereafter, opt.SampleThereafter)
}

// Test enhanced validation for new fields
func Test_Options_Validate_Enhanced_Fields(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test valid options with new fields
	validOpts := NewOptions().
		WithSampling(true, 50, 200)
	err := validOpts.Validate()
	asrt.NoError(err)

	// Test invalid sample initial when sampling is enabled
	invalidSampleInitialOpts := NewOptions()
	invalidSampleInitialOpts.EnableSampling = true
	invalidSampleInitialOpts.SampleInitial = 0
	err = invalidSampleInitialOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid sample initial")

	// Test invalid sample thereafter when sampling is enabled
	invalidSampleThereafterOpts := NewOptions()
	invalidSampleThereafterOpts.EnableSampling = true
	invalidSampleThereafterOpts.SampleThereafter = -1
	err = invalidSampleThereafterOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid sample thereafter")

	// Test that sampling validation is skipped when sampling is disabled
	disabledSamplingOpts := NewOptions()
	disabledSamplingOpts.EnableSampling = false
	disabledSamplingOpts.SampleInitial = 0
	disabledSamplingOpts.SampleThereafter = -1
	err = disabledSamplingOpts.Validate()
	asrt.NoError(err) // Should pass because sampling is disabled
}

// Test method chaining with new fields
func Test_Options_ChainedValidation_Enhanced(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test chaining all methods including new ones
	opts := NewOptions().
		WithPrefix("TEST_").
		WithDirectory("/var/log").
		WithFilename("myapp").
		WithLevel("debug").
		WithFormat("json").
		WithSampling(true, 25, 500).
		WithMaxSize(200).
		WithMaxBackups(10).
		WithCompress(true)

	// Verify all values are set correctly
	asrt.Equal("TEST_", opts.Prefix)
	asrt.Equal("/var/log", opts.Directory)
	asrt.Equal("myapp", opts.Filename)
	asrt.Equal("debug", opts.Level)
	asrt.Equal("json", opts.Format)
	asrt.True(opts.EnableSampling)
	asrt.Equal(25, opts.SampleInitial)
	asrt.Equal(500, opts.SampleThereafter)
	asrt.Equal(200, opts.MaxSize)
	asrt.Equal(10, opts.MaxBackups)
	asrt.True(opts.Compress)

	// Validation should pass
	err := opts.Validate()
	asrt.NoError(err)
}

// Test default constants for new fields
func Test_DefaultConstants_Enhanced(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that default constants are reasonable
	asrt.False(DefaultEnableSampling)
	asrt.Equal(100, DefaultSampleInitial)
	asrt.Equal(100, DefaultSampleThereafter)

	// Test that NewOptions uses these defaults
	opts := NewOptions()
	asrt.Equal(DefaultEnableSampling, opts.EnableSampling)
	asrt.Equal(DefaultSampleInitial, opts.SampleInitial)
	asrt.Equal(DefaultSampleThereafter, opts.SampleThereafter)
}

// Test edge cases for new fields
func Test_Options_Enhanced_EdgeCases(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test basic validation
	opt := NewOptions()
	err := opt.Validate()
	asrt.NoError(err)

	// Test very large sample values
	opt = NewOptions().WithSampling(true, 10000, 50000)
	asrt.True(opt.EnableSampling)
	asrt.Equal(10000, opt.SampleInitial)
	asrt.Equal(50000, opt.SampleThereafter)
	err = opt.Validate()
	asrt.NoError(err)

	// Test boundary values
	err = opt.Validate()
	asrt.NoError(err)

	opt = NewOptions().WithSampling(true, 1, 1) // Minimum valid sample values
	asrt.True(opt.EnableSampling)
	asrt.Equal(1, opt.SampleInitial)
	asrt.Equal(1, opt.SampleThereafter)
	err = opt.Validate()
	asrt.NoError(err)
}

// Test enhanced Options fields that were added for usability improvements
func TestOptions_EnhancedFields_Defaults(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	opts := NewOptions()
	asrt.NotNil(opts)

	// Test that enhanced fields have correct default values
	asrt.Equal(DefaultEnableSampling, opts.EnableSampling)
	asrt.Equal(DefaultSampleInitial, opts.SampleInitial)
	asrt.Equal(DefaultSampleThereafter, opts.SampleThereafter)
}

func TestOptions_WithSampling_ValidValues(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	testCases := []struct {
		name               string
		enable             bool
		initial            int
		thereafter         int
		expectedEnable     bool
		expectedInitial    int
		expectedThereafter int
	}{
		{
			"enable with valid values",
			true, 50, 200,
			true, 50, 200,
		},
		{
			"disable with valid values",
			false, 50, 200,
			false, 50, 200,
		},
		{
			"enable with zero initial uses default",
			true, 0, 200,
			true, DefaultSampleInitial, 200,
		},
		{
			"enable with negative initial uses default",
			true, -10, 200,
			true, DefaultSampleInitial, 200,
		},
		{
			"enable with zero thereafter uses default",
			true, 50, 0,
			true, 50, DefaultSampleThereafter,
		},
		{
			"enable with negative thereafter uses default",
			true, 50, -10,
			true, 50, DefaultSampleThereafter,
		},
		{
			"enable with both zero uses defaults",
			true, 0, 0,
			true, DefaultSampleInitial, DefaultSampleThereafter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewOptions().WithSampling(tc.enable, tc.initial, tc.thereafter)
			asrt.Equal(tc.expectedEnable, opts.EnableSampling)
			asrt.Equal(tc.expectedInitial, opts.SampleInitial)
			asrt.Equal(tc.expectedThereafter, opts.SampleThereafter)
		})
	}
}

func TestOptions_Validate_EnhancedFields(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test valid enhanced fields
	validOpts := NewOptions().
		WithSampling(true, 50, 200)

	err := validOpts.Validate()
	asrt.NoError(err)

	// Test invalid sample initial when sampling enabled
	invalidSampleInitialOpts := NewOptions()
	invalidSampleInitialOpts.EnableSampling = true
	invalidSampleInitialOpts.SampleInitial = 0
	err = invalidSampleInitialOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid sample initial")

	// Test invalid sample thereafter when sampling enabled
	invalidSampleThereafterOpts := NewOptions()
	invalidSampleThereafterOpts.EnableSampling = true
	invalidSampleThereafterOpts.SampleThereafter = -1
	err = invalidSampleThereafterOpts.Validate()
	asrt.Error(err)
	asrt.Contains(err.Error(), "invalid sample thereafter")

	// Test that sampling validation is skipped when disabled
	disabledSamplingOpts := NewOptions()
	disabledSamplingOpts.EnableSampling = false
	disabledSamplingOpts.SampleInitial = 0
	disabledSamplingOpts.SampleThereafter = -1
	err = disabledSamplingOpts.Validate()
	asrt.NoError(err) // Should pass because sampling is disabled
}

func TestOptions_EnhancedFields_MethodChaining(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test chaining all enhanced methods
	opts := NewOptions().
		WithSampling(true, 25, 500)

	asrt.True(opts.EnableSampling)
	asrt.Equal(25, opts.SampleInitial)
	asrt.Equal(500, opts.SampleThereafter)

	// Test chaining with existing methods
	opts = NewOptions().
		WithLevel("debug").
		WithFormat("json").
		WithSampling(false, 100, 1000).
		WithMaxSize(200)

	asrt.Equal("debug", opts.Level)
	asrt.Equal("json", opts.Format)
	asrt.False(opts.EnableSampling)
	asrt.Equal(100, opts.SampleInitial)
	asrt.Equal(1000, opts.SampleThereafter)
	asrt.Equal(200, opts.MaxSize)

	// Validation should pass
	err := opts.Validate()
	asrt.NoError(err)
}

func TestOptions_EnhancedFields_EdgeCases(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test maximum values
	opts := NewOptions().
		WithSampling(true, 10000, 50000) // Large sample values

	err := opts.Validate()
	asrt.NoError(err)
	asrt.Equal(10000, opts.SampleInitial)
	asrt.Equal(50000, opts.SampleThereafter)

	// Test minimum valid values
	opts = NewOptions().
		WithSampling(true, 1, 1) // Minimum sample values

	err = opts.Validate()
	asrt.NoError(err)
	asrt.Equal(1, opts.SampleInitial)
	asrt.Equal(1, opts.SampleThereafter)
}

func TestOptions_EnhancedFields_DefaultConstants(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that default constants are reasonable
	asrt.False(DefaultEnableSampling)
	asrt.Equal(100, DefaultSampleInitial)
	asrt.Equal(100, DefaultSampleThereafter)

	// Test that constants are used in NewOptions
	opts := NewOptions()
	asrt.Equal(DefaultEnableSampling, opts.EnableSampling)
	asrt.Equal(DefaultSampleInitial, opts.SampleInitial)
	asrt.Equal(DefaultSampleThereafter, opts.SampleThereafter)
}

func TestOptions_EnhancedFields_Integration(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that enhanced fields work with all existing fields
	opts := NewOptions().
		WithPrefix("ENHANCED_").
		WithDirectory("/tmp/enhanced").
		WithFilename("enhanced").
		WithLevel("debug").
		WithTimeLayout("2006-01-02 15:04:05").
		WithFormat("json").
		WithDisableCaller(true).
		WithDisableStacktrace(false).
		WithDisableSplitError(true).
		WithMaxSize(150).
		WithMaxBackups(7).
		WithCompress(true).
		WithSampling(true, 75, 750)

	// Verify all fields are set correctly
	asrt.Equal("ENHANCED_", opts.Prefix)
	asrt.Equal("/tmp/enhanced", opts.Directory)
	asrt.Equal("enhanced", opts.Filename)
	asrt.Equal("debug", opts.Level)
	asrt.Equal("2006-01-02 15:04:05", opts.TimeLayout)
	asrt.Equal("json", opts.Format)
	asrt.True(opts.DisableCaller)
	asrt.False(opts.DisableStacktrace)
	asrt.True(opts.DisableSplitError)
	asrt.Equal(150, opts.MaxSize)
	asrt.Equal(7, opts.MaxBackups)
	asrt.True(opts.Compress)
	asrt.True(opts.EnableSampling)
	asrt.Equal(75, opts.SampleInitial)
	asrt.Equal(750, opts.SampleThereafter)

	// Validation should pass
	err := opts.Validate()
	asrt.NoError(err)
}

func TestOptions_EnhancedFields_MultipleModifications(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that fields can be modified multiple times
	opts := NewOptions()

	opts = NewOptions().
		WithSampling(true, 100, 1000).
		WithSampling(false, 200, 2000).
		WithSampling(true, 50, 500)

	asrt.True(opts.EnableSampling)
	asrt.Equal(50, opts.SampleInitial)
	asrt.Equal(500, opts.SampleThereafter)
}

func TestOptions_EnhancedFields_ZeroDurationHandling(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// Test that zero duration is handled correctly in validation
	opts := NewOptions()
	err := opts.Validate()
	asrt.NoError(err) // Zero should be valid (no flushing)

	// Test that zero is preserved when set explicitly

	// But if set directly, zero should be valid
	opts = NewOptions()
	err = opts.Validate()
	asrt.NoError(err)
}

func TestOptions_EnhancedFields_SamplingDisabledValidation(t *testing.T) {
	t.Parallel()
	asrt := assert.New(t)

	// When sampling is disabled, sample values should not be validated
	opts := NewOptions()
	opts.EnableSampling = false
	opts.SampleInitial = -100    // Invalid value
	opts.SampleThereafter = -200 // Invalid value

	err := opts.Validate()
	asrt.NoError(err) // Should pass because sampling is disabled

	// When sampling is enabled, sample values should be validated
	opts.EnableSampling = true
	err = opts.Validate()
	asrt.Error(err) // Should fail because sample values are invalid
	asrt.Contains(err.Error(), "invalid sample initial")
}
