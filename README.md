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

### Advanced Configuration

```yaml
log:
  prefix: "LOG" # Log prefix
  directory: "./logs"      # Log file directory
  level: debug # Log level: debug / info / warn / error / dpanic / panic / fatal
  time-layout: "2006-01-02 15:04:05.000"
  format: console # Log format: console / json

  disable-caller: false # Disable caller: false / true
  disable-stacktrace: false # Disable stacktrace: false / true
  disable-split-error: false # Disable split error: false / true

  # Log rotation settings
  max-size: 100         # Maximum size of log files in megabytes before rotation
  max-backups: 3        # Maximum number of old log files to retain
  compress: false       # Whether to compress rotated log files: false / true
```

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kydenul/log"
	"github.com/spf13/viper"
)

const (
	defaultConfigDir  = "."
	defaultConfigName = "log"

	envPrefix = "LOG"
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get user home directory: %v", err))
	}

	viper.AddConfigPath(filepath.Join(home, defaultConfigDir)) // $HOME/defaultConfigDir
	viper.AddConfigPath(filepath.Join(".", defaultConfigDir))  // ./defaultConfigDir

	viper.SetConfigType("yaml")
	viper.SetConfigName(defaultConfigName)

	// Read matched environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file.
	// If a config file is specified, use it. Otherwise, search in defaultConfigDir.
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Failed to read viper config file: %v", err))
	}

	log.ReplaceLogger(
		log.NewLogger(log.NewOptions().
			WithPrefix(viper.GetString("log.prefix")).
			WithDirectory(viper.GetString("log.directory")).
			WithLevel(viper.GetString("log.level")).
			WithTimeLayout(viper.GetString("log.time-layout")).
			WithFormat(viper.GetString("log.format")).
			WithDisableCaller(viper.GetBool("log.disable-caller")).
			WithDisableStacktrace(viper.GetBool("log.disable-stacktrace")).
			WithDisableSplitError(viper.GetBool("log.disable-split-error")).
			WithMaxSize(viper.GetInt("log.max-size")).
			WithMaxBackups(viper.GetInt("log.max-backups")).
			WithCompress(viper.GetBool("log.compress"))))
}

func main() {
	log.Infoln("This is template project")
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
