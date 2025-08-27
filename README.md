# Log

A high-performance, structured logging package for Go applications, built on top of [zap](https://github.com/uber-go/zap) with enhanced usability features and simplified interface.

## Features

### Core Features

- Multiple log levels (Debug, Info, Warn, Error, Panic, Fatal)
- Structured logging with key-value pairs
- Printf-style logging with format strings
- Println-style logging support
- Flexible output formats (JSON and console)
- Configurable time layout
- Log file rotation by date
- Separate error log files
- Optional caller information and stack traces
- Built on top of uber-go/zap for high performance

### Enhanced Usability Features

- **Zero-configuration quick start** - Get started with one line of code
- **Environment presets** - Pre-configured settings for development, production, and testing
- **Builder pattern** - Fluent API for easy configuration
- **YAML configuration** - Load configuration from YAML files
- **HTTP middleware** - Built-in middleware for request/response logging
- **Utility functions** - Common logging patterns made simple
- **Enhanced error handling** - Better error messages and automatic fallbacks
- **Performance optimizations** - Buffering, sampling, and other performance features

## Installation

```bash
go get github.com/kydenul/log
```

## Quick Start

### Simplest Usage (Zero Configuration)

```go
package main

import "github.com/kydenul/log"

func main() {
    // Create logger with sensible defaults
    logger := log.Quick()
    logger.Info("Hello, World!")
    
    // Or use global functions (existing API still works)
    log.Info("Hello, World!")
    log.Infof("Processing item %d", 123)
}
```

### Using Environment Presets

```go
package main

import "github.com/kydenul/log"

func main() {
    // Development environment (debug level, console output, caller info)
    devLogger := log.WithPreset(log.DevelopmentPreset())
    devLogger.Debug("Development mode enabled")
    
    // Production environment (info level, JSON format, optimized)
    prodLogger := log.WithPreset(log.ProductionPreset())
    prodLogger.Info("Production service started")
    
    // Testing environment (debug level, simplified output)
    testLogger := log.WithPreset(log.TestingPreset())
    testLogger.Debug("Running tests")
}
```

### Using Builder Pattern

```go
package main

import (
    "time"
    "github.com/kydenul/log"
)

func main() {
    // Fluent configuration with builder pattern
    logger := log.NewBuilder().
        Level("debug").
        Format("json").
        Directory("./logs").
        Filename("myapp").
        BufferSize(1024).
        FlushInterval(time.Second).
        Sampling(true, 100, 1000).
        Build()
    
    logger.Info("Logger configured with builder pattern")
    
    // Or use preset and customize
    logger2 := log.NewBuilder().
        Production().                    // Start with production preset
        Level("debug").                  // Override level for debugging
        Directory("/var/log/myapp").     // Custom log directory
        Build()
    
    logger2.Debug("Custom production logger")
}
```

## Configuration

### Configuration from Files

Load configuration from YAML files:

```go
// Load from YAML file
logger, err := log.FromConfigFile("config.yaml")
if err != nil {
    log.Fatal("Failed to load config:", err)
}
logger.Info("Logger configured from file")
```

Example YAML configuration:

```yaml
# config.yaml
prefix: "MYAPP"
directory: "./logs"
filename: "app"
level: "info"
format: "json"
time-layout: "2006-01-02 15:04:05.000"

# Basic settings
disable-caller: false
disable-stacktrace: false
disable-split-error: false

# File rotation
max-size: 100
max-backups: 5
compress: true

# Performance settings
buffer-size: 2048
flush-interval: "5s"

# Sampling (reduces log volume in high-traffic scenarios)
enable-sampling: true
sample-initial: 100
sample-thereafter: 1000
```

### Advanced Configuration

For more control over configuration loading:

```go
// Load options from YAML file (without creating logger)
opts, err := log.LoadFromYAML("config.yaml")
if err != nil {
    log.Fatal("Failed to load YAML:", err)
}
logger := log.NewLog(opts)

// Load from any supported file format
opts, err := log.LoadFromFile("config.yaml") // Supports .yaml, .yml
if err != nil {
    log.Fatal("Failed to load config:", err)
}
logger := log.NewLog(opts)
```

## HTTP Middleware

Built-in HTTP middleware for automatic request/response logging:

```go
package main

import (
    "net/http"
    "github.com/kydenul/log"
)

func main() {
    logger := log.WithPreset(log.ProductionPreset())
    
    // Create HTTP middleware
    middleware := log.HTTPMiddleware(logger)
    
    // Wrap your handlers
    http.Handle("/api/users", middleware(http.HandlerFunc(usersHandler)))
    http.Handle("/api/orders", middleware(http.HandlerFunc(ordersHandler)))
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Users endpoint"))
}
```

The middleware automatically logs:

- Request start with method, URL, remote address, user agent
- Request completion with status code, duration, and timing

## Global Logger Functions

The package provides global functions that use a default logger instance:

```go
import "github.com/kydenul/log"

func main() {
    // Basic logging functions
    log.Debug("Debug message")
    log.Info("Info message")
    log.Warn("Warning message")
    log.Error("Error message")
    
    // Structured logging
    log.Infow("User action", "user_id", 123, "action", "login")
    log.Errorw("Operation failed", "error", err, "retry_count", 3)
    
    // Formatted logging
    log.Infof("Processing %d items", count)
    log.Errorf("Failed to connect to %s: %v", host, err)
    
    // Line-based logging
    log.Infoln("This", "is", "a", "line", "message")
    
    // Replace the global logger
    customLogger := log.NewBuilder().Level("debug").Build()
    log.ReplaceLogger(customLogger)
    
    // Sync all loggers before exit
    defer log.Sync()
}
```

## Utility Functions

The `logutil` package provides convenient utility functions for common logging patterns:

```go
import (
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

func main() {
    logger := log.Quick()
    
    // Error handling utilities
    err := someOperation()
    logutil.LogError(logger, err, "Operation failed")
    logutil.FatalOnError(logger, err, "Critical operation failed")
    
    // Performance timing
    defer logutil.Timer(logger, "database_query")()
    // ... database operation ...
    
    // Or time a function
    logutil.TimeFunction(logger, "data_processing", func() {
        // ... processing logic ...
    })
    
    // Conditional logging
    debugMode := true
    logutil.InfoIf(logger, debugMode, "Debug info", "key", "value")
    logutil.ErrorIf(logger, err != nil, "Error occurred", "error", err)
    
    // HTTP request logging
    logutil.LogHTTPRequest(logger, request)
    logutil.LogHTTPResponse(logger, request, 200, duration)
    
    // Application lifecycle
    logutil.LogStartup(logger, "my-service", "v1.0.0", 8080)
    logutil.LogShutdown(logger, "my-service", uptime)
    
    // Panic recovery
    defer logutil.LogPanicAsError(logger, "risky_operation")
}
```

## Key Features

### Automatic File Management

The logger automatically handles:

- **Date-based file rotation**: Creates new log files daily (e.g., `app-2024-01-15.log`)
- **Separate error logs**: Optional separate files for error-level messages
- **Custom filename support**: Use custom prefixes for log files
- **Fallback mechanisms**: Automatically falls back to safe defaults if custom filenames fail

### Performance Optimizations

- **Buffered writes**: Configurable buffer sizes for better performance
- **Sampling**: Reduce log volume in high-traffic scenarios
- **Atomic operations**: Thread-safe file operations with minimal locking
- **Memory pooling**: Reuses buffers to reduce garbage collection

### Enhanced Error Handling

- **Graceful degradation**: Continues logging even when configuration is invalid
- **Automatic recovery**: Falls back to safe defaults when file operations fail
- **Detailed error messages**: Clear error messages with suggestions for fixes
- **Validation**: Comprehensive validation of all configuration options

## Environment Presets

Choose from pre-configured environments:

### Development Preset

- Debug level logging
- Console output format
- Caller information enabled
- Stack traces enabled
- Fast flush for immediate feedback
- No log sampling

```go
logger := log.WithPreset(log.DevelopmentPreset())
```

### Production Preset

- Info level logging
- JSON output format
- Caller information disabled (performance)
- Stack traces disabled (performance)
- Larger buffer sizes
- Log sampling enabled
- File compression enabled

```go
logger := log.WithPreset(log.ProductionPreset())
```

### Testing Preset

- Debug level logging
- Console output format
- Caller information disabled (cleaner output)
- Stack traces disabled (cleaner output)
- Small buffer sizes
- Fast flush for test verification
- No log sampling

```go
logger := log.WithPreset(log.TestingPreset())
```

## Examples

### Web Service with Middleware

```go
package main

import (
    "net/http"
    "time"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

func main() {
    // Configure logger for production
    logger := log.NewBuilder().
        Production().
        Directory("/var/log/myservice").
        Build()
    
    // Log service startup
    logutil.LogStartup(logger, "my-service", "v1.2.3", 8080)
    
    // Setup middleware
    middleware := log.HTTPMiddleware(logger)
    
    // Setup routes
    http.Handle("/health", middleware(http.HandlerFunc(healthHandler)))
    http.Handle("/api/data", middleware(http.HandlerFunc(dataHandler)))
    
    // Start server
    logger.Info("Server starting on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        logger.Fatal("Server failed to start", "error", err)
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
    logger := log.Quick() // Or get from context
    
    defer logutil.Timer(logger, "data_processing")()
    
    // Simulate processing
    time.Sleep(100 * time.Millisecond)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status": "success"}`))
}
```

### Configuration-Driven Application

```go
package main

import (
    "os"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

func main() {
    // Load configuration from multiple sources
    logger, err := log.FromConfig("config.yaml", "MYAPP_")
    logutil.FatalOnError(logger, err, "Failed to initialize logger")
    
    // Application logic
    processData(logger)
}

func processData(logger *log.Log) {
    defer logutil.LogPanicAsError(logger, "data_processing")
    
    // Simulate work with error handling
    data, err := loadData()
    if logutil.CheckError(logger, err, "Failed to load data") {
        return
    }
    
    // Process data with timing
    logutil.TimeFunction(logger, "data_transformation", func() {
        transformData(data)
    })
    
    logger.Info("Data processing completed successfully")
}
```

### Testing with Logging

```go
package main

import (
    "testing"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

func TestDataProcessing(t *testing.T) {
    // Use testing preset for clean output
    logger := log.WithPreset(log.TestingPreset())
    
    // Test with logging
    logutil.InfoIf(logger, testing.Verbose(), "Starting data processing test")
    
    result := processTestData(logger)
    
    if result == nil {
        t.Error("Expected result, got nil")
    }
    
    logutil.InfoIf(logger, testing.Verbose(), "Test completed", "result", result)
}
```

## Error Handling and Validation

The library provides robust error handling and validation:

```go
// Configuration validation with automatic fixes
opts := log.NewOptions().WithLevel("invalid_level") // Will use default level
logger := log.NewLog(opts) // Logs warning but continues with safe defaults

// File operation error handling
logger := log.NewBuilder().
    Directory("/invalid/path").  // Will fall back to default directory
    Filename("invalid<>name").   // Will sanitize or fall back to default
    Build()

// YAML configuration with detailed error messages
logger, err := log.FromConfigFile("config.yaml")
if err != nil {
    // Error includes specific guidance on what went wrong
    log.Printf("Config error: %v", err)
    logger = log.Quick() // Fall back to quick setup
}
```

## Best Practices

1. **Use presets for common scenarios**: Start with `DevelopmentPreset()`, `ProductionPreset()`, or `TestingPreset()`

2. **Use structured logging**: Prefer `logger.Infow("message", "key", "value")` over `logger.Infof("message %s", value)`

3. **Handle errors gracefully**: Use `logutil.LogError()` and `logutil.CheckError()` for consistent error handling

4. **Time critical operations**: Use `logutil.Timer()` or `logutil.TimeFunction()` for performance monitoring

5. **Use HTTP middleware**: Automatically log all HTTP requests and responses

6. **Configure sampling for high-traffic services**: Enable sampling in production to manage log volume

7. **Use appropriate log levels**: Debug for development, Info for production events, Error for actual problems

8. **Always call Sync()**: Call `logger.Sync()` or `log.Sync()` before application exit to flush buffers

## Requirements

- Go 1.23.4 or higher
- Dependencies:
  - go.uber.org/zap
  - gopkg.in/natefinch/lumberjack.v2
  - gopkg.in/yaml.v3
  - github.com/stretchr/testify (for testing)

## License

This project is licensed under the terms found in the LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
