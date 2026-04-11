package app

import (
	"github.com/Strangebrewer/go-auth/token"
	"github.com/Strangebrewer/go-auth/user"
)

type Application struct {
	UserStore    *user.Store
	TokenService *token.Service
}
