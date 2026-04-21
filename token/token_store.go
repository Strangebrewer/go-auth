package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type refreshTokenDoc struct {
	ID        string     `bson:"_id"`
	UserID    string     `bson:"userId"`
	Hash      string     `bson:"hash"`
	ExpiresAt time.Time  `bson:"expiresAt"`
	CreatedAt time.Time  `bson:"createdAt"`
	RevokedAt *time.Time `bson:"revokedAt"`
}

func (d refreshTokenDoc) toDomain() (RefreshToken, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("parse token id: %w", err)
	}
	userID, err := uuid.Parse(d.UserID)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("parse token userId: %w", err)
	}
	return RefreshToken{
		ID:        id,
		UserID:    userID,
		Hash:      d.Hash,
		ExpiresAt: d.ExpiresAt,
		CreatedAt: d.CreatedAt,
		RevokedAt: d.RevokedAt,
	}, nil
}

type Store struct {
	col *mongo.Collection
}

func NewStore(db *mongo.Database) *Store {
	return &Store{col: db.Collection("refresh_tokens")}
}

func (s *Store) Create(ctx context.Context, userID uuid.UUID, hash string, expiresAt time.Time) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate id: %w", err)
	}

	doc := refreshTokenDoc{
		ID:        id.String(),
		UserID:    userID.String(),
		Hash:      hash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
		RevokedAt: nil,
	}

	_, err = s.col.InsertOne(ctx, doc)
	return err
}

func (s *Store) FindActiveByHash(ctx context.Context, hash string) (RefreshToken, error) {
	filter := bson.D{
		{Key: "hash", Value: hash},
		{Key: "revokedAt", Value: nil},
		{Key: "expiresAt", Value: bson.D{{Key: "$gt", Value: time.Now().UTC()}}},
	}

	var doc refreshTokenDoc
	err := s.col.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return RefreshToken{}, ErrInvalidToken
		}
		return RefreshToken{}, fmt.Errorf("find active token: %w", err)
	}
	return doc.toDomain()
}

func (s *Store) RevokeByID(ctx context.Context, id uuid.UUID) error {
	filter := bson.D{{Key: "_id", Value: id.String()}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "revokedAt", Value: time.Now().UTC()}}}}
	_, err := s.col.UpdateOne(ctx, filter, update)
	return err
}

func (s *Store) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	filter := bson.D{
		{Key: "userId", Value: userID.String()},
		{Key: "revokedAt", Value: nil},
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "revokedAt", Value: time.Now().UTC()}}}}
	_, err := s.col.UpdateMany(ctx, filter, update)
	return err
}
