package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/video-converter/notification/internal/config"
	"github.com/video-converter/notification/internal/consumer"
	"github.com/video-converter/notification/internal/http"
	"github.com/video-converter/notification/internal/service"
)

func main() {
	log.Println("Starting Notification Service...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Loaded configuration - SMTP Provider: %s", getProviderName(cfg.SMTP.Host))

	// Create notification service
	notificationService := service.NewNotificationService(cfg)
	
	// Test SMTP connection
	if err := notificationService.TestSMTPConnection(); err != nil {
		log.Printf("Warning: SMTP connection test failed: %v", err)
		log.Println("Service will continue but email sending may fail")
	} else {
		log.Printf("SMTP connection test successful - Provider: %s", notificationService.GetSMTPProviderInfo())
	}

	// Create and start HTTP server for health checks
	httpServer := http.NewServer(cfg, notificationService)
	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.Service.Port)
		if err := httpServer.Start(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Create and start consumer
	c := consumer.New(cfg, notificationService)
	
	if err := c.Start(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	// Wait for interrupt signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	log.Println("Shutting down notification service...")
	
	// Stop HTTP server
	if err := httpServer.Stop(); err != nil {
		log.Printf("Error stopping HTTP server: %v", err)
	}
	
	// Stop consumer
	c.Stop()
	log.Println("Notification service stopped successfully")
}

func getProviderName(host string) string {
	switch {
	case host == "smtp.gmail.com":
		return "Gmail"
	case host == "smtp-mail.outlook.com":
		return "Outlook"
	case host == "smtp.mail.yahoo.com":
		return "Yahoo"
	default:
		return "Custom SMTP"
	}
}