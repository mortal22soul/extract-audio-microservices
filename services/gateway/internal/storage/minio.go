package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type MinIOClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	mc := &MinIOClient{
		client:     client,
		bucketName: bucketName,
	}

	if err := mc.EnsureBucket(context.Background()); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"service":  "gateway",
		"endpoint": endpoint,
		"bucket":   bucketName,
	}).Info("Connected to MinIO")
	return mc, nil
}

func (mc *MinIOClient) EnsureBucket(ctx context.Context) error {
	exists, err := mc.client.BucketExists(ctx, mc.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := mc.client.MakeBucket(ctx, mc.bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", mc.bucketName, err)
		}
		logrus.WithField("bucket", mc.bucketName).Info("Created MinIO bucket")
	}
	return nil
}

// UploadFile stores a file in MinIO and returns the object key.
// objectKey should follow the pattern: videos/{videoID}/original.mp4
func (mc *MinIOClient) UploadFile(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := mc.client.PutObject(ctx, mc.bucketName, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s: %w", objectKey, err)
	}
	logrus.WithFields(logrus.Fields{
		"bucket": mc.bucketName,
		"key":    objectKey,
	}).Info("Uploaded file to MinIO")
	return nil
}

// DownloadFile retrieves a file from MinIO by object key.
func (mc *MinIOClient) DownloadFile(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	obj, err := mc.client.GetObject(ctx, mc.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", objectKey, err)
	}
	return obj, nil
}

// DeleteFile removes a file from MinIO by object key.
func (mc *MinIOClient) DeleteFile(ctx context.Context, objectKey string) error {
	err := mc.client.RemoveObject(ctx, mc.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete %s: %w", objectKey, err)
	}
	return nil
}

// VideoObjectKey returns the canonical MinIO key for an original uploaded video.
func VideoObjectKey(videoID string, filename string) string {
	return fmt.Sprintf("videos/%s/%s", videoID, filename)
}

// MP3ObjectKey returns the canonical MinIO key for a converted MP3 file.
func MP3ObjectKey(videoID string) string {
	return fmt.Sprintf("videos/%s/output.mp3", videoID)
}
