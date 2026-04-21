package token

import (
	"time"

	"github.com/google/uuid"
)

type ExchangeResult struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Hash      string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}
