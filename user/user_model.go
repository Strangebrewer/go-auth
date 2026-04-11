package user

import "github.com/google/uuid"

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
