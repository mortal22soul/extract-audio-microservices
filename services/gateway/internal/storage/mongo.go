package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/video-converter/gateway/internal/models"
)

type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoClient(uri, dbName string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Printf("Connected to MongoDB: %s/%s", uri, dbName)
	return &MongoClient{
		client:   client,
		database: client.Database(dbName),
	}, nil
}

func (mc *MongoClient) Close(ctx context.Context) error {
	return mc.client.Disconnect(ctx)
}

func (mc *MongoClient) CreateVideo(ctx context.Context, video *models.Video) error {
	video.UploadedAt = time.Now()
	video.Status = "uploaded"

	collection := mc.database.Collection("videos")
	result, err := collection.InsertOne(ctx, video)
	if err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		video.ID = models.ObjectID(oid.Hex())
	}
	return nil
}

func (mc *MongoClient) GetVideo(ctx context.Context, videoID string) (*models.Video, error) {
	oid, err := primitive.ObjectIDFromHex(videoID)
	if err != nil {
		return nil, fmt.Errorf("invalid video ID: %w", err)
	}

	collection := mc.database.Collection("videos")
	var video models.Video
	if err := collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&video); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("video not found")
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}
	return &video, nil
}

func (mc *MongoClient) GetUserVideos(ctx context.Context, userID string, limit, skip int) ([]models.Video, error) {
	collection := mc.database.Collection("videos")

	opts := options.Find().
		SetSort(bson.D{{Key: "uploadedAt", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	cursor, err := collection.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list videos: %w", err)
	}
	defer cursor.Close(ctx)

	var videos []models.Video
	if err := cursor.All(ctx, &videos); err != nil {
		return nil, fmt.Errorf("failed to decode videos: %w", err)
	}
	return videos, nil
}

func (mc *MongoClient) UpdateVideo(ctx context.Context, videoID string, update map[string]interface{}) error {
	oid, err := primitive.ObjectIDFromHex(videoID)
	if err != nil {
		return fmt.Errorf("invalid video ID: %w", err)
	}

	collection := mc.database.Collection("videos")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	if err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}
	return nil
}

func (mc *MongoClient) DeleteVideo(ctx context.Context, videoID string) error {
	oid, err := primitive.ObjectIDFromHex(videoID)
	if err != nil {
		return fmt.Errorf("invalid video ID: %w", err)
	}

	collection := mc.database.Collection("videos")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}
	return nil
}

func (mc *MongoClient) CreateConversionJob(ctx context.Context, job *models.ConversionJob) error {
	job.StartedAt = time.Now()
	job.Status = "pending"

	collection := mc.database.Collection("conversion_jobs")
	result, err := collection.InsertOne(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to create conversion job: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		job.ID = models.ObjectID(oid.Hex())
	}
	return nil
}

func (mc *MongoClient) GetConversionJob(ctx context.Context, jobID string) (*models.ConversionJob, error) {
	oid, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	collection := mc.database.Collection("conversion_jobs")
	var job models.ConversionJob
	if err := collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&job); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("conversion job not found")
		}
		return nil, fmt.Errorf("failed to get conversion job: %w", err)
	}
	return &job, nil
}
