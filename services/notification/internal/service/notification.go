package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/video-converter/notification/internal/config"
	"github.com/video-converter/notification/internal/models"
	"github.com/video-converter/notification/internal/smtp"
	"github.com/video-converter/notification/internal/templates"
)

// NotificationService handles notification processing and delivery
type NotificationService struct {
	config         *config.Config
	smtpClient     *smtp.Client
	templateEngine *templates.Engine
	preferences    map[string]*models.NotificationPreferences // In-memory cache for demo
}

// NewNotificationService creates a new notification service
func NewNotificationService(cfg *config.Config) *NotificationService {
	smtpClient := smtp.NewClient(&cfg.SMTP)
	templateEngine := templates.NewEngine()
	
	return &NotificationService{
		config:         cfg,
		smtpClient:     smtpClient,
		templateEngine: templateEngine,
		preferences:    make(map[string]*models.NotificationPreferences),
	}
}

// ProcessNotification processes a notification message
func (s *NotificationService) ProcessNotification(ctx context.Context, notification *models.NotificationMessage) error {
	log.Printf("Processing notification: %s (type: %s, user: %s)", 
		notification.ID, notification.Type, notification.UserID)
	
	// Check user preferences
	if !s.shouldSendNotification(notification) {
		log.Printf("Notification skipped due to user preferences: %s", notification.ID)
		return nil
	}
	
	// Render email template
	emailTemplate, err := s.templateEngine.RenderTemplate(notification.Type, notification.Data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}
	
	// Send email
	err = s.smtpClient.SendEmail(
		notification.Email,
		emailTemplate.Subject,
		emailTemplate.HTMLBody,
		emailTemplate.TextBody,
		emailTemplate.Attachments,
	)
	if err != nil {
		// Record delivery failure
		s.recordDeliveryStatus(notification, "failed", err.Error())
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	// Record successful delivery
	s.recordDeliveryStatus(notification, "sent", "")
	
	log.Printf("Successfully sent notification: %s", notification.ID)
	return nil
}

// shouldSendNotification checks if the notification should be sent based on user preferences
func (s *NotificationService) shouldSendNotification(notification *models.NotificationMessage) bool {
	prefs, exists := s.preferences[notification.UserID]
	if !exists {
		// Default to sending all notifications if no preferences found
		return true
	}
	
	if !prefs.EmailEnabled {
		return false
	}
	
	switch notification.Type {
	case models.NotificationTypeWelcome:
		return prefs.WelcomeEmail
	case models.NotificationTypeConversionComplete:
		return prefs.ConversionCompleteEmail
	case models.NotificationTypeConversionError:
		return prefs.ConversionErrorEmail
	default:
		return true
	}
}

// recordDeliveryStatus records the delivery status (in-memory for demo)
func (s *NotificationService) recordDeliveryStatus(notification *models.NotificationMessage, status, errorMsg string) {
	deliveryStatus := &models.DeliveryStatus{
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		Email:          notification.Email,
		Status:         status,
		Error:          errorMsg,
		SentAt:         time.Now(),
		Attempts:       1, // This would be tracked properly in a real implementation
	}
	
	// In a real implementation, this would be stored in a database
	log.Printf("Delivery status recorded: %+v", deliveryStatus)
}

// SetUserPreferences sets notification preferences for a user
func (s *NotificationService) SetUserPreferences(userID string, prefs *models.NotificationPreferences) {
	prefs.UserID = userID
	prefs.UpdatedAt = time.Now()
	if prefs.CreatedAt.IsZero() {
		prefs.CreatedAt = time.Now()
	}
	
	s.preferences[userID] = prefs
	log.Printf("Updated notification preferences for user: %s", userID)
}

// GetUserPreferences gets notification preferences for a user
func (s *NotificationService) GetUserPreferences(userID string) *models.NotificationPreferences {
	if prefs, exists := s.preferences[userID]; exists {
		return prefs
	}
	
	// Return default preferences
	return &models.NotificationPreferences{
		UserID:                  userID,
		EmailEnabled:            true,
		ConversionCompleteEmail: true,
		ConversionErrorEmail:    true,
		WelcomeEmail:            true,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
}

// TestSMTPConnection tests the SMTP connection
func (s *NotificationService) TestSMTPConnection() error {
	return s.smtpClient.TestConnection()
}

// GetSMTPProviderInfo returns information about the configured SMTP provider
func (s *NotificationService) GetSMTPProviderInfo() string {
	return s.smtpClient.GetProviderInfo()
}