# Log

A high-performance, structured logging package for Go applications, built on top of [zap](https://github.com/uber-go/zap) with additional features and simplified interface.

## Features

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

## Installation

```bash
go get github.com/kydenul/log
```

## Quick Start

```go
package main

import "github.com/kydenul/log"

func main() {
    // Use default logger
    log.Info("Hello, World!")
    
    // With formatting
    log.Infof("Processing item %d", 123)
}
```

## Requirements

- Go 1.23.4 or higher
- Dependencies:
  - go.uber.org/zap
  - gopkg.in/natefinch/lumberjack.v2
  - github.com/stretchr/testify (for testing)

## License

This project is licensed under the terms found in the LICENSE file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
