# Configuration Reference

Complete reference for all configuration options in the log library.

## Configuration Methods

### 1. Code Configuration (Builder Pattern)

```go
logger := log.NewBuilder().
    Level("debug").
    Format("json").
    Directory("./logs").
    Filename("myapp").
    Prefix("MYAPP").
    TimeLayout("2006-01-02 15:04:05.000").
    DisableCaller(false).
    DisableStacktrace(false).
    DisableSplitError(false).
    MaxSize(100).
    MaxBackups(5).
    Compress(true).
    BufferSize(2048).
    FlushInterval(5 * time.Second).
    Sampling(true, 100, 1000).
    Build()
```

### 2. Options Configuration (Traditional)

```go
opts := log.NewOptions().
    WithLevel("debug").
    WithFormat("json").
    WithDirectory("./logs").
    WithFilename("myapp").
    WithPrefix("MYAPP").
    WithTimeLayout("2006-01-02 15:04:05.000").
    WithDisableCaller(false).
    WithDisableStacktrace(false).
    WithDisableSplitError(false).
    WithMaxSize(100).
    WithMaxBackups(5).
    WithCompress(true).
    WithBufferSize(2048).
    WithFlushInterval(5 * time.Second).
    WithSampling(true, 100, 1000)

logger := log.NewLog(opts)
```

### 3. YAML Configuration

```yaml
# config.yaml
prefix: "MYAPP"
directory: "./logs"
filename: "myapp"
level: "debug"
format: "json"
time-layout: "2006-01-02 15:04:05.000"

# Output control
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

# Sampling
enable-sampling: true
sample-initial: 100
sample-thereafter: 1000
```

### 4. Environment Variables

```bash
# Basic settings
export LOG_LEVEL=debug
export LOG_FORMAT=json
export LOG_DIRECTORY=./logs
export LOG_FILENAME=myapp
export LOG_PREFIX=MYAPP
export LOG_TIME_LAYOUT="2006-01-02 15:04:05.000"

# Output control
export LOG_DISABLE_CALLER=false
export LOG_DISABLE_STACKTRACE=false
export LOG_DISABLE_SPLIT_ERROR=false

# File rotation
export LOG_MAX_SIZE=100
export LOG_MAX_BACKUPS=5
export LOG_COMPRESS=true

# Performance settings
export LOG_BUFFER_SIZE=2048
export LOG_FLUSH_INTERVAL=5s

# Sampling
export LOG_ENABLE_SAMPLING=true
export LOG_SAMPLE_INITIAL=100
export LOG_SAMPLE_THEREAFTER=1000
```

## Configuration Options

### Basic Settings

#### `level` (string)
**Default:** `"info"`  
**Valid values:** `"debug"`, `"info"`, `"warn"`, `"error"`, `"dpanic"`, `"panic"`, `"fatal"`  
**Description:** Minimum log level to output. Messages below this level are ignored.

```go
// Code
.Level("debug")

// YAML
level: debug

// Environment
LOG_LEVEL=debug
```

#### `format` (string)
**Default:** `"console"`  
**Valid values:** `"console"`, `"json"`  
**Description:** Output format for log messages.

- `console`: Human-readable format with colors (for development)
- `json`: Machine-readable JSON format (for production)

```go
// Code
.Format("json")

// YAML
format: json

// Environment
LOG_FORMAT=json
```

#### `directory` (string)
**Default:** `"./logs"`  
**Description:** Directory where log files are stored. Created automatically if it doesn't exist.

```go
// Code
.Directory("/var/log/myapp")

// YAML
directory: /var/log/myapp

// Environment
LOG_DIRECTORY=/var/log/myapp
```

#### `filename` (string)
**Default:** `"app"`  
**Description:** Base filename for log files. Actual files will have date and extension appended.

Example: `filename: "myapp"` creates files like:
- `myapp-2024-01-15.log`
- `myapp-2024-01-15_error.log`

```go
// Code
.Filename("myapp")

// YAML
filename: myapp

// Environment
LOG_FILENAME=myapp
```

#### `prefix` (string)
**Default:** `""`  
**Description:** Prefix added to all log messages.

```go
// Code
.Prefix("MYAPP")

// YAML
prefix: MYAPP

// Environment
LOG_PREFIX=MYAPP
```

#### `time-layout` (string)
**Default:** `"2006-01-02 15:04:05.000"`  
**Description:** Go time format layout for timestamps.

Common formats:
- `"2006-01-02 15:04:05.000"` - Default with milliseconds
- `"2006-01-02 15:04:05"` - Without milliseconds
- `"15:04:05"` - Time only
- `"2006/01/02 15:04:05"` - Alternative date format

```go
// Code
.TimeLayout("2006-01-02 15:04:05")

// YAML
time-layout: "2006-01-02 15:04:05"

// Environment
LOG_TIME_LAYOUT="2006-01-02 15:04:05"
```

### Output Control

#### `disable-caller` (boolean)
**Default:** `false`  
**Description:** Whether to disable caller information (file:line) in log messages.

- `false`: Include caller info (useful for debugging)
- `true`: Disable caller info (better performance)

```go
// Code
.DisableCaller(true)

// YAML
disable-caller: true

// Environment
LOG_DISABLE_CALLER=true
```

#### `disable-stacktrace` (boolean)
**Default:** `false`  
**Description:** Whether to disable stack traces for error-level and above messages.

```go
// Code
.DisableStacktrace(true)

// YAML
disable-stacktrace: true

// Environment
LOG_DISABLE_STACKTRACE=true
```

#### `disable-split-error` (boolean)
**Default:** `false`  
**Description:** Whether to disable separate error log files.

- `false`: Error messages go to both main log and separate error log
- `true`: All messages go to main log only

```go
// Code
.DisableSplitError(true)

// YAML
disable-split-error: true

// Environment
LOG_DISABLE_SPLIT_ERROR=true
```

### File Rotation

#### `max-size` (integer)
**Default:** `100`  
**Description:** Maximum size of log files in megabytes before rotation.

```go
// Code
.MaxSize(50)

// YAML
max-size: 50

// Environment
LOG_MAX_SIZE=50
```

#### `max-backups` (integer)
**Default:** `3`  
**Description:** Maximum number of old log files to retain.

```go
// Code
.MaxBackups(10)

// YAML
max-backups: 10

// Environment
LOG_MAX_BACKUPS=10
```

#### `compress` (boolean)
**Default:** `false`  
**Description:** Whether to compress rotated log files using gzip.

```go
// Code
.Compress(true)

// YAML
compress: true

// Environment
LOG_COMPRESS=true
```

### Performance Settings

#### `buffer-size` (integer)
**Default:** `1024`  
**Description:** Buffer size in bytes for log writes. Larger buffers improve performance but may delay log visibility.

Recommended values:
- Development: `512` (fast feedback)
- Production: `2048` or `4096` (better performance)
- High-traffic: `8192` or higher

```go
// Code
.BufferSize(2048)

// YAML
buffer-size: 2048

// Environment
LOG_BUFFER_SIZE=2048
```

#### `flush-interval` (duration)
**Default:** `1s`  
**Description:** How often to flush buffered logs to disk.

Format: Go duration string (`"1s"`, `"500ms"`, `"2m"`, etc.)

Recommended values:
- Development: `"100ms"` (fast feedback)
- Production: `"5s"` (better performance)
- High-traffic: `"10s"` or higher

```go
// Code
.FlushInterval(5 * time.Second)

// YAML
flush-interval: "5s"

// Environment
LOG_FLUSH_INTERVAL=5s
```

### Sampling (Log Volume Control)

#### `enable-sampling` (boolean)
**Default:** `false`  
**Description:** Whether to enable log sampling to reduce volume in high-traffic scenarios.

```go
// Code
.Sampling(true, 100, 1000)

// YAML
enable-sampling: true

// Environment
LOG_ENABLE_SAMPLING=true
```

#### `sample-initial` (integer)
**Default:** `100`  
**Description:** Number of messages to log before sampling begins.

```go
// YAML
sample-initial: 100

// Environment
LOG_SAMPLE_INITIAL=100
```

#### `sample-thereafter` (integer)
**Default:** `100`  
**Description:** After initial messages, log every Nth message.

Example: With `sample-initial: 100` and `sample-thereafter: 1000`:
- First 100 messages: all logged
- After that: every 1000th message logged

```go
// YAML
sample-thereafter: 1000

// Environment
LOG_SAMPLE_THEREAFTER=1000
```

## Environment Presets

### Development Preset

```go
log.DevelopmentPreset()
```

**Configuration:**
```yaml
level: debug
format: console
disable-caller: false
disable-stacktrace: false
disable-split-error: true
max-size: 10
max-backups: 1
compress: false
buffer-size: 512
flush-interval: "100ms"
enable-sampling: false
```

**Use case:** Local development, debugging, immediate feedback

### Production Preset

```go
log.ProductionPreset()
```

**Configuration:**
```yaml
level: info
format: json
disable-caller: true
disable-stacktrace: true
disable-split-error: false
max-size: 100
max-backups: 5
compress: true
buffer-size: 2048
flush-interval: "5s"
enable-sampling: true
sample-initial: 100
sample-thereafter: 1000
```

**Use case:** Production servers, performance-optimized, structured logging

### Testing Preset

```go
log.TestingPreset()
```

**Configuration:**
```yaml
level: debug
format: console
disable-caller: true
disable-stacktrace: true
disable-split-error: true
max-size: 1
max-backups: 1
compress: false
buffer-size: 256
flush-interval: "50ms"
enable-sampling: false
```

**Use case:** Unit tests, integration tests, clean output

## Configuration Priority

When using multiple configuration sources, the priority order is:

1. **Code configuration** (highest priority)
2. **Environment variables**
3. **Configuration file**
4. **Default values** (lowest priority)

Example:
```go
// This will merge: defaults < config.yaml < environment < code overrides
logger := log.NewBuilder().
    Production().                    // Start with production preset
    Level("debug").                  // Override level in code
    Build()
```

## Validation and Error Handling

The library validates all configuration options and provides helpful error messages:

```go
// Invalid configuration
opts := log.NewOptions().WithLevel("invalid")
logger := log.NewLog(opts) // Will use default level and log warning

// File configuration errors
logger, err := log.FromConfigFile("nonexistent.yaml")
if err != nil {
    // Handle configuration error
    log.Printf("Config error: %v", err)
    // Falls back to defaults
    logger = log.Quick()
}
```

## Best Practices

1. **Use presets as starting points**: Begin with a preset and customize as needed
2. **Environment-specific configuration**: Use environment variables for deployment differences
3. **Validate in development**: Test your configuration in development environment
4. **Monitor log volume**: Use sampling in high-traffic production environments
5. **Balance performance vs. visibility**: Adjust buffer sizes and flush intervals based on needs
6. **Plan for rotation**: Set appropriate max-size and max-backups for your storage capacity
7. **Use structured logging**: Configure JSON format for production to enable log analysis tools

## Troubleshooting

### Common Issues

**Logs not appearing immediately:**
- Reduce `flush-interval` or `buffer-size`
- Check if sampling is enabled and reducing log volume

**High memory usage:**
- Reduce `buffer-size`
- Increase `flush-interval` (but logs may be delayed)

**Disk space issues:**
- Reduce `max-size` and `max-backups`
- Enable `compress` for rotated files

**Performance issues:**
- Increase `buffer-size` and `flush-interval`
- Enable `disable-caller` and `disable-stacktrace`
- Use `json` format instead of `console`

**Missing error logs:**
- Check `disable-split-error` setting
- Verify error log files in the same directory as main logs

### Debug Configuration

To see what configuration is being used:

```go
opts, err := log.LoadConfig("config.yaml", "LOG_")
if err != nil {
    log.Printf("Config error: %v", err)
}

// Print effective configuration
fmt.Printf("Level: %s\n", opts.Level)
fmt.Printf("Format: %s\n", opts.Format)
fmt.Printf("Directory: %s\n", opts.Directory)
// ... etc
```