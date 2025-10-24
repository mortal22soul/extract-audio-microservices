package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	bucket   *gridfs.Bucket
}

type FileInfo struct {
	ID       primitive.ObjectID `bson:"_id"`
	Filename string             `bson:"filename"`
	Length   int64              `bson:"length"`
	Metadata bson.M             `bson:"metadata"`
}

func NewMongoClient(mongoURL, dbName string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)
	bucket, err := gridfs.NewBucket(database)
	if err != nil {
		return nil, fmt.Errorf("failed to create GridFS bucket: %w", err)
	}

	log.Printf("Connected to MongoDB: %s/%s", mongoURL, dbName)

	return &MongoClient{
		client:   client,
		database: database,
		bucket:   bucket,
	}, nil
}

func (mc *MongoClient) DownloadFile(ctx context.Context, fileID primitive.ObjectID) (io.ReadCloser, *FileInfo, error) {
	// Get file info first
	var fileInfo FileInfo
	err := mc.bucket.GetFilesCollection().FindOne(ctx, bson.M{"_id": fileID}).Decode(&fileInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Open download stream
	downloadStream, err := mc.bucket.OpenDownloadStream(fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open download stream: %w", err)
	}

	return downloadStream, &fileInfo, nil
}

func (mc *MongoClient) UploadFile(ctx context.Context, filename string, data io.Reader, metadata bson.M) (primitive.ObjectID, error) {
	uploadOptions := options.GridFSUpload().SetMetadata(metadata)
	
	objectID, err := mc.bucket.UploadFromStream(filename, data, uploadOptions)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("Uploaded file %s with ID: %s", filename, objectID.Hex())
	return objectID, nil
}

func (mc *MongoClient) UpdateConversionJob(ctx context.Context, jobID string, update bson.M) error {
	collection := mc.database.Collection("conversion_jobs")
	
	filter := bson.M{"_id": jobID}
	updateDoc := bson.M{"$set": update}
	
	_, err := collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update conversion job: %w", err)
	}
	
	return nil
}

func (mc *MongoClient) GetConversionJob(ctx context.Context, jobID string) (bson.M, error) {
	collection := mc.database.Collection("conversion_jobs")
	
	var job bson.M
	err := collection.FindOne(ctx, bson.M{"_id": jobID}).Decode(&job)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversion job: %w", err)
	}
	
	return job, nil
}

func (mc *MongoClient) Close(ctx context.Context) error {
	return mc.client.Disconnect(ctx)
}

func (mc *MongoClient) GetClient() *mongo.Client {
	return mc.client
}