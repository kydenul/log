package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kydenul/log"
	"github.com/kydenul/log/logutil"
)

// User represents a user in our system
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// UserService handles user operations
type UserService struct {
	logger *log.Log
	users  map[int]*User // Simple in-memory store for demo
}

func NewUserService(logger *log.Log) *UserService {
	return &UserService{
		logger: logger,
		users: map[int]*User{
			1: {ID: 1, Name: "John Doe", Email: "john@example.com", CreatedAt: time.Now()},
			2: {ID: 2, Name: "Jane Smith", Email: "jane@example.com", CreatedAt: time.Now()},
		},
	}
}

func (s *UserService) GetUser(id int) (*User, error) {
	defer logutil.Timer(s.logger, "get_user")()

	s.logger.Debugw("Getting user", "user_id", id)

	user, exists := s.users[id]
	if !exists {
		s.logger.Warnw("User not found", "user_id", id)
		return nil, nil
	}

	s.logger.Infow("User retrieved", "user_id", id, "user_name", user.Name)
	return user, nil
}

func (s *UserService) CreateUser(name, email string) (*User, error) {
	defer logutil.Timer(s.logger, "create_user")()

	// Generate new ID
	newID := len(s.users) + 1

	user := &User{
		ID:        newID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}

	s.users[newID] = user

	s.logger.Infow("User created",
		"user_id", user.ID,
		"name", user.Name,
		"email", user.Email,
	)

	return user, nil
}

func (s *UserService) ListUsers() []*User {
	defer logutil.Timer(s.logger, "list_users")()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	s.logger.Infow("Users listed", "count", len(users))
	return users
}

// HTTP Handlers
func (s *UserService) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	idStr := r.URL.Path[len("/users/"):]
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		s.logger.Warnw("Invalid user ID", "id_string", idStr, "error", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := s.GetUser(userID)
	if err != nil {
		s.logger.Errorw("Failed to get user", "user_id", userID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *UserService) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warnw("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" {
		s.logger.Warnw("Missing required fields", "name", req.Name, "email", req.Email)
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	user, err := s.CreateUser(req.Name, req.Email)
	if err != nil {
		s.logger.Errorw("Failed to create user", "name", req.Name, "email", req.Email, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (s *UserService) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users := s.ListUsers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "v1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Configure logger based on environment
	env := os.Getenv("ENVIRONMENT")
	var logger *log.Log

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
		// Default to development preset
		logger = log.WithPreset(log.DevelopmentPreset())
	}

	// Log application startup
	port := 8080
	logutil.LogStartup(logger, "web-server-example", "v1.0.0", port)

	// Create user service
	userService := NewUserService(logger)

	// Create HTTP middleware
	middleware := log.HTTPMiddleware(logger)

	// Setup routes with middleware
	http.Handle("/users", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userService.handleListUsers(w, r)
		case http.MethodPost:
			userService.handleCreateUser(w, r)
		default:
			logger.Warnw("Method not allowed", "method", r.Method, "path", r.URL.Path)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	http.Handle("/users/", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			userService.handleGetUser(w, r)
		} else {
			logger.Warnw("Method not allowed", "method", r.Method, "path", r.URL.Path)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	http.Handle("/health", middleware(http.HandlerFunc(healthHandler)))

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Warn("Shutdown signal received, initiating graceful shutdown")

		// In a real application, you would:
		// 1. Stop accepting new requests
		// 2. Wait for existing requests to complete
		// 3. Close database connections
		// 4. Clean up resources

		shutdownStart := time.Now()
		time.Sleep(1 * time.Second) // Simulate cleanup time
		shutdownDuration := time.Since(shutdownStart)

		logutil.LogShutdown(logger, "web-server-example", shutdownDuration)
		os.Exit(0)
	}()

	// Start server
	logger.Infow("Starting HTTP server",
		"port", port,
		"environment", env,
		"endpoints", []string{"/users", "/users/{id}", "/health"},
	)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal("Server failed to start", "error", err)
	}
}
