package auth

import (
	"fmt"
	"regexp"
	"strings"
)

// EmailValidator handles email validation
type EmailValidator struct {
	emailRegex *regexp.Regexp
}

// NewEmailValidator creates a new email validator
func NewEmailValidator() *EmailValidator {
	// RFC 5322 compliant email regex (simplified)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	return &EmailValidator{
		emailRegex: emailRegex,
	}
}

// ValidateEmail validates an email address
func (e *EmailValidator) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Normalize email
	email = strings.TrimSpace(strings.ToLower(email))

	// Check length
	if len(email) > 254 {
		return fmt.Errorf("email is too long (maximum 254 characters)")
	}

	// Check format
	if !e.emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	// Check for common disposable email domains (expanded list)
	disposableDomains := []string{
		"10minutemail.com", "tempmail.org", "guerrillamail.com",
		"mailinator.com", "throwaway.email", "temp-mail.org",
		"yopmail.com", "maildrop.cc", "sharklasers.com",
		"guerrillamailblock.com", "pokemail.net", "spam4.me",
		"tempail.com", "dispostable.com", "fakeinbox.com",
	}

	emailParts := strings.Split(email, "@")
	if len(emailParts) != 2 {
		return fmt.Errorf("invalid email format")
	}

	domain := emailParts[1]
	for _, disposable := range disposableDomains {
		if domain == disposable {
			return fmt.Errorf("disposable email addresses are not allowed")
		}
	}

	return nil
}

// NormalizeEmail normalizes an email address
func (e *EmailValidator) NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}