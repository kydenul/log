# Best Practices Guide

Guidelines and recommendations for using the log library effectively in production applications.

## Table of Contents

- [General Principles](#general-principles)
- [Configuration Best Practices](#configuration-best-practices)
- [Logging Patterns](#logging-patterns)
- [Performance Optimization](#performance-optimization)
- [Security Considerations](#security-considerations)
- [Monitoring and Alerting](#monitoring-and-alerting)
- [Testing with Logs](#testing-with-logs)
- [Common Pitfalls](#common-pitfalls)

## General Principles

### 1. Use Appropriate Log Levels

Choose log levels based on the importance and audience of the message:

```go
// ✅ Good: Appropriate log levels
logger.Debug("Cache miss for key", "key", cacheKey)           // Development info
logger.Info("User logged in", "user_id", userID)             // Business events
logger.Warn("Rate limit approaching", "current", 95, "max", 100) // Potential issues
logger.Error("Database connection failed", "error", err)      // Actual problems
logger.Fatal("Failed to start server", "error", err)         // Unrecoverable errors

// ❌ Bad: Wrong log levels
logger.Error("User clicked button")                           // Not an error
logger.Info("Database connection failed", "error", err)       // This is an error
logger.Debug("Payment processed", "amount", 1000)             // Important business event
```

### 2. Use Structured Logging

Always prefer structured logging over string formatting:

```go
// ✅ Good: Structured logging
logger.Infow("User action completed",
    "user_id", userID,
    "action", "purchase",
    "amount", amount,
    "duration_ms", duration.Milliseconds(),
)

// ❌ Bad: String formatting
logger.Infof("User %d completed purchase of $%.2f in %dms", 
    userID, amount, duration.Milliseconds())
```

**Benefits of structured logging:**
- Machine-readable for log analysis tools
- Consistent field names across the application
- Better performance (no string formatting)
- Easier to search and filter

### 3. Include Relevant Context

Always include enough context to understand and debug issues:

```go
// ✅ Good: Rich context
logger.Errorw("Payment processing failed",
    "user_id", userID,
    "order_id", orderID,
    "payment_method", paymentMethod,
    "amount", amount,
    "currency", currency,
    "error", err,
    "retry_count", retryCount,
)

// ❌ Bad: Insufficient context
logger.Error("Payment failed", "error", err)
```

### 4. Use Consistent Field Names

Establish and follow naming conventions across your application:

```go
// ✅ Good: Consistent naming
logger.Infow("Request started", "request_id", reqID, "user_id", userID)
logger.Infow("Database query", "request_id", reqID, "user_id", userID)
logger.Infow("Request completed", "request_id", reqID, "user_id", userID)

// ❌ Bad: Inconsistent naming
logger.Infow("Request started", "req_id", reqID, "userId", userID)
logger.Infow("Database query", "requestId", reqID, "user", userID)
logger.Infow("Request completed", "request_id", reqID, "uid", userID)
```

**Recommended field names:**
- `user_id` (not `userId`, `uid`, `user`)
- `request_id` (not `req_id`, `requestId`)
- `duration_ms` (not `duration`, `time_taken`)
- `error` (for error messages)
- `status_code` (for HTTP status codes)

## Configuration Best Practices

### 1. Use Environment-Specific Presets

Start with presets and customize as needed:

```go
// ✅ Good: Environment-specific configuration
func createLogger() *log.Log {
    env := os.Getenv("ENVIRONMENT")
    
    switch env {
    case "development":
        return log.NewBuilder().
            Development().
            Level("debug").
            Build()
    case "production":
        return log.NewBuilder().
            Production().
            Directory("/var/log/myapp").
            MaxSize(200).
            MaxBackups(10).
            Build()
    case "testing":
        return log.WithPreset(log.TestingPreset())
    default:
        return log.Quick()
    }
}
```

### 2. Configure for Your Scale

Adjust settings based on your application's traffic and requirements:

```go
// High-traffic production service
logger := log.NewBuilder().
    Production().
    BufferSize(8192).           // Larger buffer for performance
    FlushInterval(10*time.Second). // Less frequent flushes
    Sampling(true, 100, 5000).  // Aggressive sampling
    MaxSize(500).               // Larger files
    MaxBackups(20).             // More backups
    Build()

// Low-traffic service
logger := log.NewBuilder().
    Production().
    BufferSize(1024).           // Smaller buffer
    FlushInterval(time.Second). // More frequent flushes
    Sampling(false, 0, 0).      // No sampling needed
    Build()
```

### 3. Use Configuration Files for Complex Setups

For complex configurations, use YAML files:

```yaml
# config/production.yaml
level: info
format: json
directory: /var/log/myapp
filename: service
max-size: 200
max-backups: 10
compress: true
buffer-size: 4096
flush-interval: "5s"
enable-sampling: true
sample-initial: 100
sample-thereafter: 2000
```

```go
logger, err := log.FromConfigFile("config/production.yaml")
if err != nil {
    // Fallback to preset
    logger = log.WithPreset(log.ProductionPreset())
}
```

## Logging Patterns

### 1. Request Lifecycle Logging

Log the complete lifecycle of requests:

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    requestID := generateRequestID()
    start := time.Now()
    
    logger := log.Quick()
    
    // Log request start
    logger.Infow("Request started",
        "request_id", requestID,
        "method", r.Method,
        "path", r.URL.Path,
        "user_agent", r.UserAgent(),
        "ip", r.RemoteAddr,
    )
    
    // Process request
    result, err := processRequest(r, requestID, logger)
    
    // Log request completion
    duration := time.Since(start)
    if err != nil {
        logger.Errorw("Request failed",
            "request_id", requestID,
            "duration_ms", duration.Milliseconds(),
            "error", err,
        )
        http.Error(w, "Internal Server Error", 500)
        return
    }
    
    logger.Infow("Request completed",
        "request_id", requestID,
        "duration_ms", duration.Milliseconds(),
        "status", "success",
    )
    
    json.NewEncoder(w).Encode(result)
}
```

### 2. Database Operation Logging

Log database operations with timing and context:

```go
func (s *UserService) GetUser(userID int) (*User, error) {
    logger := s.logger
    
    defer logutil.Timer(logger, "get_user_db")()
    
    logger.Debugw("Fetching user from database", "user_id", userID)
    
    var user User
    err := s.db.QueryRow("SELECT * FROM users WHERE id = $1", userID).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        if err == sql.ErrNoRows {
            logger.Warnw("User not found", "user_id", userID)
            return nil, ErrUserNotFound
        }
        logger.Errorw("Database query failed", 
            "user_id", userID, 
            "query", "SELECT * FROM users WHERE id = $1",
            "error", err,
        )
        return nil, err
    }
    
    logger.Infow("User retrieved successfully", 
        "user_id", userID,
        "user_name", user.Name,
    )
    
    return &user, nil
}
```

### 3. External Service Call Logging

Log external service interactions:

```go
func (c *PaymentClient) ProcessPayment(payment *Payment) error {
    logger := c.logger
    
    defer logutil.Timer(logger, "payment_service_call")()
    
    logger.Infow("Calling payment service",
        "payment_id", payment.ID,
        "amount", payment.Amount,
        "currency", payment.Currency,
        "service_url", c.baseURL,
    )
    
    start := time.Now()
    resp, err := c.httpClient.Post(c.baseURL+"/payments", "application/json", paymentData)
    duration := time.Since(start)
    
    if err != nil {
        logger.Errorw("Payment service call failed",
            "payment_id", payment.ID,
            "duration_ms", duration.Milliseconds(),
            "error", err,
        )
        return err
    }
    defer resp.Body.Close()
    
    logger.Infow("Payment service call completed",
        "payment_id", payment.ID,
        "status_code", resp.StatusCode,
        "duration_ms", duration.Milliseconds(),
    )
    
    if resp.StatusCode != 200 {
        logger.Warnw("Payment service returned non-200 status",
            "payment_id", payment.ID,
            "status_code", resp.StatusCode,
        )
        return ErrPaymentFailed
    }
    
    return nil
}
```

### 4. Error Handling Patterns

Implement consistent error handling with logging:

```go
// ✅ Good: Consistent error handling
func processOrder(orderID int) error {
    logger := log.Quick()
    
    // Log operation start
    logger.Infow("Processing order", "order_id", orderID)
    
    // Validate order
    order, err := getOrder(orderID)
    if err != nil {
        logger.Errorw("Failed to get order", "order_id", orderID, "error", err)
        return fmt.Errorf("order validation failed: %w", err)
    }
    
    // Process payment
    if err := processPayment(order.PaymentID); err != nil {
        logger.Errorw("Payment processing failed", 
            "order_id", orderID, 
            "payment_id", order.PaymentID, 
            "error", err,
        )
        return fmt.Errorf("payment processing failed: %w", err)
    }
    
    // Update order status
    if err := updateOrderStatus(orderID, "completed"); err != nil {
        logger.Errorw("Failed to update order status", 
            "order_id", orderID, 
            "error", err,
        )
        // This might not be fatal - log but continue
    }
    
    logger.Infow("Order processed successfully", "order_id", orderID)
    return nil
}
```

## Performance Optimization

### 1. Use Appropriate Buffer Sizes

Configure buffer sizes based on your traffic patterns:

```go
// High-traffic service
logger := log.NewBuilder().
    BufferSize(8192).           // 8KB buffer
    FlushInterval(10*time.Second). // Flush every 10 seconds
    Build()

// Low-latency service (need immediate logs)
logger := log.NewBuilder().
    BufferSize(512).            // Small buffer
    FlushInterval(100*time.Millisecond). // Flush frequently
    Build()
```

### 2. Enable Sampling for High-Volume Logs

Use sampling to reduce log volume in high-traffic scenarios:

```go
// Enable sampling for high-traffic endpoints
logger := log.NewBuilder().
    Production().
    Sampling(true, 100, 1000). // Log first 100, then every 1000th
    Build()

// Disable sampling for critical operations
logger := log.NewBuilder().
    Production().
    Sampling(false, 0, 0).     // Log everything
    Build()
```

### 3. Avoid Expensive Operations in Log Calls

Don't perform expensive operations in log statements:

```go
// ❌ Bad: Expensive operation in log call
logger.Infow("User data", "user", expensiveUserDataSerialization(user))

// ✅ Good: Only log when necessary
if logger.Level() <= log.DebugLevel {
    userData := expensiveUserDataSerialization(user)
    logger.Debugw("User data", "user", userData)
}

// ✅ Better: Use lazy evaluation
logger.Debugw("User data", "user", func() interface{} {
    return expensiveUserDataSerialization(user)
})
```

### 4. Use Conditional Logging Utilities

Use the logutil conditional functions to avoid unnecessary work:

```go
// ✅ Good: Conditional logging
debugMode := os.Getenv("DEBUG") == "true"
logutil.InfoIf(logger, debugMode, "Debug information", "data", expensiveData)

// ✅ Good: Error checking with logging
if logutil.CheckError(logger, err, "Operation failed") {
    return err
}
```

## Security Considerations

### 1. Sanitize Sensitive Data

Never log sensitive information:

```go
// ❌ Bad: Logging sensitive data
logger.Infow("User login", 
    "username", username,
    "password", password,        // Never log passwords
    "credit_card", creditCard,   // Never log credit cards
    "ssn", ssn,                 // Never log SSNs
)

// ✅ Good: Sanitized logging
logger.Infow("User login",
    "username", username,
    "password_length", len(password),
    "credit_card_last4", creditCard[len(creditCard)-4:],
    "has_ssn", ssn != "",
)
```

### 2. Use Structured Logging for Audit Trails

Create consistent audit logs for security events:

```go
func logSecurityEvent(eventType, userID, details string, success bool) {
    logger := log.Quick()
    
    level := "info"
    if !success {
        level = "warn"
    }
    
    fields := []interface{}{
        "event_type", eventType,
        "user_id", userID,
        "success", success,
        "timestamp", time.Now().UTC(),
        "details", details,
    }
    
    if level == "warn" {
        logger.Warnw("Security event", fields...)
    } else {
        logger.Infow("Security event", fields...)
    }
}

// Usage
logSecurityEvent("login_attempt", userID, "successful login", true)
logSecurityEvent("password_change", userID, "password changed", true)
logSecurityEvent("failed_login", userID, "invalid password", false)
```

### 3. Implement Log Rotation and Retention

Configure appropriate log rotation and retention policies:

```go
logger := log.NewBuilder().
    MaxSize(100).        // 100MB per file
    MaxBackups(30).      // Keep 30 backup files
    Compress(true).      // Compress old files
    Build()
```

## Monitoring and Alerting

### 1. Log Metrics and Health Indicators

Log key metrics for monitoring:

```go
func logHealthMetrics(service *Service) {
    logger := service.logger
    
    metrics := service.GetMetrics()
    
    logger.Infow("Health metrics",
        "active_connections", metrics.ActiveConnections,
        "requests_per_second", metrics.RequestsPerSecond,
        "error_rate", metrics.ErrorRate,
        "memory_usage_mb", metrics.MemoryUsageMB,
        "cpu_usage_percent", metrics.CPUUsagePercent,
    )
}

// Log this periodically
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        logHealthMetrics(service)
    }
}()
```

### 2. Create Alertable Log Patterns

Use consistent patterns that monitoring systems can alert on:

```go
// ✅ Good: Consistent error patterns
logger.Errorw("ALERT: Database connection pool exhausted",
    "alert_type", "database_pool_exhausted",
    "active_connections", activeConns,
    "max_connections", maxConns,
)

logger.Errorw("ALERT: High error rate detected",
    "alert_type", "high_error_rate",
    "error_rate", errorRate,
    "threshold", threshold,
    "time_window", "5m",
)
```

### 3. Log Application Lifecycle Events

Log important application events:

```go
func main() {
    logger := log.WithPreset(log.ProductionPreset())
    
    // Log startup
    logutil.LogStartup(logger, "my-service", version, port)
    
    // Log configuration
    logger.Infow("Service configuration",
        "environment", environment,
        "database_url", sanitizeURL(databaseURL),
        "cache_enabled", cacheEnabled,
        "worker_count", workerCount,
    )
    
    // Setup graceful shutdown logging
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        logger.Warn("Shutdown signal received")
        
        // Perform graceful shutdown
        shutdownStart := time.Now()
        performGracefulShutdown()
        shutdownDuration := time.Since(shutdownStart)
        
        logutil.LogShutdown(logger, "my-service", shutdownDuration)
        os.Exit(0)
    }()
    
    // Start service
    logger.Info("Service started successfully")
    startServer()
}
```

## Testing with Logs

### 1. Use Testing Preset

Use the testing preset for clean test output:

```go
func TestUserService(t *testing.T) {
    logger := log.WithPreset(log.TestingPreset())
    service := NewUserService(mockDB, logger)
    
    // Test with logging
    user, err := service.GetUser(123)
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

### 2. Test Log Output

Test that important events are logged:

```go
func TestOrderProcessingLogsEvents(t *testing.T) {
    // Create a test logger that captures output
    var logOutput bytes.Buffer
    logger := log.NewBuilder().
        Testing().
        // Configure to write to buffer for testing
        Build()
    
    service := NewOrderService(logger)
    
    // Process order
    err := service.ProcessOrder(orderID)
    assert.NoError(t, err)
    
    // Verify log output contains expected events
    logContent := logOutput.String()
    assert.Contains(t, logContent, "Processing order")
    assert.Contains(t, logContent, "Order processed successfully")
}
```

### 3. Use Conditional Logging in Tests

Use conditional logging to control test verbosity:

```go
func TestComplexOperation(t *testing.T) {
    logger := log.WithPreset(log.TestingPreset())
    
    // Only log debug info if test is run with -v flag
    verbose := testing.Verbose()
    
    logutil.InfoIf(logger, verbose, "Starting complex operation test")
    
    // Test logic here
    
    logutil.InfoIf(logger, verbose, "Complex operation test completed")
}
```

## Common Pitfalls

### 1. Over-Logging

Don't log everything - be selective:

```go
// ❌ Bad: Too much logging
func processItems(items []Item) {
    logger := log.Quick()
    
    logger.Info("Starting to process items")
    for i, item := range items {
        logger.Infof("Processing item %d", i)
        logger.Debugf("Item details: %+v", item)
        
        result := processItem(item)
        
        logger.Infof("Item %d processed", i)
        logger.Debugf("Result: %+v", result)
    }
    logger.Info("Finished processing items")
}

// ✅ Good: Selective logging
func processItems(items []Item) {
    logger := log.Quick()
    
    logger.Infow("Processing items", "count", len(items))
    
    for _, item := range items {
        if err := processItem(item); err != nil {
            logger.Errorw("Failed to process item", 
                "item_id", item.ID, 
                "error", err,
            )
        }
    }
    
    logger.Infow("Items processing completed", "count", len(items))
}
```

### 2. Inconsistent Error Handling

Be consistent in how you handle and log errors:

```go
// ❌ Bad: Inconsistent error handling
func serviceA() error {
    err := operation()
    if err != nil {
        log.Error("Operation failed", err) // Logs and returns
        return err
    }
    return nil
}

func serviceB() error {
    err := operation()
    if err != nil {
        return err // Returns without logging
    }
    return nil
}

// ✅ Good: Consistent error handling
func serviceA() error {
    err := operation()
    if err != nil {
        // Log at the boundary, return wrapped error
        logger.Errorw("Service A operation failed", "error", err)
        return fmt.Errorf("service A failed: %w", err)
    }
    return nil
}

func serviceB() error {
    err := operation()
    if err != nil {
        // Log at the boundary, return wrapped error
        logger.Errorw("Service B operation failed", "error", err)
        return fmt.Errorf("service B failed: %w", err)
    }
    return nil
}
```

### 3. Not Using Context

Always pass context through your application:

```go
// ❌ Bad: No context
func processRequest(userID int) error {
    logger := log.Quick()
    logger.Info("Processing request", "user_id", userID)
    // ... processing ...
}

// ✅ Good: Context-aware
func processRequest(ctx context.Context, userID int) error {
    logger := logutil.WithRequestID(log.Quick(), ctx)
    logger.Info("Processing request", "user_id", userID)
    // ... processing ...
}
```

### 4. Ignoring Performance Impact

Don't ignore the performance impact of logging:

```go
// ❌ Bad: Expensive logging in hot path
func hotPath(data []byte) {
    logger := log.Quick()
    
    // This is called millions of times per second
    logger.Debugw("Processing data", "data", string(data)) // Expensive conversion
}

// ✅ Good: Conditional expensive logging
func hotPath(data []byte) {
    logger := log.Quick()
    
    // Only log if debug level is enabled
    if logger.Level() <= log.DebugLevel {
        logger.Debugw("Processing data", "data_size", len(data))
    }
}
```

## Summary

Following these best practices will help you:

1. **Create maintainable logs** that provide value during debugging and monitoring
2. **Optimize performance** while maintaining observability
3. **Ensure security** by avoiding sensitive data leaks
4. **Enable effective monitoring** with consistent, structured logs
5. **Avoid common pitfalls** that lead to log noise or performance issues

Remember: Good logging is about finding the right balance between too much information (noise) and too little information (blind spots). Focus on logging events that help you understand your application's behavior and diagnose issues when they occur.