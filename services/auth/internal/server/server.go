package server

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	pb "github.com/video-converter/shared/proto/gen/go/shared/proto"
	"github.com/video-converter/auth/internal/auth"
	"github.com/video-converter/auth/internal/service"
)

// AuthServer implements the gRPC AuthService
type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	userService *service.UserService
	rateLimiter *auth.RateLimiter
}

// New creates a new AuthServer
func New(db *gorm.DB) *AuthServer {
	return &AuthServer{
		userService: service.NewUserService(db),
		rateLimiter: auth.NewRateLimiter(),
	}
}

// ValidateToken validates a JWT token and returns user information
func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	user, err := s.userService.ValidateToken(req.Token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return &pb.TokenResponse{
			Valid:  false,
			UserId: "",
			Email:  "",
		}, nil
	}

	return &pb.TokenResponse{
		Valid:  true,
		UserId: fmt.Sprintf("%d", user.ID),
		Email:  user.Email,
		ExpiresAt: &pb.Timestamp{
			Seconds: 0, // We don't track individual token expiry in response
			Nanos:   0,
		},
	}, nil
}

// GetUserInfo retrieves user information by user ID
func (s *AuthServer) GetUserInfo(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := strconv.ParseUint(req.UserId, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	user, err := s.userService.GetUserByID(uint(userID))
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.UserResponse{
		UserId:    fmt.Sprintf("%d", user.ID),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
		CreatedAt: &pb.Timestamp{
			Seconds: user.CreatedAt.Unix(),
			Nanos:   int32(user.CreatedAt.Nanosecond()),
		},
	}, nil
}

// RefreshToken refreshes a JWT token
func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.TokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	_, _, err := s.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		log.Printf("Token refresh failed: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	}

	// Note: In a real implementation, we would return the new tokens
	// For now, we just confirm the refresh was successful
	return &pb.TokenResponse{
		Valid:  true,
		UserId: "", // We could parse this from the token if needed
		Email:  "", // We could parse this from the token if needed
	}, nil
}

// RegisterUser registers a new user
func (s *AuthServer) RegisterUser(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Validate required fields
	if req.Email == "" {
		return &pb.RegisterResponse{
			Success: false,
			Message: "Email is required",
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Email is required",
			},
		}, nil
	}

	if req.Password == "" {
		return &pb.RegisterResponse{
			Success: false,
			Message: "Password is required",
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Password is required",
			},
		}, nil
	}

	// Validate input lengths
	if len(req.Email) > 254 {
		return &pb.RegisterResponse{
			Success: false,
			Message: "Email is too long",
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Email must be no more than 254 characters",
			},
		}, nil
	}

	if len(req.FirstName) > 100 {
		return &pb.RegisterResponse{
			Success: false,
			Message: "First name is too long",
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "First name must be no more than 100 characters",
			},
		}, nil
	}

	if len(req.LastName) > 100 {
		return &pb.RegisterResponse{
			Success: false,
			Message: "Last name is too long",
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Last name must be no more than 100 characters",
			},
		}, nil
	}

	// Sanitize input (trim whitespace)
	email := strings.TrimSpace(req.Email)
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)

	user, err := s.userService.RegisterUser(email, req.Password, firstName, lastName)
	if err != nil {
		log.Printf("User registration failed: %v", err)
		
		// Determine error code based on error message
		errorCode := "REGISTRATION_FAILED"
		if strings.Contains(err.Error(), "already exists") {
			errorCode = "EMAIL_ALREADY_EXISTS"
		} else if strings.Contains(err.Error(), "invalid email") {
			errorCode = "INVALID_EMAIL"
		} else if strings.Contains(err.Error(), "password") {
			errorCode = "INVALID_PASSWORD"
		}

		return &pb.RegisterResponse{
			Success: false,
			Message: "Registration failed",
			Error: &pb.Error{
				Code:    errorCode,
				Message: err.Error(),
			},
		}, nil
	}

	log.Printf("User registered successfully: %s (ID: %d)", user.Email, user.ID)
	return &pb.RegisterResponse{
		Success: true,
		UserId:  fmt.Sprintf("%d", user.ID),
		Message: "User registered successfully",
	}, nil
}

// LoginUser authenticates a user
func (s *AuthServer) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate required fields
	if req.Email == "" {
		return &pb.LoginResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Email is required",
			},
		}, nil
	}

	if req.Password == "" {
		return &pb.LoginResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Password is required",
			},
		}, nil
	}

	// Validate input lengths
	if len(req.Email) > 254 {
		return &pb.LoginResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Email is too long",
			},
		}, nil
	}

	if len(req.Password) > 128 {
		return &pb.LoginResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Password is too long",
			},
		}, nil
	}

	// Sanitize input
	email := strings.TrimSpace(req.Email)

	// Check rate limiting
	if s.rateLimiter.IsBlocked(email) {
		log.Printf("Login blocked due to rate limiting for email: %s", email)
		return &pb.LoginResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "RATE_LIMITED",
				Message: "Too many failed login attempts. Please try again later.",
			},
		}, nil
	}

	user, accessToken, refreshToken, err := s.userService.LoginUser(email, req.Password)
	if err != nil {
		// Record failed attempt for rate limiting
		blocked := s.rateLimiter.RecordAttempt(email)
		
		// Log failed login attempt (but don't log the password)
		log.Printf("Login failed for email %s: %v", email, err)
		
		errorMessage := "Invalid credentials"
		if blocked {
			errorMessage = "Too many failed attempts. Account temporarily blocked."
		} else {
			remaining := s.rateLimiter.GetRemainingAttempts(email)
			if remaining <= 2 {
				errorMessage = fmt.Sprintf("Invalid credentials. %d attempts remaining.", remaining)
			}
		}
		
		return &pb.LoginResponse{
			Success: false,
			Error: &pb.Error{
				Code:    "AUTHENTICATION_FAILED",
				Message: errorMessage,
			},
		}, nil
	}

	// Record successful login (resets rate limiting)
	s.rateLimiter.RecordSuccess(email)
	log.Printf("User logged in successfully: %s (ID: %d)", user.Email, user.ID)
	return &pb.LoginResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &pb.UserResponse{
			UserId:    fmt.Sprintf("%d", user.ID),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			IsActive:  user.IsActive,
			CreatedAt: &pb.Timestamp{
				Seconds: user.CreatedAt.Unix(),
				Nanos:   int32(user.CreatedAt.Nanosecond()),
			},
		},
	}, nil
}

// LogoutUser logs out a user by invalidating their tokens
func (s *AuthServer) LogoutUser(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.RefreshToken == "" {
		return &pb.LogoutResponse{
			Success: false,
			Message: "Refresh token is required",
			Error: &pb.Error{
				Code:    "INVALID_ARGUMENT",
				Message: "Refresh token is required",
			},
		}, nil
	}

	// Logout user (invalidate tokens)
	err := s.userService.LogoutUser(req.RefreshToken, req.AccessToken)
	if err != nil {
		log.Printf("Logout failed: %v", err)
		return &pb.LogoutResponse{
			Success: false,
			Message: "Logout failed",
			Error: &pb.Error{
				Code:    "LOGOUT_FAILED",
				Message: err.Error(),
			},
		}, nil
	}

	log.Printf("User logged out successfully")
	return &pb.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}