package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/video-converter/notification/internal/models"
)

// SendWelcomeEmail sends a welcome email to a new user
func (s *NotificationService) SendWelcomeEmail(userID, email, firstName string) error {
	// Generate unsubscribe token
	unsubscribeToken := s.generateUnsubscribeToken(userID)
	
	notification := &models.NotificationMessage{
		ID:     s.generateNotificationID(),
		Type:   models.NotificationTypeWelcome,
		UserID: userID,
		Email:  email,
		Data: map[string]interface{}{
			"FirstName":      firstName,
			"DashboardURL":   s.getDashboardURL(),
			"UnsubscribeURL": s.getUnsubscribeURL(unsubscribeToken),
		},
		Priority: 3, // Medium priority
	}
	
	return s.ProcessNotification(nil, notification)
}

// SendConversionCompleteEmail sends a notification when video conversion is complete
func (s *NotificationService) SendConversionCompleteEmail(userID, email, firstName, videoName, jobID string, duration, fileSize, conversionTime string) error {
	// Generate unsubscribe token
	unsubscribeToken := s.generateUnsubscribeToken(userID)
	
	notification := &models.NotificationMessage{
		ID:     s.generateNotificationID(),
		Type:   models.NotificationTypeConversionComplete,
		UserID: userID,
		Email:  email,
		Data: map[string]interface{}{
			"FirstName":      firstName,
			"VideoName":      videoName,
			"Duration":       duration,
			"FileSize":       fileSize,
			"ConversionTime": conversionTime,
			"Quality":        "192 kbps MP3",
			"DownloadURL":    s.getDownloadURL(jobID),
			"DashboardURL":   s.getDashboardURL(),
			"UnsubscribeURL": s.getUnsubscribeURL(unsubscribeToken),
		},
		Priority: 2, // High priority
	}
	
	return s.ProcessNotification(nil, notification)
}

// SendConversionErrorEmail sends a notification when video conversion fails
func (s *NotificationService) SendConversionErrorEmail(userID, email, firstName, videoName, jobID, errorMessage string) error {
	// Generate unsubscribe token
	unsubscribeToken := s.generateUnsubscribeToken(userID)
	
	notification := &models.NotificationMessage{
		ID:     s.generateNotificationID(),
		Type:   models.NotificationTypeConversionError,
		UserID: userID,
		Email:  email,
		Data: map[string]interface{}{
			"FirstName":    firstName,
			"VideoName":    videoName,
			"ErrorMessage": errorMessage,
			"ErrorTime":    time.Now().Format("2006-01-02 15:04:05 UTC"),
			"JobID":        jobID,
			"UploadURL":    s.getUploadURL(),
			"SupportURL":   s.getSupportURL(),
			"UnsubscribeURL": s.getUnsubscribeURL(unsubscribeToken),
		},
		Priority: 1, // Highest priority
	}
	
	return s.ProcessNotification(nil, notification)
}

// HandleUnsubscribe handles unsubscribe requests
func (s *NotificationService) HandleUnsubscribe(token string) error {
	userID, err := s.validateUnsubscribeToken(token)
	if err != nil {
		return fmt.Errorf("invalid unsubscribe token: %w", err)
	}
	
	// Get current preferences
	prefs := s.GetUserPreferences(userID)
	
	// Disable all email notifications
	prefs.EmailEnabled = false
	prefs.ConversionCompleteEmail = false
	prefs.ConversionErrorEmail = false
	prefs.WelcomeEmail = false
	prefs.UnsubscribeToken = token
	
	// Save preferences
	s.SetUserPreferences(userID, prefs)
	
	return nil
}

// UpdateNotificationPreferences updates user notification preferences
func (s *NotificationService) UpdateNotificationPreferences(userID string, emailEnabled, conversionComplete, conversionError, welcome bool) error {
	prefs := s.GetUserPreferences(userID)
	
	prefs.EmailEnabled = emailEnabled
	prefs.ConversionCompleteEmail = conversionComplete
	prefs.ConversionErrorEmail = conversionError
	prefs.WelcomeEmail = welcome
	
	s.SetUserPreferences(userID, prefs)
	
	return nil
}

// generateNotificationID generates a unique notification ID
func (s *NotificationService) generateNotificationID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("notif_%s_%d", hex.EncodeToString(bytes)[:8], time.Now().Unix())
}

// generateUnsubscribeToken generates a secure unsubscribe token
func (s *NotificationService) generateUnsubscribeToken(userID string) string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return fmt.Sprintf("%s_%s", userID, hex.EncodeToString(bytes))
}

// validateUnsubscribeToken validates and extracts user ID from unsubscribe token
func (s *NotificationService) validateUnsubscribeToken(token string) (string, error) {
	// Simple validation - in production, this would be more secure
	if len(token) < 10 {
		return "", fmt.Errorf("token too short")
	}
	
	// Extract user ID (everything before the first underscore)
	for i, char := range token {
		if char == '_' && i > 0 {
			return token[:i], nil
		}
	}
	
	return "", fmt.Errorf("invalid token format")
}

// URL generation helpers - these would be configurable in production
func (s *NotificationService) getDashboardURL() string {
	return "https://videoconverter.com/dashboard"
}

func (s *NotificationService) getDownloadURL(jobID string) string {
	return fmt.Sprintf("https://videoconverter.com/api/v1/videos/%s/download", jobID)
}

func (s *NotificationService) getUploadURL() string {
	return "https://videoconverter.com/upload"
}

func (s *NotificationService) getSupportURL() string {
	return "https://videoconverter.com/support"
}

func (s *NotificationService) getUnsubscribeURL(token string) string {
	return fmt.Sprintf("https://videoconverter.com/unsubscribe?token=%s", token)
}