package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/video-converter/converter/internal/config"
	"github.com/video-converter/converter/internal/monitoring"
	"github.com/video-converter/converter/internal/server"
	"github.com/video-converter/converter/internal/worker"
)

func main() {
	log.Println("Starting Converter Service...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded: MaxWorkers=%d, TempDir=%s", cfg.MaxWorkers, cfg.TempDir)

	// Create worker
	w, err := worker.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}

	// Create health checker
	healthChecker := monitoring.NewHealthChecker(
		w.GetMongoClient().GetClient(),
		w.GetRedisClient().GetClient(),
		w.GetRabbitClient().GetConnection(),
	)

	// Create HTTP server for health checks
	httpServer := server.New(8082, healthChecker)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Start worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Start()
	}()

	// Start health monitoring
	wg.Add(1)
	go func() {
		defer wg.Done()
		healthChecker.StartHealthMonitoring(ctx, 30*time.Second)
	}()

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.Start(ctx); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down converter service...")
	
	// Cancel context to stop all goroutines
	cancel()
	
	// Stop worker
	w.Stop()

	// Wait for all goroutines to finish
	wg.Wait()

	log.Println("Converter service shutdown complete")
}