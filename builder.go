package log

// Builder provides a fluent interface for configuring and creating Log instances
// It wraps the existing Options struct and provides chainable methods for configuration
type Builder struct{ opts *Options }

// NewBuilder creates a new Builder instance with default options
// Returns a Builder that can be used to configure a logger through method chaining
func NewBuilder() *Builder {
	return &Builder{opts: NewOptions()} // Use existing function to get default options
}

// Level sets the log level (debug, info, warn, error, dpanic, panic, fatal)
// Returns the Builder for method chaining
func (b *Builder) Level(level string) *Builder {
	b.opts.WithLevel(level) // Use existing method
	return b
}

// Format sets the log format (console or json)
// Returns the Builder for method chaining
func (b *Builder) Format(format string) *Builder {
	b.opts.WithFormat(format) // Use existing method
	return b
}

// Directory sets the log file directory
// Returns the Builder for method chaining
func (b *Builder) Directory(dir string) *Builder {
	b.opts.WithDirectory(dir) // Use existing method
	return b
}

// Filename sets the log filename prefix
// Returns the Builder for method chaining
func (b *Builder) Filename(filename string) *Builder {
	b.opts.WithFilename(filename) // Use existing method
	return b
}

// Prefix sets the log prefix
// Returns the Builder for method chaining
func (b *Builder) Prefix(prefix string) *Builder {
	b.opts.WithPrefix(prefix) // Use existing method
	return b
}

// TimeLayout sets the time layout format
// Returns the Builder for method chaining
func (b *Builder) TimeLayout(layout string) *Builder {
	b.opts.WithTimeLayout(layout) // Use existing method
	return b
}

// DisableCaller sets whether to disable caller information
// Returns the Builder for method chaining
func (b *Builder) DisableCaller(disable bool) *Builder {
	b.opts.WithDisableCaller(disable) // Use existing method
	return b
}

// DisableStacktrace sets whether to disable stack traces
// Returns the Builder for method chaining
func (b *Builder) DisableStacktrace(disable bool) *Builder {
	b.opts.WithDisableStacktrace(disable) // Use existing method
	return b
}

// DisableSplitError sets whether to disable separate error log files
// Returns the Builder for method chaining
func (b *Builder) DisableSplitError(disable bool) *Builder {
	b.opts.WithDisableSplitError(disable) // Use existing method
	return b
}

// MaxSize sets the maximum size of log files in megabytes before rotation
// Returns the Builder for method chaining
func (b *Builder) MaxSize(size int) *Builder {
	b.opts.WithMaxSize(size) // Use existing method
	return b
}

// MaxBackups sets the maximum number of old log files to retain
// Returns the Builder for method chaining
func (b *Builder) MaxBackups(backups int) *Builder {
	b.opts.WithMaxBackups(backups) // Use existing method
	return b
}

// Compress sets whether to compress rotated log files
// Returns the Builder for method chaining
func (b *Builder) Compress(compress bool) *Builder {
	b.opts.WithCompress(compress) // Use existing method
	return b
}

// Sampling configures log sampling settings
// Returns the Builder for method chaining
func (b *Builder) Sampling(enable bool, initial, thereafter int) *Builder {
	b.opts.WithSampling(enable, initial, thereafter) // Use existing method
	return b
}

// ConsoleOutput sets whether to output logs to console
// When disabled, logs are only written to files
// Returns the Builder for method chaining
func (b *Builder) ConsoleOutput(enable bool) *Builder {
	b.opts.WithConsoleOutput(enable) // Use existing method
	return b
}

// Development applies the development preset configuration
// This configures the logger for development environment with debug level,
// console output, caller info enabled, and fast flush
// Returns the Builder for method chaining
func (b *Builder) Development() *Builder {
	DevelopmentPreset().Apply(b.opts) // Use existing preset
	return b
}

// Production applies the production preset configuration
// This configures the logger for production environment with info level,
// JSON format, optimized for performance and storage
// Returns the Builder for method chaining
func (b *Builder) Production() *Builder {
	ProductionPreset().Apply(b.opts) // Use existing preset
	return b
}

// Testing applies the testing preset configuration
// This configures the logger for testing environment with debug level,
// simplified output, and minimal file operations
// Returns the Builder for method chaining
func (b *Builder) Testing() *Builder {
	TestingPreset().Apply(b.opts) // Use existing preset
	return b
}

// Build creates and returns a new Log instance with the configured options
// This method calls the existing NewLog() function with the built options
func (b *Builder) Build() *Log {
	return NewLog(b.opts) // Call existing function
}
