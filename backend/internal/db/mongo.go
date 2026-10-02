package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func ConnectMongo(uri, dbName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %w", err)
	}

	// Ping database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb at %s: %w", uri, err)
	}

	database := client.Database(dbName)
	m := &MongoDB{
		Client:   client,
		Database: database,
	}

	// Initialize indexes
	if err := m.createIndexes(); err != nil {
		log.Printf("Warning: error creating mongo indexes: %v", err)
	}

	return m, nil
}

func (m *MongoDB) Projects() *mongo.Collection {
	return m.Database.Collection("projects")
}

func (m *MongoDB) Studies() *mongo.Collection {
	return m.Database.Collection("studies")
}

func (m *MongoDB) Close(ctx context.Context) error {
	if m.Client != nil {
		return m.Client.Disconnect(ctx)
	}
	return nil
}

func (m *MongoDB) createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Projects collection indexes
	projectIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "projectId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := m.Projects().Indexes().CreateOne(ctx, projectIdx); err != nil {
		log.Printf("Warning: failed to create projectId index: %v", err)
	}

	// Studies collection indexes
	studyIdIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "studyId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := m.Studies().Indexes().CreateOne(ctx, studyIdIdx); err != nil {
		log.Printf("Warning: failed to create studyId index: %v", err)
	}

	// Compound index for fast querying by project and decision
	studyCompoundIdx := mongo.IndexModel{
		Keys: bson.D{
			{Key: "projectId", Value: 1},
			{Key: "decision", Value: 1},
		},
	}
	if _, err := m.Studies().Indexes().CreateOne(ctx, studyCompoundIdx); err != nil {
		log.Printf("Warning: failed to create compound index on studies: %v", err)
	}

	return nil
}
