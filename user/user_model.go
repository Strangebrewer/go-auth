package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Disabled     bool
}

type PublicUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User         PublicUser `json:"user"`
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
}

func publicUser(id uuid.UUID, email string) PublicUser {
	return PublicUser{
		ID:    id.String(),
		Email: email,
	}
}
