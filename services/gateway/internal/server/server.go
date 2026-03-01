package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/video-converter/gateway/internal/clients"
	"github.com/video-converter/gateway/internal/config"
	"github.com/video-converter/gateway/internal/middleware"
	"github.com/video-converter/gateway/internal/models"
	"github.com/video-converter/gateway/internal/storage"
)

type Server struct {
	router      *gin.Engine
	config      *config.Config
	mongodb     *storage.MongoClient
	minio       *storage.MinIOClient
	grpcClients *clients.GRPCClients
	validator   *validator.Validate
	rateLimiter *middleware.RateLimiter
}

func New() *Server {
	// Load configuration
	cfg := config.Load()

	// Initialize MongoDB
	mongodb, err := storage.NewMongoClient(cfg.MongoURI, "videoconverter")
	if err != nil {
		log.Printf("Warning: Failed to connect to MongoDB: %v", err)
		mongodb = nil
	}

	// Initialize MinIO
	minioClient, err := storage.NewMinIOClient(
		cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey,
		cfg.MinIOBucket, cfg.MinIOUseSSL,
	)
	if err != nil {
		log.Printf("Warning: Failed to connect to MinIO: %v", err)
		minioClient = nil
	}

	// Initialize gRPC clients
	grpcClients, err := clients.NewGRPCClients(cfg.AuthGRPCAddr, cfg.AnalyticsGRPCAddr)
	if err != nil {
		log.Printf("Warning: Failed to connect to gRPC services: %v", err)
		// Continue without gRPC clients for now
	}

	// Initialize validator
	validate := validator.New()

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerMinute, cfg.RateLimit.BurstSize)
	rateLimiter.CleanupVisitors()

	// Set Gin mode
	if cfg.Port == "8080" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(rateLimiter.Middleware())

	server := &Server{
		router:      router,
		config:      cfg,
		mongodb:     mongodb,
		minio:       minioClient,
		grpcClients: grpcClients,
		validator:   validate,
		rateLimiter: rateLimiter,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", s.login)
			auth.POST("/register", s.register)
		}

		// Protected routes (authentication required)
		if s.grpcClients != nil {
			protected := v1.Group("")
			protected.Use(middleware.JWTAuth(s.grpcClients))
			{
				// Video management routes
				videos := protected.Group("/videos")
				{
					videos.POST("/upload", s.uploadVideo)
					videos.GET("", s.listVideos)
					videos.GET("/:id", s.getVideo)
					videos.GET("/:id/status", s.getVideoStatus)
					videos.GET("/:id/download", s.downloadVideo)
					videos.GET("/:id/analytics", s.getVideoAnalytics)
					videos.DELETE("/:id", s.deleteVideo)
				}

				// Analytics routes
				analytics := protected.Group("/analytics")
				{
					analytics.GET("/recommendations", s.getRecommendations)
				}

				// User routes
				user := protected.Group("/user")
				{
					user.GET("/profile", s.getUserProfile)
					user.POST("/logout", s.logout)
				}
			}
		}
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	status := gin.H{
		"status":  "healthy",
		"service": "gateway",
		"mongodb": "connected",
	}

	if s.grpcClients != nil {
		status["grpc"] = "connected"
	} else {
		status["grpc"] = "disconnected"
	}

	c.JSON(http.StatusOK, status)
}

// Authentication handlers
func (s *Server) login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := s.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	// For now, return a stub response since gRPC is disabled
	if s.grpcClients == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "Authentication service unavailable",
			Code:  "AUTH_SERVICE_UNAVAILABLE",
		})
		return
	}

	// Will call auth service via gRPC when proto issues are resolved
	// For now, return a mock successful login
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"access_token": "mock-jwt-token",
		"user": gin.H{
			"id":    "mock-user-id",
			"email": req.Email,
		},
	})
}

func (s *Server) register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := s.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	// For now, return a stub response since gRPC is disabled
	if s.grpcClients == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "Authentication service unavailable",
			Code:  "AUTH_SERVICE_UNAVAILABLE",
		})
		return
	}

	// Will call auth service via gRPC when proto issues are resolved
	// For now, return a mock successful registration
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"user_id": "mock-user-id",
		"message": "User registered successfully",
	})
}

func (s *Server) uploadVideo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(s.config.MaxFileSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Failed to parse multipart form",
			Code:    "INVALID_FORM",
			Details: err.Error(),
		})
		return
	}

	// Get the file from form
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "No video file provided",
			Code:    "MISSING_FILE",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > s.config.MaxFileSize {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "File size exceeds maximum allowed",
			Code:    "FILE_TOO_LARGE",
			Details: fmt.Sprintf("Max size: %d bytes", s.config.MaxFileSize),
		})
		return
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !s.isValidVideoType(contentType) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid file type",
			Code:    "INVALID_FILE_TYPE",
			Details: "Only video files are allowed",
		})
		return
	}

	// Ensure storage is available
	if s.mongodb == nil || s.minio == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "Storage service unavailable",
			Code:  "STORAGE_UNAVAILABLE",
		})
		return
	}

	// Create video record
	video := &models.Video{
		UserID:           userID.(string),
		OriginalFilename: header.Filename,
		MimeType:         contentType,
		Size:             header.Size,
		Status:           "uploading",
	}

	if err := s.mongodb.CreateVideo(c.Request.Context(), video); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create video record",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	objectKey := storage.VideoObjectKey(string(video.ID), video.OriginalFilename)
	if err := s.minio.UploadFile(c.Request.Context(), objectKey, file, header.Size, contentType); err != nil {
		s.mongodb.UpdateVideo(c.Request.Context(), string(video.ID), map[string]interface{}{"status": "failed"})
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to upload to storage",
			Code:    "UPLOAD_ERROR",
			Details: err.Error(),
		})
		return
	}

	video.Status = "uploaded"
	if err := s.mongodb.UpdateVideo(c.Request.Context(), string(video.ID), map[string]interface{}{"status": "uploaded"}); err != nil {
		log.Printf("Failed to update video status after upload: %v", err)
	}

	c.JSON(http.StatusCreated, models.UploadResponse{
		VideoID:  string(video.ID),
		Filename: video.OriginalFilename,
		Size:     video.Size,
		Status:   video.Status,
		Message:  "Video uploaded successfully",
	})
}

func (s *Server) listVideos(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	// Parse query parameters
	limit := s.parseIntQuery(c, "limit", 10)
	offset := s.parseIntQuery(c, "offset", 0)
	status := c.Query("status") // Optional filter by status

	// Ensure storage is available
	if s.mongodb == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "Storage service unavailable",
			Code:  "STORAGE_UNAVAILABLE",
		})
		return
	}

	videos, err := s.mongodb.GetUserVideos(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve videos",
			Code:  "DB_ERROR",
		})
		return
	}

	// Simple status filter (since GetUserVideos doesn't accept status filter right now)
	var filteredVideos []models.Video
	for _, v := range videos {
		if status == "" || v.Status == status {
			filteredVideos = append(filteredVideos, v)
		}
	}

	c.JSON(http.StatusOK, models.VideoListResponse{
		Videos: filteredVideos,
		Total:  len(filteredVideos),
	})
}

func (s *Server) getVideo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Video ID is required",
			Code:  "MISSING_VIDEO_ID",
		})
		return
	}

	if s.mongodb == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "Storage service unavailable",
			Code:  "STORAGE_UNAVAILABLE",
		})
		return
	}

	video, err := s.mongodb.GetVideo(c.Request.Context(), videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Video not found",
			Code:  "NOT_FOUND",
		})
		return
	}

	if video.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied",
			Code:  "FORBIDDEN",
		})
		return
	}

	c.JSON(http.StatusOK, video)
}

func (s *Server) getVideoStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Video ID is required",
			Code:  "MISSING_VIDEO_ID",
		})
		return
	}

	if s.mongodb == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "Storage service unavailable",
			Code:  "STORAGE_UNAVAILABLE",
		})
		return
	}

	video, err := s.mongodb.GetVideo(c.Request.Context(), videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Video not found",
			Code:  "NOT_FOUND",
		})
		return
	}

	if video.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied",
			Code:  "FORBIDDEN",
		})
		return
	}

	response := models.VideoStatusResponse{
		VideoID:         string(video.ID),
		Status:          video.Status,
		ConversionJobID: video.ConversionJobID,
		Analytics:       video.Analytics,
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) downloadVideo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Video ID is required",
			Code:  "MISSING_VIDEO_ID",
		})
		return
	}

	if s.mongodb == nil || s.minio == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "Storage service unavailable",
			Code:  "STORAGE_UNAVAILABLE",
		})
		return
	}

	video, err := s.mongodb.GetVideo(c.Request.Context(), videoID)
	if err != nil || video.UserID != userID.(string) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Video not found or access denied",
			Code:  "NOT_FOUND",
		})
		return
	}

	objectKey := storage.VideoObjectKey(string(video.ID), video.OriginalFilename)
	reader, err := s.minio.DownloadFile(c.Request.Context(), objectKey)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "File not found in storage",
			Code:  "FILE_NOT_FOUND",
		})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", video.OriginalFilename))
	c.Header("Content-Type", video.MimeType)
	c.Header("Content-Length", strconv.FormatInt(video.Size, 10))
	c.DataFromReader(http.StatusOK, video.Size, video.MimeType, reader, nil)
}

func (s *Server) deleteVideo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Video ID is required",
			Code:  "MISSING_VIDEO_ID",
		})
		return
	}

	if s.mongodb == nil || s.minio == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "Storage service unavailable",
			Code:  "STORAGE_UNAVAILABLE",
		})
		return
	}

	video, err := s.mongodb.GetVideo(c.Request.Context(), videoID)
	if err != nil || video.UserID != userID.(string) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Video not found or access denied",
			Code:  "NOT_FOUND",
		})
		return
	}

	objectKey := storage.VideoObjectKey(string(video.ID), video.OriginalFilename)
	s.minio.DeleteFile(c.Request.Context(), objectKey)

	mp3Key := storage.MP3ObjectKey(string(video.ID))
	s.minio.DeleteFile(c.Request.Context(), mp3Key)

	s.mongodb.DeleteVideo(c.Request.Context(), videoID)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Video deleted successfully",
		"video_id": videoID,
		"user_id":  userID,
	})
}

func (s *Server) getUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	userEmail, _ := c.Get("userEmail")

	// For now, return stub data since gRPC is disabled
	if s.grpcClients == nil {
		c.JSON(http.StatusOK, gin.H{
			"user_id":    userID,
			"email":      userEmail,
			"first_name": "Stub",
			"last_name":  "User",
			"is_active":  true,
		})
		return
	}

	// Will call auth service via gRPC to get full user profile
	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"email":      userEmail,
		"first_name": "Stub",
		"last_name":  "User",
		"is_active":  true,
	})
}

func (s *Server) logout(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Authorization header required",
			Code:  "MISSING_AUTH_HEADER",
		})
		return
	}

	// For now, return success since gRPC is disabled
	if s.grpcClients == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Logged out successfully",
		})
		return
	}

	// Will call auth service via gRPC to invalidate token
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

func (s *Server) Run(addr string) error {
	log.Printf("Starting Gateway Service on %s", addr)
	return s.router.Run(addr)
}

func (s *Server) Shutdown() {
	if s.mongodb != nil {
		s.mongodb.Close(context.Background())
	}
	if s.grpcClients != nil {
		s.grpcClients.Close()
	}
}

// Helper methods

func (s *Server) isValidVideoType(contentType string) bool {
	validTypes := []string{
		"video/mp4",
		"video/avi",
		"video/quicktime",
		"video/x-msvideo",
		"video/x-matroska",
		"video/webm",
	}

	for _, validType := range validTypes {
		if strings.Contains(contentType, validType) {
			return true
		}
	}
	return false
}

func (s *Server) parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

// Analytics handlers

func (s *Server) getVideoAnalytics(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Video ID is required",
			Code:  "MISSING_VIDEO_ID",
		})
		return
	}

	// For now, return stub analytics since gRPC is disabled
	if s.grpcClients == nil {
		mockAnalytics := models.VideoAnalytics{
			QualityScore: 8.5,
			SafetyScore:  9.2,
			Tags:         []string{"music", "entertainment", "high-quality"},
			Thumbnails:   []string{"/thumbnails/thumb1.jpg", "/thumbnails/thumb2.jpg", "/thumbnails/thumb3.jpg"},
		}

		c.JSON(http.StatusOK, gin.H{
			"video_id":  videoID,
			"user_id":   userID,
			"analytics": mockAnalytics,
		})
		return
	}

	// Will call analytics service via gRPC when proto issues are resolved
	c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
		Error: "Analytics service unavailable",
		Code:  "ANALYTICS_SERVICE_UNAVAILABLE",
	})
}

func (s *Server) getRecommendations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
			Code:  "NOT_AUTHENTICATED",
		})
		return
	}

	limit := s.parseIntQuery(c, "limit", 5)

	// For now, return stub recommendations since gRPC is disabled
	if s.grpcClients == nil {
		mockRecommendations := []gin.H{
			{
				"video_id":         "rec-video-1",
				"title":           "Similar Music Video",
				"similarity_score": 0.85,
				"tags":            []string{"music", "pop"},
				"thumbnail_url":   "/thumbnails/rec1.jpg",
			},
			{
				"video_id":         "rec-video-2",
				"title":           "Entertainment Content",
				"similarity_score": 0.78,
				"tags":            []string{"entertainment", "comedy"},
				"thumbnail_url":   "/thumbnails/rec2.jpg",
			},
		}

		// Apply limit
		if limit < len(mockRecommendations) {
			mockRecommendations = mockRecommendations[:limit]
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":         userID,
			"recommendations": mockRecommendations,
			"total":          len(mockRecommendations),
		})
		return
	}

	// Will call analytics service via gRPC when proto issues are resolved
	c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
		Error: "Analytics service unavailable",
		Code:  "ANALYTICS_SERVICE_UNAVAILABLE",
	})
}