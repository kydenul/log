package logutil

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/kydenul/log"
)

// mockLogger implements log.Logger for testing
type mockLogger struct {
	logs []logEntry
}

type logEntry struct {
	level   string
	message string
	fields  []any
}

func newMockLogger() *mockLogger {
	return &mockLogger{logs: make([]logEntry, 0)}
}

func (m *mockLogger) Sync() {}

func (m *mockLogger) Debug(args ...any) {
	m.logs = append(m.logs, logEntry{level: "debug", message: "", fields: args})
}

func (m *mockLogger) Debugf(template string, args ...any) {
	m.logs = append(m.logs, logEntry{level: "debugf", message: template, fields: args})
}

func (m *mockLogger) Debugw(msg string, keysAndValues ...any) {
	m.logs = append(m.logs, logEntry{level: "debugw", message: msg, fields: keysAndValues})
}

func (m *mockLogger) Debugln(args ...any) {
	m.logs = append(m.logs, logEntry{level: "debugln", message: "", fields: args})
}

func (m *mockLogger) Info(args ...any) {
	m.logs = append(m.logs, logEntry{level: "info", message: "", fields: args})
}

func (m *mockLogger) Infof(template string, args ...any) {
	m.logs = append(m.logs, logEntry{level: "infof", message: template, fields: args})
}

func (m *mockLogger) Infow(msg string, keysAndValues ...any) {
	m.logs = append(m.logs, logEntry{level: "infow", message: msg, fields: keysAndValues})
}

func (m *mockLogger) Infoln(args ...any) {
	m.logs = append(m.logs, logEntry{level: "infoln", message: "", fields: args})
}

func (m *mockLogger) Warn(args ...any) {
	m.logs = append(m.logs, logEntry{level: "warn", message: "", fields: args})
}

func (m *mockLogger) Warnf(template string, args ...any) {
	m.logs = append(m.logs, logEntry{level: "warnf", message: template, fields: args})
}

func (m *mockLogger) Warnw(msg string, keysAndValues ...any) {
	m.logs = append(m.logs, logEntry{level: "warnw", message: msg, fields: keysAndValues})
}

func (m *mockLogger) Warnln(args ...any) {
	m.logs = append(m.logs, logEntry{level: "warnln", message: "", fields: args})
}

func (m *mockLogger) Error(args ...any) {
	m.logs = append(m.logs, logEntry{level: "error", message: "", fields: args})
}

func (m *mockLogger) Errorf(template string, args ...any) {
	m.logs = append(m.logs, logEntry{level: "errorf", message: template, fields: args})
}

func (m *mockLogger) Errorw(msg string, keysAndValues ...any) {
	m.logs = append(m.logs, logEntry{level: "errorw", message: msg, fields: keysAndValues})
}

func (m *mockLogger) Errorln(args ...any) {
	m.logs = append(m.logs, logEntry{level: "errorln", message: "", fields: args})
}

func (m *mockLogger) Panic(args ...any) {
	m.logs = append(m.logs, logEntry{level: "panic", message: "", fields: args})
	panic(args)
}

func (m *mockLogger) Panicf(template string, args ...any) {
	m.logs = append(m.logs, logEntry{level: "panicf", message: template, fields: args})
	panic(args)
}

func (m *mockLogger) Panicw(msg string, keysAndValues ...any) {
	m.logs = append(m.logs, logEntry{level: "panicw", message: msg, fields: keysAndValues})
	panic(keysAndValues)
}

func (m *mockLogger) Panicln(args ...any) {
	m.logs = append(m.logs, logEntry{level: "panicln", message: "", fields: args})
	panic(args)
}

func (m *mockLogger) Fatal(args ...any) {
	m.logs = append(m.logs, logEntry{level: "fatal", message: "", fields: args})
	os.Exit(1)
}

func (m *mockLogger) Fatalf(template string, args ...any) {
	m.logs = append(m.logs, logEntry{level: "fatalf", message: template, fields: args})
	os.Exit(1)
}

func (m *mockLogger) Fatalw(msg string, keysAndValues ...any) {
	m.logs = append(m.logs, logEntry{level: "fatalw", message: msg, fields: keysAndValues})
	os.Exit(1)
}

func (m *mockLogger) Fatalln(args ...any) {
	m.logs = append(m.logs, logEntry{level: "fatalln", message: "", fields: args})
	os.Exit(1)
}

func (m *mockLogger) getLastLog() *logEntry {
	if len(m.logs) == 0 {
		return nil
	}
	return &m.logs[len(m.logs)-1]
}

func (m *mockLogger) hasField(key string, value any) bool {
	for _, entry := range m.logs {
		for i := 0; i < len(entry.fields)-1; i += 2 {
			if entry.fields[i] == key && entry.fields[i+1] == value {
				return true
			}
		}
	}
	return false
}

func TestLogHTTPRequest(t *testing.T) {
	tests := []struct {
		name   string
		logger log.Logger
		req    *http.Request
		want   bool // whether log should be created
	}{
		{
			name:   "valid request",
			logger: newMockLogger(),
			req:    httptest.NewRequest("GET", "http://example.com/test", nil),
			want:   true,
		},
		{
			name:   "nil logger",
			logger: nil,
			req:    httptest.NewRequest("GET", "http://example.com/test", nil),
			want:   false,
		},
		{
			name:   "nil request",
			logger: newMockLogger(),
			req:    nil,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, ok := tt.logger.(*mockLogger)
			if tt.want && !ok {
				t.Fatal("Expected mockLogger for positive test cases")
			}

			LogHTTPRequest(tt.logger, tt.req)

			if tt.want {
				if len(mock.logs) == 0 {
					t.Error("Expected log entry, got none")
					return
				}

				lastLog := mock.getLastLog()
				if lastLog.level != "infow" {
					t.Errorf("Expected level 'infow', got '%s'", lastLog.level)
				}

				if lastLog.message != "HTTP请求" {
					t.Errorf("Expected message 'HTTP请求', got '%s'", lastLog.message)
				}

				if !mock.hasField("method", "GET") {
					t.Error("Expected method field with value 'GET'")
				}
			}
		})
	}
}

func TestLogHTTPResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantLevel  string
	}{
		{"success response", 200, "infow"},
		{"redirect response", 301, "warnw"},
		{"client error", 404, "errorw"},
		{"server error", 500, "errorw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockLogger()
			req := httptest.NewRequest("GET", "http://example.com/test", nil)

			LogHTTPResponse(mock, req, tt.statusCode, 100*time.Millisecond)

			if len(mock.logs) == 0 {
				t.Error("Expected log entry, got none")
				return
			}

			lastLog := mock.getLastLog()
			if lastLog.level != tt.wantLevel {
				t.Errorf("Expected level '%s', got '%s'", tt.wantLevel, lastLog.level)
			}

			if !mock.hasField("status_code", tt.statusCode) {
				t.Errorf("Expected status_code field with value %d", tt.statusCode)
			}

			if !mock.hasField("duration_ms", int64(100)) {
				t.Error("Expected duration_ms field with value 100")
			}
		})
	}
}

func TestLogError(t *testing.T) {
	tests := []struct {
		name   string
		logger log.Logger
		err    error
		msg    string
		want   bool
	}{
		{
			name:   "valid error",
			logger: newMockLogger(),
			err:    errors.New("test error"),
			msg:    "test message",
			want:   true,
		},
		{
			name:   "nil error",
			logger: newMockLogger(),
			err:    nil,
			msg:    "test message",
			want:   false,
		},
		{
			name:   "nil logger",
			logger: nil,
			err:    errors.New("test error"),
			msg:    "test message",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, ok := tt.logger.(*mockLogger)
			if tt.want && !ok {
				t.Fatal("Expected mockLogger for positive test cases")
			}

			LogError(tt.logger, tt.err, tt.msg)

			if tt.want {
				if len(mock.logs) == 0 {
					t.Error("Expected log entry, got none")
					return
				}

				lastLog := mock.getLastLog()
				if lastLog.level != "errorw" {
					t.Errorf("Expected level 'errorw', got '%s'", lastLog.level)
				}

				if lastLog.message != tt.msg {
					t.Errorf("Expected message '%s', got '%s'", tt.msg, lastLog.message)
				}

				if !mock.hasField("error", "test error") {
					t.Error("Expected error field with value 'test error'")
				}
			} else if mock != nil && len(mock.logs) > 0 {
				t.Error("Expected no log entry, got one")
			}
		})
	}
}

func TestTimer(t *testing.T) {
	mock := newMockLogger()

	// Test timer function
	timerFunc := Timer(mock, "test_operation")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Call the timer function
	timerFunc()

	if len(mock.logs) == 0 {
		t.Error("Expected log entry, got none")
		return
	}

	lastLog := mock.getLastLog()
	if lastLog.level != "infow" {
		t.Errorf("Expected level 'infow', got '%s'", lastLog.level)
	}

	if lastLog.message != "操作耗时" {
		t.Errorf("Expected message '操作耗时', got '%s'", lastLog.message)
	}

	if !mock.hasField("operation", "test_operation") {
		t.Error("Expected operation field with value 'test_operation'")
	}

	// Check that duration_ms field exists and is reasonable
	found := false
	for i := 0; i < len(lastLog.fields)-1; i += 2 {
		if lastLog.fields[i] == "duration_ms" {
			if duration, ok := lastLog.fields[i+1].(int64); ok && duration >= 10 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("Expected duration_ms field with reasonable value")
	}
}

func TestTimeFunction(t *testing.T) {
	mock := newMockLogger()
	executed := false

	TimeFunction(mock, "test_function", func() {
		executed = true
		time.Sleep(5 * time.Millisecond)
	})

	if !executed {
		t.Error("Expected function to be executed")
	}

	if len(mock.logs) == 0 {
		t.Error("Expected log entry, got none")
		return
	}

	if !mock.hasField("operation", "test_function") {
		t.Error("Expected operation field with value 'test_function'")
	}
}

func TestInfoIf(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		want      bool
	}{
		{"true condition", true, true},
		{"false condition", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockLogger()

			InfoIf(mock, tt.condition, "test message", "key", "value")

			if tt.want {
				if len(mock.logs) == 0 {
					t.Error("Expected log entry, got none")
					return
				}

				lastLog := mock.getLastLog()
				if lastLog.level != "infow" {
					t.Errorf("Expected level 'infow', got '%s'", lastLog.level)
				}
			} else {
				if len(mock.logs) > 0 {
					t.Error("Expected no log entry, got one")
				}
			}
		})
	}
}

func TestErrorIf(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		want      bool
	}{
		{"true condition", true, true},
		{"false condition", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockLogger()

			ErrorIf(mock, tt.condition, "test message", "key", "value")

			if tt.want {
				if len(mock.logs) == 0 {
					t.Error("Expected log entry, got none")
					return
				}

				lastLog := mock.getLastLog()
				if lastLog.level != "errorw" {
					t.Errorf("Expected level 'errorw', got '%s'", lastLog.level)
				}
			} else {
				if len(mock.logs) > 0 {
					t.Error("Expected no log entry, got one")
				}
			}
		})
	}
}

func TestWithRequestID(t *testing.T) {
	mock := newMockLogger()

	// Test with request ID in context
	ctx := context.WithValue(context.Background(), "request_id", "test-123")
	wrappedLogger := WithRequestID(ctx, mock)

	// Test that the wrapped logger adds request_id to log calls
	wrappedLogger.Info("test message")

	if len(mock.logs) == 0 {
		t.Error("Expected log entry, got none")
		return
	}

	if !mock.hasField("request_id", "test-123") {
		t.Error("Expected request_id field with value 'test-123'")
	}

	// Test with no request ID in context
	mock2 := newMockLogger()
	ctx2 := context.Background()
	wrappedLogger2 := WithRequestID(ctx2, mock2)

	if wrappedLogger2 != mock2 {
		t.Error("Expected original logger when no request ID in context")
	}
}

func TestLogPanicAsError(t *testing.T) {
	mock := newMockLogger()

	// Test function that panics
	func() {
		defer LogPanicAsError(mock, "test_operation")
		panic("test panic")
	}()

	if len(mock.logs) == 0 {
		t.Error("Expected log entry, got none")
		return
	}

	lastLog := mock.getLastLog()
	if lastLog.level != "errorw" {
		t.Errorf("Expected level 'errorw', got '%s'", lastLog.level)
	}

	if !mock.hasField("operation", "test_operation") {
		t.Error("Expected operation field with value 'test_operation'")
	}

	if !mock.hasField("panic", "test panic") {
		t.Error("Expected panic field with value 'test panic'")
	}
}

func TestCheckError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"with error", errors.New("test error"), true},
		{"without error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockLogger()

			result := CheckError(mock, tt.err, "test message")

			if result != tt.want {
				t.Errorf("Expected %v, got %v", tt.want, result)
			}

			if tt.want {
				if len(mock.logs) == 0 {
					t.Error("Expected log entry, got none")
				}
			} else {
				if len(mock.logs) > 0 {
					t.Error("Expected no log entry, got one")
				}
			}
		})
	}
}

func TestLogStartup(t *testing.T) {
	mock := newMockLogger()

	LogStartup(mock, "test-app", "1.0.0", 8080)

	if len(mock.logs) == 0 {
		t.Error("Expected log entry, got none")
		return
	}

	lastLog := mock.getLastLog()
	if lastLog.level != "infow" {
		t.Errorf("Expected level 'infow', got '%s'", lastLog.level)
	}

	if lastLog.message != "应用启动" {
		t.Errorf("Expected message '应用启动', got '%s'", lastLog.message)
	}

	if !mock.hasField("app_name", "test-app") {
		t.Error("Expected app_name field with value 'test-app'")
	}

	if !mock.hasField("version", "1.0.0") {
		t.Error("Expected version field with value '1.0.0'")
	}

	if !mock.hasField("port", 8080) {
		t.Error("Expected port field with value 8080")
	}
}

func TestLogShutdown(t *testing.T) {
	mock := newMockLogger()
	uptime := 5 * time.Minute

	LogShutdown(mock, "test-app", uptime)

	if len(mock.logs) == 0 {
		t.Error("Expected log entry, got none")
		return
	}

	lastLog := mock.getLastLog()
	if lastLog.level != "infow" {
		t.Errorf("Expected level 'infow', got '%s'", lastLog.level)
	}

	if lastLog.message != "应用关闭" {
		t.Errorf("Expected message '应用关闭', got '%s'", lastLog.message)
	}

	if !mock.hasField("app_name", "test-app") {
		t.Error("Expected app_name field with value 'test-app'")
	}

	if !mock.hasField("uptime", uptime.String()) {
		t.Error("Expected uptime field with correct value")
	}
}

// Test nil logger handling for all functions
func TestNilLoggerHandling(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	err := errors.New("test error")

	// These should not panic with nil logger
	LogHTTPRequest(nil, req)
	LogHTTPResponse(nil, req, 200, time.Millisecond)
	LogError(nil, err, "test")

	timerFunc := Timer(nil, "test")
	timerFunc() // Should not panic

	TimeFunction(nil, "test", func() {})
	InfoIf(nil, true, "test")
	ErrorIf(nil, true, "test")
	DebugIf(nil, true, "test")
	WarnIf(nil, true, "test")

	wrappedLogger := WithRequestID(context.Background(), nil)
	if wrappedLogger != nil {
		t.Error("Expected nil logger to remain nil")
	}

	LogStartup(nil, "test", "1.0.0", 8080)
	LogShutdown(nil, "test", time.Minute)

	result := CheckError(nil, err, "test")
	if !result {
		t.Error("Expected CheckError to return true even with nil logger")
	}

	// LogPanicAsError with nil logger should still recover
	func() {
		defer LogPanicAsError(nil, "test")
		panic("test")
	}()
}

func TestLogutilIntegration(t *testing.T) {
	// Create a logger instance
	logger := log.NewLog(nil)

	// Test HTTP request logging
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	LogHTTPRequest(logger, req)

	// Test HTTP response logging
	LogHTTPResponse(logger, req, 200, 100*time.Millisecond)

	// Test error logging
	err := errors.New("test error")
	LogError(logger, err, "Test error occurred")

	// Test timer
	timerFunc := Timer(logger, "test_operation")
	time.Sleep(10 * time.Millisecond)
	timerFunc()

	// Test time function
	TimeFunction(logger, "test_function", func() {
		time.Sleep(5 * time.Millisecond)
	})

	// Test conditional logging
	InfoIf(logger, true, "This should be logged")
	InfoIf(logger, false, "This should not be logged")

	ErrorIf(logger, true, "Error condition met")
	ErrorIf(logger, false, "Error condition not met")

	// Test request ID logging
	ctx := context.WithValue(context.Background(), "request_id", "test-123")
	requestLogger := WithRequestID(ctx, logger)
	requestLogger.Info("Request with ID")

	// Test panic recovery
	func() {
		defer LogPanicAsError(logger, "test_panic")
		// This would normally panic, but it's caught and logged
		// panic("test panic")
	}()

	// Test error checking
	hasError := CheckError(logger, err, "Checking error")
	if !hasError {
		t.Error("Expected CheckError to return true for non-nil error")
	}

	hasError = CheckError(logger, nil, "Checking nil error")
	if hasError {
		t.Error("Expected CheckError to return false for nil error")
	}

	// Test startup and shutdown logging
	LogStartup(logger, "test-app", "1.0.0", 8080)
	LogShutdown(logger, "test-app", 5*time.Minute)

	t.Log("All logutil integration tests passed")
}

func TestLogutilWithNilLogger(t *testing.T) {
	// Test that all functions handle nil logger gracefully
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	err := errors.New("test error")

	// These should not panic
	LogHTTPRequest(nil, req)
	LogHTTPResponse(nil, req, 200, time.Millisecond)
	LogError(nil, err, "test")

	timerFunc := Timer(nil, "test")
	timerFunc()

	TimeFunction(nil, "test", func() {})
	InfoIf(nil, true, "test")
	ErrorIf(nil, true, "test")

	wrappedLogger := WithRequestID(context.Background(), nil)
	if wrappedLogger != nil {
		t.Error("Expected nil logger to remain nil")
	}

	LogStartup(nil, "test", "1.0.0", 8080)
	LogShutdown(nil, "test", time.Minute)

	t.Log("All nil logger tests passed")
}
