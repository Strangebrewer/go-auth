package token

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Strangebrewer/go-auth/db/generated"
)

type Store struct {
	q *db.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: db.New(pool)}
}

func (s *Store) Create(ctx context.Context, userID uuid.UUID, hash string, expiresAt time.Time) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate id: %w", err)
	}

	_, err = s.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		ID:        id,
		UserID:    userID,
		Hash:      hash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	})
	return err
}

func (s *Store) FindActiveByHash(ctx context.Context, hash string) (db.RefreshToken, error) {
	return s.q.GetActiveRefreshTokenByHash(ctx, hash)
}

func (s *Store) RevokeByID(ctx context.Context, id uuid.UUID) error {
	return s.q.RevokeRefreshToken(ctx, id)
}

func (s *Store) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return s.q.RevokeAllUserRefreshTokens(ctx, userID)
}
