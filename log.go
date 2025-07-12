package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/kydenul/log/internal"
)

var (
	mu sync.Mutex

	// Global logger instance
	logger *ZiwiLog
	// Log prefix
	logPrefix string
)

// Initialize global logger instance
func init() {
	logger = NewLogger(NewOptions())

	internal.SetupAutoSync(logger.Sync)
}

// Ensure ZiwiLog implements Logger interface
var _ Logger = &ZiwiLog{}

// ZiwiLog is the implement of Logger interface.
// It wraps zap.Logger.
type ZiwiLog struct {
	zapcore.Encoder

	log      *zap.Logger
	logDir   string // log file directory
	file     *lumberjack.Logger
	errFile  *lumberjack.Logger
	currDate string // current date
	opts     *Options
}

// NewLogger creates a new logger instance. It will initialize the global logger instance with the specified options.
//
// Returns:
//
//   - *ZiwiLog: The new logger instance.
func NewLogger(opts *Options) *ZiwiLog {
	// 1. Lock to ensure thread safety
	mu.Lock()
	defer mu.Unlock()

	// 2. If opts is nil, use default options
	if opts == nil {
		opts = NewOptions()
	}

	// 3. Set log prefix
	logPrefix = opts.Prefix

	// 4. Set time layout, Default time layout
	timeLayout := DefaultTimeLayout
	if err := internal.ValidateTimeLayout(opts.TimeLayout); err == nil {
		timeLayout = opts.TimeLayout
	}

	// 5. Create our custom ZiwiLog with the base encoder
	logger := &ZiwiLog{
		Encoder: internal.NewBaseEncoder(opts.Format, timeLayout),
		opts:    opts,
		logDir:  opts.Directory,
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

// EncodeEntry encodes the entry and fields into a buffer.
func (l *ZiwiLog) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := l.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, fmt.Errorf("EncodeEntry error: %w", err)
	}

	// Add log prefix
	originalData := buf.String()
	buf.Reset()
	buf.AppendString(logPrefix + originalData)

	// Check and update log files
	now := time.Now().Format(time.DateOnly)
	if err := l.setupLogFiles(now); err != nil {
		return nil, err
	}

	// Write to main log file
	data := buf.Bytes()
	_, _ = l.file.Write(data)

	// For error level logs, also write to error log file
	if entry.Level == zapcore.ErrorLevel && l.errFile != nil {
		_, _ = l.errFile.Write(data)
	}

	return buf, nil
}

// setupLogFiles ensures log files are properly configured
func (l *ZiwiLog) setupLogFiles(date string) error {
	// If the date hasn't changed and the file exists, no need to reconfigure
	if l.currDate == date && l.file != nil && (l.errFile != nil || !l.opts.DisableSplitError) {
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
func Sync() { logger.Sync() }

// Sync flushs any buffered log entries. Applications should take care to call Sync before exiting.
func (l *ZiwiLog) Sync() {
	_ = l.log.Sync()
	if l.file != nil {
		_ = l.file.Close()
	}
	if l.errFile != nil {
		_ = l.errFile.Close()
	}
}

func Debug(args ...any) { logger.log.Sugar().Debug(args...) }

func (l *ZiwiLog) Debug(args ...any) { l.log.Sugar().Debug(args...) }

func Info(args ...any) { logger.log.Sugar().Info(args...) }

func (l *ZiwiLog) Info(args ...any) { l.log.Sugar().Info(args...) }

func Warn(args ...any) { logger.log.Sugar().Warn(args...) }

func (l *ZiwiLog) Warn(args ...any) { l.log.Sugar().Warn(args...) }

func Error(args ...any) { logger.log.Sugar().Error(args...) }

func (l *ZiwiLog) Error(args ...any) { l.log.Sugar().Error(args...) }

func Panic(args ...any) { logger.log.Sugar().Panic(args...) }

func (l *ZiwiLog) Panic(args ...any) { l.log.Sugar().Panic(args...) }

func Fatal(args ...any) { logger.log.Sugar().Fatal(args...) }

func (l *ZiwiLog) Fatal(args ...any) { l.log.Sugar().Fatal(args...) }

func Debugln(args ...any) { logger.log.Sugar().Debugln(args...) }

func (l *ZiwiLog) Debugln(args ...any) { l.log.Sugar().Debugln(args...) }

func Infoln(args ...any) { logger.log.Sugar().Infoln(args...) }

func (l *ZiwiLog) Infoln(args ...any) { l.log.Sugar().Infoln(args...) }

func Warnln(args ...any) { logger.log.Sugar().Warnln(args...) }

func (l *ZiwiLog) Warnln(args ...any) { l.log.Sugar().Warnln(args...) }

func Errorln(args ...any) { logger.log.Sugar().Errorln(args...) }

func (l *ZiwiLog) Errorln(args ...any) { l.log.Sugar().Errorln(args...) }

func Panicln(args ...any) { logger.log.Sugar().Panicln(args...) }

func (l *ZiwiLog) Panicln(args ...any) { l.log.Sugar().Panicln(args...) }

func Fatalln(args ...any) { logger.log.Sugar().Fatalln(args...) }

func (l *ZiwiLog) Fatalln(args ...any) { l.log.Sugar().Fatalln(args...) }

// Debugw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Debugw(msg string, ksAndvs ...any) { logger.log.Sugar().Debugw(msg, ksAndvs...) }

// Debugw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Debugw(msg string, ksAndvs ...any) { l.log.Sugar().Debugw(msg, ksAndvs...) }

// Infow logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Infow(msg string, ksAndvs ...any) { logger.log.Sugar().Infow(msg, ksAndvs...) }

// Infow logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Infow(msg string, keysAndValues ...any) { l.log.Sugar().Infow(msg, keysAndValues...) }

// Warnw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Warnw(msg string, keysAndValues ...any) { logger.log.Sugar().Warnw(msg, keysAndValues...) }

// Warnw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Warnw(msg string, keysAndValues ...any) { l.log.Sugar().Warnw(msg, keysAndValues...) }

// Errorw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func Errorw(msg string, keysAndValues ...any) { logger.log.Sugar().Errorw(msg, keysAndValues...) }

// Errorw logs a message with some additional context.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Errorw(msg string, ksAndVs ...any) { l.log.Sugar().Errorw(msg, ksAndVs...) }

// Panicw logs a message with some additional context, then panics.
// The variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...any) { logger.log.Sugar().Panicw(msg, keysAndValues...) }

// Panicw logs a message with some additional context, then panics.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Panicw(msg string, ksAndvs ...any) { l.log.Sugar().Panicw(msg, ksAndvs...) }

// Fatalw logs a message with some additional context, then calls os.Exit.
// The variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, ksAndvs ...any) { logger.log.Sugar().Fatalw(msg, ksAndvs...) }

// Fatalw logs a message with some additional context, then calls os.Exit.
// The variadic key-value pairs are treated as they are in With.
func (l *ZiwiLog) Fatalw(msg string, ksAndvs ...any) { l.log.Sugar().Fatalw(msg, ksAndvs...) }

// Debugf formats the message according to the format specifier and logs it.
func Debugf(template string, args ...any) { logger.log.Sugar().Debugf(template, args...) }

// Debugf formats the message according to the format specifier and logs it.
func (l *ZiwiLog) Debugf(template string, args ...any) { l.log.Sugar().Debugf(template, args...) }

// Infof formats the message according to the format specifier and logs it.
func Infof(template string, args ...any) { logger.log.Sugar().Infof(template, args...) }

// Infof formats the message according to the format specifier and logs it.
func (l *ZiwiLog) Infof(template string, args ...any) { l.log.Sugar().Infof(template, args...) }

// Warnf formats the message according to the format specifier and logs it.
func Warnf(template string, args ...any) { logger.log.Sugar().Warnf(template, args...) }

// Warnf formats the message according to the format specifier and logs it.
func (l *ZiwiLog) Warnf(template string, args ...any) { l.log.Sugar().Warnf(template, args...) }

// Errorf formats the message according to the format specifier and logs it.
func Errorf(template string, args ...any) { logger.log.Sugar().Errorf(template, args...) }

// Errorf formats the message according to the format specifier and logs it.
func (l *ZiwiLog) Errorf(template string, args ...any) { l.log.Sugar().Errorf(template, args...) }

// Panicf formats the message according to the format specifier and panics.
func Panicf(template string, args ...any) { logger.log.Sugar().Panicf(template, args...) }

// Panicf formats the message according to the format specifier and panics.
func (l *ZiwiLog) Panicf(template string, args ...any) { l.log.Sugar().Panicf(template, args...) }

// Fatalf formats the message according to the format specifier and calls os.Exit.
func Fatalf(template string, args ...any) { logger.log.Sugar().Fatalf(template, args...) }

// Fatalf formats the message according to the format specifier and calls os.Exit.
func (l *ZiwiLog) Fatalf(template string, args ...any) { l.log.Sugar().Fatalf(template, args...) }
