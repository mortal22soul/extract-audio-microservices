package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/video-converter/notification/internal/config"
	"github.com/video-converter/notification/internal/service"
)

// Server represents the HTTP server for health checks and management
type Server struct {
	config              *config.Config
	notificationService *service.NotificationService
	server              *http.Server
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, notificationService *service.NotificationService) *Server {
	return &Server{
		config:              cfg,
		notificationService: notificationService,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	
	// Health check endpoints
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/health/ready", s.readinessHandler)
	mux.HandleFunc("/health/live", s.livenessHandler)
	
	// Metrics endpoint (basic)
	mux.HandleFunc("/metrics", s.metricsHandler)
	
	// Notification management endpoints
	mux.HandleFunc("/unsubscribe", s.unsubscribeHandler)
	mux.HandleFunc("/preferences/", s.preferencesHandler)
	
	// Test endpoints (for development)
	mux.HandleFunc("/test/welcome", s.testWelcomeHandler)
	mux.HandleFunc("/test/conversion-complete", s.testConversionCompleteHandler)
	mux.HandleFunc("/test/conversion-error", s.testConversionErrorHandler)
	
	s.server = &http.Server{
		Addr:         ":" + s.config.Service.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	return s.server.ListenAndServe()
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// healthHandler handles general health checks
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "notification-service",
		"version":   "1.0.0",
	}
	
	// Test SMTP connection
	if err := s.notificationService.TestSMTPConnection(); err != nil {
		health["status"] = "degraded"
		health["smtp_error"] = err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		health["smtp_provider"] = s.notificationService.GetSMTPProviderInfo()
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// readinessHandler handles readiness probes
func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check if service is ready to accept traffic
	ready := map[string]interface{}{
		"ready":     true,
		"timestamp": time.Now().UTC(),
	}
	
	// Test SMTP connection for readiness
	if err := s.notificationService.TestSMTPConnection(); err != nil {
		ready["ready"] = false
		ready["error"] = "SMTP connection failed"
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ready)
}

// livenessHandler handles liveness probes
func (s *Server) livenessHandler(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check
	liveness := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now().UTC(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(liveness)
}

// metricsHandler provides basic metrics
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"service":           "notification-service",
		"uptime_seconds":    time.Since(time.Now()).Seconds(), // This would be tracked properly
		"smtp_provider":     s.notificationService.GetSMTPProviderInfo(),
		"notifications_sent": 0, // This would be tracked properly
		"notifications_failed": 0, // This would be tracked properly
		"timestamp":         time.Now().UTC(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// unsubscribeHandler handles unsubscribe requests
func (s *Server) unsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing unsubscribe token", http.StatusBadRequest)
		return
	}
	
	err := s.notificationService.HandleUnsubscribe(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to unsubscribe: %v", err), http.StatusBadRequest)
		return
	}
	
	// Return a simple HTML page
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Unsubscribed</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .container { max-width: 500px; margin: 0 auto; }
        .success { color: #4CAF50; }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="success">✅ Successfully Unsubscribed</h1>
        <p>You have been unsubscribed from all email notifications.</p>
        <p>You can update your preferences anytime in your account settings.</p>
    </div>
</body>
</html>
	`)
}

// preferencesHandler handles notification preferences
func (s *Server) preferencesHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	userID := r.URL.Path[len("/preferences/"):]
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case http.MethodGet:
		prefs := s.notificationService.GetUserPreferences(userID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prefs)
		
	case http.MethodPut:
		var updateReq struct {
			EmailEnabled            bool `json:"email_enabled"`
			ConversionCompleteEmail bool `json:"conversion_complete_email"`
			ConversionErrorEmail    bool `json:"conversion_error_email"`
			WelcomeEmail            bool `json:"welcome_email"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		err := s.notificationService.UpdateNotificationPreferences(
			userID,
			updateReq.EmailEnabled,
			updateReq.ConversionCompleteEmail,
			updateReq.ConversionErrorEmail,
			updateReq.WelcomeEmail,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update preferences: %v", err), http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Test handlers for development
func (s *Server) testWelcomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		UserID    string `json:"user_id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	err := s.notificationService.SendWelcomeEmail(req.UserID, req.Email, req.FirstName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send welcome email: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

func (s *Server) testConversionCompleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		UserID         string `json:"user_id"`
		Email          string `json:"email"`
		FirstName      string `json:"first_name"`
		VideoName      string `json:"video_name"`
		JobID          string `json:"job_id"`
		Duration       string `json:"duration"`
		FileSize       string `json:"file_size"`
		ConversionTime string `json:"conversion_time"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	err := s.notificationService.SendConversionCompleteEmail(
		req.UserID, req.Email, req.FirstName, req.VideoName, req.JobID,
		req.Duration, req.FileSize, req.ConversionTime,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send conversion complete email: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

func (s *Server) testConversionErrorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		UserID       string `json:"user_id"`
		Email        string `json:"email"`
		FirstName    string `json:"first_name"`
		VideoName    string `json:"video_name"`
		JobID        string `json:"job_id"`
		ErrorMessage string `json:"error_message"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	err := s.notificationService.SendConversionErrorEmail(
		req.UserID, req.Email, req.FirstName, req.VideoName, req.JobID, req.ErrorMessage,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send conversion error email: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}