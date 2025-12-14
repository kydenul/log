# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A high-performance, structured logging package for Go applications built on uber-go/zap. The library provides dual calling modes (instance methods and global functions), environment presets, builder pattern configuration, and multi-format config file support.

## Common Commands

```bash
# Run all tests
make test
# Or directly:
go test -v ./...

# Run a single test
go test -v -run TestFunctionName ./...

# Format code (requires gofumpt)
make fumpt

# Run linter (requires golangci-lint)
make lint

# Tidy dependencies
make tidy

# Full build (clean, tidy, format, lint)
make build

# Build without linter (for environments without golangci-lint)
make compile
```

## Architecture

### Core Components

- **log.go** - Main logger implementation (`Log` struct) wrapping zap.Logger with custom encoder, file rotation via lumberjack, global logger management via atomic.Value
- **options.go** - Configuration options with validation, defaults, and mapstructure tags for Viper
- **builder.go** - Fluent builder pattern API for logger configuration
- **logger.go** - `Logger` interface definition and filename sanitization utilities
- **internal/log.go** - Base encoder implementation, time layout validation, auto-sync signal handling

### Key Design Patterns

**Dual Calling Modes**: Every logger creation method (`NewLog`, `Quick`, `WithPreset`, `FromConfigFile`, `Builder.Build`) automatically calls `ReplaceLogger()` to set itself as the global default. This enables both `logger.Info()` and `log.Info()` to work with the same configuration.

**Preset System**: `DevelopmentPreset()`, `ProductionPreset()`, `TestingPreset()` provide pre-configured Options via the `Preset.Apply()` method.

**Configuration Loading**: Uses Viper for multi-format support. `LoadFromFile()` auto-detects format by extension (.yaml, .json, .toml). Format-specific functions (`LoadFromYAML`, etc.) are wrappers.

### Subpackages

- **logutil/** - Utility functions: error handling (`LogError`, `FatalOnError`), timing (`Timer`, `TimeFunction`), conditional logging (`InfoIf`), HTTP helpers, lifecycle logging
- **internal/** - Internal encoder and sync utilities (not exported)

### File Naming Convention

Log files follow the pattern `{filename}-{date}.log` with error logs as `{filename}-{date}_error.log`. Filenames are sanitized to remove invalid characters.

## Dependencies

- go.uber.org/zap - Core logging
- gopkg.in/natefinch/lumberjack.v2 - File rotation
- github.com/spf13/viper - Configuration parsing
- github.com/stretchr/testify - Testing
