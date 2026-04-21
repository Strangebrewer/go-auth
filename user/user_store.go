package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
	ErrNotFound           = errors.New("user not found")
)

type userDoc struct {
	ID           string    `bson:"_id"`
	Email        string    `bson:"email"`
	PasswordHash string    `bson:"passwordHash"`
	CreatedAt    time.Time `bson:"createdAt"`
	UpdatedAt    time.Time `bson:"updatedAt"`
	Disabled     bool      `bson:"disabled"`
}

func (d userDoc) toDomain() (User, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return User{}, fmt.Errorf("parse user id: %w", err)
	}
	return User{
		ID:           id,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		Disabled:     d.Disabled,
	}, nil
}

type Store struct {
	col *mongo.Collection
}

func NewStore(db *mongo.Database) *Store {
	return &Store{col: db.Collection("users")}
}

func (s *Store) Create(ctx context.Context, email, password string) (User, error) {
	email = normalizeEmail(email)

	hash, err := hashPassword(password)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return User{}, fmt.Errorf("generate id: %w", err)
	}

	now := time.Now().UTC()
	doc := userDoc{
		ID:           id.String(),
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    now,
		UpdatedAt:    now,
		Disabled:     false,
	}

	if _, err := s.col.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return User{}, ErrEmailExists
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return doc.toDomain()
}

func (s *Store) FindByEmail(ctx context.Context, email string) (User, error) {
	var doc userDoc
	err := s.col.FindOne(ctx, bson.D{{Key: "email", Value: normalizeEmail(email)}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("find user by email: %w", err)
	}
	return doc.toDomain()
}

func (s *Store) FindByID(ctx context.Context, id uuid.UUID) (User, error) {
	var doc userDoc
	err := s.col.FindOne(ctx, bson.D{{Key: "_id", Value: id.String()}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("find user by id: %w", err)
	}
	return doc.toDomain()
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
