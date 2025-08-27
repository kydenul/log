# Web Server Example

This example demonstrates how to use the enhanced log library in a web server application.

## Features Demonstrated

- **Environment-specific configuration** - Different logging setups for development vs production
- **HTTP middleware** - Automatic request/response logging
- **Structured logging** - Key-value pairs for better searchability
- **Performance timing** - Automatic timing of operations
- **Error handling** - Consistent error logging patterns
- **Application lifecycle** - Startup and shutdown logging
- **Graceful shutdown** - Proper cleanup with logging

## Running the Example

### Development Mode (Default)

```bash
go run main.go
```

This uses the development preset with:
- Debug level logging
- Console output format
- Caller information enabled
- Fast flush for immediate feedback

### Production Mode

```bash
ENVIRONMENT=production go run main.go
```

This uses production-optimized settings with:
- Info level logging
- JSON output format
- File compression enabled
- Larger buffers for better performance

## API Endpoints

### List Users
```bash
curl http://localhost:8080/users
```

### Get User by ID
```bash
curl http://localhost:8080/users/1
```

### Create User
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Johnson", "email": "alice@example.com"}'
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Log Output Examples

### Request Logging (Automatic via Middleware)

```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.123Z",
  "msg": "HTTP请求开始",
  "method": "GET",
  "url": "/users/1",
  "remote_addr": "127.0.0.1:54321",
  "user_agent": "curl/7.68.0",
  "host": "localhost:8080"
}
```

### Business Logic Logging

```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.125Z",
  "msg": "User retrieved",
  "user_id": 1,
  "user_name": "John Doe"
}
```

### Performance Timing

```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.127Z",
  "msg": "操作耗时",
  "operation": "get_user",
  "duration_ms": 2,
  "duration": "2.1ms"
}
```

### Request Completion

```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.128Z",
  "msg": "HTTP请求完成",
  "method": "GET",
  "url": "/users/1",
  "status_code": 200,
  "duration_ms": 5,
  "duration_ns": 5234567,
  "remote_addr": "127.0.0.1:54321"
}
```

## Key Patterns Demonstrated

### 1. Service-Level Logging

```go
type UserService struct {
    logger *log.Log
    // ... other fields
}

func (s *UserService) GetUser(id int) (*User, error) {
    defer logutil.Timer(s.logger, "get_user")()
    
    s.logger.Debugw("Getting user", "user_id", id)
    // ... business logic
    s.logger.Infow("User retrieved", "user_id", id, "user_name", user.Name)
}
```

### 2. HTTP Handler Logging

```go
func (s *UserService) handleGetUser(w http.ResponseWriter, r *http.Request) {
    // Extract and validate input
    userID, err := strconv.Atoi(idStr)
    if err != nil {
        s.logger.Warnw("Invalid user ID", "id_string", idStr, "error", err)
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // Call business logic
    user, err := s.GetUser(userID)
    if err != nil {
        s.logger.Errorw("Failed to get user", "user_id", userID, "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
}
```

### 3. Environment-Specific Configuration

```go
switch env {
case "production":
    logger = log.NewBuilder().
        Production().
        Directory("/var/log/web-server").
        MaxSize(100).
        MaxBackups(10).
        Build()
case "development":
    logger = log.WithPreset(log.DevelopmentPreset())
default:
    logger = log.WithPreset(log.DevelopmentPreset())
}
```

### 4. Graceful Shutdown with Logging

```go
go func() {
    <-c
    logger.Warn("Shutdown signal received, initiating graceful shutdown")
    
    shutdownStart := time.Now()
    // Perform cleanup...
    shutdownDuration := time.Since(shutdownStart)
    
    logutil.LogShutdown(logger, "web-server-example", shutdownDuration)
    os.Exit(0)
}()
```

## Testing the Example

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Make some requests:**
   ```bash
   # List users
   curl http://localhost:8080/users
   
   # Get specific user
   curl http://localhost:8080/users/1
   
   # Create new user
   curl -X POST http://localhost:8080/users \
     -H "Content-Type: application/json" \
     -d '{"name": "Test User", "email": "test@example.com"}'
   
   # Invalid request (to see error logging)
   curl http://localhost:8080/users/invalid
   ```

3. **Check the logs:**
   - Console output shows all log messages
   - Log files are created in `./logs/` directory
   - Error logs are separated into `*_error.log` files

4. **Test graceful shutdown:**
   - Press `Ctrl+C` to trigger shutdown
   - Observe shutdown logging in console

## Log Files

The example creates several log files:

- `logs/web-server-example-YYYY-MM-DD.log` - Main application logs
- `logs/web-server-example-YYYY-MM-DD_error.log` - Error logs only
- Files rotate automatically when they reach the configured size limit

## Customization

You can customize the logging behavior by:

1. **Changing log levels:**
   ```bash
   LOG_LEVEL=debug go run main.go
   ```

2. **Using different formats:**
   ```bash
   LOG_FORMAT=json go run main.go
   ```

3. **Custom log directory:**
   ```bash
   LOG_DIRECTORY=/tmp/logs go run main.go
   ```

4. **Loading from config file:**
   Create a `config.yaml` file and use `log.FromConfigFile("config.yaml")`

This example provides a solid foundation for building production web services with comprehensive logging.