package logutil_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/kydenul/log"
	"github.com/kydenul/log/logutil"
)

func ExampleLogHTTPRequest() {
	logger := log.NewLog(nil)
	req := httptest.NewRequest("GET", "http://example.com/api/users", nil)
	req.Header.Set("User-Agent", "MyApp/1.0")

	logutil.LogHTTPRequest(logger, req)
	// Output logs HTTP request details including method, URL, and user agent
}

func ExampleLogHTTPResponse() {
	logger := log.NewLog(nil)
	req := httptest.NewRequest("GET", "http://example.com/api/users", nil)

	// Simulate processing time
	duration := 150 * time.Millisecond
	statusCode := 200

	logutil.LogHTTPResponse(logger, req, statusCode, duration)
	// Output logs HTTP response details including status code and duration
}

func ExampleLogError() {
	logger := log.NewLog(nil)
	err := errors.New("database connection failed")

	logutil.LogError(logger, err, "Failed to connect to database")
	// Only logs if err is not nil
}

func ExampleTimer() {
	logger := log.NewLog(nil)

	// Use Timer to measure operation duration
	defer logutil.Timer(logger, "database_query")()

	// Simulate database operation
	time.Sleep(50 * time.Millisecond)
	// When the function returns, it will log the operation duration
}

func ExampleTimeFunction() {
	logger := log.NewLog(nil)

	// Use TimeFunction to measure and log function execution time
	logutil.TimeFunction(logger, "data_processing", func() {
		// Simulate data processing
		time.Sleep(100 * time.Millisecond)
	})
	// Automatically logs the execution time of the function
}

func ExampleInfoIf() {
	logger := log.NewLog(nil)
	debugMode := true

	logutil.InfoIf(logger, debugMode, "Debug mode is enabled", "config", "debug")
	// Only logs if debugMode is true
}

func ExampleErrorIf() {
	logger := log.NewLog(nil)
	hasError := true

	logutil.ErrorIf(logger, hasError, "Validation failed", "field", "email")
	// Only logs if hasError is true
}

func ExampleWithRequestID() {
	logger := log.NewLog(nil)

	// Create context with request ID
	ctx := context.WithValue(context.Background(), "request_id", "req-123-456")

	// Wrap logger to automatically include request ID
	requestLogger := logutil.WithRequestID(ctx, logger)

	// All log calls will now include the request ID
	requestLogger.Info("Processing user request")
	requestLogger.Error("Failed to process request")
}

func ExampleLogPanicAsError() {
	logger := log.NewLog(nil)

	func() {
		defer logutil.LogPanicAsError(logger, "background_task")

		// This will panic but be caught and logged as an error
		panic("unexpected error occurred")
	}()
	// Function continues execution after panic is caught and logged
}

func ExampleCheckError() {
	logger := log.NewLog(nil)

	err := performOperation()
	if logutil.CheckError(logger, err, "Operation failed") {
		// Handle error case
		return
	}

	// Continue with success case
}

func performOperation() error {
	// Simulate an operation that might fail
	return errors.New("simulated error")
}

func ExampleLogStartup() {
	logger := log.NewLog(nil)

	logutil.LogStartup(logger, "my-web-service", "v1.2.3", 8080)
	// Logs application startup information including name, version, and port
}

func ExampleLogShutdown() {
	logger := log.NewLog(nil)
	startTime := time.Now()

	// Simulate application running
	time.Sleep(100 * time.Millisecond)

	uptime := time.Since(startTime)
	logutil.LogShutdown(logger, "my-web-service", uptime)
	// Logs application shutdown information including uptime
}

// Example of using logutil in an HTTP handler
func Example_httpHandler() {
	logger := log.NewLog(nil)

	handler := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log incoming request
		logutil.LogHTTPRequest(logger, r)

		// Add request ID to context and create request-scoped logger
		ctx := context.WithValue(r.Context(), "request_id", "req-789")
		requestLogger := logutil.WithRequestID(ctx, logger)

		// Use timer to measure processing time
		defer logutil.Timer(requestLogger, "request_processing")()

		// Simulate request processing
		err := processRequest(r)
		if logutil.CheckError(requestLogger, err, "Failed to process request") {
			http.Error(w, "Internal Server Error", 500)
			logutil.LogHTTPResponse(logger, r, 500, time.Since(start))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte("OK"))

		// Log response
		logutil.LogHTTPResponse(logger, r, 200, time.Since(start))
	}

	// Create test request
	req := httptest.NewRequest("POST", "http://example.com/api/process", nil)
	w := httptest.NewRecorder()

	handler(w, req)
}

func processRequest(r *http.Request) error {
	// Simulate request processing that might fail
	_ = r
	return nil
}
