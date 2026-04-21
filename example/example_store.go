package example

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Store struct {
	col *mongo.Collection
}

func NewStore(db *mongo.Database) *Store {
	return &Store{col: db.Collection("examples")}
}

func (s *Store) GetAll(ctx context.Context, userID string) ([]*Example, error) {
	return nil, nil
}

func (s *Store) GetOne(ctx context.Context, id, userID string) (*Example, error) {
	return nil, nil
}

func (s *Store) Create(ctx context.Context, userID string, req CreateExampleRequest) (*Example, error) {
	return nil, nil
}

func (s *Store) Update(ctx context.Context, id, userID string, req UpdateExampleRequest) (*Example, error) {
	return nil, nil
}

func (s *Store) Delete(ctx context.Context, id, userID string) error {
	return nil
}
