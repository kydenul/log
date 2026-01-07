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

- **Dual calling modes** - Use both instance methods (`logger.Info()`) and global functions (`log.Info()`) seamlessly
- **Zero-configuration quick start** - Get started with one line of code
- **Environment presets** - Pre-configured settings for development, production, and testing
- **Builder pattern** - Fluent API for easy configuration
- **Multi-format configuration** - Load configuration from YAML, JSON, TOML, and other formats
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
    
    // Both calling modes work with the same configuration:
    
    // Instance method calls
    logger.Info("Hello from instance method!")
    logger.Warn("Warning from instance method")
    logger.Errorf("Error from instance: %v", err)
    
    // Global function calls (automatically use the same logger)
    log.Info("Hello from global function!")
    log.Warn("Warning from global function")
    log.Errorf("Error from global: %v", err)
    
    // Both produce identical output with the same configuration
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
        ConsoleOutput(false).           // Disable console output
        Sampling(true, 100, 1000).
        Build()
    
    logger.Info("Logger configured with builder pattern")
    
    // Or use preset and customize
    logger2 := log.NewBuilder().
        Production().                    // Start with production preset
        Level("debug").                  // Override level for debugging
        Directory("/var/log/myapp").     // Custom log directory
        ConsoleOutput(false).            // Disable console for production
        Build()
    
    logger2.Debug("Custom production logger")
}
```

## Configuration

### Powered by Viper

The logging library uses [Viper](https://github.com/spf13/viper) for configuration management, providing:

- **Multiple format support**: YAML, JSON, TOML, HCL, INI, and more
- **Automatic format detection**: Based on file extension
- **Environment variable support**: Can be extended to read from environment variables
- **Configuration validation**: Built-in validation with helpful error messages
- **Hot reloading capability**: Can be extended for runtime configuration updates

### Supported Configuration Formats

| Format | Extensions | Example |
|--------|------------|---------|
| YAML   | `.yaml`, `.yml` | `config.yaml` |
| JSON   | `.json` | `config.json` |
| TOML   | `.toml` | `config.toml` |

All formats support the same configuration options with automatic conversion between formats.

### Configuration File Formats

The library supports two configuration formats:

1. **Nested configuration with `KLOG` key (Recommended)** - Configuration under a `KLOG` top-level key
2. **Direct configuration** - Configuration at the root level

> **Recommendation**: Use the `KLOG` nested configuration format. This approach allows you to combine logger configuration with other application settings in a single file, keeping your configuration organized and avoiding key conflicts.

#### Nested Configuration with `KLOG` Key (Recommended)

**YAML configuration (config.yaml):**

```yaml
# config.yaml - Recommended format with KLOG key
# This allows combining with other application settings

# Other application settings can coexist
app:
  name: "my-service"
  port: 8080

# Logger configuration under KLOG key
KLOG:
  prefix: "ZIWI_"
  directory: "./logs"
  filename: "ziwi"
  level: "info"
  format: "json"
  time_layout: "2006-01-02 15:04:05.000"

  # Basic settings
  disable_caller: false
  disable_stacktrace: false
  disable_split_error: false

  # File rotation
  max_size: 100
  max_backups: 5
  compress: true

  # Console output control
  console_output: true

  # Sampling (reduces log volume in high-traffic scenarios)
  enable_sampling: true
  sample_initial: 100
  sample_thereafter: 1000
```

**JSON configuration (config.json):**

```json
{
  "app": {
    "name": "my-service",
    "port": 8080
  },
  "KLOG": {
    "prefix": "ZIWI_",
    "directory": "./logs",
    "filename": ziwi",
    "level": "info",
    "format": "json",
    "time_layout": "2006-01-02 15:04:05.000",
    "disable_caller": false,
    "disable_stacktrace": false,
    "disable_split_error": false,
    "max_size": 100,
    "max_backups": 5,
    "compress": true,
    "console_output": true,
    "enable_sampling": true,
    "sample_initial": 100,
    "sample_thereafter": 1000
  }
}
```

**TOML configuration (config.toml):**

```toml
# config.toml - Recommended format with KLOG section

# Other application settings
[app]
name = "my-service"
port = 8080

# Logger configuration
[KLOG]
prefix = "ZIWI_"
directory = "./logs"
filename = "ziwi"
level = "info"
format = "json"
time_layout = "2006-01-02 15:04:05.000"

# Basic settings
disable_caller = false
disable_stacktrace = false
disable_split_error = false

# File rotation
max_size = 100
max_backups = 5
compress = true

# Console output control
console_output = true

# Sampling
enable_sampling = true
sample_initial = 100
sample_thereafter = 1000
```

#### Direct Configuration (Alternative)

If you prefer a dedicated configuration file for logging only, you can use direct configuration without the `KLOG` key:

**YAML configuration (log-config.yaml):**

```yaml
# log-config.yaml - Direct configuration (no KLOG key)
prefix: "ZIWI_"
directory: "./logs"
filename: "ziwi"
level: "info"
format: "json"
time_layout: "2006-01-02 15:04:05.000"

# Basic settings
disable_caller: false
disable_stacktrace: false
disable_split_error: false

# File rotation
max_size: 100
max_backups: 5
compress: true

# Console output control
console_output: true

# Sampling
enable_sampling: true
sample_initial: 100
sample_thereafter: 1000
```

### Configuration Options Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `prefix` | string | `"ZIWI_"` | Log message prefix |
| `directory` | string | `$HOME/logs` | Log file directory |
| `filename` | string | `""` | Log filename prefix (e.g., `app` -> `app-2024-01-15.log`) |
| `level` | string | `"info"` | Log level: `debug`, `info`, `warn`, `error`, `dpanic`, `panic`, `fatal` |
| `format` | string | `"console"` | Output format: `console` or `json` |
| `time_layout` | string | `"2006-01-02 15:04:05.000"` | Time format layout |
| `disable_caller` | bool | `false` | Disable caller information |
| `disable_stacktrace` | bool | `false` | Disable stack traces |
| `disable_split_error` | bool | `true` | Disable separate error log files |
| `max_size` | int | `100` | Maximum log file size in MB |
| `max_backups` | int | `3` | Maximum number of old log files to retain |
| `compress` | bool | `false` | Compress rotated log files |
| `console_output` | bool | `true` | Enable console output |
| `enable_sampling` | bool | `false` | Enable log sampling |
| `sample_initial` | int | `100` | Initial sample count per second |
| `sample_thereafter` | int | `100` | Sample count after initial burst |

### Advanced Configuration

For more control over configuration loading:

```go
// Load options from YAML file (without creating logger)
opts, err := log.LoadFromYAML("config.yaml")
if err != nil {
    log.Fatal("Failed to load YAML:", err)
}
logger := log.NewLog(opts)

// Load from any supported file format (auto-detected by extension)
opts, err := log.LoadFromFile("config.yaml") // Supports .yaml, .yml, .json, .toml
if err != nil {
    log.Fatal("Failed to load config:", err)
}
logger := log.NewLog(opts)

// Examples with different formats
opts, err = log.LoadFromFile("config.json")  // JSON format
opts, err = log.LoadFromFile("config.toml")  // TOML format
opts, err = log.LoadFromFile("config.yml")   // YAML format
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

## Dual Calling Modes

One of the key features of this logging library is **dual calling modes** - you can use both instance methods and global functions seamlessly with the same configuration.

### How It Works

When you create a logger using any method (`NewLog()`, `Quick()`, `WithPreset()`, `FromConfigFile()`, `Builder.Build()`), it automatically becomes the global default logger. This means:

- **Instance calls** like `logger.Info()` work directly on your logger instance
- **Global calls** like `log.Info()` automatically use the same logger configuration
- **Both produce identical output** with the same formatting, file destinations, and settings

### Usage Examples

```go
package main

import "github.com/kydenul/log"

func main() {
    // Create a custom logger
    logger := log.NewBuilder().
        Level("debug").
        Format("json").
        Directory("./logs").
        Prefix("[MyApp] ").
        Build()
    
    // Method 1: Instance method calls
    logger.Info("User logged in", "user_id", 123)
    logger.Warn("Rate limit approaching", "current", 95, "limit", 100)
    logger.Error("Database connection failed", "error", err)
    
    // Method 2: Global function calls (same configuration automatically)
    log.Info("User logged in", "user_id", 123)
    log.Warn("Rate limit approaching", "current", 95, "limit", 100)
    log.Error("Database connection failed", "error", err)
    
    // Both methods produce identical output:
    // [MyApp] {"level":"info","ts":"2024-01-15T10:30:45.123Z","msg":"User logged in","user_id":123}
}
```

**Complete Example**: See [`example/dual-calling/main.go`](example/dual-calling/main.go) for a comprehensive demonstration of dual calling modes with different logger creation methods.

### All Logger Creation Methods Support Dual Calling

```go
// 1. Direct creation
logger := log.NewLog(opts)
logger.Info("Instance call")
log.Info("Global call")  // Uses same config

// 2. Quick setup
logger := log.Quick()
logger.Info("Instance call")
log.Info("Global call")  // Uses same config

// 3. Environment presets
logger := log.WithPreset(log.ProductionPreset())
logger.Info("Instance call")
log.Info("Global call")  // Uses same config

// 4. Configuration files
logger, _ := log.FromConfigFile("config.yaml")
logger.Info("Instance call")
log.Info("Global call")  // Uses same config

// 5. Builder pattern
logger := log.NewBuilder().Level("debug").Build()
logger.Info("Instance call")
log.Info("Global call")  // Uses same config
```

### Multiple Logger Instances

When you create multiple logger instances, the **most recently created** logger becomes the global default:

```go
// Create first logger
logger1 := log.WithPreset(log.DevelopmentPreset())
log.Info("Uses logger1 config")  // Development preset

// Create second logger
logger2 := log.WithPreset(log.ProductionPreset())
log.Info("Uses logger2 config")  // Production preset (now global)

// Instance methods still work independently
logger1.Info("Still uses development config")
logger2.Info("Still uses production config")
```

### Benefits of Dual Calling Modes

1. **Flexibility**: Choose the calling style that fits your code structure
2. **Migration friendly**: Existing code using global functions continues to work
3. **Consistent configuration**: No need to pass logger instances everywhere
4. **Team preferences**: Different team members can use their preferred style
5. **Library integration**: Easy to integrate with existing libraries that expect global functions

### When to Use Each Mode

**Use instance methods (`logger.Info()`) when:**

- You want explicit control over which logger to use
- Working with dependency injection patterns
- Building libraries that accept logger parameters
- You have multiple loggers with different configurations

**Use global functions (`log.Info()`) when:**

- You have a single logger configuration for your application
- Migrating from other logging libraries
- You prefer the simplicity of global functions
- Working with existing code that uses global logging

## Global Logger Functions

The package provides global functions that automatically use the current default logger instance:

```go
import "github.com/kydenul/log"

func main() {
    // Configure the logger (automatically becomes global default)
    logger := log.NewBuilder().
        Level("debug").
        Format("json").
        Directory("./logs").
        Build()
    
    // Now both calling modes work with the same configuration:
    
    // Global functions (use the logger configuration above)
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
    
    // Instance methods (same output as global functions)
    logger.Debug("Debug message")
    logger.Info("Info message")
    logger.Infow("User action", "user_id", 123, "action", "login")
    
    // Manually replace the global logger (optional)
    customLogger := log.NewBuilder().Level("info").Build()
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

### Console Output Control

The logger provides flexible control over console output while maintaining file logging:

- **Independent control**: Console output can be enabled/disabled independently of file logging
- **Default behavior**: Console output is enabled by default (`console_output: true`)
- **Production optimization**: Disable console output in production to reduce performance overhead
- **File logging preserved**: When console output is disabled, all logs still write to files

**Usage examples:**

```go
// Enable console output (default behavior)
logger := log.NewBuilder().
    ConsoleOutput(true).
    Build()

// Disable console output (logs only to files)
logger := log.NewBuilder().
    ConsoleOutput(false).
    Build()

// Production setup with no console output
prodLogger := log.NewBuilder().
    Production().
    ConsoleOutput(false).  // Override preset to disable console
    Build()

// Development setup with console output
devLogger := log.NewBuilder().
    Development().
    ConsoleOutput(true).   // Explicitly enable (already default)
    Build()
```

**Configuration file examples:**

```yaml
# Enable console output (default)
console_output: true

# Disable console output (production)
console_output: false
```

```json
{
  "console_output": false
}
```

### Automatic File Management

The logger automatically handles:

- **Date-based file rotation**: Creates new log files daily (e.g., `app-2024-01-15.log`)
- **Separate error logs**: Optional separate files for error-level messages
- **Custom filename support**: Use custom prefixes for log files
- **Fallback mechanisms**: Automatically falls back to safe defaults if custom filenames fail

### Performance Optimizations

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
- **Console output enabled** (for immediate feedback)
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
- **Console output enabled** (can be overridden)
- Caller information disabled (performance)
- Stack traces disabled (performance)
- Log sampling enabled
- File compression enabled

```go
logger := log.WithPreset(log.ProductionPreset())
```

### Testing Preset

- Debug level logging
- Console output format
- **Console output enabled** (for test visibility)
- Caller information disabled (cleaner output)
- Stack traces disabled (cleaner output)
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
        ConsoleOutput(false).            // Disable console output for production
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
    // Load configuration from different formats
    var logger *log.Log
    var err error
    
    // Try different configuration formats
    if _, err := os.Stat("config.yaml"); err == nil {
        logger, err = log.FromConfigFile("config.yaml")
    } else if _, err := os.Stat("config.json"); err == nil {
        logger, err = log.FromConfigFile("config.json")
    } else if _, err := os.Stat("config.toml"); err == nil {
        logger, err = log.FromConfigFile("config.toml")
    } else {
        // Fallback to default configuration
        logger = log.Quick()
    }
    
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

## Best Practices

1. **Choose your calling mode consistently**:
   - Use **instance methods** (`logger.Info()`) when you need explicit control or multiple logger configurations
   - Use **global functions** (`log.Info()`) for simple applications with a single logger configuration
   - **Mix both modes** as needed - they work seamlessly together

2. **Use presets for common scenarios**: Start with `DevelopmentPreset()`, `ProductionPreset()`, or `TestingPreset()`

3. **Control console output appropriately**:
   - **Development**: Keep console output enabled for immediate feedback
   - **Production**: Consider disabling console output (`ConsoleOutput(false)`) to reduce performance overhead
   - **Containers**: Enable console output if using container log aggregation, disable if using file-based logging

4. **Use structured logging**: Prefer `logger.Infow("message", "key", "value")` over `logger.Infof("message %s", value)`

5. **Handle errors gracefully**: Use `logutil.LogError()` and `logutil.CheckError()` for consistent error handling

6. **Time critical operations**: Use `logutil.Timer()` or `logutil.TimeFunction()` for performance monitoring

7. **Use HTTP middleware**: Automatically log all HTTP requests and responses

8. **Configure sampling for high-traffic services**: Enable sampling in production to manage log volume

9. **Use appropriate log levels**: Debug for development, Info for production events, Error for actual problems

10. **Always call Sync()**: Call `logger.Sync()` or `log.Sync()` before application exit to flush buffers

11. **Understand global logger behavior**: When creating multiple loggers, the most recent one becomes the global default. Use `log.ReplaceLogger()` if you need explicit control

## Requirements

- Go 1.23.4 or higher
- Dependencies:
  - go.uber.org/zap
  - gopkg.in/natefinch/lumberjack.v2
  - github.com/spf13/viper (replaces gopkg.in/yaml.v3)
  - github.com/stretchr/testify (for testing)

## License

This project is licensed under the terms found in the LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
