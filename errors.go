package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConfigError represents a configuration error with detailed information
type ConfigError struct {
	Field   string // The configuration field that has an error
	Value   any    // The invalid value
	Reason  string // Human-readable reason for the error
	Wrapped error  // The underlying error, if any
}

// Error implements the error interface
func (e *ConfigError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("配置字段 %s 错误: %s (值: %v): %v", e.Field, e.Reason, e.Value, e.Wrapped)
	}
	return fmt.Sprintf("配置字段 %s 错误: %s (值: %v)", e.Field, e.Reason, e.Value)
}

// Unwrap returns the wrapped error for error chain support
func (e *ConfigError) Unwrap() error {
	return e.Wrapped
}

// NewConfigError creates a new ConfigError with the specified details
func NewConfigError(field string, value any, reason string, wrapped error) *ConfigError {
	return &ConfigError{
		Field:   field,
		Value:   value,
		Reason:  reason,
		Wrapped: wrapped,
	}
}

// Predefined error variables for common configuration errors
var (
	ErrInvalidLevel      = errors.New("无效的日志级别")
	ErrInvalidFormat     = errors.New("无效的日志格式")
	ErrInvalidDirectory  = errors.New("无效的日志目录")
	ErrInvalidFilename   = errors.New("无效的文件名")
	ErrPermissionDenied  = errors.New("权限被拒绝")
	ErrInvalidTimeLayout = errors.New("无效的时间格式")
	ErrInvalidMaxSize    = errors.New("无效的最大文件大小")
	ErrInvalidMaxBackups = errors.New("无效的最大备份数量")

	ErrInvalidSampling = errors.New("无效的采样配置")
)

// ValidateOptions validates configuration options and returns a fixed configuration
// along with any validation errors encountered. This function provides automatic
// error recovery by using safe default values when invalid configurations are detected.
func ValidateOptions(opts *Options) (*Options, error) {
	if opts == nil {
		return NewOptions(), nil
	}

	// Create a copy of the options to avoid modifying the original
	fixed := *opts
	var errs []error

	// Validate and fix each field
	if err := validateLevel(&fixed); err != nil {
		errs = append(errs, err)
	}

	if err := validateFormat(&fixed); err != nil {
		errs = append(errs, err)
	}

	if err := validateDirectory(&fixed); err != nil {
		errs = append(errs, err)
	}

	if err := validateFilename(&fixed); err != nil {
		errs = append(errs, err)
	}

	if err := validateTimeLayout(&fixed); err != nil {
		errs = append(errs, err)
	}

	if err := validateMaxSize(&fixed); err != nil {
		errs = append(errs, err)
	}

	if err := validateMaxBackups(&fixed); err != nil {
		errs = append(errs, err)
	}

	if err := validateSampling(&fixed); err != nil {
		errs = append(errs, err)
	}

	// If there were validation errors, return the fixed configuration with error details
	if len(errs) > 0 {
		return &fixed, fmt.Errorf("配置验证失败，已自动修复: %v", errs)
	}

	return &fixed, nil
}

// validateLevel validates and fixes the log level
func validateLevel(opts *Options) error {
	if opts.Level == "" || !isValidLevelString(opts.Level) {
		originalValue := opts.Level
		opts.Level = DefaultLevel.String()
		return NewConfigError("Level", originalValue, "使用默认级别 "+DefaultLevel.String(), ErrInvalidLevel)
	}
	return nil
}

// validateFormat validates and fixes the log format
func validateFormat(opts *Options) error {
	if opts.Format != FormatConsole && opts.Format != FormatJSON {
		originalValue := opts.Format
		opts.Format = DefaultFormat
		return NewConfigError("Format", originalValue, "Use Default Log Format "+DefaultFormat, ErrInvalidFormat)
	}
	return nil
}

// validateDirectory validates and fixes the log directory
func validateDirectory(opts *Options) error {
	if opts.Directory == "" {
		opts.Directory = DefaultDirectory
		return NewConfigError("Directory", "", "Use default directory "+DefaultDirectory, ErrInvalidDirectory)
	}

	// Check if directory is accessible
	if err := ensureDirectoryExists(opts.Directory); err != nil {
		originalValue := opts.Directory
		opts.Directory = DefaultDirectory
		return NewConfigError("Directory", originalValue,
			"Can't access to directory, use default directory",
			fmt.Errorf("%w: %v", ErrPermissionDenied, err),
		)
	}

	return nil
}

// validateFilename validates and fixes the filename
func validateFilename(opts *Options) error {
	if opts.Filename != "" {
		sanitized := sanitizeFilename(opts.Filename)
		if sanitized == "" {
			originalValue := opts.Filename
			opts.Filename = DefaultFilename
			return NewConfigError("Filename", originalValue,
				"The filename contains illegal characters and is set to default filename: "+DefaultFilename,
				ErrInvalidFilename)
		}
		if sanitized != opts.Filename {
			originalValue := opts.Filename
			opts.Filename = sanitized
			return NewConfigError("Filename", originalValue,
				"The filename is sanitized: "+sanitized, ErrInvalidFilename)
		}
	}
	return nil
}

// validateTimeLayout validates and fixes the time layout
func validateTimeLayout(opts *Options) error {
	if opts.TimeLayout == "" {
		opts.TimeLayout = DefaultTimeLayout
		return NewConfigError("TimeLayout", "", "使用默认时间格式", ErrInvalidTimeLayout)
	}

	// Test the time layout by trying to format current time
	if _, err := time.Parse(opts.TimeLayout, time.Now().Format(opts.TimeLayout)); err != nil {
		originalValue := opts.TimeLayout
		opts.TimeLayout = DefaultTimeLayout
		return NewConfigError("TimeLayout", originalValue, "时间格式无效，使用默认格式",
			fmt.Errorf("%w: %v", ErrInvalidTimeLayout, err))
	}

	return nil
}

// validateMaxSize validates and fixes the maximum file size
func validateMaxSize(opts *Options) error {
	if opts.MaxSize <= 0 {
		originalValue := opts.MaxSize
		opts.MaxSize = DefaultMaxSize
		return NewConfigError(
			"MaxSize",
			originalValue,
			fmt.Sprintf("使用默认最大文件大小 %dMB", DefaultMaxSize),
			ErrInvalidMaxSize,
		)
	}
	return nil
}

// validateMaxBackups validates and fixes the maximum number of backups
func validateMaxBackups(opts *Options) error {
	if opts.MaxBackups <= 0 {
		originalValue := opts.MaxBackups
		opts.MaxBackups = DefaultMaxBackups
		return NewConfigError(
			"MaxBackups",
			originalValue,
			fmt.Sprintf("使用默认最大备份数量 %d", DefaultMaxBackups),
			ErrInvalidMaxBackups,
		)
	}
	return nil
}

// validateSampling validates and fixes the sampling configuration
func validateSampling(opts *Options) error {
	if opts.EnableSampling {
		var errs []error

		if opts.SampleInitial <= 0 {
			originalValue := opts.SampleInitial
			opts.SampleInitial = DefaultSampleInitial
			errs = append(errs, NewConfigError("SampleInitial", originalValue,
				fmt.Sprintf("使用默认初始采样数 %d", DefaultSampleInitial), ErrInvalidSampling))
		}

		if opts.SampleThereafter <= 0 {
			originalValue := opts.SampleThereafter
			opts.SampleThereafter = DefaultSampleThereafter
			errs = append(errs, NewConfigError("SampleThereafter", originalValue,
				fmt.Sprintf("使用默认后续采样数 %d", DefaultSampleThereafter), ErrInvalidSampling))
		}

		if len(errs) > 0 {
			return fmt.Errorf("采样配置错误: %v", errs)
		}
	}
	return nil
}

// ensureDirectoryExists creates the directory if it doesn't exist and checks permissions
func ensureDirectoryExists(dir string) error {
	// Clean the path
	dir = filepath.Clean(dir)

	// Check if directory already exists
	if info, err := os.Stat(dir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("路径存在但不是目录: %s", dir)
		}
		// Check if we can write to the directory
		testFile := filepath.Join(dir, ".write_test")
		if f, err := os.Create(testFile); err != nil { //nolint:gosec
			return fmt.Errorf("目录不可写: %s", dir)
		} else {
			f.Close()           //nolint:gosec
			os.Remove(testFile) //nolint:gosec
		}
		return nil
	}

	// Try to create the directory
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec
		return fmt.Errorf("无法创建目录: %s: %v", dir, err)
	}

	return nil
}

// Note: isValidLevelString function is already defined in options.go

// RecoverFromConfigError attempts to recover from configuration errors by
// providing safe fallback values and logging the issues
func RecoverFromConfigError(err error) *Options {
	if err == nil {
		return NewOptions()
	}

	// Log the configuration error to stderr
	fmt.Fprintf(os.Stderr, "配置错误，使用默认配置: %v\n", err)

	// Return safe default options
	return NewOptions()
}

// ValidateAndFixOptions is a convenience function that validates options
// and automatically recovers from errors by returning fixed options
func ValidateAndFixOptions(opts *Options) *Options {
	fixed, err := ValidateOptions(opts)
	if err != nil {
		// Log validation errors but continue with fixed configuration
		fmt.Fprintf(os.Stderr, "配置验证警告: %v\n", err)
	}
	return fixed
}

// IsConfigError checks if an error is a ConfigError
func IsConfigError(err error) bool {
	var configErr *ConfigError
	return errors.As(err, &configErr)
}

// GetConfigErrors extracts all ConfigError instances from an error chain
func GetConfigErrors(err error) []*ConfigError {
	var configErrors []*ConfigError

	// Handle wrapped errors that contain multiple ConfigErrors
	if err != nil {
		errStr := err.Error()
		// This is a simple approach - in a more sophisticated implementation,
		// you might want to use a custom error type that can hold multiple errors
		if strings.Contains(errStr, "配置验证失败") {
			// For now, just return a single ConfigError representing the overall failure
			configErrors = append(configErrors, &ConfigError{
				Field:  "multiple",
				Reason: errStr,
			})
		} else {
			// Check if it's a single ConfigError
			var configErr *ConfigError
			if errors.As(err, &configErr) {
				configErrors = append(configErrors, configErr)
			}
		}
	}

	return configErrors
}
