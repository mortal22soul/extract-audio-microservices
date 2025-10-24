package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/video-converter/converter/internal/monitoring"
)

type Server struct {
	port          int
	healthChecker *monitoring.HealthChecker
	server        *http.Server
}

func New(port int, healthChecker *monitoring.HealthChecker) *Server {
	return &Server{
		port:          port,
		healthChecker: healthChecker,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/health/ready", s.handleReadiness)
	mux.HandleFunc("/health/live", s.handleLiveness)
	
	// Metrics endpoint
	mux.HandleFunc("/metrics", s.handleMetrics)
	
	// Root endpoint
	mux.HandleFunc("/", s.handleRoot)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	log.Printf("Starting HTTP server on port %d", s.port)

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return s.server.Shutdown(shutdownCtx)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service":   "converter-service",
		"version":   "1.0.0",
		"timestamp": time.Now(),
		"endpoints": []string{
			"/health",
			"/health/ready",
			"/health/live",
			"/metrics",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	health := s.healthChecker.CheckHealth(ctx)

	w.Header().Set("Content-Type", "application/json")
	
	if health.Overall == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(health)
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health := s.healthChecker.CheckHealth(ctx)

	w.Header().Set("Content-Type", "application/json")

	// Service is ready if all dependencies are healthy
	if health.Overall == "healthy" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "not ready",
			"timestamp": time.Now().Format(time.RFC3339),
			"issues": health.Services,
		})
	}
}

func (s *Server) handleLiveness(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - service is alive if it can respond
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get queue metrics
	queueMetrics, err := s.healthChecker.GetQueueMetrics("video.conversion")
	if err != nil {
		log.Printf("Failed to get queue metrics: %v", err)
	}

	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"queue_metrics": queueMetrics,
		"system_health": s.healthChecker.CheckHealth(ctx),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}