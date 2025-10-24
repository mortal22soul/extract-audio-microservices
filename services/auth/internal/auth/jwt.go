package auth

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	blacklist       *TokenBlacklist
}

// NewJWTManager creates a new JWT manager
func NewJWTManager() *JWTManager {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production"
	}

	return &JWTManager{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  15 * time.Minute,  // Access tokens expire in 15 minutes
		refreshTokenTTL: 7 * 24 * time.Hour, // Refresh tokens expire in 7 days
		blacklist:       NewTokenBlacklist(),
	}
}

// GenerateAccessToken generates a new access token
func (j *JWTManager) GenerateAccessToken(userID uint, email string) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.accessTokenTTL)
	
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "video-converter-auth",
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken generates a new refresh token
func (j *JWTManager) GenerateRefreshToken(userID uint, email string) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.refreshTokenTTL)
	
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "video-converter-auth",
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Check if token is blacklisted first
	tokenHash := j.HashToken(tokenString)
	if j.blacklist.IsBlacklisted(tokenHash) {
		return nil, fmt.Errorf("token is blacklisted")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// HashToken creates a SHA256 hash of the token for storage
func (j *JWTManager) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// BlacklistToken adds a token to the blacklist
func (j *JWTManager) BlacklistToken(tokenString string) error {
	// Parse token to get expiry time
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		// Even if token is invalid, we should blacklist it
		tokenHash := j.HashToken(tokenString)
		j.blacklist.BlacklistToken(tokenHash, time.Now().Add(j.accessTokenTTL))
		return nil
	}

	tokenHash := j.HashToken(tokenString)
	expiresAt := time.Unix(claims.ExpiresAt.Unix(), 0)
	j.blacklist.BlacklistToken(tokenHash, expiresAt)
	
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (j *JWTManager) IsTokenBlacklisted(tokenString string) bool {
	tokenHash := j.HashToken(tokenString)
	return j.blacklist.IsBlacklisted(tokenHash)
}