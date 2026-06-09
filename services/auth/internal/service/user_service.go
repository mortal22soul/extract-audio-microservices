package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/video-converter/auth/internal/auth"
	"github.com/video-converter/auth/internal/models"
)

// UserService handles user-related business logic
type UserService struct {
	client            *mongo.Client
	dbName            string
	users             *mongo.Collection
	sessions          *mongo.Collection
	jwtManager        *auth.JWTManager
	passwordValidator *auth.PasswordValidator
	emailValidator    *auth.EmailValidator
}

// NewUserService creates a new user service with Redis-backed JWT manager
func NewUserService(client *mongo.Client, dbName string, redisClient *redis.Client) (*UserService, error) {
	db := client.Database(dbName)

	jwtMgr, err := auth.NewJWTManager(redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}

	return &UserService{
		client:            client,
		dbName:            dbName,
		users:             db.Collection("users"),
		sessions:          db.Collection("user_sessions"),
		jwtManager:        jwtMgr,
		passwordValidator: auth.NewPasswordValidator(),
		emailValidator:    auth.NewEmailValidator(),
	}, nil
}

// RegisterUser registers a new user
func (s *UserService) RegisterUser(email, password, firstName, lastName string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.emailValidator.ValidateEmail(email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	email = s.emailValidator.NormalizeEmail(email)

	// Check if user already exists
	count, err := s.users.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	hashedPassword, err := s.passwordValidator.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           primitive.NewObjectID(),
		Email:        email,
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = s.users.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// LoginUser authenticates a user and returns tokens
func (s *UserService) LoginUser(email, password string) (*models.User, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.emailValidator.ValidateEmail(email); err != nil {
		return nil, "", "", fmt.Errorf("invalid email: %w", err)
	}

	email = s.emailValidator.NormalizeEmail(email)

	var user models.User
	err := s.users.FindOne(ctx, bson.M{"email": email, "is_active": true}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, "", "", fmt.Errorf("invalid credentials")
		}
		return nil, "", "", fmt.Errorf("failed to find user: %w", err)
	}

	if err := s.passwordValidator.VerifyPassword(password, user.PasswordHash); err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	accessToken, _, err := s.jwtManager.GenerateAccessTokenWithID(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshTokenWithID(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.LimitConcurrentSessions(user.ID, 5); err != nil {
		return nil, "", "", fmt.Errorf("failed to limit concurrent sessions: %w", err)
	}

	session := &models.UserSession{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		TokenHash: s.jwtManager.HashToken(refreshToken),
		ExpiresAt: refreshExpiresAt,
		CreatedAt: time.Now(),
	}

	_, err = s.sessions.InsertOne(ctx, session)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create session: %w", err)
	}

	go s.cleanupExpiredSessions(user.ID)

	return &user, accessToken, refreshToken, nil
}

// ValidateToken validates a JWT token
func (s *UserService) ValidateToken(tokenString string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Retrieve string UserID from claims and convert to ObjectID
	var objID primitive.ObjectID
	if claims.UserIDStr != "" {
		objID, err = primitive.ObjectIDFromHex(claims.UserIDStr)
	} else {
		// Temporary fallback if old uint ID is present
		objID, err = primitive.ObjectIDFromHex(fmt.Sprintf("%d", claims.UserID))
	}

	if err != nil {
		return nil, fmt.Errorf("invalid user id in token: %w", err)
	}

	var user models.User
	err = s.users.FindOne(ctx, bson.M{"_id": objID, "is_active": true}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *UserService) RefreshToken(refreshTokenString string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	claims, err := s.jwtManager.ValidateToken(refreshTokenString)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return "", "", fmt.Errorf("invalid token type")
	}

	var objID primitive.ObjectID
	if claims.UserIDStr != "" {
		objID, err = primitive.ObjectIDFromHex(claims.UserIDStr)
	} else {
		objID, err = primitive.ObjectIDFromHex(fmt.Sprintf("%d", claims.UserID))
	}
	if err != nil {
		return "", "", fmt.Errorf("invalid user id in token")
	}

	tokenHash := s.jwtManager.HashToken(refreshTokenString)
	var session models.UserSession
	err = s.sessions.FindOne(ctx, bson.M{
		"user_id":    objID,
		"token_hash": tokenHash,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&session)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", "", fmt.Errorf("invalid or expired refresh token")
		}
		return "", "", fmt.Errorf("failed to find session: %w", err)
	}

	var user models.User
	err = s.users.FindOne(ctx, bson.M{"_id": objID, "is_active": true}).Decode(&user)
	if err != nil {
		return "", "", fmt.Errorf("user not found or inactive")
	}

	newAccessToken, _, err := s.jwtManager.GenerateAccessTokenWithID(user.ID.Hex(), user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, newRefreshExpiresAt, err := s.jwtManager.GenerateRefreshTokenWithID(user.ID.Hex(), user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	_, err = s.sessions.UpdateOne(
		ctx,
		bson.M{"_id": session.ID},
		bson.M{"$set": bson.M{
			"token_hash": s.jwtManager.HashToken(newRefreshToken),
			"expires_at": newRefreshExpiresAt,
		}},
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to update session: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := s.users.FindOne(ctx, bson.M{"_id": userID, "is_active": true}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// LogoutUser invalidates a refresh token and optionally an access token
func (s *UserService) LogoutUser(refreshTokenString, accessTokenString string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if accessTokenString != "" {
		if err := s.jwtManager.BlacklistToken(accessTokenString); err != nil {
			fmt.Printf("Failed to blacklist access token: %v\n", err)
		}
	}

	tokenHash := s.jwtManager.HashToken(refreshTokenString)
	_, err := s.sessions.DeleteMany(ctx, bson.M{"token_hash": tokenHash})
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// LogoutAllSessions invalidates all sessions for a user
func (s *UserService) LogoutAllSessions(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.sessions.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return fmt.Errorf("failed to delete all sessions: %w", err)
	}
	return nil
}

// GetActiveSessions returns the number of active sessions for a user
func (s *UserService) GetActiveSessions(userID primitive.ObjectID) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := s.sessions.CountDocuments(ctx, bson.M{
		"user_id":    userID,
		"expires_at": bson.M{"$gt": time.Now()},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count active sessions: %w", err)
	}
	return count, nil
}

// LimitConcurrentSessions limits the number of concurrent sessions per user
func (s *UserService) LimitConcurrentSessions(userID primitive.ObjectID, maxSessions int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := s.GetActiveSessions(userID)
	if err != nil {
		return err
	}

	if int(count) >= maxSessions {
		sessionsToRemove := int(count) - maxSessions + 1

		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "created_at", Value: 1}})
		findOptions.SetLimit(int64(sessionsToRemove))

		cursor, err := s.sessions.Find(ctx, bson.M{
			"user_id":    userID,
			"expires_at": bson.M{"$gt": time.Now()},
		}, findOptions)
		
		if err != nil {
			return fmt.Errorf("failed to find old sessions: %w", err)
		}
		defer cursor.Close(ctx)

		var oldSessions []models.UserSession
		if err = cursor.All(ctx, &oldSessions); err != nil {
			return fmt.Errorf("failed to decode old sessions: %w", err)
		}

		for _, session := range oldSessions {
			_, err := s.sessions.DeleteOne(ctx, bson.M{"_id": session.ID})
			if err != nil {
				return fmt.Errorf("failed to delete old session: %w", err)
			}
		}
	}

	return nil
}

// cleanupExpiredSessions removes expired sessions for a user
func (s *UserService) cleanupExpiredSessions(userID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.sessions.DeleteMany(ctx, bson.M{
		"user_id":    userID,
		"expires_at": bson.M{"$lt": time.Now()},
	})
}