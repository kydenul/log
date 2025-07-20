package log

import (
	"fmt"
	"os"
	"path/filepath"

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

	DefaultDisableCaller     = false
	DefaultDisableStacktrace = false
	DefaultDisableSplitError = true

	DefaultMaxSize    = 100   // 100MB
	DefaultMaxBackups = 3     // Keep 3 old log files
	DefaultCompress   = false // Not compress rotated log files
)

// Options for logger
type Options struct {
	Prefix    string // Log Prefix, e.g ZIWI
	Directory string // Log File Directory, e.g logs

	Level      string // Log Level, "debug", "info", "warn", "error"
	TimeLayout string // Time Layout, e.g "2006-01-02 15:04:05.000"
	Format     string // Log Format, "console", "json"

	DisableCaller     bool // Disable caller information
	DisableStacktrace bool // Disable stack traces
	DisableSplitError bool // Disable separate error log files

	// Log rotation settings
	MaxSize    int  // Maximum size of log files in megabytes before rotation
	MaxBackups int  // Maximum number of old log files to retain
	Compress   bool // Whether to compress rotated log files
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
func NewOptions() *Options {
	opt := &Options{
		Prefix:    DefaultPrefix,
		Directory: DefaultDirectory,

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
	if format == "" || (format != "console" && format != "json") {
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

	if !isValidLevelString(opt.Level) {
		return fmt.Errorf("invalid level: %s, expected: debug, info, warn, error, dpanic, panic or fatal", opt.Level)
	}

	if err := internal.ValidateTimeLayout(opt.TimeLayout); err != nil {
		return fmt.Errorf("invalid time layout: %s, expected: valid time layout", opt.TimeLayout)
	}

	if opt.Format != "console" && opt.Format != "json" {
		return fmt.Errorf("invalid format: %s, expected: console or json", opt.Format)
	}

	if opt.MaxSize <= 0 {
		return fmt.Errorf("invalid max size: %d, expected: > 0", opt.MaxSize)
	}

	if opt.MaxBackups <= 0 {
		return fmt.Errorf("invalid max backups: %d, expected: > 0", opt.MaxBackups)
	}

	return nil
}
