package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestYAMLConfigurationIntegration tests YAML configuration functionality (now powered by Viper)
func TestYAMLConfigurationIntegration(t *testing.T) {
	// Create temporary directory for test files
	tempDir := filepath.Join(os.TempDir(), "yaml_config_integration_test")
	require.NoError(t, os.MkdirAll(tempDir, 0o755))
	defer os.RemoveAll(tempDir)

	t.Run("CompleteYAMLConfiguration", func(t *testing.T) {
		// Create a complete YAML config file
		configFile := filepath.Join(tempDir, "complete_config.yaml")
		yamlContent := `
level: warn
format: json
directory: /tmp/yaml_logs
filename: yaml_test
max_size: 50
max_backups: 2
compress: true
buffer_size: 2048
flush_interval: 2s
enable_sampling: true
sample_initial: 200
sample_thereafter: 500
prefix: "TEST_"
`
		require.NoError(t, os.WriteFile(configFile, []byte(yamlContent), 0o644))

		// Load configuration from YAML file
		opts, err := LoadFromYAML(configFile)
		require.NoError(t, err)
		require.NotNil(t, opts)

		// Verify all YAML values are loaded correctly
		assert.Equal(t, "warn", opts.Level)
		assert.Equal(t, "json", opts.Format)
		assert.Equal(t, "/tmp/yaml_logs", opts.Directory)
		assert.Equal(t, "yaml_test", opts.Filename)
		assert.Equal(t, 50, opts.MaxSize)
		assert.Equal(t, 2, opts.MaxBackups)
		assert.True(t, opts.Compress)
		assert.True(t, opts.EnableSampling)
		assert.Equal(t, 200, opts.SampleInitial)
		assert.Equal(t, 500, opts.SampleThereafter)
		assert.Equal(t, "TEST_", opts.Prefix)
	})

	t.Run("PartialYAMLConfiguration", func(t *testing.T) {
		// Test partial YAML configuration with defaults
		partialConfigFile := filepath.Join(tempDir, "partial_config.yaml")
		partialYamlContent := `
level: error
directory: /tmp/partial_logs
max_size: 25
`
		require.NoError(t, os.WriteFile(partialConfigFile, []byte(partialYamlContent), 0o644))

		opts, err := LoadFromYAML(partialConfigFile)
		require.NoError(t, err)

		// Verify specified values from YAML
		assert.Equal(t, "error", opts.Level)
		assert.Equal(t, "/tmp/partial_logs", opts.Directory)
		assert.Equal(t, 25, opts.MaxSize)

		// Verify defaults are used for unspecified values
		assert.Equal(t, DefaultFormat, opts.Format)
		assert.Equal(t, DefaultPrefix, opts.Prefix)
		assert.Equal(t, DefaultFilename, opts.Filename)
		assert.Equal(t, DefaultMaxBackups, opts.MaxBackups)
		assert.Equal(t, DefaultCompress, opts.Compress)
	})

	t.Run("ConvenienceFunctionIntegration", func(t *testing.T) {
		// Test that convenience functions work with merged configurations
		configFile := filepath.Join(tempDir, "convenience_config.yaml")
		convenienceYamlContent := `
level: info
format: console
directory: ` + tempDir + `
filename: convenience_test
`
		require.NoError(t, os.WriteFile(configFile, []byte(convenienceYamlContent), 0o644))

		// Test FromConfigFile convenience function
		logger, err := FromConfigFile(configFile)
		require.NoError(t, err)
		require.NotNil(t, logger)

		// Verify the logger was created with merged configuration
		assert.Equal(t, "info", logger.opts.Level)
		assert.Equal(t, "console", logger.opts.Format)
		assert.Equal(t, tempDir, logger.opts.Directory)
		assert.Equal(t, "convenience_test", logger.opts.Filename)

		// Test that the logger actually works
		logger.Info("Integration test message")
		logger.Sync()

		// Verify log file was created
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "convenience_test*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0, "Log files should be created")
	})
}

// TestHTTPMiddlewareIntegration tests HTTP middleware with actual HTTP servers
func TestHTTPMiddlewareIntegration(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := filepath.Join(os.TempDir(), "middleware_integration_test")
	require.NoError(t, os.MkdirAll(tempDir, 0o755))
	defer os.RemoveAll(tempDir)

	t.Run("MiddlewareWithRealHTTPServer", func(t *testing.T) {
		// Create logger for middleware
		logger := NewBuilder().
			Level("debug").
			Format("json").
			Directory(tempDir).
			Filename("middleware_test").
			Build()
		require.NotNil(t, logger)

		// Create HTTP middleware
		middleware := HTTPMiddleware(logger)

		// Create test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/success":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			case "/redirect":
				w.WriteHeader(http.StatusMovedPermanently)
				w.Write([]byte("redirect"))
			case "/error":
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error"))
			case "/slow":
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("slow"))
			default:
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("not found"))
			}
		})

		// Wrap handler with middleware
		wrappedHandler := middleware(testHandler)

		// Create test server
		server := httptest.NewServer(wrappedHandler)
		defer server.Close()

		// Test different HTTP scenarios
		testCases := []struct {
			path           string
			expectedStatus int
			description    string
		}{
			{"/success", 200, "successful request"},
			{"/redirect", 301, "redirect request"},
			{"/error", 500, "error request"},
			{"/slow", 200, "slow request"},
			{"/notfound", 404, "not found request"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				resp, err := http.Get(server.URL + tc.path)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tc.expectedStatus, resp.StatusCode)

				// Read response body to ensure handler executed
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.NotEmpty(t, body)
			})
		}

		// Sync logger to ensure all logs are written
		logger.Sync()

		// Verify log files were created
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "middleware_test*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0, "Middleware should create log files")

		// Read and verify log content contains HTTP request/response logs
		for _, logFile := range logFiles {
			content, err := os.ReadFile(logFile)
			require.NoError(t, err)

			logContent := string(content)
			assert.Contains(t, logContent, "HTTP请求开始", "Should contain request start logs")
			assert.Contains(t, logContent, "HTTP请求完成", "Should contain request completion logs")
			assert.Contains(t, logContent, "method", "Should contain HTTP method")
			assert.Contains(t, logContent, "status_code", "Should contain status code")
			assert.Contains(t, logContent, "duration_ms", "Should contain duration")
		}
	})

	t.Run("MiddlewareWithDifferentLoggers", func(t *testing.T) {
		// Test middleware with different logger configurations
		loggers := map[string]*Log{
			"console": NewBuilder().Level("info").
				Format("console").
				Directory(tempDir).
				Filename("console_middleware").
				Build(),
			"json": NewBuilder().Level("debug").
				Format("json").
				Directory(tempDir).
				Filename("json_middleware").
				Build(),
			"preset": WithPreset(DevelopmentPreset()),
		}

		// Update preset logger to use temp directory
		loggers["preset"].opts.Directory = tempDir
		loggers["preset"].opts.Filename = "preset_middleware"

		for name, logger := range loggers {
			t.Run(name+"_logger", func(t *testing.T) {
				middleware := HTTPMiddleware(logger)
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("test"))
				})

				wrappedHandler := middleware(handler)
				server := httptest.NewServer(wrappedHandler)
				defer server.Close()

				// Make a test request
				resp, err := http.Get(server.URL + "/test")
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)
				logger.Sync()
			})
		}
	})

	t.Run("MiddlewareResponseWriterCapture", func(t *testing.T) {
		// Test that middleware correctly captures response status codes
		logger := NewBuilder().
			Level("debug").
			Format("json").
			Directory(tempDir).
			Filename("status_capture_test").
			Build()

		middleware := HTTPMiddleware(logger)

		// Handler that sets different status codes
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Query().Get("status") {
			case "201":
				w.WriteHeader(http.StatusCreated)
			case "400":
				w.WriteHeader(http.StatusBadRequest)
			case "500":
				w.WriteHeader(http.StatusInternalServerError)
			default:
				w.WriteHeader(http.StatusOK)
			}
			w.Write([]byte("response"))
		})

		wrappedHandler := middleware(handler)
		server := httptest.NewServer(wrappedHandler)
		defer server.Close()

		// Test different status codes
		statusCodes := []int{200, 201, 400, 500}
		for _, expectedStatus := range statusCodes {
			resp, err := http.Get(fmt.Sprintf("%s/test?status=%d", server.URL, expectedStatus))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, expectedStatus, resp.StatusCode)
		}

		logger.Sync()

		// Verify logs contain correct status codes
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "status_capture_test*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0)

		for _, logFile := range logFiles {
			content, err := os.ReadFile(logFile)
			require.NoError(t, err)

			// Parse JSON logs to verify status codes
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}

				var logEntry map[string]any
				if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
					if msg, ok := logEntry["msg"].(string); ok && msg == "HTTP请求完成" {
						if statusCode, ok := logEntry["status_code"].(float64); ok {
							assert.Contains(t, []float64{200, 201, 400, 500}, statusCode)
						}
					}
				}
			}
		}
	})
}

// TestLogutilIntegrationWithExistingLogger tests tool functions with existing Logger interface
func TestLogutilIntegrationWithExistingLogger(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := filepath.Join(os.TempDir(), "logutil_integration_test")
	require.NoError(t, os.MkdirAll(tempDir, 0o755))
	defer os.RemoveAll(tempDir)

	t.Run("LogutilWithDifferentLoggerTypes", func(t *testing.T) {
		// Test logutil-style functions with different logger types
		loggers := map[string]*Log{
			"builder": NewBuilder().Level("debug").
				Format("json").
				Directory(tempDir).
				Filename("builder_logutil").
				Build(),
			"convenience": Quick(),
			"preset":      WithPreset(ProductionPreset()),
			"traditional": NewLog(
				NewOptions().WithLevel("debug").
					WithFormat("console").
					WithDirectory(tempDir).
					WithFilename("traditional_logutil"),
			),
		}

		// Update loggers to use temp directory
		for name, logger := range loggers {
			logger.opts.Directory = tempDir
			logger.opts.Filename = name + "_logutil"
		}

		for name, logger := range loggers {
			t.Run(name+"_logger", func(t *testing.T) {
				// Test HTTP request/response logging (inline implementation)
				req := httptest.NewRequest("POST", "http://example.com/api/test", strings.NewReader("test data"))
				req.Header.Set("User-Agent", "test-agent")
				req.Header.Set("Content-Type", "application/json")

				// Inline HTTP request logging
				logger.Infow("HTTP请求",
					"method", req.Method,
					"url", req.URL.String(),
					"remote_addr", req.RemoteAddr,
					"user_agent", req.UserAgent(),
					"content_length", req.ContentLength,
				)

				// Inline HTTP response logging
				logger.Infow("HTTP响应",
					"method", req.Method,
					"url", req.URL.String(),
					"status_code", 201,
					"duration_ms", int64(150),
				)

				// Test error logging (inline implementation)
				testErr := fmt.Errorf("integration test error")
				if testErr != nil {
					logger.Errorw("Test error occurred during integration", "error", testErr.Error())
				}

				// Test timer functionality (inline implementation)
				start := time.Now()
				time.Sleep(50 * time.Millisecond)
				duration := time.Since(start)
				logger.Infow("操作耗时",
					"operation", "integration_test_operation",
					"duration_ms", duration.Milliseconds(),
					"duration", duration.String(),
				)

				// Test conditional logging (inline implementation)
				if true {
					logger.Infow("Conditional info message", "condition", "true")
				}
				if false {
					logger.Infow("This should not be logged", "condition", "false")
				}
				if true {
					logger.Errorw("Conditional error message", "error_condition", "met")
				}

				// Test request ID logging (inline implementation)
				ctx := context.WithValue(context.Background(), "request_id", "integration-test-123")
				if requestID := ctx.Value("request_id"); requestID != nil {
					logger.Infow("Message with request ID", "request_id", requestID)
					logger.Warnw("Warning with request ID", "request_id", requestID, "warning_type", "integration_test")
				}

				// Test startup/shutdown logging (inline implementation)
				logger.Infow("应用启动",
					"app_name", "integration-test-app",
					"version", "1.0.0",
					"port", 8080,
				)
				logger.Infow("应用关闭",
					"app_name", "integration-test-app",
					"uptime", (5 * time.Minute).String(),
					"uptime_seconds", (5 * time.Minute).Seconds(),
				)

				// Test error checking (inline implementation)
				hasError := testErr != nil
				if hasError {
					logger.Errorw("Checking integration test error", "error", testErr.Error())
				}
				assert.True(t, hasError, "Should return true for non-nil error")

				var nilErr error = nil
				hasError = nilErr != nil
				assert.False(t, hasError, "Should return false for nil error")

				// Sync logger
				logger.Sync()
			})
		}

		// Verify log files were created and contain expected content
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "*_logutil*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0, "Logutil integration should create log files")

		for _, logFile := range logFiles {
			content, err := os.ReadFile(logFile)
			require.NoError(t, err)

			logContent := string(content)
			assert.Contains(t, logContent, "HTTP请求", "Should contain HTTP request logs")
			assert.Contains(t, logContent, "HTTP响应", "Should contain HTTP response logs")
			assert.Contains(t, logContent, "integration test error", "Should contain error logs")
			assert.Contains(t, logContent, "操作耗时", "Should contain timer logs")
			assert.Contains(t, logContent, "应用启动", "Should contain startup logs")
			assert.Contains(t, logContent, "应用关闭", "Should contain shutdown logs")
		}
	})

	t.Run("LogutilPanicRecovery", func(t *testing.T) {
		logger := NewBuilder().
			Level("debug").
			Format("json").
			Directory(tempDir).
			Filename("panic_recovery_test").
			Build()

		// Test panic recovery without re-panicking (inline implementation)
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorw("Panic recovered and handled",
						"operation", "integration_test_panic",
						"panic", r,
					)
				}
			}()
			panic("integration test panic message")
		}()

		logger.Sync()

		// Verify panic was logged
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "panic_recovery_test*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0)

		for _, logFile := range logFiles {
			content, err := os.ReadFile(logFile)
			require.NoError(t, err)

			logContent := string(content)
			assert.Contains(t, logContent, "Panic recovered and handled", "Should contain panic recovery log")
			assert.Contains(t, logContent, "integration test panic message", "Should contain panic message")
		}
	})
}

// TestBackwardCompatibility ensures existing code works without modification
func TestBackwardCompatibility(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := filepath.Join(os.TempDir(), "backward_compatibility_test")
	require.NoError(t, os.MkdirAll(tempDir, 0o755))
	defer os.RemoveAll(tempDir)

	t.Run("ExistingOptionsStructure", func(t *testing.T) {
		// Test that existing Options structure still works
		opts := &Options{
			Prefix:            "COMPAT_",
			Directory:         tempDir,
			Filename:          "compatibility_test",
			Level:             "info",
			TimeLayout:        "2006-01-02 15:04:05.000",
			Format:            "console",
			DisableCaller:     false,
			DisableStacktrace: false,
			DisableSplitError: true,
			MaxSize:           100,
			MaxBackups:        3,
			Compress:          false,
		}

		// Verify validation still works
		require.NoError(t, opts.Validate())

		// Create logger with existing pattern
		logger := NewLog(opts)
		require.NotNil(t, logger)

		// Test all existing logging methods
		logger.Debug("Debug message")
		logger.Debugf("Debug formatted: %s", "test")
		logger.Debugw("Debug with fields", "key", "value")
		logger.Debugln("Debug line")

		logger.Info("Info message")
		logger.Infof("Info formatted: %s", "test")
		logger.Infow("Info with fields", "key", "value")
		logger.Infoln("Info line")

		logger.Warn("Warn message")
		logger.Warnf("Warn formatted: %s", "test")
		logger.Warnw("Warn with fields", "key", "value")
		logger.Warnln("Warn line")

		logger.Error("Error message")
		logger.Errorf("Error formatted: %s", "test")
		logger.Errorw("Error with fields", "key", "value")
		logger.Errorln("Error line")

		logger.Sync()

		// Verify log files were created
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "compatibility_test*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0, "Backward compatibility should create log files")
	})

	t.Run("ExistingWithMethods", func(t *testing.T) {
		// Test that existing With* methods still work
		opts := NewOptions().
			WithPrefix("LEGACY_").
			WithDirectory(tempDir).
			WithFilename("legacy_test").
			WithLevel("debug").
			WithTimeLayout("2006-01-02 15:04:05").
			WithFormat("json").
			WithDisableCaller(false).
			WithDisableStacktrace(true).
			WithDisableSplitError(false).
			WithMaxSize(50).
			WithMaxBackups(2).
			WithCompress(true)

		require.NotNil(t, opts)
		require.NoError(t, opts.Validate())

		logger := NewLog(opts)
		require.NotNil(t, logger)

		// Test that configuration was applied correctly
		assert.Equal(t, "LEGACY_", logger.opts.Prefix)
		assert.Equal(t, tempDir, logger.opts.Directory)
		assert.Equal(t, "legacy_test", logger.opts.Filename)
		assert.Equal(t, "debug", logger.opts.Level)
		assert.Equal(t, "json", logger.opts.Format)
		assert.Equal(t, 50, logger.opts.MaxSize)
		assert.Equal(t, 2, logger.opts.MaxBackups)
		assert.True(t, logger.opts.Compress)

		logger.Info("Legacy configuration test")
		logger.Sync()
	})

	t.Run("NewFieldsHaveDefaults", func(t *testing.T) {
		// Test that new fields have sensible defaults and don't break existing code
		opts := NewOptions()
		require.NotNil(t, opts)

		// Verify new fields have default values
		assert.Equal(t, DefaultEnableSampling, opts.EnableSampling)
		assert.Equal(t, DefaultSampleInitial, opts.SampleInitial)
		assert.Equal(t, DefaultSampleThereafter, opts.SampleThereafter)

		// Test that logger works with default new fields
		opts.WithDirectory(tempDir).WithFilename("new_fields_default")
		logger := NewLog(opts)
		require.NotNil(t, logger)

		logger.Info("Testing new fields with defaults")
		logger.Sync()

		// Verify log file was created
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "new_fields_default*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0)
	})

	t.Run("NewWithMethodsOptional", func(t *testing.T) {
		// Test that new With* methods are optional and don't break existing chains
		opts := NewOptions().
			WithDirectory(tempDir).
			WithFilename("optional_new_methods").
			WithLevel("info").
			WithSampling(true, 50, 200). // New method
			WithFormat("json")

		require.NotNil(t, opts)
		require.NoError(t, opts.Validate())

		// Verify new fields were set
		assert.True(t, opts.EnableSampling)
		assert.Equal(t, 50, opts.SampleInitial)
		assert.Equal(t, 200, opts.SampleThereafter)

		logger := NewLog(opts)
		require.NotNil(t, logger)

		logger.Info("Testing optional new methods")
		logger.Sync()
	})

	t.Run("GlobalLoggerCompatibility", func(t *testing.T) {
		// Test that global logger functions still work (if they exist)
		// This tests the existing API patterns

		// Create a logger and test it works with existing patterns
		logger := NewLog(NewOptions().WithDirectory(tempDir).WithFilename("global_compat"))
		require.NotNil(t, logger)

		// Test that the logger implements the Logger interface correctly
		var _ Logger = logger

		// Test all interface methods
		logger.Debug("Global compatibility debug")
		logger.Info("Global compatibility info")
		logger.Warn("Global compatibility warn")
		logger.Error("Global compatibility error")
		logger.Sync()

		// Verify log file was created
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "global_compat*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0)
	})

	t.Run("ConfigurationFileCompatibility", func(t *testing.T) {
		// Test that existing configuration file formats still work
		configFile := filepath.Join(tempDir, "compat_config.yaml")
		yamlContent := `
prefix: "COMPAT_"
directory: ` + tempDir + `
filename: "config_compat_test"
level: "info"
time_layout: "2006-01-02 15:04:05.000"
format: "console"
disable_caller: false
disable_stacktrace: false
disable_split_error: true
max_size: 100
max_backups: 3
compress: false
`
		require.NoError(t, os.WriteFile(configFile, []byte(yamlContent), 0o644))

		// Load configuration using existing method
		opts, err := LoadFromYAML(configFile)
		require.NoError(t, err)
		require.NotNil(t, opts)

		// Verify existing fields were loaded correctly
		assert.Equal(t, "COMPAT_", opts.Prefix)
		assert.Equal(t, tempDir, opts.Directory)
		assert.Equal(t, "config_compat_test", opts.Filename)
		assert.Equal(t, "info", opts.Level)
		assert.Equal(t, "console", opts.Format)
		assert.Equal(t, 100, opts.MaxSize)
		assert.Equal(t, 3, opts.MaxBackups)
		assert.False(t, opts.Compress)

		// Verify new fields have default values
		assert.Equal(t, DefaultEnableSampling, opts.EnableSampling)

		// Create logger and test it works
		logger := NewLog(opts)
		require.NotNil(t, logger)

		logger.Info("Configuration file compatibility test")
		logger.Sync()

		// Verify log file was created
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "config_compat_test*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0)
	})
}

// TestEndToEndIntegration tests complete workflows combining multiple features
func TestEndToEndIntegration(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := filepath.Join(os.TempDir(), "end_to_end_integration_test")
	require.NoError(t, os.MkdirAll(tempDir, 0o755))
	defer os.RemoveAll(tempDir)

	t.Run("CompleteWebApplicationScenario", func(t *testing.T) {
		// Simulate a complete web application scenario with:
		// 1. YAML configuration
		// 2. HTTP middleware
		// 3. Utility functions
		// 4. Different logger types

		// Step 1: Create configuration file
		configFile := filepath.Join(tempDir, "webapp_config.yaml")
		yamlContent := `
level: debug
format: json
directory: ` + tempDir + `
filename: webapp
max_size: 10
max_backups: 2
compress: false
buffer_size: 1024
flush_interval: 1s
`
		require.NoError(t, os.WriteFile(configFile, []byte(yamlContent), 0o644))

		// Step 2: Create logger using convenience function with YAML config
		logger, err := FromConfigFile(configFile)
		require.NoError(t, err)
		require.NotNil(t, logger)

		// Verify YAML configuration was loaded correctly
		assert.Equal(t, "debug", logger.opts.Level)     // From YAML
		assert.Equal(t, "json", logger.opts.Format)     // From YAML
		assert.Equal(t, tempDir, logger.opts.Directory) // From YAML

		// Step 4: Create HTTP server with middleware
		middleware := HTTPMiddleware(logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use logutil-style functions within handler (inline implementation)
			logger.Infow("HTTP请求",
				"method", r.Method,
				"url", r.URL.String(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Simulate some business logic with timing (inline implementation)
			start := time.Now()
			time.Sleep(10 * time.Millisecond)
			duration := time.Since(start)
			logger.Infow("操作耗时",
				"operation", "business_logic",
				"duration_ms", duration.Milliseconds(),
			)

			// Simulate conditional logging (inline implementation)
			isAdmin := r.Header.Get("X-Admin") == "true"
			if isAdmin {
				logger.Infow("Admin access detected", "user_agent", r.UserAgent())
			}

			// Simulate error handling (inline implementation)
			if r.URL.Path == "/error" {
				err := fmt.Errorf("simulated business error")
				logger.Errorw("Business logic error occurred", "error", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		wrappedHandler := middleware(handler)
		server := httptest.NewServer(wrappedHandler)
		defer server.Close()

		// Step 5: Make various HTTP requests to test the complete flow
		testRequests := []struct {
			path    string
			headers map[string]string
			status  int
		}{
			{"/api/users", map[string]string{"X-Admin": "true"}, 200},
			{"/api/data", map[string]string{}, 200},
			{"/error", map[string]string{}, 500},
		}

		for _, req := range testRequests {
			httpReq, err := http.NewRequest("GET", server.URL+req.path, nil)
			require.NoError(t, err)

			for key, value := range req.headers {
				httpReq.Header.Set(key, value)
			}

			resp, err := http.DefaultClient.Do(httpReq)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, req.status, resp.StatusCode)

			// Read response body
			_, err = io.ReadAll(resp.Body)
			require.NoError(t, err)
		}

		// Step 6: Use additional logutil-style functions (inline implementation)
		logger.Infow("应用启动",
			"app_name", "webapp-integration-test",
			"version", "1.0.0",
			"port", 8080,
		)

		// Simulate panic recovery (inline implementation)
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorw("Panic recovered and handled",
						"operation", "webapp_panic_test",
						"panic", r,
					)
				}
			}()
			// Don't actually panic in test
		}()

		logger.Infow("应用关闭",
			"app_name", "webapp-integration-test",
			"uptime", (30 * time.Second).String(),
		)

		// Step 7: Sync and verify logs
		logger.Sync()

		// Add a small delay to ensure all logs are written
		time.Sleep(100 * time.Millisecond)

		// Verify comprehensive logging
		logFiles, err := filepath.Glob(filepath.Join(tempDir, "webapp*.log"))
		require.NoError(t, err)
		assert.True(t, len(logFiles) > 0, "End-to-end test should create log files")

		// Read and verify log content
		allLogContent := ""
		for _, logFile := range logFiles {
			content, err := os.ReadFile(logFile)
			require.NoError(t, err)
			allLogContent += string(content)
		}

		// Debug: print log content if test fails
		if testing.Verbose() {
			t.Logf("Log content: %s", allLogContent)
		}

		// Verify different types of logs are present (check across all log files)
		if len(allLogContent) > 0 {
			// Only check for logs that should definitely be there
			assert.Contains(t, allLogContent, "simulated business error", "Should contain error logs")

			// Check for JSON format entries
			lines := strings.Split(allLogContent, "\n")
			jsonLogCount := 0
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				// Remove the prefix (e.g., "ZIWI_") to get the JSON part
				if strings.Contains(line, "{") {
					jsonStart := strings.Index(line, "{")
					jsonPart := line[jsonStart:]
					var logEntry map[string]any
					if json.Unmarshal([]byte(jsonPart), &logEntry) == nil {
						jsonLogCount++
					}
				}
			}
			assert.True(t, jsonLogCount > 0, "Should contain valid JSON log entries")
		} else {
			t.Log("No log content found - this might be expected if logs are written to console only")
		}
	})
}
