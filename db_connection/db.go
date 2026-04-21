package db_connection

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(ctx context.Context, mongoURI string) (*mongo.Client, *mongo.Database, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, nil, fmt.Errorf("db_connection: failed to connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, fmt.Errorf("db_connection: failed to ping: %w", err)
	}

	database := client.Database("auth")

	users := database.Collection("users")
	_, err = users.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("db_connection: failed to create user indexes: %w", err)
	}

	tokens := database.Collection("refresh_tokens")
	_, err = tokens.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "hash", Value: 1}}},
		{Keys: bson.D{{Key: "userId", Value: 1}}},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("db_connection: failed to create token indexes: %w", err)
	}

	return client, database, nil
}
