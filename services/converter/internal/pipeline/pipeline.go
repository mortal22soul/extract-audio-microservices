package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/video-converter/converter/internal/ffmpeg"
	"github.com/video-converter/converter/internal/filemanager"
	"github.com/video-converter/converter/internal/storage"
)

type Pipeline struct {
	mongoClient *storage.MongoClient
	minioClient *storage.MinIOClient
	redisClient *storage.RedisClient
	ffmpeg      *ffmpeg.FFmpeg
	fileManager *filemanager.FileManager
}

type ConversionRequest struct {
	JobID    string
	UserID   string
	VideoID  string
	Filename string
	Options  ConversionOptions
}

type ConversionOptions struct {
	AudioBitrate    string
	AudioSampleRate string
	Quality         string // "low", "medium", "high"
}

type ProgressCallback func(progress int, status, message string)

func New(mongoClient *storage.MongoClient, minioClient *storage.MinIOClient,
	redisClient *storage.RedisClient, ffmpeg *ffmpeg.FFmpeg,
	fileManager *filemanager.FileManager) *Pipeline {
	return &Pipeline{
		mongoClient: mongoClient,
		minioClient: minioClient,
		redisClient: redisClient,
		ffmpeg:      ffmpeg,
		fileManager: fileManager,
	}
}

func (p *Pipeline) ProcessVideo(ctx context.Context, req ConversionRequest, onProgress ProgressCallback) error {
	log.Printf("Starting video processing pipeline for job: %s", req.JobID)

	// Step 1: Download video from MinIO
	onProgress(10, "processing", "Downloading video from storage...")
	inputPath, err := p.downloadVideo(ctx, req.VideoID, req.Filename, req.JobID)
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
		adjustedProgress := 25 + (progress * 60 / 100)
		onProgress(adjustedProgress, "processing", fmt.Sprintf("Converting audio... %d%%", progress))
	})
	if err != nil {
		return fmt.Errorf("failed to convert video: %w", err)
	}
	defer p.fileManager.DeleteFile(outputPath)

	// Step 5: Upload converted MP3 to MinIO
	onProgress(90, "processing", "Uploading converted audio file...")
	mp3ObjectKey, err := p.uploadMP3(ctx, outputPath, req)
	if err != nil {
		return fmt.Errorf("failed to upload MP3: %w", err)
	}

	// Step 6: Update conversion job with MP3 object key
	onProgress(95, "processing", "Updating video metadata...")
	if err := p.updateConversionJob(ctx, req.JobID, mp3ObjectKey, videoInfo); err != nil {
		log.Printf("Failed to update conversion job: %v", err)
	}

	onProgress(100, "completed", "Conversion completed successfully")
	log.Printf("Video processing pipeline completed for job: %s, MP3 key: %s", req.JobID, mp3ObjectKey)
	return nil
}

func (p *Pipeline) downloadVideo(ctx context.Context, videoID, filename, jobID string) (string, error) {
	objectKey := storage.VideoObjectKey(videoID, filename)

	videoStream, err := p.minioClient.DownloadFile(ctx, objectKey)
	if err != nil {
		return "", fmt.Errorf("failed to download from MinIO: %w", err)
	}
	defer videoStream.Close()

	inputExt := p.fileManager.GetFileExtension(filename)
	inputPath, err := p.fileManager.CreateTempFile(jobID+"_input", inputExt)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	if err := p.fileManager.SaveStreamToFile(videoStream, inputPath); err != nil {
		p.fileManager.DeleteFile(inputPath)
		return "", fmt.Errorf("failed to save video to temp file: %w", err)
	}

	log.Printf("Downloaded video %s from MinIO to %s", objectKey, inputPath)
	return inputPath, nil
}

func (p *Pipeline) prepareConversionOptions(opts ConversionOptions, videoInfo *ffmpeg.VideoInfo) ffmpeg.ConversionOptions {
	conversionOpts := ffmpeg.ConversionOptions{
		AudioBitrate:    opts.AudioBitrate,
		AudioSampleRate: opts.AudioSampleRate,
	}

	if conversionOpts.AudioBitrate == "" {
		switch opts.Quality {
		case "high":
			conversionOpts.AudioBitrate = "320k"
		case "medium":
			conversionOpts.AudioBitrate = "192k"
		case "low":
			conversionOpts.AudioBitrate = "128k"
		default:
			conversionOpts.AudioBitrate = "192k"
		}
	}

	if conversionOpts.AudioSampleRate == "" {
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

	outputPath, err := p.fileManager.CreateTempFile(jobID+"_output", ".mp3")
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}

	opts.InputPath = inputPath
	opts.OutputPath = outputPath
	opts.OnProgress = onProgress

	if err := p.ffmpeg.ConvertToMP3(ctx, opts); err != nil {
		p.fileManager.DeleteFile(outputPath)
		return "", fmt.Errorf("FFmpeg conversion failed: %w", err)
	}

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

// uploadMP3 uploads the converted MP3 to MinIO and returns the object key.
func (p *Pipeline) uploadMP3(ctx context.Context, mp3Path string, req ConversionRequest) (string, error) {
	mp3File, err := p.fileManager.OpenFile(mp3Path)
	if err != nil {
		return "", fmt.Errorf("failed to open MP3 file: %w", err)
	}
	defer mp3File.Close()

	mp3Size, err := p.fileManager.GetFileSize(mp3Path)
	if err != nil {
		return "", fmt.Errorf("failed to get MP3 file size: %w", err)
	}

	objectKey := storage.MP3ObjectKey(req.VideoID)

	if err := p.minioClient.UploadFile(ctx, objectKey, mp3File, mp3Size, "audio/mpeg"); err != nil {
		return "", fmt.Errorf("failed to upload MP3 to MinIO: %w", err)
	}

	log.Printf("Uploaded MP3 to MinIO: %s (size: %d bytes)", objectKey, mp3Size)
	return objectKey, nil
}

func (p *Pipeline) updateConversionJob(ctx context.Context, jobID, mp3ObjectKey string,
	videoInfo *ffmpeg.VideoInfo) error {

	update := bson.M{
		"mp3_object_key": mp3ObjectKey,
		"status":         "completed",
		"updated_at":     time.Now(),
		"processing_info": bson.M{
			"duration":   videoInfo.Duration.Seconds(),
			"resolution": videoInfo.Resolution,
			"bitrate":    videoInfo.Bitrate,
			"format":     videoInfo.Format,
		},
	}

	return p.mongoClient.UpdateConversionJob(ctx, jobID, update)
}

// GetEstimatedDuration estimates conversion time based on video duration and quality.
func (p *Pipeline) GetEstimatedDuration(videoDuration time.Duration, quality string) time.Duration {
	var conversionRate float64
	switch quality {
	case "high":
		conversionRate = 2.0
	case "medium":
		conversionRate = 3.0
	case "low":
		conversionRate = 4.0
	default:
		conversionRate = 3.0
	}

	estimatedTime := time.Duration(float64(videoDuration) / conversionRate)
	bufferTime := 30 * time.Second
	return estimatedTime + bufferTime
}