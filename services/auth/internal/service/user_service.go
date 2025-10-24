package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/video-converter/auth/internal/auth"
	"github.com/video-converter/auth/internal/models"
)

// UserService handles user-related business logic
type UserService struct {
	db                *gorm.DB
	jwtManager        *auth.JWTManager
	passwordValidator *auth.PasswordValidator
	emailValidator    *auth.EmailValidator
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db:                db,
		jwtManager:        auth.NewJWTManager(),
		passwordValidator: auth.NewPasswordValidator(),
		emailValidator:    auth.NewEmailValidator(),
	}
}

// RegisterUser registers a new user
func (s *UserService) RegisterUser(email, password, firstName, lastName string) (*models.User, error) {
	// Validate email
	if err := s.emailValidator.ValidateEmail(email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Normalize email
	email = s.emailValidator.NormalizeEmail(email)

	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password
	hashedPassword, err := s.passwordValidator.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        email,
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		IsActive:     true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// LoginUser authenticates a user and returns tokens
func (s *UserService) LoginUser(email, password string) (*models.User, string, string, error) {
	// Validate email format
	if err := s.emailValidator.ValidateEmail(email); err != nil {
		return nil, "", "", fmt.Errorf("invalid email: %w", err)
	}

	// Normalize email
	email = s.emailValidator.NormalizeEmail(email)

	// Find user
	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", "", fmt.Errorf("invalid credentials")
		}
		return nil, "", "", fmt.Errorf("failed to find user: %w", err)
	}

	// Verify password
	if err := s.passwordValidator.VerifyPassword(password, user.PasswordHash); err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	accessToken, _, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Limit concurrent sessions (max 5 sessions per user)
	if err := s.LimitConcurrentSessions(user.ID, 5); err != nil {
		return nil, "", "", fmt.Errorf("failed to limit concurrent sessions: %w", err)
	}

	// Store refresh token session
	session := &models.UserSession{
		UserID:    user.ID,
		TokenHash: s.jwtManager.HashToken(refreshToken),
		ExpiresAt: refreshExpiresAt,
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, "", "", fmt.Errorf("failed to create session: %w", err)
	}

	// Clean up expired sessions
	go s.cleanupExpiredSessions(user.ID)

	return &user, accessToken, refreshToken, nil
}

// ValidateToken validates a JWT token
func (s *UserService) ValidateToken(tokenString string) (*models.User, error) {
	// Parse and validate token
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Only accept access tokens for validation
	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Find user
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", claims.UserID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *UserService) RefreshToken(refreshTokenString string) (string, string, error) {
	// Parse and validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshTokenString)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Only accept refresh tokens
	if claims.TokenType != "refresh" {
		return "", "", fmt.Errorf("invalid token type")
	}

	// Check if session exists and is valid
	tokenHash := s.jwtManager.HashToken(refreshTokenString)
	var session models.UserSession
	if err := s.db.Where("user_id = ? AND token_hash = ? AND expires_at > ?", 
		claims.UserID, tokenHash, time.Now()).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", fmt.Errorf("invalid or expired refresh token")
		}
		return "", "", fmt.Errorf("failed to find session: %w", err)
	}

	// Find user
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", claims.UserID, true).First(&user).Error; err != nil {
		return "", "", fmt.Errorf("user not found or inactive")
	}

	// Generate new tokens
	newAccessToken, _, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, newRefreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Invalidate old refresh token by updating session
	// This implements refresh token rotation for better security
	session.TokenHash = s.jwtManager.HashToken(newRefreshToken)
	session.ExpiresAt = newRefreshExpiresAt
	if err := s.db.Save(&session).Error; err != nil {
		return "", "", fmt.Errorf("failed to update session: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// LogoutUser invalidates a refresh token and optionally an access token (logout)
func (s *UserService) LogoutUser(refreshTokenString, accessTokenString string) error {
	// Blacklist access token if provided
	if accessTokenString != "" {
		if err := s.jwtManager.BlacklistToken(accessTokenString); err != nil {
			// Log error but don't fail the logout
			fmt.Printf("Failed to blacklist access token: %v\n", err)
		}
	}

	// Parse refresh token to get user ID
	claims, err := s.jwtManager.ValidateToken(refreshTokenString)
	if err != nil {
		// Even if token is invalid, we should try to clean it up
		tokenHash := s.jwtManager.HashToken(refreshTokenString)
		s.db.Where("token_hash = ?", tokenHash).Delete(&models.UserSession{})
		return nil
	}

	// Delete the session
	tokenHash := s.jwtManager.HashToken(refreshTokenString)
	if err := s.db.Where("user_id = ? AND token_hash = ?", claims.UserID, tokenHash).Delete(&models.UserSession{}).Error; err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// LogoutAllSessions invalidates all sessions for a user
func (s *UserService) LogoutAllSessions(userID uint) error {
	if err := s.db.Where("user_id = ?", userID).Delete(&models.UserSession{}).Error; err != nil {
		return fmt.Errorf("failed to delete all sessions: %w", err)
	}
	return nil
}

// GetActiveSessions returns the number of active sessions for a user
func (s *UserService) GetActiveSessions(userID uint) (int64, error) {
	var count int64
	if err := s.db.Model(&models.UserSession{}).Where("user_id = ? AND expires_at > ?", userID, time.Now()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active sessions: %w", err)
	}
	return count, nil
}

// LimitConcurrentSessions limits the number of concurrent sessions per user
func (s *UserService) LimitConcurrentSessions(userID uint, maxSessions int) error {
	// Get current session count
	count, err := s.GetActiveSessions(userID)
	if err != nil {
		return err
	}

	// If we're at or over the limit, remove oldest sessions
	if int(count) >= maxSessions {
		sessionsToRemove := int(count) - maxSessions + 1
		
		// Find oldest sessions
		var oldSessions []models.UserSession
		if err := s.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
			Order("created_at ASC").
			Limit(sessionsToRemove).
			Find(&oldSessions).Error; err != nil {
			return fmt.Errorf("failed to find old sessions: %w", err)
		}

		// Delete old sessions
		for _, session := range oldSessions {
			if err := s.db.Delete(&session).Error; err != nil {
				return fmt.Errorf("failed to delete old session: %w", err)
			}
		}
	}

	return nil
}

// cleanupExpiredSessions removes expired sessions for a user
func (s *UserService) cleanupExpiredSessions(userID uint) {
	s.db.Where("user_id = ? AND expires_at < ?", userID, time.Now()).Delete(&models.UserSession{})
}