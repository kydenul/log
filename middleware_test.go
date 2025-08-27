package log

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockLogger is a test logger that captures log messages
type mockLogger struct {
	messages []string
	fields   []map[string]any
}

func (m *mockLogger) Sync() {}

func (m *mockLogger) Debug(args ...any) {}

func (m *mockLogger) Debugf(template string, args ...any) {}

func (m *mockLogger) Debugw(msg string, keysAndValues ...any) {}

func (m *mockLogger) Debugln(args ...any) {}

func (m *mockLogger) Info(args ...any) {}

func (m *mockLogger) Infof(template string, args ...any) {}

func (m *mockLogger) Infoln(args ...any) {}

func (m *mockLogger) Warn(args ...any) {}

func (m *mockLogger) Warnf(template string, args ...any) {}

func (m *mockLogger) Warnw(msg string, keysAndValues ...any) {}

func (m *mockLogger) Warnln(args ...any) {}

func (m *mockLogger) Error(args ...any) {}

func (m *mockLogger) Errorf(template string, args ...any) {}

func (m *mockLogger) Errorw(msg string, keysAndValues ...any) {}

func (m *mockLogger) Errorln(args ...any) {}

func (m *mockLogger) Panic(args ...any) {}

func (m *mockLogger) Panicf(template string, args ...any) {}

func (m *mockLogger) Panicw(msg string, keysAndValues ...any) {}

func (m *mockLogger) Panicln(args ...any) {}

func (m *mockLogger) Fatal(args ...any) {}

func (m *mockLogger) Fatalf(template string, args ...any) {}

func (m *mockLogger) Fatalw(msg string, keysAndValues ...any) {}

func (m *mockLogger) Fatalln(args ...any) {}

func (m *mockLogger) Infow(msg string, keysAndValues ...any) {
	m.messages = append(m.messages, msg)

	// Convert keysAndValues to map
	fields := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i].(string)
			value := keysAndValues[i+1]
			fields[key] = value
		}
	}
	m.fields = append(m.fields, fields)
}

func TestHTTPMiddleware(t *testing.T) {
	// Create a mock logger
	mockLog := &mockLogger{}

	// Create the middleware
	middleware := HTTPMiddleware(mockLog)

	// Create a simple handler that returns 200 OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// Wrap the handler with middleware
	wrappedHandler := middleware(handler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	wrappedHandler.ServeHTTP(rr, req)

	// Verify the response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != "Hello, World!" {
		t.Errorf("Expected body 'Hello, World!', got '%s'", rr.Body.String())
	}

	// Verify logging
	if len(mockLog.messages) != 2 {
		t.Errorf("Expected 2 log messages, got %d", len(mockLog.messages))
	}

	// Check request start log
	if mockLog.messages[0] != "HTTP请求开始" {
		t.Errorf("Expected first message to be 'HTTP请求开始', got '%s'", mockLog.messages[0])
	}

	// Check request completion log
	if mockLog.messages[1] != "HTTP请求完成" {
		t.Errorf("Expected second message to be 'HTTP请求完成', got '%s'", mockLog.messages[1])
	}

	// Verify request start fields
	startFields := mockLog.fields[0]
	if startFields["method"] != "GET" {
		t.Errorf("Expected method 'GET', got '%v'", startFields["method"])
	}
	if startFields["url"] != "/test?param=value" {
		t.Errorf("Expected url '/test?param=value', got '%v'", startFields["url"])
	}
	if startFields["user_agent"] != "test-agent" {
		t.Errorf("Expected user_agent 'test-agent', got '%v'", startFields["user_agent"])
	}
	if startFields["remote_addr"] != "127.0.0.1:12345" {
		t.Errorf("Expected remote_addr '127.0.0.1:12345', got '%v'", startFields["remote_addr"])
	}

	// Verify request completion fields
	endFields := mockLog.fields[1]
	if endFields["method"] != "GET" {
		t.Errorf("Expected method 'GET', got '%v'", endFields["method"])
	}
	if endFields["url"] != "/test?param=value" {
		t.Errorf("Expected url '/test?param=value', got '%v'", endFields["url"])
	}
	if endFields["status_code"] != 200 {
		t.Errorf("Expected status_code 200, got '%v'", endFields["status_code"])
	}

	// Verify duration fields exist and are reasonable
	if _, ok := endFields["duration_ms"]; !ok {
		t.Error("Expected duration_ms field in completion log")
	}
	if _, ok := endFields["duration_ns"]; !ok {
		t.Error("Expected duration_ns field in completion log")
	}
}

func TestHTTPMiddleware_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Not Found", http.StatusNotFound},
		{"Internal Server Error", http.StatusInternalServerError},
		{"Created", http.StatusCreated},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLog := &mockLogger{}
			middleware := HTTPMiddleware(mockLog)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			})

			wrappedHandler := middleware(handler)
			req := httptest.NewRequest("POST", "/api/test", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			// Verify status code in response
			if rr.Code != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, rr.Code)
			}

			// Verify status code in log
			if len(mockLog.fields) < 2 {
				t.Fatal("Expected at least 2 log entries")
			}

			endFields := mockLog.fields[1]
			if endFields["status_code"] != tc.statusCode {
				t.Errorf("Expected logged status_code %d, got %v", tc.statusCode, endFields["status_code"])
			}
		})
	}
}

func TestHTTPMiddleware_TimingAccuracy(t *testing.T) {
	mockLog := &mockLogger{}
	middleware := HTTPMiddleware(mockLog)

	// Handler that sleeps for a known duration
	sleepDuration := 10 * time.Millisecond
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleepDuration)
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)
	req := httptest.NewRequest("GET", "/slow", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	wrappedHandler.ServeHTTP(rr, req)
	actualDuration := time.Since(start)

	// Verify timing in logs
	if len(mockLog.fields) < 2 {
		t.Fatal("Expected at least 2 log entries")
	}

	endFields := mockLog.fields[1]
	loggedDurationMs, ok := endFields["duration_ms"].(int64)
	if !ok {
		t.Fatal("Expected duration_ms to be int64")
	}

	// The logged duration should be at least as long as our sleep
	// but allow some tolerance for execution overhead
	expectedMinMs := sleepDuration.Milliseconds()
	if loggedDurationMs < expectedMinMs {
		t.Errorf("Expected duration_ms >= %d, got %d", expectedMinMs, loggedDurationMs)
	}

	// Should be reasonably close to actual duration (within 50ms tolerance)
	actualMs := actualDuration.Milliseconds()
	if loggedDurationMs > actualMs+50 {
		t.Errorf("Logged duration %dms seems too high compared to actual %dms", loggedDurationMs, actualMs)
	}
}

func TestResponseWriter_StatusCodeCapture(t *testing.T) {
	// Test the responseWriter directly
	rr := httptest.NewRecorder()
	wrapped := &responseWriter{
		ResponseWriter: rr,
		statusCode:     200, // Default
	}

	// Test default status code
	if wrapped.statusCode != 200 {
		t.Errorf("Expected default status code 200, got %d", wrapped.statusCode)
	}

	// Test WriteHeader
	wrapped.WriteHeader(404)
	if wrapped.statusCode != 404 {
		t.Errorf("Expected status code 404 after WriteHeader, got %d", wrapped.statusCode)
	}

	// Test Write method
	data := []byte("test data")
	n, err := wrapped.Write(data)
	if err != nil {
		t.Errorf("Unexpected error from Write: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Test Header method
	wrapped.Header().Set("X-Test", "value")
	if wrapped.Header().Get("X-Test") != "value" {
		t.Error("Header method not working correctly")
	}
}

func TestHTTPMiddleware_WithRealLogger(t *testing.T) {
	// Test with actual logger to ensure compatibility
	opts := NewOptions()
	opts.Level = "info"
	logger := NewLog(opts)

	middleware := HTTPMiddleware(logger)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(handler)
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	// This should not panic or error
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "OK") {
		t.Errorf("Expected response body to contain 'OK', got '%s'", rr.Body.String())
	}
}
