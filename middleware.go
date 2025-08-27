package log

import (
	"net/http"
	"time"
)

// HTTPMiddleware creates an HTTP middleware that logs request and response information.
// It logs the start of each request and completion with timing information.
//
// Parameters:
//   - logger: The Logger instance to use for logging
//
// Returns:
//   - func(http.Handler) http.Handler: A middleware function that can be used with standard HTTP handlers
//
// Usage:
//
//	logger := log.NewLog(log.NewOptions())
//	middleware := log.HTTPMiddleware(logger)
//	http.Handle("/", middleware(yourHandler))
func HTTPMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Log request start
			logger.Infow("HTTP请求开始",
				"method", r.Method,
				"url", r.URL.String(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"host", r.Host,
			)

			// Wrap the ResponseWriter to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     200, // Default status code
			}

			// Execute the next handler
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Log request completion
			logger.Infow("HTTP请求完成",
				"method", r.Method,
				"url", r.URL.String(),
				"status_code", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"duration_ns", duration.Nanoseconds(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code.
// It implements the http.ResponseWriter interface and additionally tracks the HTTP status code
// that was written to the response.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and calls the underlying ResponseWriter's WriteHeader.
// This method is called automatically by the HTTP server when writing the response.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write calls the underlying ResponseWriter's Write method.
// If WriteHeader hasn't been called yet, this will trigger an implicit WriteHeader(200).
func (rw *responseWriter) Write(data []byte) (int, error) {
	return rw.ResponseWriter.Write(data)
}

// Header returns the header map that will be sent by WriteHeader.
func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}
