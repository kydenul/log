// Package logutil provides utility functions for common logging tasks.
// It works with any logger that implements the log.Logger interface.
package logutil

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/kydenul/log"
)

// LogHTTPRequest logs details about an HTTP request.
// It records the HTTP method, URL, remote address, and user agent.
func LogHTTPRequest(logger log.Logger, req *http.Request) {
	if logger == nil || req == nil {
		return
	}

	logger.Infow("HTTP请求",
		"method", req.Method,
		"url", req.URL.String(),
		"remote_addr", req.RemoteAddr,
		"user_agent", req.UserAgent(),
		"content_length", req.ContentLength,
	)
}

// LogHTTPResponse logs details about an HTTP response.
// It records the HTTP method, URL, status code, and response duration.
func LogHTTPResponse(logger log.Logger, req *http.Request, statusCode int, duration time.Duration) {
	if logger == nil || req == nil {
		return
	}

	level := "info"
	if statusCode >= 400 {
		level = "error"
	} else if statusCode >= 300 {
		level = "warn"
	}

	fields := []any{
		"method", req.Method,
		"url", req.URL.String(),
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
	}

	switch level {
	case "error":
		logger.Errorw("HTTP响应", fields...)
	case "warn":
		logger.Warnw("HTTP响应", fields...)
	default:
		logger.Infow("HTTP响应", fields...)
	}
}

// LogError logs an error if it's not nil.
// This is a convenience function to avoid repetitive error checking in code.
func LogError(logger log.Logger, err error, msg string) {
	if logger == nil || err == nil {
		return
	}

	logger.Errorw(msg, "error", err.Error())
}

// FatalOnError logs an error and exits the program if the error is not nil.
// This should be used sparingly, only for unrecoverable errors during startup.
func FatalOnError(logger log.Logger, err error, msg string) {
	if err == nil {
		return
	}

	if logger != nil {
		logger.Fatalw(msg, "error", err.Error())
	} else {
		// Fallback if logger is nil
		log.Fatalf("%s: %v", msg, err)
	}
}

// Timer returns a function that, when called, logs the elapsed time since Timer was called.
// This is useful for measuring and logging the duration of operations.
//
// Example usage:
//
//	defer Timer(logger, "database_query")()
func Timer(logger log.Logger, name string) func() {
	if logger == nil {
		return func() {} // No-op if logger is nil
	}

	start := time.Now()
	return func() {
		duration := time.Since(start)
		logger.Infow("操作耗时",
			"operation", name,
			"duration_ms", duration.Milliseconds(),
			"duration", duration.String(),
		)
	}
}

// TimeFunction executes a function and logs its execution time.
// This is a convenience wrapper around Timer for simple function timing.
func TimeFunction(logger log.Logger, name string, fn func()) {
	if fn == nil {
		return
	}

	if logger == nil {
		fn() // Just execute the function if no logger
		return
	}

	defer Timer(logger, name)()
	fn()
}

// InfoIf logs an info message only if the condition is true.
// This helps reduce conditional logging boilerplate in application code.
func InfoIf(logger log.Logger, condition bool, msg string, args ...any) {
	if logger == nil || !condition {
		return
	}

	if len(args) > 0 {
		logger.Infow(msg, args...)
	} else {
		logger.Info(msg)
	}
}

// ErrorIf logs an error message only if the condition is true.
// This helps reduce conditional logging boilerplate in application code.
func ErrorIf(logger log.Logger, condition bool, msg string, args ...any) {
	if logger == nil || !condition {
		return
	}

	if len(args) > 0 {
		logger.Errorw(msg, args...)
	} else {
		logger.Error(msg)
	}
}

// DebugIf logs a debug message only if the condition is true.
// This helps reduce conditional logging boilerplate in application code.
func DebugIf(logger log.Logger, condition bool, msg string, args ...any) {
	if logger == nil || !condition {
		return
	}

	if len(args) > 0 {
		logger.Debugw(msg, args...)
	} else {
		logger.Debug(msg)
	}
}

// WarnIf logs a warning message only if the condition is true.
// This helps reduce conditional logging boilerplate in application code.
func WarnIf(logger log.Logger, condition bool, msg string, args ...any) {
	if logger == nil || !condition {
		return
	}

	if len(args) > 0 {
		logger.Warnw(msg, args...)
	} else {
		logger.Warn(msg)
	}
}

// WithRequestID extracts a request ID from context and returns a logger wrapper
// that automatically includes the request ID in all log messages.
// If no request ID is found in context, returns the original logger.
func WithRequestID(ctx context.Context, logger log.Logger) log.Logger {
	if logger == nil || ctx == nil {
		return logger
	}

	// Try common request ID keys
	requestIDKeys := []string{"request_id", "requestId", "req_id", "trace_id", "traceId"}

	for _, key := range requestIDKeys {
		if requestID := ctx.Value(key); requestID != nil {
			// Create a wrapper that adds request_id to all log calls
			return &requestIDLogger{
				logger:    logger,
				requestID: requestID,
			}
		}
	}

	return logger
}

// requestIDLogger is a wrapper that automatically adds request ID to log messages
type requestIDLogger struct {
	logger    log.Logger
	requestID any
}

func (r *requestIDLogger) Sync() {
	r.logger.Sync()
}

func (r *requestIDLogger) Debug(args ...any) {
	r.logger.Debugw("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Debugf(template string, args ...any) {
	r.logger.Debugw(template, "request_id", r.requestID, "formatted_args", args)
}

func (r *requestIDLogger) Debugw(msg string, keysAndValues ...any) {
	r.logger.Debugw(msg, append([]any{"request_id", r.requestID}, keysAndValues...)...)
}

func (r *requestIDLogger) Debugln(args ...any) {
	r.logger.Debugw("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Info(args ...any) {
	r.logger.Infow("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Infof(template string, args ...any) {
	r.logger.Infow(template, "request_id", r.requestID, "formatted_args", args)
}

func (r *requestIDLogger) Infow(msg string, keysAndValues ...any) {
	r.logger.Infow(msg, append([]any{"request_id", r.requestID}, keysAndValues...)...)
}

func (r *requestIDLogger) Infoln(args ...any) {
	r.logger.Infow("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Warn(args ...any) {
	r.logger.Warnw("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Warnf(template string, args ...any) {
	r.logger.Warnw(template, "request_id", r.requestID, "formatted_args", args)
}

func (r *requestIDLogger) Warnw(msg string, keysAndValues ...any) {
	r.logger.Warnw(msg, append([]any{"request_id", r.requestID}, keysAndValues...)...)
}

func (r *requestIDLogger) Warnln(args ...any) {
	r.logger.Warnw("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Error(args ...any) {
	r.logger.Errorw("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Errorf(template string, args ...any) {
	r.logger.Errorw(template, "request_id", r.requestID, "formatted_args", args)
}

func (r *requestIDLogger) Errorw(msg string, keysAndValues ...any) {
	r.logger.Errorw(msg, append([]any{"request_id", r.requestID}, keysAndValues...)...)
}

func (r *requestIDLogger) Errorln(args ...any) {
	r.logger.Errorw("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Panic(args ...any) {
	r.logger.Panicw("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Panicf(template string, args ...any) {
	r.logger.Panicw(template, "request_id", r.requestID, "formatted_args", args)
}

func (r *requestIDLogger) Panicw(msg string, keysAndValues ...any) {
	r.logger.Panicw(msg, append([]any{"request_id", r.requestID}, keysAndValues...)...)
}

func (r *requestIDLogger) Panicln(args ...any) {
	r.logger.Panicw("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Fatal(args ...any) {
	r.logger.Fatalw("", append([]any{"request_id", r.requestID}, args...)...)
}

func (r *requestIDLogger) Fatalf(template string, args ...any) {
	r.logger.Fatalw(template, "request_id", r.requestID, "formatted_args", args)
}

func (r *requestIDLogger) Fatalw(msg string, keysAndValues ...any) {
	r.logger.Fatalw(msg, append([]any{"request_id", r.requestID}, keysAndValues...)...)
}

func (r *requestIDLogger) Fatalln(args ...any) {
	r.logger.Fatalw("", append([]any{"request_id", r.requestID}, args...)...)
}

// LogPanic recovers from a panic and logs it as an error.
// This should be used with defer in functions that might panic.
//
// Example usage:
//
//	defer LogPanic(logger, "processing_request")
func LogPanic(logger log.Logger, operation string) {
	if r := recover(); r != nil {
		if logger != nil {
			logger.Errorw("Panic recovered",
				"operation", operation,
				"panic", r,
			)
		}

		// Re-panic to maintain original behavior
		panic(r)
	}
}

// LogPanicAsError recovers from a panic and logs it as an error without re-panicking.
// This is useful when you want to handle panics gracefully.
//
// Example usage:
//
//	defer LogPanicAsError(logger, "background_task")
func LogPanicAsError(logger log.Logger, operation string) {
	if r := recover(); r != nil {
		if logger != nil {
			logger.Errorw("Panic recovered and handled",
				"operation", operation,
				"panic", r,
			)
		}
	}
}

// Must logs a fatal error and exits if err is not nil.
// This is similar to FatalOnError but with a shorter name for convenience.
func Must(logger log.Logger, err error, msg string) {
	FatalOnError(logger, err, msg)
}

// CheckError logs an error if it's not nil and returns whether an error occurred.
// This is useful for error checking in conditional statements.
func CheckError(logger log.Logger, err error, msg string) bool {
	if err != nil {
		LogError(logger, err, msg)
		return true
	}
	return false
}

// LogStartup logs application startup information.
// This is a convenience function for logging common startup details.
func LogStartup(logger log.Logger, appName, version string, port int) {
	if logger == nil {
		return
	}

	hostname, _ := os.Hostname()
	pid := os.Getpid()

	logger.Infow("应用启动",
		"app_name", appName,
		"version", version,
		"port", port,
		"hostname", hostname,
		"pid", pid,
	)
}

// LogShutdown logs application shutdown information.
// This is a convenience function for logging shutdown details.
func LogShutdown(logger log.Logger, appName string, uptime time.Duration) {
	if logger == nil {
		return
	}

	logger.Infow("应用关闭",
		"app_name", appName,
		"uptime", uptime.String(),
		"uptime_seconds", uptime.Seconds(),
	)
}
