package pipeline

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/video-converter/converter/internal/ffmpeg"
	"github.com/video-converter/converter/internal/filemanager"
	"github.com/video-converter/converter/internal/storage"
)

type Pipeline struct {
	mongoClient *storage.MongoClient
	redisClient *storage.RedisClient
	ffmpeg      *ffmpeg.FFmpeg
	fileManager *filemanager.FileManager
}

type ConversionRequest struct {
	JobID     string
	UserID    string
	VideoID   string
	Filename  string
	Options   ConversionOptions
}

type ConversionOptions struct {
	AudioBitrate    string
	AudioSampleRate string
	Quality         string // "low", "medium", "high"
}

type ProgressCallback func(progress int, status, message string)

func New(mongoClient *storage.MongoClient, redisClient *storage.RedisClient, 
	ffmpeg *ffmpeg.FFmpeg, fileManager *filemanager.FileManager) *Pipeline {
	return &Pipeline{
		mongoClient: mongoClient,
		redisClient: redisClient,
		ffmpeg:      ffmpeg,
		fileManager: fileManager,
	}
}

func (p *Pipeline) ProcessVideo(ctx context.Context, req ConversionRequest, onProgress ProgressCallback) error {
	log.Printf("Starting video processing pipeline for job: %s", req.JobID)

	// Step 1: Download video from GridFS
	onProgress(10, "processing", "Downloading video from storage...")
	inputPath, fileInfo, err := p.downloadVideo(ctx, req.VideoID, req.JobID)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}
	defer p.fileManager.DeleteFile(inputPath)

	// Step 2: Analyze video properties
	onProgress(20, "processing", "Analyzing video properties...")
	videoInfo, err := p.ffmpeg.GetVideoInfo(ctx, inputPath)
	if err != nil {
		return fmt.Errorf("failed to analyze video: %w", err)
	}

	log.Printf("Video analysis for job %s: duration=%v, resolution=%s, bitrate=%d",
		req.JobID, videoInfo.Duration, videoInfo.Resolution, videoInfo.Bitrate)

	// Step 3: Prepare conversion options
	conversionOpts := p.prepareConversionOptions(req.Options, videoInfo)

	// Step 4: Convert video to MP3
	onProgress(25, "processing", "Starting audio extraction...")
	outputPath, err := p.convertToMP3(ctx, inputPath, req.JobID, conversionOpts, func(progress int) {
		// Map FFmpeg progress (0-100) to our progress range (25-85)
		adjustedProgress := 25 + (progress * 60 / 100)
		onProgress(adjustedProgress, "processing", fmt.Sprintf("Converting audio... %d%%", progress))
	})
	if err != nil {
		return fmt.Errorf("failed to convert video: %w", err)
	}
	defer p.fileManager.DeleteFile(outputPath)

	// Step 5: Upload converted MP3 to GridFS
	onProgress(90, "processing", "Uploading converted audio file...")
	mp3FileID, err := p.uploadMP3(ctx, outputPath, req, fileInfo)
	if err != nil {
		return fmt.Errorf("failed to upload MP3: %w", err)
	}

	// Step 6: Update video document with MP3 reference
	onProgress(95, "processing", "Updating video metadata...")
	if err := p.updateVideoDocument(ctx, req.VideoID, mp3FileID, videoInfo); err != nil {
		log.Printf("Failed to update video document: %v", err)
		// Don't fail the conversion for this
	}

	onProgress(100, "completed", "Conversion completed successfully")
	log.Printf("Video processing pipeline completed for job: %s, MP3 ID: %s", req.JobID, mp3FileID.Hex())
	return nil
}

func (p *Pipeline) downloadVideo(ctx context.Context, videoID, jobID string) (string, *storage.FileInfo, error) {
	// Parse video ID
	objectID, err := primitive.ObjectIDFromHex(videoID)
	if err != nil {
		return "", nil, fmt.Errorf("invalid video ID: %w", err)
	}

	// Download from GridFS
	videoStream, fileInfo, err := p.mongoClient.DownloadFile(ctx, objectID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download from GridFS: %w", err)
	}
	defer videoStream.Close()

	// Create temporary file
	inputExt := p.fileManager.GetFileExtension(fileInfo.Filename)
	inputPath, err := p.fileManager.CreateTempFile(jobID+"_input", inputExt)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Save to temporary file
	if err := p.fileManager.SaveStreamToFile(videoStream, inputPath); err != nil {
		p.fileManager.DeleteFile(inputPath)
		return "", nil, fmt.Errorf("failed to save video to temp file: %w", err)
	}

	log.Printf("Downloaded video %s to %s (size: %d bytes)", videoID, inputPath, fileInfo.Length)
	return inputPath, fileInfo, nil
}

func (p *Pipeline) prepareConversionOptions(opts ConversionOptions, videoInfo *ffmpeg.VideoInfo) ffmpeg.ConversionOptions {
	conversionOpts := ffmpeg.ConversionOptions{
		AudioBitrate:    opts.AudioBitrate,
		AudioSampleRate: opts.AudioSampleRate,
	}

	// Set quality-based defaults if not specified
	if conversionOpts.AudioBitrate == "" {
		switch opts.Quality {
		case "high":
			conversionOpts.AudioBitrate = "320k"
		case "medium":
			conversionOpts.AudioBitrate = "192k"
		case "low":
			conversionOpts.AudioBitrate = "128k"
		default:
			conversionOpts.AudioBitrate = "192k" // Default
		}
	}

	if conversionOpts.AudioSampleRate == "" {
		// Use higher sample rate for high quality
		if opts.Quality == "high" {
			conversionOpts.AudioSampleRate = "48000"
		} else {
			conversionOpts.AudioSampleRate = "44100"
		}
	}

	return conversionOpts
}

func (p *Pipeline) convertToMP3(ctx context.Context, inputPath, jobID string, 
	opts ffmpeg.ConversionOptions, onProgress func(int)) (string, error) {
	
	// Create output file
	outputPath, err := p.fileManager.CreateTempFile(jobID+"_output", ".mp3")
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}

	// Set up conversion options
	opts.InputPath = inputPath
	opts.OutputPath = outputPath
	opts.OnProgress = onProgress

	// Perform conversion
	if err := p.ffmpeg.ConvertToMP3(ctx, opts); err != nil {
		p.fileManager.DeleteFile(outputPath)
		return "", fmt.Errorf("FFmpeg conversion failed: %w", err)
	}

	// Verify output file exists and has content
	size, err := p.fileManager.GetFileSize(outputPath)
	if err != nil {
		p.fileManager.DeleteFile(outputPath)
		return "", fmt.Errorf("failed to verify output file: %w", err)
	}

	if size == 0 {
		p.fileManager.DeleteFile(outputPath)
		return "", fmt.Errorf("conversion produced empty file")
	}

	log.Printf("Successfully converted to MP3: %s (size: %d bytes)", outputPath, size)
	return outputPath, nil
}

func (p *Pipeline) uploadMP3(ctx context.Context, mp3Path string, req ConversionRequest, 
	originalFileInfo *storage.FileInfo) (primitive.ObjectID, error) {
	
	// Open MP3 file
	mp3File, err := p.fileManager.OpenFile(mp3Path)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to open MP3 file: %w", err)
	}
	defer mp3File.Close()

	// Generate MP3 filename
	mp3Filename := p.generateMP3Filename(originalFileInfo.Filename)

	// Get file size
	mp3Size, err := p.fileManager.GetFileSize(mp3Path)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to get MP3 file size: %w", err)
	}

	// Prepare metadata
	metadata := bson.M{
		"original_video_id":   originalFileInfo.ID,
		"original_filename":   originalFileInfo.Filename,
		"job_id":             req.JobID,
		"user_id":            req.UserID,
		"conversion_date":    time.Now(),
		"audio_bitrate":      req.Options.AudioBitrate,
		"audio_sample_rate":  req.Options.AudioSampleRate,
		"quality":            req.Options.Quality,
		"file_size":          mp3Size,
		"content_type":       "audio/mpeg",
	}

	// Upload to GridFS
	mp3FileID, err := p.mongoClient.UploadFile(ctx, mp3Filename, mp3File, metadata)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to upload to GridFS: %w", err)
	}

	log.Printf("Uploaded MP3 file: %s (ID: %s, size: %d bytes)", 
		mp3Filename, mp3FileID.Hex(), mp3Size)
	return mp3FileID, nil
}

func (p *Pipeline) updateVideoDocument(ctx context.Context, videoID string, mp3FileID primitive.ObjectID, 
	videoInfo *ffmpeg.VideoInfo) error {
	
	update := bson.M{
		"mp3_file_id":    mp3FileID,
		"status":         "completed",
		"updated_at":     time.Now(),
		"processing_info": bson.M{
			"duration":    videoInfo.Duration.Seconds(),
			"resolution":  videoInfo.Resolution,
			"bitrate":     videoInfo.Bitrate,
			"format":      videoInfo.Format,
		},
	}

	return p.mongoClient.UpdateConversionJob(ctx, videoID, update)
}

func (p *Pipeline) generateMP3Filename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	nameWithoutExt := originalFilename[:len(originalFilename)-len(ext)]
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.mp3", nameWithoutExt, timestamp)
}

// GetEstimatedDuration estimates conversion time based on video duration and quality
func (p *Pipeline) GetEstimatedDuration(videoDuration time.Duration, quality string) time.Duration {
	// Base conversion rate: typically 2-5x faster than real-time for audio extraction
	var conversionRate float64
	
	switch quality {
	case "high":
		conversionRate = 2.0 // Slower for higher quality
	case "medium":
		conversionRate = 3.0
	case "low":
		conversionRate = 4.0 // Faster for lower quality
	default:
		conversionRate = 3.0
	}

	estimatedTime := time.Duration(float64(videoDuration) / conversionRate)
	
	// Add buffer time for file operations
	bufferTime := 30 * time.Second
	return estimatedTime + bufferTime
}