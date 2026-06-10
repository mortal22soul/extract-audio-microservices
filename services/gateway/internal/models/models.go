package models

import (
	"time"
)

// Temporary ObjectID type until MongoDB driver is fixed
type ObjectID string

// Video represents a video document in MongoDB
type Video struct {
	ID               ObjectID       `bson:"_id,omitempty" json:"id"`
	UserID           string         `bson:"userId" json:"userId"`
	OriginalFilename string         `bson:"originalFilename" json:"originalFilename"`
	MimeType         string         `bson:"mimeType" json:"mimeType"`
	Size             int64          `bson:"size" json:"size"`
	UploadedAt       time.Time      `bson:"uploadedAt" json:"uploadedAt"`
	Status           string         `bson:"status" json:"status"`
	ConversionJobID  string         `bson:"conversionJobId,omitempty" json:"conversionJobId,omitempty"`
	MP3FileID        ObjectID       `bson:"mp3FileId,omitempty" json:"mp3FileId,omitempty"`
	Metadata         VideoMetadata  `bson:"metadata" json:"metadata"`
	Analytics        VideoAnalytics `bson:"analytics,omitempty" json:"analytics,omitempty"`
}

// VideoMetadata contains video file metadata
type VideoMetadata struct {
	Duration   int    `bson:"duration" json:"duration"`
	Resolution string `bson:"resolution" json:"resolution"`
	Codec      string `bson:"codec" json:"codec"`
	Bitrate    int    `bson:"bitrate" json:"bitrate"`
}

// VideoAnalytics contains ML analysis results
type VideoAnalytics struct {
	Thumbnails   []string `bson:"thumbnails" json:"thumbnails"`
	QualityScore float32  `bson:"qualityScore" json:"qualityScore"`
	SafetyScore  float32  `bson:"safetyScore" json:"safetyScore"`
	Tags         []string `bson:"tags" json:"tags"`
}

// ConversionJob represents a video conversion job
type ConversionJob struct {
	ID          ObjectID   `bson:"_id,omitempty" json:"id"`
	VideoID     ObjectID   `bson:"videoId" json:"videoId"`
	UserID      string     `bson:"userId" json:"userId"`
	Status      string     `bson:"status" json:"status"`
	Progress    int        `bson:"progress" json:"progress"`
	StartedAt   time.Time  `bson:"startedAt" json:"startedAt"`
	CompletedAt *time.Time `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
	ErrorMsg    string     `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`
}

// Request/Response models for API

type UploadResponse struct {
	VideoID  string `json:"videoId"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

type VideoListResponse struct {
	Videos []Video `json:"videos"`
	Total  int     `json:"total"`
}

type VideoStatusResponse struct {
	VideoID         string         `json:"videoId"`
	Status          string         `json:"status"`
	Progress        int            `json:"progress"`
	ConversionJobID string         `json:"conversionJobId,omitempty"`
	Analytics       VideoAnalytics `json:"analytics,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Validation tags for request validation
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}