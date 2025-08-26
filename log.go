package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/kydenul/log/internal"
)

const (
	MaxRetries = 3                     // Maximum retries for file write operations
	BriefDelay = time.Millisecond * 10 // Brief delay before retry
)

var (
	// Global logger instance using atomic.Value for lock-free access
	defaultLogger atomic.Value // *ZiwiLog
	// Log prefix
	logPrefix string

	// Buffer pool to reduce memory allocations
	bufferPool = sync.Pool{
		New: func() any {
			return &buffer.Buffer{}
		},
	}
)

// DefaultLogger returns the default global logger instance
func DefaultLogger() *Log {
	logger, ok := defaultLogger.Load().(*Log)
	if ok {
		return logger
	}

	return NewLog(nil)
}

// ReplaceLogger replaces the default logger with a new instance
func ReplaceLogger(l *Log) {
	if l != nil {
		defaultLogger.Store(l)
	}
}

// Initialize global logger instance
func init() {
	logger := NewLog(NewOptions())
	defaultLogger.Store(logger)

	internal.SetupAutoSync(logger.Sync)
}

// Ensure ZiwiLog implements Logger interface
var _ Logger = &Log{}

// Log is the implement of Logger interface.
// It wraps zap.Logger.
type Log struct {
	zapcore.Encoder

	log       *zap.Logger
	logDir    string // log file directory
	file      *lumberjack.Logger
	errFile   *lumberjack.Logger
	currDate  string // current date
	dateCheck int64  // atomic timestamp for date checking optimization
	opts      *Options
	mu        sync.RWMutex // protects file operations
}

// NewLog creates a new logger instance. It will initialize the global logger instance with the specified options.
//
// Returns:
//
//   - *Log: The new logger instance.
func NewLog(opts *Options) *Log {
	// 1. If opts is nil, use default options
	if opts == nil {
		opts = NewOptions()
	}

	// 2. Validate options once and fix invalid values
	if err := opts.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid logger options: %v. Using fallback values.\n", err)

		// Fix invalid options with defaults
		if opts.Directory == "" {
			opts.Directory = DefaultDirectory
		}
		if opts.Level == "" || !isValidLevel(opts.Level) {
			opts.Level = DefaultLevel.String()
		}
		if opts.Format != DefaultFormat && opts.Format != "json" {
			opts.Format = DefaultFormat
		}
		if opts.MaxSize <= 0 {
			opts.MaxSize = DefaultMaxSize
		}
		if opts.MaxBackups <= 0 {
			opts.MaxBackups = DefaultMaxBackups
		}
	}

	// 3. Set log prefix
	logPrefix = opts.Prefix

	// 4. Set time layout, Default time layout
	timeLayout := DefaultTimeLayout
	if err := internal.ValidateTimeLayout(opts.TimeLayout); err == nil {
		timeLayout = opts.TimeLayout
	} else {
		fmt.Fprintf(os.Stderr,
			"Invalid time layout '%s', using default: %s\n", opts.TimeLayout, DefaultTimeLayout)
	}

	// 5. Create our custom ZiwiLog with the base encoder
	logger := &Log{
		Encoder:   internal.NewBaseEncoder(opts.Format, timeLayout),
		opts:      opts,
		logDir:    opts.Directory,
		dateCheck: time.Now().Unix(),
	}

	// 6. Create the zap logger with our custom core, ZiwiLog encoder
	zapLevel := DefaultLevel
	_ = zapLevel.UnmarshalText([]byte(opts.Level))
	log := zap.New(
		zapcore.NewCore(
			logger,                     // Our custom encoder
			zapcore.AddSync(os.Stdout), // Output to stdout
			zap.NewAtomicLevelAt(zapLevel),
		),
		zap.AddStacktrace(zapcore.PanicLevel),
		zap.AddCallerSkip(1),
		zap.WithCaller(!opts.DisableCaller),
	)

	// 7. Assign the zap logger to our ZiwiLog
	logger.log = log
	zap.RedirectStdLog(logger.log)

	return logger
}

// isValidLevel checks if the provided level is valid
func isValidLevel(level string) bool {
	return slices.Contains(
		[]string{
			zapcore.DebugLevel.String(),
			zapcore.InfoLevel.String(),
			zapcore.WarnLevel.String(),
			zapcore.ErrorLevel.String(),
			zapcore.DPanicLevel.String(),
			zapcore.PanicLevel.String(),
			zapcore.FatalLevel.String(),
		}, level)
}

// EncodeEntry encodes the entry and fields into a buffer.
func (l *Log) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// Get buffer from base encoder
	buf, err := l.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, fmt.Errorf("EncodeEntry error: %w", err)
	}

	// Optimize prefix addition using buffer operations instead of string concatenation
	if logPrefix != "" {
		// Get a temporary buffer from pool for prefix operation
		tempBuf, _ := bufferPool.Get().(*buffer.Buffer)
		tempBuf.Reset()
		defer bufferPool.Put(tempBuf)

		// Write prefix + original content efficiently
		tempBuf.AppendString(logPrefix)
		_, _ = tempBuf.Write(buf.Bytes())

		// Replace original buffer content
		buf.Reset()
		_, _ = buf.Write(tempBuf.Bytes())
	}

	// Optimized date checking - only check every few seconds
	now := time.Now()
	currentTimestamp := now.Unix()
	if currentTimestamp-atomic.LoadInt64(&l.dateCheck) >= 3600 { // Check every hour
		if err := l.setupLogFiles(now.Format(time.DateOnly)); err != nil {
			return nil, err
		}
		atomic.StoreInt64(&l.dateCheck, currentTimestamp)
	} else {
		// Quick check if files exist, setup if needed
		l.mu.RLock()
		fileExists := l.file != nil
		l.mu.RUnlock()
		if !fileExists {
			if err := l.setupLogFiles(now.Format(time.DateOnly)); err != nil {
				return nil, err
			}
		}
	}

	// Write to main log file with error handling
	data := buf.Bytes()
	if err := l.writeToFile(l.file, data); err != nil {
		// Log write errors to stderr as fallback
		fmt.Fprintf(os.Stderr, "Failed to write to log file: %v\n", err)
	}

	// For error level logs, also write to error log file
	if entry.Level == zapcore.ErrorLevel && !l.opts.DisableSplitError {
		l.mu.RLock()
		errFile := l.errFile
		l.mu.RUnlock()
		if errFile != nil {
			if err := l.writeToFile(errFile, data); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write to error log file: %v\n", err)
			}
		}
	}

	return buf, nil
}

// writeToFile writes data to the specified file with retry logic
func (l *Log) writeToFile(file *lumberjack.Logger, data []byte) error {
	if file == nil {
		return errors.New("file is nil")
	}

	// Simple retry logic for file write
	for retries := range MaxRetries {
		if _, err := file.Write(data); err != nil {
			if retries == MaxRetries-1 {
				return fmt.Errorf("failed to write after retries: %w", err)
			}

			time.Sleep(BriefDelay) // Brief delay before retry
			continue
		}
		return nil
	}
	return nil
}

// testFileCreation tests if a lumberjack logger can successfully create and write to its file.
// This is used to validate that the filename and path are valid before committing to use them.
func (l *Log) testFileCreation(logger *lumberjack.Logger) error {
	if logger == nil {
		return errors.New("logger is nil")
	}

	// Test by writing a small test message
	testData := []byte("# Log file test\n")
	if _, err := logger.Write(testData); err != nil {
		return fmt.Errorf("failed to write test data to log file '%s': %w", logger.Filename, err)
	}

	return nil
}

// generateFileName generates the appropriate filename based on the Filename field and log type.
// It implements different naming rules for main log files and error log files while ensuring
// backward compatibility when Filename is empty.
//
// Parameters:
//   - date: The date string to use in the filename (e.g., "2025-07-20")
//   - isErrorLog: Whether this is for an error log file
//
// Returns:
//   - string: The generated filename (without directory path)
//
// File naming rules:
//   - Main log with Filename: "{filename}-{date}.log"
//   - Main log without Filename: "{date}.log" (backward compatible)
//   - Error log with Filename: "{filename}-{date}_error.log"
//   - Error log without Filename: "{date}_error.log" (backward compatible)
func (l *Log) generateFileName(date string, isErrorLog bool) string {
	var baseName string

	if l.opts.Filename != "" {
		// Use custom prefix - sanitize it first to ensure it's safe
		sanitized := sanitizeFilename(l.opts.Filename)
		if sanitized != "" {
			baseName = sanitized + "-" + date
		} else {
			// Fallback to default format if sanitization results in empty string
			baseName = date
		}
	} else {
		// Default format (backward compatible)
		baseName = date
	}

	if isErrorLog {
		return baseName + "_error.log"
	}
	return baseName + ".log"
}

// setupLogFiles ensures log files are properly configured with thread safety.
// It handles file creation errors and implements fallback mechanisms to ensure
// logging continues even when custom filenames cause issues.
func (l *Log) setupLogFiles(date string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// If the date hasn't changed and the file exists, no need to reconfigure
	if l.currDate == date &&
		l.file != nil &&
		(l.errFile != nil || l.opts.DisableSplitError) {
		return nil
	}

	// Ensure log directory exists
	if err := os.MkdirAll(l.logDir, 0o755); err != nil { //nolint:gosec
		return fmt.Errorf("create log dir error: %w", err)
	}

	// Set main log file using the new filename generation logic with error handling
	if l.currDate != date || l.file == nil {
		fileName := l.generateFileName(date, false)
		fullPath := filepath.Join(l.logDir, fileName)

		// Create lumberjack logger with error handling
		mainLogger := &lumberjack.Logger{
			Filename:   fullPath,
			MaxSize:    l.opts.MaxSize,    // megabytes
			MaxBackups: l.opts.MaxBackups, // number of backups
			Compress:   l.opts.Compress,   // compress rotated files
		}

		// Test file creation by attempting to write to it
		if err := l.testFileCreation(mainLogger); err != nil {
			// Fallback to default filename format if custom filename fails
			fmt.Fprintf(os.Stderr, "Failed to create log file with custom filename '%s': %v. Falling back to default format.\n", fileName, err)

			// Generate fallback filename (without custom prefix)
			fallbackFileName := date + ".log"
			fallbackPath := filepath.Join(l.logDir, fallbackFileName)
			mainLogger = &lumberjack.Logger{
				Filename:   fallbackPath,
				MaxSize:    l.opts.MaxSize,
				MaxBackups: l.opts.MaxBackups,
				Compress:   l.opts.Compress,
			}

			// Test fallback file creation
			if err := l.testFileCreation(mainLogger); err != nil {
				return fmt.Errorf("failed to create fallback log file: %w", err)
			}
		}

		l.file = mainLogger
	}

	// Set error log file (if needed) using the new filename generation logic with error handling
	if !l.opts.DisableSplitError && (l.currDate != date || l.errFile == nil) {
		errFileName := l.generateFileName(date, true)
		errFullPath := filepath.Join(l.logDir, errFileName)

		// Create error log lumberjack logger with error handling
		errLogger := &lumberjack.Logger{
			Filename:   errFullPath,
			MaxSize:    l.opts.MaxSize,    // megabytes
			MaxBackups: l.opts.MaxBackups, // number of backups
			Compress:   l.opts.Compress,   // compress rotated files
		}

		// Test error file creation
		if err := l.testFileCreation(errLogger); err != nil {
			// Fallback to default error filename format if custom filename fails
			fmt.Fprintf(os.Stderr, "Failed to create error log file with custom filename '%s': %v. Falling back to default format.\n", errFileName, err)

			// Generate fallback error filename (without custom prefix)
			fallbackErrFileName := date + "_error.log"
			fallbackErrPath := filepath.Join(l.logDir, fallbackErrFileName)
			errLogger = &lumberjack.Logger{
				Filename:   fallbackErrPath,
				MaxSize:    l.opts.MaxSize,
				MaxBackups: l.opts.MaxBackups,
				Compress:   l.opts.Compress,
			}

			// Test fallback error file creation
			if err := l.testFileCreation(errLogger); err != nil {
				return fmt.Errorf("failed to create fallback error log file: %w", err)
			}
		}

		l.errFile = errLogger
	}

	// Update current date only after successful file setup
	l.currDate = date
	return nil
}

// Sync flushs any buffered log entries. Applications should take care to call Sync before exiting.
func Sync() {
	DefaultLogger().Sync()
}

// Sync flushs any buffered log entries. Applications should take care to call Sync before exiting.
func (l *Log) Sync() {
	_ = l.log.Sync()

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		_ = l.file.Close()
	}
	if l.errFile != nil {
		_ = l.errFile.Close()
	}
}

func Debug(args ...any) {
	DefaultLogger().log.Sugar().Debug(args...)
}

func (l *Log) Debug(args ...any) {
	l.log.Sugar().Debug(args...)
}

func Info(args ...any) {
	DefaultLogger().log.Sugar().Info(args...)
}

func (l *Log) Info(args ...any) {
	l.log.Sugar().Info(args...)
}

func Warn(args ...any) {
	DefaultLogger().log.Sugar().Warn(args...)
}

func (l *Log) Warn(args ...any) {
	l.log.Sugar().Warn(args...)
}

func Error(args ...any) {
	DefaultLogger().log.Sugar().Error(args...)
}

func (l *Log) Error(args ...any) {
	l.log.Sugar().Error(args...)
}

func Panic(args ...any) {
	DefaultLogger().log.Sugar().Panic(args...)
}

func (l *Log) Panic(args ...any) {
	l.log.Sugar().Panic(args...)
}

func Fatal(args ...any) {
	DefaultLogger().log.Sugar().Fatal(args...)
}

func (l *Log) Fatal(args ...any) {
	l.log.Sugar().Fatal(args...)
}

func Debugln(args ...any) {
	DefaultLogger().log.Sugar().Debugln(args...)
}

func (l *Log) Debugln(args ...any) {
	l.log.Sugar().Debugln(args...)
}

func Infoln(args ...any) {
	DefaultLogger().log.Sugar().Infoln(args...)
}

func (l *Log) Infoln(args ...any) {
	l.log.Sugar().Infoln(args...)
}

func Warnln(args ...any) {
	DefaultLogger().log.Sugar().Warnln(args...)
}

func (l *Log) Warnln(args ...any) {
	l.log.Sugar().Warnln(args...)
}

func Errorln(args ...any) { DefaultLogger().log.Sugar().Errorln(args...) }

func (l *Log) Errorln(args ...any) {
	l.log.Sugar().Errorln(args...)
}

func Panicln(args ...any) {
	DefaultLogger().log.Sugar().Panicln(args...)
}

func (l *Log) Panicln(args ...any) {
	l.log.Sugar().Panicln(args...)
}

func Fatalln(args ...any) {
	DefaultLogger().log.Sugar().Fatalln(args...)
}

func (l *Log) Fatalln(args ...any) {
	l.log.Sugar().Fatalln(args...)
}

// Debugw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Debugw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Debugw(msg, keysAndValues...)
}

// Debugw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *Log) Debugw(msg string, keysAndValues ...any) {
	l.log.Sugar().Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Infow(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Infow(msg, keysAndValues...)
}

// Infow logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *Log) Infow(msg string, keysAndValues ...any) {
	l.log.Sugar().Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Warnw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Warnw(msg, keysAndValues...)
}

// Warnw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *Log) Warnw(msg string, keysAndValues ...any) {
	l.log.Sugar().Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Errorw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Errorw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *Log) Errorw(msg string, keysAndValues ...any) {
	l.log.Sugar().Errorw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics.
// The variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Panicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics.
// The variadic key-value pairs are treated as they are in With.
func (l *Log) Panicw(msg string, keysAndValues ...any) {
	l.log.Sugar().Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit.
// The variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Fatalw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit.
// The variadic key-value pairs are treated as they are in With.
func (l *Log) Fatalw(msg string, keysAndValues ...any) {
	l.log.Sugar().Fatalw(msg, keysAndValues...)
}

// Debugf formats the message according to the format specifier and logs it.
func Debugf(template string, args ...any) {
	DefaultLogger().log.Sugar().Debugf(template, args...)
}

// Debugf formats the message according to the format specifier and logs it.
func (l *Log) Debugf(template string, args ...any) {
	l.log.Sugar().Debugf(template, args...)
}

// Infof formats the message according to the format specifier and logs it.
func Infof(template string, args ...any) {
	DefaultLogger().log.Sugar().Infof(template, args...)
}

// Infof formats the message according to the format specifier and logs it.
func (l *Log) Infof(template string, args ...any) {
	l.log.Sugar().Infof(template, args...)
}

// Warnf formats the message according to the format specifier and logs it.
func Warnf(template string, args ...any) {
	DefaultLogger().log.Sugar().Warnf(template, args...)
}

// Warnf formats the message according to the format specifier and logs it.
func (l *Log) Warnf(template string, args ...any) {
	l.log.Sugar().Warnf(template, args...)
}

// Errorf formats the message according to the format specifier and logs it.
func Errorf(template string, args ...any) {
	DefaultLogger().log.Sugar().Errorf(template, args...)
}

// Errorf formats the message according to the format specifier and logs it.
func (l *Log) Errorf(template string, args ...any) {
	l.log.Sugar().Errorf(template, args...)
}

// Panicf formats the message according to the format specifier and panics.
func Panicf(template string, args ...any) {
	DefaultLogger().log.Sugar().Panicf(template, args...)
}

// Panicf formats the message according to the format specifier and panics.
func (l *Log) Panicf(template string, args ...any) {
	l.log.Sugar().Panicf(template, args...)
}

// Fatalf formats the message according to the format specifier and calls os.Exit.
func Fatalf(template string, args ...any) {
	DefaultLogger().log.Sugar().Fatalf(template, args...)
}

// Fatalf formats the message according to the format specifier and calls os.Exit.
func (l *Log) Fatalf(template string, args ...any) {
	l.log.Sugar().Fatalf(template, args...)
}
