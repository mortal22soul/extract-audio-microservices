package storage

import (
	"fmt"
	"io"
	"log"

	"github.com/video-converter/gateway/internal/models"
)

// Temporary stub types until MongoDB driver issue is resolved
type ObjectID string
type Client struct{}
type Database struct{}
type Bucket struct{}
type DownloadStream struct{}
type File struct{}

type MongoDB struct {
	client   *Client
	database *Database
	bucket   *Bucket
}

func NewMongoDB(uri string) (*MongoDB, error) {
	// Temporary stub implementation
	log.Println("MongoDB stub - connection simulated")
	
	return &MongoDB{
		client:   &Client{},
		database: &Database{},
		bucket:   &Bucket{},
	}, nil
}

func (m *MongoDB) Close() error {
	// Temporary stub implementation
	return nil
}

// Stub implementations - will be replaced with real MongoDB implementation

func (m *MongoDB) CreateVideo(video *models.Video) error {
	return fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) GetVideo(videoID ObjectID) (*models.Video, error) {
	return nil, fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) GetUserVideos(userID string, limit, skip int) ([]models.Video, error) {
	return nil, fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) UpdateVideo(videoID ObjectID, update map[string]interface{}) error {
	return fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) DeleteVideo(videoID ObjectID) error {
	return fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) UploadFile(filename string, data io.Reader, metadata map[string]interface{}) (ObjectID, error) {
	return "", fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) DownloadFile(fileID ObjectID) (*DownloadStream, error) {
	return nil, fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) DeleteFile(fileID ObjectID) error {
	return fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) GetFileInfo(fileID ObjectID) (*File, error) {
	return nil, fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) CreateConversionJob(job *models.ConversionJob) error {
	return fmt.Errorf("MongoDB stub - not implemented")
}

func (m *MongoDB) GetConversionJob(jobID ObjectID) (*models.ConversionJob, error) {
	return nil, fmt.Errorf("MongoDB stub - not implemented")
}