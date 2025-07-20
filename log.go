package log

import (
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
func DefaultLogger() *ZiwiLog {
	return defaultLogger.Load().(*ZiwiLog)
}

// ReplaceLogger replaces the default logger with a new instance
func ReplaceLogger(l *ZiwiLog) {
	if l != nil {
		defaultLogger.Store(l)
	}
}

// Initialize global logger instance
func init() {
	logger := NewLogger(NewOptions())
	defaultLogger.Store(logger)

	internal.SetupAutoSync(logger.Sync)
}

// Ensure ZiwiLog implements Logger interface
var _ Logger = &ZiwiLog{}

// ZiwiLog is the implement of Logger interface.
// It wraps zap.Logger.
type ZiwiLog struct {
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

// NewLogger creates a new logger instance. It will initialize the global logger instance with the specified options.
//
// Returns:
//
//   - *ZiwiLog: The new logger instance.
func NewLogger(opts *Options) *ZiwiLog {
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
		if opts.Format != "console" && opts.Format != "json" {
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
	logger := &ZiwiLog{
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
func (l *ZiwiLog) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// Get buffer from base encoder
	buf, err := l.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, fmt.Errorf("EncodeEntry error: %w", err)
	}

	// Optimize prefix addition using buffer operations instead of string concatenation
	if logPrefix != "" {
		// Get a temporary buffer from pool for prefix operation
		tempBuf := bufferPool.Get().(*buffer.Buffer)
		tempBuf.Reset()
		defer bufferPool.Put(tempBuf)

		// Write prefix + original content efficiently
		tempBuf.AppendString(logPrefix)
		tempBuf.Write(buf.Bytes())

		// Replace original buffer content
		buf.Reset()
		buf.Write(tempBuf.Bytes())
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
func (l *ZiwiLog) writeToFile(file *lumberjack.Logger, data []byte) error {
	if file == nil {
		return fmt.Errorf("file is nil")
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

// setupLogFiles ensures log files are properly configured with thread safety
func (l *ZiwiLog) setupLogFiles(date string) error {
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

	// Set main log file
	if l.currDate != date || l.file == nil {
		fileName := filepath.Join(l.logDir, date+".log")
		l.file = &lumberjack.Logger{
			Filename:   fileName,
			MaxSize:    l.opts.MaxSize,    // megabytes
			MaxBackups: l.opts.MaxBackups, // number of backups
			Compress:   l.opts.Compress,   // compress rotated files
		}
	}

	// Set error log file (if needed)
	if !l.opts.DisableSplitError && (l.currDate != date || l.errFile == nil) {
		errFileName := filepath.Join(l.logDir, date+"_error.log")
		l.errFile = &lumberjack.Logger{
			Filename:   errFileName,
			MaxSize:    l.opts.MaxSize,    // megabytes
			MaxBackups: l.opts.MaxBackups, // number of backups
			Compress:   l.opts.Compress,   // compress rotated files
		}
	}

	// Update current date
	l.currDate = date
	return nil
}

// Sync flushs any buffered log entries. Applications should take care to call Sync before exiting.
func Sync() {
	DefaultLogger().Sync()
}

// Sync flushs any buffered log entries. Applications should take care to call Sync before exiting.
func (l *ZiwiLog) Sync() {
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

func (l *ZiwiLog) Debug(args ...any) {
	l.log.Sugar().Debug(args...)
}

func Info(args ...any) {
	DefaultLogger().log.Sugar().Info(args...)
}

func (l *ZiwiLog) Info(args ...any) {
	l.log.Sugar().Info(args...)
}

func Warn(args ...any) {
	DefaultLogger().log.Sugar().Warn(args...)
}

func (l *ZiwiLog) Warn(args ...any) {
	l.log.Sugar().Warn(args...)
}

func Error(args ...any) {
	DefaultLogger().log.Sugar().Error(args...)
}

func (l *ZiwiLog) Error(args ...any) {
	l.log.Sugar().Error(args...)
}

func Panic(args ...any) {
	DefaultLogger().log.Sugar().Panic(args...)
}

func (l *ZiwiLog) Panic(args ...any) {
	l.log.Sugar().Panic(args...)
}

func Fatal(args ...any) {
	DefaultLogger().log.Sugar().Fatal(args...)
}

func (l *ZiwiLog) Fatal(args ...any) {
	l.log.Sugar().Fatal(args...)
}

func Debugln(args ...any) {
	DefaultLogger().log.Sugar().Debugln(args...)
}

func (l *ZiwiLog) Debugln(args ...any) {
	l.log.Sugar().Debugln(args...)
}

func Infoln(args ...any) {
	DefaultLogger().log.Sugar().Infoln(args...)
}

func (l *ZiwiLog) Infoln(args ...any) {
	l.log.Sugar().Infoln(args...)
}

func Warnln(args ...any) {
	DefaultLogger().log.Sugar().Warnln(args...)
}

func (l *ZiwiLog) Warnln(args ...any) {
	l.log.Sugar().Warnln(args...)
}

func Errorln(args ...any) { DefaultLogger().log.Sugar().Errorln(args...) }

func (l *ZiwiLog) Errorln(args ...any) {
	l.log.Sugar().Errorln(args...)
}

func Panicln(args ...any) {
	DefaultLogger().log.Sugar().Panicln(args...)
}

func (l *ZiwiLog) Panicln(args ...any) {
	l.log.Sugar().Panicln(args...)
}

func Fatalln(args ...any) {
	DefaultLogger().log.Sugar().Fatalln(args...)
}

func (l *ZiwiLog) Fatalln(args ...any) {
	l.log.Sugar().Fatalln(args...)
}

// Debugw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Debugw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Debugw(msg, keysAndValues...)
}

// Debugw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Debugw(msg string, keysAndValues ...any) {
	l.log.Sugar().Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Infow(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Infow(msg, keysAndValues...)
}

// Infow logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Infow(msg string, keysAndValues ...any) {
	l.log.Sugar().Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Warnw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Warnw(msg, keysAndValues...)
}

// Warnw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Warnw(msg string, keysAndValues ...any) {
	l.log.Sugar().Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Errorw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Errorw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Errorw(msg string, keysAndValues ...any) {
	l.log.Sugar().Errorw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics.
// The variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Panicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Panicw(msg string, keysAndValues ...any) {
	l.log.Sugar().Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit.
// The variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, keysAndValues ...any) {
	DefaultLogger().log.Sugar().Fatalw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Fatalw(msg string, keysAndValues ...any) {
	l.log.Sugar().Fatalw(msg, keysAndValues...)
}

// Debugf formats the message according to the format specifier and logs it.
func Debugf(template string, args ...any) {
	DefaultLogger().log.Sugar().Debugf(template, args...)
}

// Debugf formats the message according to the format specifier and logs it.
func (l *ZiwiLog) Debugf(template string, args ...any) {
	l.log.Sugar().Debugf(template, args...)
}

// Infof formats the message according to the format specifier and logs it.
func Infof(template string, args ...any) {
	DefaultLogger().log.Sugar().Infof(template, args...)
}

// Infof formats the message according to the format specifier and logs it.
func (l *ZiwiLog) Infof(template string, args ...any) {
	l.log.Sugar().Infof(template, args...)
}

// Warnf formats the message according to the format specifier and logs it.
func Warnf(template string, args ...any) {
	DefaultLogger().log.Sugar().Warnf(template, args...)
}

// Warnf formats the message according to the format specifier and logs it.
func (l *ZiwiLog) Warnf(template string, args ...any) {
	l.log.Sugar().Warnf(template, args...)
}

// Errorf formats the message according to the format specifier and logs it.
func Errorf(template string, args ...any) {
	DefaultLogger().log.Sugar().Errorf(template, args...)
}

// Errorf formats the message according to the format specifier and logs it.
func (l *ZiwiLog) Errorf(template string, args ...any) {
	l.log.Sugar().Errorf(template, args...)
}

// Panicf formats the message according to the format specifier and panics.
func Panicf(template string, args ...any) {
	DefaultLogger().log.Sugar().Panicf(template, args...)
}

// Panicf formats the message according to the format specifier and panics.
func (l *ZiwiLog) Panicf(template string, args ...any) {
	l.log.Sugar().Panicf(template, args...)
}

// Fatalf formats the message according to the format specifier and calls os.Exit.
func Fatalf(template string, args ...any) {
	DefaultLogger().log.Sugar().Fatalf(template, args...)
}

// Fatalf formats the message according to the format specifier and calls os.Exit.
func (l *ZiwiLog) Fatalf(template string, args ...any) {
	l.log.Sugar().Fatalf(template, args...)
}
