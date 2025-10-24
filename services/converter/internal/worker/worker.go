package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/video-converter/converter/internal/config"
	"github.com/video-converter/converter/internal/ffmpeg"
	"github.com/video-converter/converter/internal/filemanager"
	"github.com/video-converter/converter/internal/messaging"
	"github.com/video-converter/converter/internal/pipeline"
	"github.com/video-converter/converter/internal/storage"
)

type Worker struct {
	config      *config.Config
	workerPool  chan struct{}
	stopCh      chan struct{}
	wg          sync.WaitGroup
	
	// Clients
	mongoClient   *storage.MongoClient
	redisClient   *storage.RedisClient
	rabbitClient  *messaging.RabbitMQClient
	ffmpeg        *ffmpeg.FFmpeg
	fileManager   *filemanager.FileManager
	pipeline      *pipeline.Pipeline
}

func New(cfg *config.Config) (*Worker, error) {
	// Initialize MongoDB client
	mongoClient, err := storage.NewMongoClient(cfg.MongoURL, cfg.MongoDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Initialize Redis client
	redisClient, err := storage.NewRedisClient(cfg.RedisURL)
	if err != nil {
		mongoClient.Close(context.Background())
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	// Initialize RabbitMQ client
	rabbitConfig := messaging.RabbitMQConfig{
		URL:                  cfg.RabbitMQURL,
		ConversionQueue:      cfg.ConversionQueue,
		NotificationExchange: cfg.NotificationExchange,
	}
	rabbitClient, err := messaging.NewRabbitMQClient(rabbitConfig)
	if err != nil {
		mongoClient.Close(context.Background())
		redisClient.Close()
		return nil, fmt.Errorf("failed to create RabbitMQ client: %w", err)
	}

	// Initialize FFmpeg
	ffmpegClient := ffmpeg.New(cfg.FFmpegPath, cfg.AudioBitrate, cfg.AudioSampleRate)
	
	// Validate FFmpeg installation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ffmpegClient.ValidateInstallation(ctx); err != nil {
		mongoClient.Close(context.Background())
		redisClient.Close()
		rabbitClient.Close()
		return nil, fmt.Errorf("FFmpeg validation failed: %w", err)
	}

	// Initialize file manager
	fileManager, err := filemanager.New(cfg.TempDir)
	if err != nil {
		mongoClient.Close(context.Background())
		redisClient.Close()
		rabbitClient.Close()
		return nil, fmt.Errorf("failed to create file manager: %w", err)
	}

	// Initialize pipeline
	pipelineClient := pipeline.New(mongoClient, redisClient, ffmpegClient, fileManager)

	return &Worker{
		config:      cfg,
		workerPool:  make(chan struct{}, cfg.MaxWorkers),
		stopCh:      make(chan struct{}),
		mongoClient: mongoClient,
		redisClient: redisClient,
		rabbitClient: rabbitClient,
		ffmpeg:      ffmpegClient,
		fileManager: fileManager,
		pipeline:    pipelineClient,
	}, nil
}

func (w *Worker) Start() {
	log.Println("Converter worker started")
	
	// Start cleanup routine
	w.wg.Add(1)
	go w.cleanupRoutine()
	
	// Create message handler
	messageHandler := messaging.NewMessageHandler(w.rabbitClient, w)
	
	// Start consuming conversion jobs
	ctx := context.Background()
	if err := messageHandler.StartConsuming(ctx); err != nil {
		log.Printf("Failed to consume conversion jobs: %v", err)
	}
}

func (w *Worker) Stop() {
	log.Println("Stopping converter worker...")
	close(w.stopCh)
	w.wg.Wait()
	
	// Close all clients
	if w.mongoClient != nil {
		w.mongoClient.Close(context.Background())
	}
	if w.redisClient != nil {
		w.redisClient.Close()
	}
	if w.rabbitClient != nil {
		w.rabbitClient.Close()
	}
	
	log.Println("Converter worker stopped")
}

// ProcessConversion implements the ConversionProcessor interface
func (w *Worker) ProcessConversion(ctx context.Context, msg messaging.ConversionMessage) error {
	return w.processConversionJob(ctx, msg)
}

func (w *Worker) processConversionJob(ctx context.Context, msg messaging.ConversionMessage) error {
	// Acquire worker from pool
	select {
	case w.workerPool <- struct{}{}:
		defer func() { <-w.workerPool }()
	case <-w.stopCh:
		return fmt.Errorf("worker is shutting down")
	case <-ctx.Done():
		return ctx.Err()
	}

	w.wg.Add(1)
	defer w.wg.Done()

	log.Printf("Processing conversion job: %s for user: %s", msg.JobID, msg.UserID)
	
	// Update job status to processing
	err := w.updateJobStatus(ctx, msg.JobID, "processing", 0, "")
	if err != nil {
		log.Printf("Failed to update job status: %v", err)
	}

	// Publish progress update
	w.publishProgress(ctx, msg.JobID, msg.UserID, 0, "processing", "Starting conversion...")

	// Process the conversion
	if err := w.convertVideo(ctx, msg); err != nil {
		log.Printf("Conversion failed for job %s: %v", msg.JobID, err)
		
		// Update job status to failed
		w.updateJobStatus(ctx, msg.JobID, "failed", 0, err.Error())
		w.publishProgress(ctx, msg.JobID, msg.UserID, 0, "failed", err.Error())
		
		// Send failure notification
		w.sendNotification(ctx, messaging.NotificationMessage{
			Type:    "conversion_failed",
			UserID:  msg.UserID,
			JobID:   msg.JobID,
			Subject: "Video Conversion Failed",
			Body:    fmt.Sprintf("Failed to convert video: %s", err.Error()),
		})
		
		return err
	}

	// Update job status to completed
	w.updateJobStatus(ctx, msg.JobID, "completed", 100, "")
	w.publishProgress(ctx, msg.JobID, msg.UserID, 100, "completed", "Conversion completed successfully")
	
	// Send success notification
	w.sendNotification(ctx, messaging.NotificationMessage{
		Type:    "conversion_completed",
		UserID:  msg.UserID,
		JobID:   msg.JobID,
		Subject: "Video Conversion Completed",
		Body:    "Your video has been successfully converted to MP3.",
		Data: map[string]interface{}{
			"video_id": msg.VideoID,
			"job_id":   msg.JobID,
		},
	})

	log.Printf("Successfully completed conversion job: %s", msg.JobID)
	return nil
}

func (w *Worker) convertVideo(ctx context.Context, msg messaging.ConversionMessage) error {
	// Prepare conversion request
	conversionReq := pipeline.ConversionRequest{
		JobID:    msg.JobID,
		UserID:   msg.UserID,
		VideoID:  msg.VideoID,
		Filename: msg.Filename,
		Options: pipeline.ConversionOptions{
			AudioBitrate:    w.config.AudioBitrate,
			AudioSampleRate: w.config.AudioSampleRate,
			Quality:         "medium", // Default quality
		},
	}

	// Process video using pipeline
	return w.pipeline.ProcessVideo(ctx, conversionReq, func(progress int, status, message string) {
		w.publishProgress(ctx, msg.JobID, msg.UserID, progress, status, message)
	})
}



func (w *Worker) updateJobStatus(ctx context.Context, jobID, status string, progress int, errorMsg string) error {
	update := bson.M{
		"status":     status,
		"progress":   progress,
		"updated_at": time.Now(),
	}
	
	if errorMsg != "" {
		update["error_message"] = errorMsg
	}
	
	if status == "completed" {
		update["completed_at"] = time.Now()
	}

	return w.mongoClient.UpdateConversionJob(ctx, jobID, update)
}

func (w *Worker) publishProgress(ctx context.Context, jobID, userID string, progress int, status, message string) {
	update := storage.ProgressUpdate{
		JobID:    jobID,
		UserID:   userID,
		Progress: progress,
		Status:   status,
		Message:  message,
	}

	if err := w.redisClient.PublishProgress(ctx, update); err != nil {
		log.Printf("Failed to publish progress update: %v", err)
	}

	if err := w.redisClient.SetJobProgress(ctx, jobID, progress); err != nil {
		log.Printf("Failed to set job progress in Redis: %v", err)
	}
}

func (w *Worker) sendNotification(ctx context.Context, notification messaging.NotificationMessage) {
	if err := w.rabbitClient.PublishNotification(ctx, notification); err != nil {
		log.Printf("Failed to send notification: %v", err)
	}
}

func (w *Worker) cleanupRoutine() {
	defer w.wg.Done()
	
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			// Cleanup temp files older than 2 hours
			if err := w.fileManager.CleanupTempFiles(2 * time.Hour); err != nil {
				log.Printf("Failed to cleanup temp files: %v", err)
			}
		}
	}
}

// Getter methods for health checking
func (w *Worker) GetMongoClient() *storage.MongoClient {
	return w.mongoClient
}

func (w *Worker) GetRedisClient() *storage.RedisClient {
	return w.redisClient
}

func (w *Worker) GetRabbitClient() *messaging.RabbitMQClient {
	return w.rabbitClient
}