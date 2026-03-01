package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoClient(mongoURL, dbName string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Printf("Connected to MongoDB: %s/%s", mongoURL, dbName)
	return &MongoClient{
		client:   client,
		database: client.Database(dbName),
	}, nil
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
	if err := collection.FindOne(ctx, bson.M{"_id": jobID}).Decode(&job); err != nil {
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
