package log

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    interface{}
		reason   string
		wrapped  error
		expected string
	}{
		{
			name:     "simple config error",
			field:    "Level",
			value:    "invalid",
			reason:   "无效的日志级别",
			wrapped:  nil,
			expected: "配置字段 Level 错误: 无效的日志级别 (值: invalid)",
		},
		{
			name:     "wrapped config error",
			field:    "Directory",
			value:    "/invalid/path",
			reason:   "目录不可访问",
			wrapped:  errors.New("permission denied"),
			expected: "配置字段 Directory 错误: 目录不可访问 (值: /invalid/path): permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigError(tt.field, tt.value, tt.reason, tt.wrapped)
			if err.Error() != tt.expected {
				t.Errorf("ConfigError.Error() = %v, want %v", err.Error(), tt.expected)
			}

			if tt.wrapped != nil && !errors.Is(err, tt.wrapped) {
				t.Errorf("ConfigError should wrap the original error")
			}
		})
	}
}

func TestValidateOptions(t *testing.T) {
	tests := []struct {
		name        string
		input       *Options
		expectError bool
		checkField  func(*Options) bool
	}{
		{
			name:        "nil options",
			input:       nil,
			expectError: false,
			checkField: func(opts *Options) bool {
				return opts != nil && opts.Level == DefaultLevel.String()
			},
		},
		{
			name: "invalid level",
			input: &Options{
				Level:      "invalid_level",
				Directory:  DefaultDirectory,
				Format:     DefaultFormat,
				MaxSize:    DefaultMaxSize,
				MaxBackups: DefaultMaxBackups,
			},
			expectError: true,
			checkField: func(opts *Options) bool {
				return opts.Level == DefaultLevel.String()
			},
		},
		{
			name: "invalid format",
			input: &Options{
				Level:      DefaultLevel.String(),
				Directory:  DefaultDirectory,
				Format:     "invalid_format",
				MaxSize:    DefaultMaxSize,
				MaxBackups: DefaultMaxBackups,
			},
			expectError: true,
			checkField: func(opts *Options) bool {
				return opts.Format == DefaultFormat
			},
		},
		{
			name: "empty directory",
			input: &Options{
				Level:      DefaultLevel.String(),
				Directory:  "",
				Format:     DefaultFormat,
				MaxSize:    DefaultMaxSize,
				MaxBackups: DefaultMaxBackups,
			},
			expectError: true,
			checkField: func(opts *Options) bool {
				return opts.Directory == DefaultDirectory
			},
		},
		{
			name: "invalid max size",
			input: &Options{
				Level:      DefaultLevel.String(),
				Directory:  DefaultDirectory,
				Format:     DefaultFormat,
				MaxSize:    -1,
				MaxBackups: DefaultMaxBackups,
			},
			expectError: true,
			checkField: func(opts *Options) bool {
				return opts.MaxSize == DefaultMaxSize
			},
		},
		{
			name: "invalid buffer size",
			input: &Options{
				Level:      DefaultLevel.String(),
				Directory:  DefaultDirectory,
				Format:     DefaultFormat,
				MaxSize:    DefaultMaxSize,
				MaxBackups: DefaultMaxBackups,
				BufferSize: -1,
			},
			expectError: true,
			checkField: func(opts *Options) bool {
				return opts.BufferSize == DefaultBufferSize
			},
		},
		{
			name: "invalid sampling config",
			input: &Options{
				Level:            DefaultLevel.String(),
				Directory:        DefaultDirectory,
				Format:           DefaultFormat,
				MaxSize:          DefaultMaxSize,
				MaxBackups:       DefaultMaxBackups,
				EnableSampling:   true,
				SampleInitial:    -1,
				SampleThereafter: -1,
			},
			expectError: true,
			checkField: func(opts *Options) bool {
				return opts.SampleInitial == DefaultSampleInitial && opts.SampleThereafter == DefaultSampleThereafter
			},
		},
		{
			name: "valid options",
			input: &Options{
				Level:      DefaultLevel.String(),
				Directory:  DefaultDirectory,
				Format:     DefaultFormat,
				TimeLayout: DefaultTimeLayout,
				MaxSize:    DefaultMaxSize,
				MaxBackups: DefaultMaxBackups,
			},
			expectError: false,
			checkField: func(opts *Options) bool {
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateOptions(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("ValidateOptions() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("ValidateOptions() unexpected error: %v", err)
			}

			if result == nil {
				t.Errorf("ValidateOptions() returned nil options")
				return
			}

			if !tt.checkField(result) {
				t.Errorf("ValidateOptions() field validation failed")
			}
		})
	}
}

func TestValidateLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectError   bool
		expectedLevel string
	}{
		{"valid debug", "debug", false, "debug"},
		{"valid info", "info", false, "info"},
		{"valid warn", "warn", false, "warn"},
		{"valid error", "error", false, "error"},
		{"invalid level", "invalid", true, DefaultLevel.String()},
		{"empty level", "", true, DefaultLevel.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{Level: tt.level}
			err := validateLevel(opts)

			if tt.expectError && err == nil {
				t.Errorf("validateLevel() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("validateLevel() unexpected error: %v", err)
			}

			if opts.Level != tt.expectedLevel {
				t.Errorf("validateLevel() level = %v, want %v", opts.Level, tt.expectedLevel)
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name           string
		format         string
		expectError    bool
		expectedFormat string
	}{
		{"valid console", "console", false, "console"},
		{"valid json", "json", false, "json"},
		{"invalid format", "xml", true, DefaultFormat},
		{"empty format", "", true, DefaultFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{Format: tt.format}
			err := validateFormat(opts)

			if tt.expectError && err == nil {
				t.Errorf("validateFormat() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("validateFormat() unexpected error: %v", err)
			}

			if opts.Format != tt.expectedFormat {
				t.Errorf("validateFormat() format = %v, want %v", opts.Format, tt.expectedFormat)
			}
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "log_test")
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name              string
		directory         string
		expectError       bool
		expectedDirectory string
	}{
		{"empty directory", "", true, DefaultDirectory},
		{"valid directory", tempDir, false, tempDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{Directory: tt.directory}
			err := validateDirectory(opts)

			if tt.expectError && err == nil {
				t.Errorf("validateDirectory() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("validateDirectory() unexpected error: %v", err)
			}

			if opts.Directory != tt.expectedDirectory {
				t.Errorf("validateDirectory() directory = %v, want %v", opts.Directory, tt.expectedDirectory)
			}
		})
	}
}

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name             string
		filename         string
		expectError      bool
		expectedFilename string
	}{
		{"valid filename", "myapp", false, "myapp"},
		{"empty filename", "", false, ""},
		{"filename with unsafe chars", "my/app*log", true, "my_app_log"},
		{"filename with only unsafe chars", "/*?", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{Filename: tt.filename}
			err := validateFilename(opts)

			if tt.expectError && err == nil {
				t.Errorf("validateFilename() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("validateFilename() unexpected error: %v", err)
			}

			if opts.Filename != tt.expectedFilename {
				t.Errorf("validateFilename() filename = %v, want %v", opts.Filename, tt.expectedFilename)
			}
		})
	}
}

func TestValidateTimeLayout(t *testing.T) {
	tests := []struct {
		name               string
		timeLayout         string
		expectError        bool
		expectedTimeLayout string
	}{
		{"valid layout", "2006-01-02 15:04:05", false, "2006-01-02 15:04:05"},
		{"default layout", DefaultTimeLayout, false, DefaultTimeLayout},
		{"empty layout", "", true, DefaultTimeLayout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{TimeLayout: tt.timeLayout}
			err := validateTimeLayout(opts)

			if tt.expectError && err == nil {
				t.Errorf("validateTimeLayout() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("validateTimeLayout() unexpected error: %v", err)
			}

			if opts.TimeLayout != tt.expectedTimeLayout {
				t.Errorf("validateTimeLayout() timeLayout = %v, want %v", opts.TimeLayout, tt.expectedTimeLayout)
			}
		})
	}
}

func TestValidateSampling(t *testing.T) {
	tests := []struct {
		name                     string
		enableSampling           bool
		sampleInitial            int
		sampleThereafter         int
		expectError              bool
		expectedSampleInitial    int
		expectedSampleThereafter int
	}{
		{
			name:                     "sampling disabled",
			enableSampling:           false,
			sampleInitial:            -1,
			sampleThereafter:         -1,
			expectError:              false,
			expectedSampleInitial:    -1,
			expectedSampleThereafter: -1,
		},
		{
			name:                     "valid sampling config",
			enableSampling:           true,
			sampleInitial:            100,
			sampleThereafter:         1000,
			expectError:              false,
			expectedSampleInitial:    100,
			expectedSampleThereafter: 1000,
		},
		{
			name:                     "invalid sampling config",
			enableSampling:           true,
			sampleInitial:            -1,
			sampleThereafter:         -1,
			expectError:              true,
			expectedSampleInitial:    DefaultSampleInitial,
			expectedSampleThereafter: DefaultSampleThereafter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{
				EnableSampling:   tt.enableSampling,
				SampleInitial:    tt.sampleInitial,
				SampleThereafter: tt.sampleThereafter,
			}
			err := validateSampling(opts)

			if tt.expectError && err == nil {
				t.Errorf("validateSampling() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("validateSampling() unexpected error: %v", err)
			}

			if opts.SampleInitial != tt.expectedSampleInitial {
				t.Errorf("validateSampling() sampleInitial = %v, want %v", opts.SampleInitial, tt.expectedSampleInitial)
			}

			if opts.SampleThereafter != tt.expectedSampleThereafter {
				t.Errorf("validateSampling() sampleThereafter = %v, want %v", opts.SampleThereafter, tt.expectedSampleThereafter)
			}
		})
	}
}

func TestRecoverFromConfigError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, true},
		{"config error", NewConfigError("Level", "invalid", "test", nil), true},
		{"other error", errors.New("some error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RecoverFromConfigError(tt.err)
			if (result != nil) != tt.expected {
				t.Errorf("RecoverFromConfigError() = %v, want non-nil: %v", result, tt.expected)
			}
		})
	}
}

func TestValidateAndFixOptions(t *testing.T) {
	// Test with invalid options
	opts := &Options{
		Level:     "invalid",
		Directory: "",
		Format:    "xml",
		MaxSize:   -1,
	}

	result := ValidateAndFixOptions(opts)

	if result == nil {
		t.Errorf("ValidateAndFixOptions() returned nil")
		return
	}

	if result.Level != DefaultLevel.String() {
		t.Errorf("ValidateAndFixOptions() level = %v, want %v", result.Level, DefaultLevel.String())
	}

	if result.Directory != DefaultDirectory {
		t.Errorf("ValidateAndFixOptions() directory = %v, want %v", result.Directory, DefaultDirectory)
	}

	if result.Format != DefaultFormat {
		t.Errorf("ValidateAndFixOptions() format = %v, want %v", result.Format, DefaultFormat)
	}

	if result.MaxSize != DefaultMaxSize {
		t.Errorf("ValidateAndFixOptions() maxSize = %v, want %v", result.MaxSize, DefaultMaxSize)
	}
}

func TestIsConfigError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"config error", NewConfigError("test", "value", "reason", nil), true},
		{"other error", errors.New("not a config error"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConfigError(tt.err)
			if result != tt.expected {
				t.Errorf("IsConfigError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetConfigErrors(t *testing.T) {
	configErr := NewConfigError("test", "value", "reason", nil)
	otherErr := errors.New("not a config error")

	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"config error", configErr, 1},
		{"other error", otherErr, 0},
		{"nil error", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetConfigErrors(tt.err)
			if len(result) != tt.expected {
				t.Errorf("GetConfigErrors() returned %d errors, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestEnsureDirectoryExists(t *testing.T) {
	// Test with a temporary directory
	tempDir := filepath.Join(os.TempDir(), "log_test_ensure")
	defer os.RemoveAll(tempDir)

	// Test creating new directory
	err := ensureDirectoryExists(tempDir)
	if err != nil {
		t.Errorf("ensureDirectoryExists() failed to create directory: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("ensureDirectoryExists() did not create directory")
	}

	// Test with existing directory
	err = ensureDirectoryExists(tempDir)
	if err != nil {
		t.Errorf("ensureDirectoryExists() failed with existing directory: %v", err)
	}

	// Test with file instead of directory
	testFile := filepath.Join(tempDir, "testfile")
	if f, err := os.Create(testFile); err == nil {
		f.Close()
		err = ensureDirectoryExists(testFile)
		if err == nil {
			t.Errorf("ensureDirectoryExists() should fail when path is a file")
		}
	}
}
