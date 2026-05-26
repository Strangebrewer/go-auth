package demo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ipAttemptDoc struct {
	IP        string    `bson:"ip"`
	Count     int       `bson:"count"`
	FirstSeen time.Time `bson:"firstSeen"`
}

type Store struct {
	col *mongo.Collection
}

func NewStore(db *mongo.Database) *Store {
	return &Store{col: db.Collection("demo_ip_attempts")}
}

// CheckAndIncrementIP atomically increments the attempt count for the given IP
// and returns true if the new count exceeds limit.
func (s *Store) CheckAndIncrementIP(ctx context.Context, ip string, limit int) (bool, error) {
	now := time.Now().UTC()

	filter := bson.D{{Key: "ip", Value: ip}}
	update := bson.D{
		{Key: "$inc", Value: bson.D{{Key: "count", Value: 1}}},
		{Key: "$setOnInsert", Value: bson.D{{Key: "firstSeen", Value: now}}},
	}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var doc ipAttemptDoc
	err := s.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err != nil {
		return false, fmt.Errorf("check ip attempts: %w", err)
	}

	return doc.Count > limit, nil
}
