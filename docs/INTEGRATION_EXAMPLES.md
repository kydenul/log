# Integration Examples

Examples of integrating the log library with popular Go web frameworks and tools.

## Table of Contents

- [Standard HTTP Server](#standard-http-server)
- [Gin Framework](#gin-framework)
- [Echo Framework](#echo-framework)
- [Fiber Framework](#fiber-framework)
- [Chi Router](#chi-router)
- [gRPC Server](#grpc-server)
- [Database Integration](#database-integration)
- [Background Workers](#background-workers)
- [Microservices](#microservices)

## Standard HTTP Server

### Basic HTTP Server with Middleware

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
    logger := log.WithPreset(log.ProductionPreset())
    
    // Log application startup
    logutil.LogStartup(logger, "http-server", "v1.0.0", 8080)
    
    // Create middleware
    middleware := log.HTTPMiddleware(logger)
    
    // Setup routes with middleware
    http.Handle("/", middleware(http.HandlerFunc(homeHandler)))
    http.Handle("/api/users", middleware(http.HandlerFunc(usersHandler)))
    http.Handle("/api/health", middleware(http.HandlerFunc(healthHandler)))
    
    // Start server
    logger.Info("Starting HTTP server", "port", 8080)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        logger.Fatal("Server failed to start", "error", err)
    }
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    logger := log.Quick() // Or inject via context
    
    defer logutil.Timer(logger, "home_handler")()
    
    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("<h1>Welcome to our service</h1>"))
    
    logger.Info("Home page served")
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    logger := log.Quick()
    
    switch r.Method {
    case http.MethodGet:
        handleGetUsers(w, r, logger)
    case http.MethodPost:
        handleCreateUser(w, r, logger)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
        logger.Warn("Method not allowed", "method", r.Method, "path", r.URL.Path)
    }
}

func handleGetUsers(w http.ResponseWriter, r *http.Request, logger *log.Log) {
    defer logutil.Timer(logger, "get_users")()
    
    // Simulate database query
    time.Sleep(50 * time.Millisecond)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`[{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]`))
    
    logger.Info("Users retrieved", "count", 2)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request, logger *log.Log) {
    defer logutil.Timer(logger, "create_user")()
    
    // Simulate user creation
    time.Sleep(100 * time.Millisecond)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(`{"id": 3, "name": "New User"}`))
    
    logger.Info("User created", "user_id", 3)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}
```

## Gin Framework

### Gin with Custom Logging Middleware

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

func main() {
    // Configure logger
    logger := log.WithPreset(log.ProductionPreset())
    
    // Create Gin router
    r := gin.New()
    
    // Add custom logging middleware
    r.Use(GinLoggingMiddleware(logger))
    r.Use(gin.Recovery())
    
    // Routes
    r.GET("/", homeHandler)
    r.GET("/users/:id", getUserHandler)
    r.POST("/users", createUserHandler)
    r.GET("/health", healthHandler)
    
    // Start server
    logutil.LogStartup(logger, "gin-server", "v1.0.0", 8080)
    logger.Fatal("Server error", "error", r.Run(":8080"))
}

// GinLoggingMiddleware creates a Gin middleware for logging
func GinLoggingMiddleware(logger *log.Log) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery
        
        // Log request start
        logger.Infow("Request started",
            "method", c.Request.Method,
            "path", path,
            "query", raw,
            "ip", c.ClientIP(),
            "user_agent", c.Request.UserAgent(),
        )
        
        // Process request
        c.Next()
        
        // Log request completion
        latency := time.Since(start)
        status := c.Writer.Status()
        
        logLevel := "info"
        if status >= 500 {
            logLevel = "error"
        } else if status >= 400 {
            logLevel = "warn"
        }
        
        fields := []interface{}{
            "method", c.Request.Method,
            "path", path,
            "status", status,
            "latency_ms", latency.Milliseconds(),
            "ip", c.ClientIP(),
            "size", c.Writer.Size(),
        }
        
        switch logLevel {
        case "error":
            logger.Errorw("Request completed", fields...)
        case "warn":
            logger.Warnw("Request completed", fields...)
        default:
            logger.Infow("Request completed", fields...)
        }
    }
}

func homeHandler(c *gin.Context) {
    logger := log.Quick()
    defer logutil.Timer(logger, "home_handler")()
    
    c.JSON(200, gin.H{
        "message": "Welcome to Gin API",
        "version": "v1.0.0",
    })
}

func getUserHandler(c *gin.Context) {
    logger := log.Quick()
    defer logutil.Timer(logger, "get_user")()
    
    userID := c.Param("id")
    
    // Simulate database lookup
    time.Sleep(30 * time.Millisecond)
    
    logger.Infow("User retrieved", "user_id", userID)
    
    c.JSON(200, gin.H{
        "id":   userID,
        "name": "John Doe",
    })
}

func createUserHandler(c *gin.Context) {
    logger := log.Quick()
    defer logutil.Timer(logger, "create_user")()
    
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    
    if err := c.ShouldBindJSON(&user); err != nil {
        logger.Errorw("Invalid request body", "error", err)
        c.JSON(400, gin.H{"error": "Invalid request body"})
        return
    }
    
    // Simulate user creation
    time.Sleep(100 * time.Millisecond)
    
    logger.Infow("User created", "name", user.Name, "email", user.Email)
    
    c.JSON(201, gin.H{
        "id":    123,
        "name":  user.Name,
        "email": user.Email,
    })
}

func healthHandler(c *gin.Context) {
    c.JSON(200, gin.H{
        "status":    "healthy",
        "timestamp": time.Now().Format(time.RFC3339),
    })
}
```

## Echo Framework

### Echo with Structured Logging

```go
package main

import (
    "net/http"
    "time"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

func main() {
    // Configure logger
    logger := log.WithPreset(log.ProductionPreset())
    
    // Create Echo instance
    e := echo.New()
    
    // Middleware
    e.Use(EchoLoggingMiddleware(logger))
    e.Use(middleware.Recover())
    
    // Routes
    e.GET("/", homeHandler)
    e.GET("/users/:id", getUserHandler)
    e.POST("/users", createUserHandler)
    e.GET("/health", healthHandler)
    
    // Start server
    logutil.LogStartup(logger, "echo-server", "v1.0.0", 8080)
    logger.Fatal("Server error", "error", e.Start(":8080"))
}

// EchoLoggingMiddleware creates an Echo middleware for logging
func EchoLoggingMiddleware(logger *log.Log) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            start := time.Now()
            req := c.Request()
            
            // Log request start
            logger.Infow("Request started",
                "method", req.Method,
                "uri", req.RequestURI,
                "ip", c.RealIP(),
                "user_agent", req.UserAgent(),
            )
            
            // Process request
            err := next(c)
            
            // Log request completion
            res := c.Response()
            latency := time.Since(start)
            
            logLevel := "info"
            status := res.Status
            if status >= 500 {
                logLevel = "error"
            } else if status >= 400 {
                logLevel = "warn"
            }
            
            fields := []interface{}{
                "method", req.Method,
                "uri", req.RequestURI,
                "status", status,
                "latency_ms", latency.Milliseconds(),
                "ip", c.RealIP(),
                "size", res.Size,
            }
            
            if err != nil {
                fields = append(fields, "error", err.Error())
            }
            
            switch logLevel {
            case "error":
                logger.Errorw("Request completed", fields...)
            case "warn":
                logger.Warnw("Request completed", fields...)
            default:
                logger.Infow("Request completed", fields...)
            }
            
            return err
        }
    }
}

func homeHandler(c echo.Context) error {
    logger := log.Quick()
    defer logutil.Timer(logger, "home_handler")()
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "message": "Welcome to Echo API",
        "version": "v1.0.0",
    })
}

func getUserHandler(c echo.Context) error {
    logger := log.Quick()
    defer logutil.Timer(logger, "get_user")()
    
    userID := c.Param("id")
    
    // Simulate database lookup
    time.Sleep(30 * time.Millisecond)
    
    logger.Infow("User retrieved", "user_id", userID)
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "id":   userID,
        "name": "John Doe",
    })
}

func createUserHandler(c echo.Context) error {
    logger := log.Quick()
    defer logutil.Timer(logger, "create_user")()
    
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    
    if err := c.Bind(&user); err != nil {
        logger.Errorw("Invalid request body", "error", err)
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body",
        })
    }
    
    // Simulate user creation
    time.Sleep(100 * time.Millisecond)
    
    logger.Infow("User created", "name", user.Name, "email", user.Email)
    
    return c.JSON(http.StatusCreated, map[string]interface{}{
        "id":    123,
        "name":  user.Name,
        "email": user.Email,
    })
}

func healthHandler(c echo.Context) error {
    return c.JSON(http.StatusOK, map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now().Format(time.RFC3339),
    })
}
```

## Fiber Framework

### Fiber with Request ID Logging

```go
package main

import (
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/requestid"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

func main() {
    // Configure logger
    logger := log.WithPreset(log.ProductionPreset())
    
    // Create Fiber app
    app := fiber.New()
    
    // Middleware
    app.Use(requestid.New())
    app.Use(FiberLoggingMiddleware(logger))
    
    // Routes
    app.Get("/", homeHandler)
    app.Get("/users/:id", getUserHandler)
    app.Post("/users", createUserHandler)
    app.Get("/health", healthHandler)
    
    // Start server
    logutil.LogStartup(logger, "fiber-server", "v1.0.0", 8080)
    logger.Fatal("Server error", "error", app.Listen(":8080"))
}

// FiberLoggingMiddleware creates a Fiber middleware for logging
func FiberLoggingMiddleware(logger *log.Log) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        
        // Get request ID
        requestID := c.Locals("requestid")
        
        // Log request start
        logger.Infow("Request started",
            "request_id", requestID,
            "method", c.Method(),
            "path", c.Path(),
            "ip", c.IP(),
            "user_agent", c.Get("User-Agent"),
        )
        
        // Process request
        err := c.Next()
        
        // Log request completion
        latency := time.Since(start)
        status := c.Response().StatusCode()
        
        logLevel := "info"
        if status >= 500 {
            logLevel = "error"
        } else if status >= 400 {
            logLevel = "warn"
        }
        
        fields := []interface{}{
            "request_id", requestID,
            "method", c.Method(),
            "path", c.Path(),
            "status", status,
            "latency_ms", latency.Milliseconds(),
            "ip", c.IP(),
            "size", len(c.Response().Body()),
        }
        
        if err != nil {
            fields = append(fields, "error", err.Error())
        }
        
        switch logLevel {
        case "error":
            logger.Errorw("Request completed", fields...)
        case "warn":
            logger.Warnw("Request completed", fields...)
        default:
            logger.Infow("Request completed", fields...)
        }
        
        return err
    }
}

func homeHandler(c *fiber.Ctx) error {
    logger := log.Quick()
    requestID := c.Locals("requestid")
    
    defer logutil.Timer(logger, "home_handler")()
    
    logger.Infow("Home handler called", "request_id", requestID)
    
    return c.JSON(fiber.Map{
        "message": "Welcome to Fiber API",
        "version": "v1.0.0",
    })
}

func getUserHandler(c *fiber.Ctx) error {
    logger := log.Quick()
    requestID := c.Locals("requestid")
    
    defer logutil.Timer(logger, "get_user")()
    
    userID := c.Params("id")
    
    // Simulate database lookup
    time.Sleep(30 * time.Millisecond)
    
    logger.Infow("User retrieved", "request_id", requestID, "user_id", userID)
    
    return c.JSON(fiber.Map{
        "id":   userID,
        "name": "John Doe",
    })
}

func createUserHandler(c *fiber.Ctx) error {
    logger := log.Quick()
    requestID := c.Locals("requestid")
    
    defer logutil.Timer(logger, "create_user")()
    
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    
    if err := c.BodyParser(&user); err != nil {
        logger.Errorw("Invalid request body", "request_id", requestID, "error", err)
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    // Simulate user creation
    time.Sleep(100 * time.Millisecond)
    
    logger.Infow("User created", "request_id", requestID, "name", user.Name, "email", user.Email)
    
    return c.Status(201).JSON(fiber.Map{
        "id":    123,
        "name":  user.Name,
        "email": user.Email,
    })
}

func healthHandler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status":    "healthy",
        "timestamp": time.Now().Format(time.RFC3339),
    })
}
```

## Chi Router

### Chi with Structured Logging and Request Context

```go
package main

import (
    "context"
    "net/http"
    "time"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

type contextKey string

const loggerKey contextKey = "logger"

func main() {
    // Configure logger
    logger := log.WithPreset(log.ProductionPreset())
    
    // Create Chi router
    r := chi.NewRouter()
    
    // Middleware
    r.Use(middleware.RequestID)
    r.Use(ChiLoggingMiddleware(logger))
    r.Use(middleware.Recoverer)
    
    // Routes
    r.Get("/", homeHandler)
    r.Get("/users/{id}", getUserHandler)
    r.Post("/users", createUserHandler)
    r.Get("/health", healthHandler)
    
    // Start server
    logutil.LogStartup(logger, "chi-server", "v1.0.0", 8080)
    logger.Fatal("Server error", "error", http.ListenAndServe(":8080", r))
}

// ChiLoggingMiddleware creates a Chi middleware for logging
func ChiLoggingMiddleware(logger *log.Log) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Get request ID from Chi middleware
            requestID := middleware.GetReqID(r.Context())
            
            // Create request-scoped logger
            requestLogger := logger // In a real app, you might create a wrapper with request ID
            
            // Add logger to context
            ctx := context.WithValue(r.Context(), loggerKey, requestLogger)
            r = r.WithContext(ctx)
            
            // Log request start
            requestLogger.Infow("Request started",
                "request_id", requestID,
                "method", r.Method,
                "path", r.URL.Path,
                "ip", r.RemoteAddr,
                "user_agent", r.UserAgent(),
            )
            
            // Wrap ResponseWriter to capture status
            ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
            
            // Process request
            next.ServeHTTP(ww, r)
            
            // Log request completion
            latency := time.Since(start)
            status := ww.Status()
            
            logLevel := "info"
            if status >= 500 {
                logLevel = "error"
            } else if status >= 400 {
                logLevel = "warn"
            }
            
            fields := []interface{}{
                "request_id", requestID,
                "method", r.Method,
                "path", r.URL.Path,
                "status", status,
                "latency_ms", latency.Milliseconds(),
                "ip", r.RemoteAddr,
                "size", ww.BytesWritten(),
            }
            
            switch logLevel {
            case "error":
                requestLogger.Errorw("Request completed", fields...)
            case "warn":
                requestLogger.Warnw("Request completed", fields...)
            default:
                requestLogger.Infow("Request completed", fields...)
            }
        })
    }
}

// getLoggerFromContext extracts logger from request context
func getLoggerFromContext(r *http.Request) *log.Log {
    if logger, ok := r.Context().Value(loggerKey).(*log.Log); ok {
        return logger
    }
    return log.Quick() // Fallback
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    logger := getLoggerFromContext(r)
    requestID := middleware.GetReqID(r.Context())
    
    defer logutil.Timer(logger, "home_handler")()
    
    logger.Infow("Home handler called", "request_id", requestID)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message": "Welcome to Chi API", "version": "v1.0.0"}`))
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
    logger := getLoggerFromContext(r)
    requestID := middleware.GetReqID(r.Context())
    
    defer logutil.Timer(logger, "get_user")()
    
    userID := chi.URLParam(r, "id")
    
    // Simulate database lookup
    time.Sleep(30 * time.Millisecond)
    
    logger.Infow("User retrieved", "request_id", requestID, "user_id", userID)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"id": "` + userID + `", "name": "John Doe"}`))
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    logger := getLoggerFromContext(r)
    requestID := middleware.GetReqID(r.Context())
    
    defer logutil.Timer(logger, "create_user")()
    
    // Simulate user creation
    time.Sleep(100 * time.Millisecond)
    
    logger.Infow("User created", "request_id", requestID)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(`{"id": 123, "name": "New User"}`))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}
```

## Database Integration

### Database Operations with Logging

```go
package main

import (
    "database/sql"
    "time"
    _ "github.com/lib/pq"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

type UserService struct {
    db     *sql.DB
    logger *log.Log
}

func NewUserService(db *sql.DB, logger *log.Log) *UserService {
    return &UserService{
        db:     db,
        logger: logger,
    }
}

func (s *UserService) GetUser(id int) (*User, error) {
    defer logutil.Timer(s.logger, "get_user_db")()
    
    s.logger.Debugw("Getting user from database", "user_id", id)
    
    var user User
    query := "SELECT id, name, email, created_at FROM users WHERE id = $1"
    
    err := s.db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            s.logger.Warnw("User not found", "user_id", id)
            return nil, ErrUserNotFound
        }
        s.logger.Errorw("Database query failed", "user_id", id, "error", err)
        return nil, err
    }
    
    s.logger.Infow("User retrieved successfully", "user_id", id, "user_name", user.Name)
    return &user, nil
}

func (s *UserService) CreateUser(name, email string) (*User, error) {
    defer logutil.Timer(s.logger, "create_user_db")()
    
    s.logger.Infow("Creating new user", "name", name, "email", email)
    
    var user User
    query := `
        INSERT INTO users (name, email, created_at) 
        VALUES ($1, $2, $3) 
        RETURNING id, name, email, created_at`
    
    err := s.db.QueryRow(query, name, email, time.Now()).Scan(
        &user.ID, &user.Name, &user.Email, &user.CreatedAt)
    if err != nil {
        s.logger.Errorw("Failed to create user", "name", name, "email", email, "error", err)
        return nil, err
    }
    
    s.logger.Infow("User created successfully", 
        "user_id", user.ID, "name", user.Name, "email", user.Email)
    return &user, nil
}

func (s *UserService) UpdateUser(id int, name, email string) error {
    defer logutil.Timer(s.logger, "update_user_db")()
    
    s.logger.Infow("Updating user", "user_id", id, "name", name, "email", email)
    
    query := "UPDATE users SET name = $1, email = $2 WHERE id = $3"
    result, err := s.db.Exec(query, name, email, id)
    if err != nil {
        s.logger.Errorw("Failed to update user", "user_id", id, "error", err)
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        s.logger.Errorw("Failed to get rows affected", "user_id", id, "error", err)
        return err
    }
    
    if rowsAffected == 0 {
        s.logger.Warnw("No user updated (user not found)", "user_id", id)
        return ErrUserNotFound
    }
    
    s.logger.Infow("User updated successfully", "user_id", id, "rows_affected", rowsAffected)
    return nil
}

func (s *UserService) DeleteUser(id int) error {
    defer logutil.Timer(s.logger, "delete_user_db")()
    
    s.logger.Infow("Deleting user", "user_id", id)
    
    query := "DELETE FROM users WHERE id = $1"
    result, err := s.db.Exec(query, id)
    if err != nil {
        s.logger.Errorw("Failed to delete user", "user_id", id, "error", err)
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        s.logger.Errorw("Failed to get rows affected", "user_id", id, "error", err)
        return err
    }
    
    if rowsAffected == 0 {
        s.logger.Warnw("No user deleted (user not found)", "user_id", id)
        return ErrUserNotFound
    }
    
    s.logger.Infow("User deleted successfully", "user_id", id)
    return nil
}

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

var ErrUserNotFound = errors.New("user not found")

func main() {
    // Configure logger
    logger := log.WithPreset(log.ProductionPreset())
    
    // Connect to database
    db, err := sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")
    logutil.FatalOnError(logger, err, "Failed to connect to database")
    defer db.Close()
    
    // Test database connection
    err = db.Ping()
    logutil.FatalOnError(logger, err, "Failed to ping database")
    
    logger.Info("Database connection established")
    
    // Create user service
    userService := NewUserService(db, logger)
    
    // Example usage
    user, err := userService.CreateUser("John Doe", "john@example.com")
    if logutil.CheckError(logger, err, "Failed to create user") {
        return
    }
    
    logger.Infow("Created user", "user", user)
}
```

## Background Workers

### Worker Pool with Logging

```go
package main

import (
    "context"
    "sync"
    "time"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

type Job struct {
    ID   int
    Data string
}

type Worker struct {
    id     int
    jobs   <-chan Job
    logger *log.Log
    wg     *sync.WaitGroup
}

func NewWorker(id int, jobs <-chan Job, logger *log.Log, wg *sync.WaitGroup) *Worker {
    return &Worker{
        id:     id,
        jobs:   jobs,
        logger: logger,
        wg:     wg,
    }
}

func (w *Worker) Start(ctx context.Context) {
    defer w.wg.Done()
    
    w.logger.Infow("Worker started", "worker_id", w.id)
    
    for {
        select {
        case job, ok := <-w.jobs:
            if !ok {
                w.logger.Infow("Worker stopping (channel closed)", "worker_id", w.id)
                return
            }
            w.processJob(job)
            
        case <-ctx.Done():
            w.logger.Infow("Worker stopping (context cancelled)", "worker_id", w.id)
            return
        }
    }
}

func (w *Worker) processJob(job Job) {
    defer logutil.Timer(w.logger, "process_job")()
    defer logutil.LogPanicAsError(w.logger, "job_processing")
    
    w.logger.Infow("Processing job", "worker_id", w.id, "job_id", job.ID)
    
    // Simulate work
    time.Sleep(time.Duration(100+job.ID*10) * time.Millisecond)
    
    // Simulate occasional errors
    if job.ID%10 == 0 {
        w.logger.Errorw("Job failed", "worker_id", w.id, "job_id", job.ID, "reason", "simulated_error")
        return
    }
    
    w.logger.Infow("Job completed", "worker_id", w.id, "job_id", job.ID)
}

type WorkerPool struct {
    workerCount int
    jobs        chan Job
    logger      *log.Log
    wg          sync.WaitGroup
}

func NewWorkerPool(workerCount int, logger *log.Log) *WorkerPool {
    return &WorkerPool{
        workerCount: workerCount,
        jobs:        make(chan Job, workerCount*2), // Buffer jobs
        logger:      logger,
    }
}

func (wp *WorkerPool) Start(ctx context.Context) {
    wp.logger.Infow("Starting worker pool", "worker_count", wp.workerCount)
    
    // Start workers
    for i := 1; i <= wp.workerCount; i++ {
        wp.wg.Add(1)
        worker := NewWorker(i, wp.jobs, wp.logger, &wp.wg)
        go worker.Start(ctx)
    }
    
    wp.logger.Infow("Worker pool started", "worker_count", wp.workerCount)
}

func (wp *WorkerPool) AddJob(job Job) {
    select {
    case wp.jobs <- job:
        wp.logger.Debugw("Job queued", "job_id", job.ID)
    default:
        wp.logger.Warnw("Job queue full, dropping job", "job_id", job.ID)
    }
}

func (wp *WorkerPool) Stop() {
    wp.logger.Info("Stopping worker pool")
    close(wp.jobs)
    wp.wg.Wait()
    wp.logger.Info("Worker pool stopped")
}

func main() {
    // Configure logger
    logger := log.WithPreset(log.ProductionPreset())
    
    // Log application startup
    logutil.LogStartup(logger, "worker-pool", "v1.0.0", 0)
    
    // Create context with cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Create and start worker pool
    pool := NewWorkerPool(5, logger)
    pool.Start(ctx)
    
    // Add jobs
    go func() {
        for i := 1; i <= 50; i++ {
            job := Job{
                ID:   i,
                Data: fmt.Sprintf("job-data-%d", i),
            }
            pool.AddJob(job)
            time.Sleep(50 * time.Millisecond)
        }
    }()
    
    // Run for a while
    time.Sleep(10 * time.Second)
    
    // Graceful shutdown
    logger.Info("Initiating graceful shutdown")
    cancel()
    pool.Stop()
    
    // Log application shutdown
    logutil.LogShutdown(logger, "worker-pool", 10*time.Second)
}
```

## Microservices

### Service-to-Service Communication with Distributed Tracing

```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

// ServiceClient handles communication with other services
type ServiceClient struct {
    baseURL string
    client  *http.Client
    logger  *log.Log
}

func NewServiceClient(baseURL string, logger *log.Log) *ServiceClient {
    return &ServiceClient{
        baseURL: baseURL,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
        logger: logger,
    }
}

func (sc *ServiceClient) GetUser(ctx context.Context, userID int) (*User, error) {
    defer logutil.Timer(sc.logger, "service_call_get_user")()
    
    url := fmt.Sprintf("%s/users/%d", sc.baseURL, userID)
    
    // Extract trace ID from context (if available)
    traceID := getTraceIDFromContext(ctx)
    
    sc.logger.Infow("Making service call",
        "service", "user-service",
        "method", "GET",
        "url", url,
        "trace_id", traceID,
    )
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        sc.logger.Errorw("Failed to create request", "error", err, "trace_id", traceID)
        return nil, err
    }
    
    // Add trace ID to headers
    if traceID != "" {
        req.Header.Set("X-Trace-ID", traceID)
    }
    
    start := time.Now()
    resp, err := sc.client.Do(req)
    duration := time.Since(start)
    
    if err != nil {
        sc.logger.Errorw("Service call failed",
            "service", "user-service",
            "url", url,
            "duration_ms", duration.Milliseconds(),
            "error", err,
            "trace_id", traceID,
        )
        return nil, err
    }
    defer resp.Body.Close()
    
    sc.logger.Infow("Service call completed",
        "service", "user-service",
        "url", url,
        "status", resp.StatusCode,
        "duration_ms", duration.Milliseconds(),
        "trace_id", traceID,
    )
    
    if resp.StatusCode != http.StatusOK {
        sc.logger.Warnw("Service returned non-200 status",
            "service", "user-service",
            "status", resp.StatusCode,
            "trace_id", traceID,
        )
        return nil, fmt.Errorf("service returned status %d", resp.StatusCode)
    }
    
    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        sc.logger.Errorw("Failed to decode response", "error", err, "trace_id", traceID)
        return nil, err
    }
    
    sc.logger.Infow("User retrieved from service",
        "user_id", user.ID,
        "user_name", user.Name,
        "trace_id", traceID,
    )
    
    return &user, nil
}

func (sc *ServiceClient) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
    defer logutil.Timer(sc.logger, "service_call_create_order")()
    
    url := fmt.Sprintf("%s/orders", sc.baseURL)
    traceID := getTraceIDFromContext(ctx)
    
    sc.logger.Infow("Creating order via service",
        "service", "order-service",
        "user_id", order.UserID,
        "amount", order.Amount,
        "trace_id", traceID,
    )
    
    body, err := json.Marshal(order)
    if err != nil {
        sc.logger.Errorw("Failed to marshal order", "error", err, "trace_id", traceID)
        return nil, err
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
    if err != nil {
        sc.logger.Errorw("Failed to create request", "error", err, "trace_id", traceID)
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if traceID != "" {
        req.Header.Set("X-Trace-ID", traceID)
    }
    
    start := time.Now()
    resp, err := sc.client.Do(req)
    duration := time.Since(start)
    
    if err != nil {
        sc.logger.Errorw("Service call failed",
            "service", "order-service",
            "duration_ms", duration.Milliseconds(),
            "error", err,
            "trace_id", traceID,
        )
        return nil, err
    }
    defer resp.Body.Close()
    
    sc.logger.Infow("Order creation call completed",
        "service", "order-service",
        "status", resp.StatusCode,
        "duration_ms", duration.Milliseconds(),
        "trace_id", traceID,
    )
    
    if resp.StatusCode != http.StatusCreated {
        sc.logger.Errorw("Failed to create order",
            "service", "order-service",
            "status", resp.StatusCode,
            "trace_id", traceID,
        )
        return nil, fmt.Errorf("failed to create order, status: %d", resp.StatusCode)
    }
    
    var createdOrder Order
    if err := json.NewDecoder(resp.Body).Decode(&createdOrder); err != nil {
        sc.logger.Errorw("Failed to decode response", "error", err, "trace_id", traceID)
        return nil, err
    }
    
    sc.logger.Infow("Order created successfully",
        "order_id", createdOrder.ID,
        "user_id", createdOrder.UserID,
        "amount", createdOrder.Amount,
        "trace_id", traceID,
    )
    
    return &createdOrder, nil
}

// OrderService handles order processing
type OrderService struct {
    userClient *ServiceClient
    logger     *log.Log
}

func NewOrderService(userServiceURL string, logger *log.Log) *OrderService {
    return &OrderService{
        userClient: NewServiceClient(userServiceURL, logger),
        logger:     logger,
    }
}

func (os *OrderService) ProcessOrder(ctx context.Context, orderReq *OrderRequest) (*Order, error) {
    defer logutil.Timer(os.logger, "process_order")()
    defer logutil.LogPanicAsError(os.logger, "order_processing")
    
    traceID := getTraceIDFromContext(ctx)
    
    os.logger.Infow("Processing order",
        "user_id", orderReq.UserID,
        "amount", orderReq.Amount,
        "trace_id", traceID,
    )
    
    // 1. Validate user exists
    user, err := os.userClient.GetUser(ctx, orderReq.UserID)
    if err != nil {
        os.logger.Errorw("Failed to validate user",
            "user_id", orderReq.UserID,
            "error", err,
            "trace_id", traceID,
        )
        return nil, fmt.Errorf("user validation failed: %w", err)
    }
    
    os.logger.Infow("User validated",
        "user_id", user.ID,
        "user_name", user.Name,
        "trace_id", traceID,
    )
    
    // 2. Create order
    order := &Order{
        UserID: orderReq.UserID,
        Amount: orderReq.Amount,
        Status: "pending",
    }
    
    // Simulate order processing
    time.Sleep(100 * time.Millisecond)
    
    order.ID = generateOrderID()
    order.Status = "completed"
    order.CreatedAt = time.Now()
    
    os.logger.Infow("Order processed successfully",
        "order_id", order.ID,
        "user_id", order.UserID,
        "amount", order.Amount,
        "status", order.Status,
        "trace_id", traceID,
    )
    
    return order, nil
}

// HTTP handlers with distributed tracing
func orderHandler(orderService *OrderService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        logger := log.Quick()
        
        // Extract or generate trace ID
        traceID := r.Header.Get("X-Trace-ID")
        if traceID == "" {
            traceID = generateTraceID()
        }
        
        // Add trace ID to context
        ctx := context.WithValue(r.Context(), "trace_id", traceID)
        
        // Add trace ID to response headers
        w.Header().Set("X-Trace-ID", traceID)
        
        logger.Infow("Order request received",
            "method", r.Method,
            "path", r.URL.Path,
            "trace_id", traceID,
        )
        
        var orderReq OrderRequest
        if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
            logger.Errorw("Invalid request body", "error", err, "trace_id", traceID)
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }
        
        order, err := orderService.ProcessOrder(ctx, &orderReq)
        if err != nil {
            logger.Errorw("Order processing failed", "error", err, "trace_id", traceID)
            http.Error(w, "Order processing failed", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(order)
        
        logger.Infow("Order request completed",
            "order_id", order.ID,
            "trace_id", traceID,
        )
    }
}

// Helper functions
func getTraceIDFromContext(ctx context.Context) string {
    if traceID, ok := ctx.Value("trace_id").(string); ok {
        return traceID
    }
    return ""
}

func generateTraceID() string {
    return fmt.Sprintf("trace-%d", time.Now().UnixNano())
}

func generateOrderID() int {
    return int(time.Now().UnixNano() % 1000000)
}

// Data structures
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type Order struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    Amount    float64   `json:"amount"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

type OrderRequest struct {
    UserID int     `json:"user_id"`
    Amount float64 `json:"amount"`
}

func main() {
    // Configure logger
    logger := log.WithPreset(log.ProductionPreset())
    
    // Create order service
    orderService := NewOrderService("http://user-service:8080", logger)
    
    // Setup HTTP server
    http.HandleFunc("/orders", orderHandler(orderService))
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "healthy"}`))
    })
    
    // Start server
    logutil.LogStartup(logger, "order-service", "v1.0.0", 8080)
    logger.Fatal("Server error", "error", http.ListenAndServe(":8080", nil))
}
```

These integration examples demonstrate how to use the enhanced log library with various Go frameworks and patterns. Each example shows:

1. **Proper logger configuration** for the specific use case
2. **Middleware integration** for automatic request/response logging
3. **Structured logging** with relevant context information
4. **Error handling** with appropriate log levels
5. **Performance monitoring** using timing utilities
6. **Request tracing** for distributed systems

Choose the example that best matches your architecture and customize it for your specific needs.