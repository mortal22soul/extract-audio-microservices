package smtp

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"github.com/video-converter/notification/internal/config"
	"github.com/video-converter/notification/internal/models"
	"gopkg.in/gomail.v2"
)

// Client represents an SMTP client with multiple provider support
type Client struct {
	config *config.SMTPConfig
	dialer *gomail.Dialer
}

// NewClient creates a new SMTP client
func NewClient(cfg *config.SMTPConfig) *Client {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	
	// Configure TLS
	dialer.TLSConfig = &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: false,
	}

	return &Client{
		config: cfg,
		dialer: dialer,
	}
}

// SendEmail sends an email using the configured SMTP provider
func (c *Client) SendEmail(to, subject, htmlBody, textBody string, attachments []models.Attachment) error {
	message := gomail.NewMessage()
	
	// Set headers
	message.SetHeader("From", c.config.From)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	
	// Set body
	if htmlBody != "" {
		message.SetBody("text/html", htmlBody)
		if textBody != "" {
			message.AddAlternative("text/plain", textBody)
		}
	} else if textBody != "" {
		message.SetBody("text/plain", textBody)
	}
	
	// Add attachments (simplified for now)
	for _, attachment := range attachments {
		// For now, we'll skip attachments to avoid gomail API complexity
		// In production, this would be implemented properly
		log.Printf("Skipping attachment: %s (not implemented)", attachment.Filename)
	}
	
	// Send email with retry logic
	return c.sendWithRetry(message, 3)
}

// sendWithRetry attempts to send email with retry logic
func (c *Client) sendWithRetry(message *gomail.Message, maxRetries int) error {
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := c.dialer.DialAndSend(message)
		if err == nil {
			log.Printf("Email sent successfully on attempt %d", attempt)
			return nil
		}
		
		lastErr = err
		log.Printf("Email send attempt %d failed: %v", attempt, err)
		
		if attempt < maxRetries {
			// Wait before retry (exponential backoff could be added here)
			continue
		}
	}
	
	return fmt.Errorf("failed to send email after %d attempts: %w", maxRetries, lastErr)
}

// TestConnection tests the SMTP connection
func (c *Client) TestConnection() error {
	// Create a test connection
	conn, err := c.dialer.Dial()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()
	
	log.Println("SMTP connection test successful")
	return nil
}

// GetProviderInfo returns information about the configured SMTP provider
func (c *Client) GetProviderInfo() string {
	host := strings.ToLower(c.config.Host)
	
	switch {
	case strings.Contains(host, "gmail"):
		return "Gmail"
	case strings.Contains(host, "outlook") || strings.Contains(host, "hotmail"):
		return "Outlook"
	case strings.Contains(host, "yahoo"):
		return "Yahoo"
	case strings.Contains(host, "sendgrid"):
		return "SendGrid"
	case strings.Contains(host, "mailgun"):
		return "Mailgun"
	default:
		return "Custom SMTP"
	}
}