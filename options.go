package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap/zapcore"

	"github.com/kydenul/log/internal"
)

// DefaultDirectory returns the default log directory, which is typically the user's home directory joined with "logs".
var DefaultDirectory = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get user home directory: %v", err))
	}
	return filepath.Join(home, "logs")
}()

const (
	DefaultPrefix     = "ZIWI_"
	DefaultLevel      = zapcore.InfoLevel
	DefaultTimeLayout = "2006-01-02 15:04:05.000"
	DefaultFormat     = "console" // console style
	DefaultFilename   = ""        // Default filename prefix

	DefaultDisableCaller     = false
	DefaultDisableStacktrace = false
	DefaultDisableSplitError = true

	DefaultMaxSize    = 100   // 100MB
	DefaultMaxBackups = 3     // Keep 3 old log files
	DefaultCompress   = false // Not compress rotated log files

	// Defaults for sampling functionality
	DefaultEnableSampling   = false // Sampling disabled by default
	DefaultSampleInitial    = 100   // Initial sample count
	DefaultSampleThereafter = 100   // Subsequent sample count

	FormatConsole = "console"
	FormatJSON    = "json"

	LevelDebug = "debug"
	LevelInfo  = "info"
)

// Options for logger
type Options struct {
	Prefix     string `json:"prefix,omitempty"      yaml:"prefix,omitempty"`      // Log Prefix
	Directory  string `json:"directory,omitempty"   yaml:"directory,omitempty"`   // Log File Directory
	Filename   string `json:"filename,omitempty"    yaml:"filename,omitempty"`    // Log filename prefix
	Level      string `json:"level,omitempty"       yaml:"level,omitempty"`       // Log Level
	TimeLayout string `json:"time_layout,omitempty" yaml:"time_layout,omitempty"` // Time Layout
	Format     string `json:"format,omitempty"      yaml:"format,omitempty"`      // Log Format

	DisableCaller     bool `json:"disable_caller,omitempty"      yaml:"disable_caller,omitempty"`
	DisableStacktrace bool `json:"disable_stacktrace,omitempty"  yaml:"disable_stacktrace,omitempty"`
	DisableSplitError bool `json:"disable_split_error,omitempty" yaml:"disable_split_error,omitempty"`

	// -----------------
	// Log rotation settings
	// -----------------

	MaxSize    int  `json:"max_size,omitempty"    yaml:"max_size,omitempty"`    // Maximum size of log files in megabytes
	MaxBackups int  `json:"max_backups,omitempty" yaml:"max_backups,omitempty"` // Maximum number of old log files
	Compress   bool `json:"compress,omitempty"    yaml:"compress,omitempty"`    // Whether to compress rotated log files

	// -----------------
	// Sampling settings
	// -----------------

	EnableSampling   bool `json:"enable_sampling,omitempty"   yaml:"enable_sampling,omitempty"`
	SampleInitial    int  `json:"sample_initial,omitempty"    yaml:"sample_initial,omitempty"`
	SampleThereafter int  `json:"sample_thereafter,omitempty" yaml:"sample_thereafter,omitempty"`
}

// NewOptions return the default Options.
//
// Default:
//
//	Prefix:    "ZIWI_",
//	Directory: "$HOME/logs",
//
//	Level:      "info",
//	TimeLayout: "2006-01-02 15:04:05.000",
//	Format:     "console",
//
//	DisableCaller:     false,
//	DisableStacktrace: false,
//	DisableSplitError: false,
//
//	// Default log rotation settings
//	MaxSize:    100, // 100MB
//	MaxBackups: 3,   // Keep 3 old log files
//	Compress:   false,
//
//	// Sampling settings
//	EnableSampling:   false, // Sampling disabled by default
//	SampleInitial:    100,   // Initial sample count
//	SampleThereafter: 100,   // Subsequent sample count
func NewOptions() *Options {
	opt := &Options{
		Prefix:    DefaultPrefix,
		Directory: DefaultDirectory,
		Filename:  DefaultFilename,

		Level:      DefaultLevel.String(),
		TimeLayout: DefaultTimeLayout,
		Format:     DefaultFormat,

		DisableCaller:     DefaultDisableCaller,
		DisableStacktrace: DefaultDisableStacktrace,
		DisableSplitError: DefaultDisableSplitError,

		// Default log rotation settings
		MaxSize:    DefaultMaxSize,
		MaxBackups: DefaultMaxBackups,
		Compress:   DefaultCompress,

		// Sampling settings
		EnableSampling:   DefaultEnableSampling,
		SampleInitial:    DefaultSampleInitial,
		SampleThereafter: DefaultSampleThereafter,
	}

	if err := opt.Validate(); err != nil {
		fmt.Printf("invalid options: %s", err)
		return nil
	}

	return opt
}

func (opt *Options) WithPrefix(prefix string) *Options {
	opt.Prefix = prefix
	return opt
}

func (opt *Options) WithDirectory(dir string) *Options {
	if dir == "" {
		opt.Directory = DefaultDirectory
	} else {
		opt.Directory = dir
	}
	return opt
}

func (opt *Options) WithFilename(filename string) *Options {
	opt.Filename = filename
	return opt
}

func (opt *Options) WithLevel(level string) *Options {
	if level == "" || !isValidLevelString(level) {
		opt.Level = DefaultLevel.String()
	} else {
		opt.Level = level
	}
	return opt
}

func (opt *Options) WithTimeLayout(timeLayout string) *Options {
	if timeLayout == "" {
		opt.TimeLayout = DefaultTimeLayout
	} else {
		opt.TimeLayout = timeLayout
	}
	return opt
}

func (opt *Options) WithFormat(format string) *Options {
	if format == "" || (format != DefaultFormat && format != "json") {
		opt.Format = DefaultFormat
	} else {
		opt.Format = format
	}
	return opt
}

func (opt *Options) WithDisableCaller(disableCaller bool) *Options {
	opt.DisableCaller = disableCaller
	return opt
}

func (opt *Options) WithDisableStacktrace(disableStacktrace bool) *Options {
	opt.DisableStacktrace = disableStacktrace
	return opt
}

func (opt *Options) WithDisableSplitError(disableSplitError bool) *Options {
	opt.DisableSplitError = disableSplitError
	return opt
}

func (opt *Options) WithMaxSize(maxSize int) *Options {
	if maxSize <= 0 {
		opt.MaxSize = DefaultMaxSize
	} else {
		opt.MaxSize = maxSize
	}
	return opt
}

func (opt *Options) WithMaxBackups(maxBackups int) *Options {
	if maxBackups <= 0 {
		opt.MaxBackups = DefaultMaxBackups
	} else {
		opt.MaxBackups = maxBackups
	}
	return opt
}

func (opt *Options) WithCompress(compress bool) *Options {
	opt.Compress = compress
	return opt
}

func (opt *Options) WithSampling(enable bool, initial, thereafter int) *Options {
	opt.EnableSampling = enable
	if initial > 0 {
		opt.SampleInitial = initial
	} else {
		opt.SampleInitial = DefaultSampleInitial
	}
	if thereafter > 0 {
		opt.SampleThereafter = thereafter
	} else {
		opt.SampleThereafter = DefaultSampleThereafter
	}
	return opt
}

// isValidLevelString checks if the provided level string is valid
func isValidLevelString(level string) bool {
	return level == zapcore.DebugLevel.String() ||
		level == zapcore.InfoLevel.String() ||
		level == zapcore.WarnLevel.String() ||
		level == zapcore.ErrorLevel.String() ||
		level == zapcore.DPanicLevel.String() ||
		level == zapcore.PanicLevel.String() ||
		level == zapcore.FatalLevel.String()
}

func (opt *Options) Validate() error {
	if opt.Directory == "" {
		return fmt.Errorf("invalid directory: %s, expected: not empty", opt.Directory)
	}

	// Validate filename if provided
	if opt.Filename != "" {
		sanitized := sanitizeFilename(opt.Filename)
		if sanitized == "" {
			return fmt.Errorf("invalid filename: %s, results in empty name after sanitization", opt.Filename)
		}
	}

	if !isValidLevelString(opt.Level) {
		return fmt.Errorf("invalid level: %s, expected: debug, info, warn, error, dpanic, panic or fatal", opt.Level)
	}

	if err := internal.ValidateTimeLayout(opt.TimeLayout); err != nil {
		return fmt.Errorf("invalid time layout: %s, expected: valid time layout", opt.TimeLayout)
	}

	if opt.Format != DefaultFormat && opt.Format != "json" {
		return fmt.Errorf("invalid format: %s, expected: console or json", opt.Format)
	}

	if opt.MaxSize <= 0 {
		return fmt.Errorf("invalid max size: %d, expected: > 0", opt.MaxSize)
	}

	if opt.MaxBackups <= 0 {
		return fmt.Errorf("invalid max backups: %d, expected: > 0", opt.MaxBackups)
	}

	// Validate sampling settings
	if opt.EnableSampling {
		if opt.SampleInitial <= 0 {
			return fmt.Errorf("invalid sample initial: %d, expected: > 0", opt.SampleInitial)
		}
		if opt.SampleThereafter <= 0 {
			return fmt.Errorf("invalid sample thereafter: %d, expected: > 0", opt.SampleThereafter)
		}
	}

	return nil
}

// sanitizeFilename cleans and validates a filename by removing unsafe characters,
// limiting length, and ensuring the filename is valid for filesystem use.
// It returns the sanitized filename or an empty string if the input results in an invalid filename.
func sanitizeFilename(filename string) string {
	if filename == "" {
		return ""
	}

	// Remove unsafe characters that are not allowed in filenames
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	cleaned := filename
	for _, char := range unsafe {
		cleaned = strings.ReplaceAll(cleaned, char, "_")
	}

	// Additional validation: ensure no control characters
	for _, r := range cleaned {
		if r < 32 || r == 127 { // ASCII control characters
			cleaned = strings.ReplaceAll(cleaned, string(r), "_")
		}
	}

	// Clean whitespace characters
	cleaned = strings.TrimSpace(cleaned)

	// Check if filename is only dots and/or underscores after cleaning
	if strings.Trim(cleaned, "._") == "" {
		return ""
	}

	// Ensure filename doesn't start with a dot (hidden files)
	if strings.HasPrefix(cleaned, ".") {
		cleaned = "_" + cleaned[1:]
	}

	// Limit length to prevent filesystem issues (max 100 characters)
	if len(cleaned) > 100 {
		cleaned = cleaned[:100]
	}

	// Ensure filename is not empty after cleaning
	if cleaned == "" {
		return ""
	}

	return cleaned
}
