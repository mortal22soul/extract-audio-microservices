package models

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeWelcome            NotificationType = "welcome"
	NotificationTypeConversionComplete NotificationType = "conversion_complete"
	NotificationTypeConversionError    NotificationType = "conversion_error"
)

// NotificationMessage represents a notification message from RabbitMQ
type NotificationMessage struct {
	ID       string                 `json:"id"`
	Type     NotificationType       `json:"type"`
	UserID   string                 `json:"user_id"`
	Email    string                 `json:"email"`
	Data     map[string]interface{} `json:"data"`
	Priority int                    `json:"priority"` // 1-5, 1 being highest
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	Subject     string
	HTMLBody    string
	TextBody    string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Content  []byte
	MimeType string
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID                   string    `json:"user_id"`
	EmailEnabled             bool      `json:"email_enabled"`
	ConversionCompleteEmail  bool      `json:"conversion_complete_email"`
	ConversionErrorEmail     bool      `json:"conversion_error_email"`
	WelcomeEmail             bool      `json:"welcome_email"`
	UnsubscribeToken         string    `json:"unsubscribe_token"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// DeliveryStatus represents the status of email delivery
type DeliveryStatus struct {
	NotificationID string    `json:"notification_id"`
	UserID         string    `json:"user_id"`
	Email          string    `json:"email"`
	Status         string    `json:"status"` // sent, failed, bounced
	Error          string    `json:"error,omitempty"`
	SentAt         time.Time `json:"sent_at"`
	Attempts       int       `json:"attempts"`
}