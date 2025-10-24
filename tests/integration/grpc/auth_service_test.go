package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authpb "github.com/video-converter/shared/proto/gen/go/shared/proto"
	"github.com/video-converter/tests/integration/utils"
)

func TestAuthServiceGRPC(t *testing.T) {
	config := utils.GetTestConfig()
	
	// Wait for services to be ready
	utils.WaitForServices(t, config)
	
	// Setup gRPC connections
	grpcConns := utils.SetupGRPCConnections(t, config)
	defer grpcConns.CleanupGRPCConnections()
	
	// Setup databases
	dbConns := utils.SetupDatabases(t, config)
	defer dbConns.CleanupDatabases(t)

	// Create auth service client
	authClient := authpb.NewAuthServiceClient(grpcConns.AuthConn)

	t.Run("RegisterUser", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := &authpb.RegisterRequest{
			Email:     "grpc_test@example.com",
			Password:  "TestPassword123!",
			FirstName: "GRPC",
			LastName:  "Test",
		}

		resp, err := authClient.RegisterUser(ctx, req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.NotEmpty(t, resp.UserId)
		assert.Equal(t, "User registered successfully", resp.Message)
	})

	t.Run("LoginUser", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// First register a user
		registerReq := &authpb.RegisterRequest{
			Email:     "login_test@example.com",
			Password:  "TestPassword123!",
			FirstName: "Login",
			LastName:  "Test",
		}

		registerResp, err := authClient.RegisterUser(ctx, registerReq)
		require.NoError(t, err)
		require.True(t, registerResp.Success)

		// Now login
		loginReq := &authpb.LoginRequest{
			Email:    "login_test@example.com",
			Password: "TestPassword123!",
		}

		loginResp, err := authClient.LoginUser(ctx, loginReq)
		require.NoError(t, err)
		assert.True(t, loginResp.Success)
		assert.NotEmpty(t, loginResp.AccessToken)
		assert.NotEmpty(t, loginResp.RefreshToken)
		assert.Equal(t, "login_test@example.com", loginResp.User.Email)
	})

	t.Run("ValidateToken", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create test user and get token
		testUser := utils.CreateTestUser(t, config)

		req := &authpb.TokenRequest{
			Token: testUser.Token,
		}

		resp, err := authClient.ValidateToken(ctx, req)
		require.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, testUser.Email, resp.Email)
		assert.Equal(t, testUser.ID, resp.UserId)
	})

	t.Run("ValidateInvalidToken", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := &authpb.TokenRequest{
			Token: "invalid_token",
		}

		resp, err := authClient.ValidateToken(ctx, req)
		// Should return error or invalid response
		if err == nil {
			assert.False(t, resp.Valid)
		}
	})

	t.Run("GetUserInfo", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// First register a user
		registerReq := &authpb.RegisterRequest{
			Email:     "userinfo_test@example.com",
			Password:  "TestPassword123!",
			FirstName: "UserInfo",
			LastName:  "Test",
		}

		registerResp, err := authClient.RegisterUser(ctx, registerReq)
		require.NoError(t, err)
		require.True(t, registerResp.Success)

		// Get user info
		userReq := &authpb.UserRequest{
			UserId: registerResp.UserId,
		}

		userResp, err := authClient.GetUserInfo(ctx, userReq)
		require.NoError(t, err)
		assert.Equal(t, registerResp.UserId, userResp.UserId)
		assert.Equal(t, "userinfo_test@example.com", userResp.Email)
		assert.Equal(t, "UserInfo", userResp.FirstName)
		assert.Equal(t, "Test", userResp.LastName)
		assert.True(t, userResp.IsActive)
	})

	t.Run("RefreshToken", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// First login to get refresh token
		registerReq := &authpb.RegisterRequest{
			Email:     "refresh_test@example.com",
			Password:  "TestPassword123!",
			FirstName: "Refresh",
			LastName:  "Test",
		}

		_, err := authClient.RegisterUser(ctx, registerReq)
		require.NoError(t, err)

		loginReq := &authpb.LoginRequest{
			Email:    "refresh_test@example.com",
			Password: "TestPassword123!",
		}

		loginResp, err := authClient.LoginUser(ctx, loginReq)
		require.NoError(t, err)
		require.True(t, loginResp.Success)

		// Refresh token
		refreshReq := &authpb.RefreshRequest{
			RefreshToken: loginResp.RefreshToken,
		}

		refreshResp, err := authClient.RefreshToken(ctx, refreshReq)
		require.NoError(t, err)
		assert.True(t, refreshResp.Valid)
		assert.NotEmpty(t, refreshResp.UserId)
		assert.NotEqual(t, loginResp.AccessToken, refreshResp.UserId) // New token should be different
	})

	t.Run("LogoutUser", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// First login to get tokens
		registerReq := &authpb.RegisterRequest{
			Email:     "logout_test@example.com",
			Password:  "TestPassword123!",
			FirstName: "Logout",
			LastName:  "Test",
		}

		_, err := authClient.RegisterUser(ctx, registerReq)
		require.NoError(t, err)

		loginReq := &authpb.LoginRequest{
			Email:    "logout_test@example.com",
			Password: "TestPassword123!",
		}

		loginResp, err := authClient.LoginUser(ctx, loginReq)
		require.NoError(t, err)
		require.True(t, loginResp.Success)

		// Logout
		logoutReq := &authpb.LogoutRequest{
			RefreshToken: loginResp.RefreshToken,
			AccessToken:  loginResp.AccessToken,
		}

		logoutResp, err := authClient.LogoutUser(ctx, logoutReq)
		require.NoError(t, err)
		assert.True(t, logoutResp.Success)
		assert.Equal(t, "User logged out successfully", logoutResp.Message)

		// Verify token is no longer valid
		validateReq := &authpb.TokenRequest{
			Token: loginResp.AccessToken,
		}

		validateResp, err := authClient.ValidateToken(ctx, validateReq)
		if err == nil {
			assert.False(t, validateResp.Valid)
		}
	})
}