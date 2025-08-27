# Quick Start Guide

Get up and running with the log library in under 5 minutes.

## Installation

```bash
go get github.com/kydenul/log
```

## 1. Zero Configuration (Fastest Start)

```go
package main

import "github.com/kydenul/log"

func main() {
    // Create logger with sensible defaults
    logger := log.Quick()
    
    // Start logging immediately
    logger.Info("Hello, World!")
    logger.Debug("This won't show (default level is info)")
    logger.Error("This is an error")
    logger.Warn("This is a warning")
}
```

**What you get:**
- Info level logging
- Console output format
- Logs to `./logs/` directory
- Automatic file rotation

## 2. Choose Your Environment (Recommended)

### Development Environment

```go
package main

import "github.com/kydenul/log"

func main() {
    // Perfect for development
    logger := log.WithPreset(log.DevelopmentPreset())
    
    logger.Debug("Debug messages are visible")
    logger.Info("Nice console formatting")
    logger.Error("Includes caller information")
}
```

**Development preset includes:**
- Debug level (see all messages)
- Console format (human-readable)
- Caller information (file:line)
- Stack traces on errors
- Fast flush (immediate output)

### Production Environment

```go
package main

import "github.com/kydenul/log"

func main() {
    // Optimized for production
    logger := log.WithPreset(log.ProductionPreset())
    
    logger.Info("Service started")
    logger.Error("This will be in JSON format")
}
```

**Production preset includes:**
- Info level (reduces noise)
- JSON format (machine-readable)
- No caller info (better performance)
- File compression
- Log sampling (handles high volume)
- Larger buffers (better performance)

### Testing Environment

```go
package main

import "github.com/kydenul/log"

func main() {
    // Clean output for tests
    logger := log.WithPreset(log.TestingPreset())
    
    logger.Debug("Test debug info")
    logger.Info("Clean, simple output")
}
```

**Testing preset includes:**
- Debug level (detailed test info)
- Console format (readable)
- No caller info (cleaner output)
- Small files (easy cleanup)
- Fast flush (immediate verification)

## 3. Configuration from YAML Files

Create a configuration file and load it:

```yaml
# config.yaml
level: debug
format: json
directory: ./logs
filename: myapp
```

```go
package main

import "github.com/kydenul/log"

func main() {
    // Load configuration from YAML file
    logger, err := log.FromConfigFile("config.yaml")
    if err != nil {
        // Falls back to quick setup on error
        logger = log.Quick()
    }
    
    logger.Debug("Configured from YAML file")
}
```

## 4. Advanced YAML Configuration

Create a comprehensive configuration file:

```yaml
# config.yaml
level: debug
format: console
directory: ./logs
filename: myapp
prefix: "MYAPP_"

# File rotation
max-size: 10
max-backups: 3
compress: true

# Performance settings
buffer-size: 1024
flush-interval: "1s"

# Sampling for high-traffic scenarios
enable-sampling: false
sample-initial: 100
sample-thereafter: 1000
```

```go
package main

import "github.com/kydenul/log"

func main() {
    logger, err := log.FromConfigFile("config.yaml")
    if err != nil {
        log.Printf("Config error: %v", err)
        logger = log.Quick() // Fallback
    }
    
    logger.Info("Configured from comprehensive YAML file")
}
```

## 5. Builder Pattern (Most Flexible)

```go
package main

import (
    "time"
    "github.com/kydenul/log"
)

func main() {
    logger := log.NewBuilder().
        Level("debug").
        Format("json").
        Directory("./logs").
        Filename("myapp").
        MaxSize(50).
        MaxBackups(5).
        Compress(true).
        BufferSize(2048).
        FlushInterval(time.Second).
        Build()
    
    logger.Info("Fully customized logger")
}
```

## 6. HTTP Server with Middleware

```go
package main

import (
    "net/http"
    "github.com/kydenul/log"
)

func main() {
    logger := log.WithPreset(log.ProductionPreset())
    
    // Add logging middleware
    middleware := log.HTTPMiddleware(logger)
    
    http.Handle("/", middleware(http.HandlerFunc(handler)))
    
    logger.Info("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Hello, World!"))
}
```

**What the middleware logs:**
- Request start: method, URL, remote address
- Request completion: status code, duration

## 7. Using Utilities

```go
package main

import (
    "errors"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

func main() {
    logger := log.Quick()
    
    // Error handling
    err := errors.New("something went wrong")
    logutil.LogError(logger, err, "Operation failed")
    
    // Performance timing
    defer logutil.Timer(logger, "main_function")()
    
    // Conditional logging
    debugMode := true
    logutil.InfoIf(logger, debugMode, "Debug mode is enabled")
    
    // Application lifecycle
    logutil.LogStartup(logger, "my-app", "v1.0.0", 8080)
}
```

## Next Steps

1. **Read the full README** for comprehensive documentation
2. **Check the examples** in the `example/` directory
3. **Explore the logutil package** for more utility functions
4. **Configure for your environment** using presets or custom configuration
5. **Add HTTP middleware** for automatic request logging
6. **Use structured logging** with key-value pairs for better searchability

## Common Patterns

### Error Handling
```go
result, err := someOperation()
if err != nil {
    logger.Errorw("Operation failed", "error", err, "input", input)
    return err
}
```

### Performance Monitoring
```go
defer logutil.Timer(logger, "database_query")()
// ... database operation ...
```

### Structured Data
```go
logger.Infow("User login",
    "user_id", userID,
    "ip_address", req.RemoteAddr,
    "user_agent", req.UserAgent(),
    "timestamp", time.Now(),
)
```

### Context-Aware Logging
```go
// In HTTP handlers
requestLogger := logutil.WithRequestID(logger, req.Context())
requestLogger.Info("Processing request") // Automatically includes request ID
```

That's it! You're now ready to use the enhanced logging library. Choose the approach that best fits your needs and start logging!