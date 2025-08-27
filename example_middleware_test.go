package log_test

import (
	"fmt"
	"net/http"

	"github.com/kydenul/log"
)

// ExampleHTTPMiddleware_usage shows how to use the HTTP middleware with a real HTTP server
func ExampleHTTPMiddleware_usage() {
	// Create logger
	logger := log.NewLog(log.NewOptions())

	// Create your handlers
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello!"))
	})

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "API response"}`))
	})

	// Create middleware
	middleware := log.HTTPMiddleware(logger)

	// Set up routes with middleware
	mux := http.NewServeMux()
	mux.Handle("/hello", middleware(helloHandler))
	mux.Handle("/api/", middleware(apiHandler))

	// In a real application, you would start the server like this:
	// log.Fatal(http.ListenAndServe(":8080", mux))

	fmt.Println("Server configured with logging middleware")
	// Output: Server configured with logging middleware
}
