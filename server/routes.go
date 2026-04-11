package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Strangebrewer/go-auth/app"
	"github.com/Strangebrewer/go-auth/health"
	"github.com/Strangebrewer/go-auth/token"
	"github.com/Strangebrewer/go-auth/user"
)

func registerRoutes(r chi.Router, application *app.Application, authMiddleware func(http.Handler) http.Handler) {
	r.Get("/health", health.Handler)
	r.Mount("/users", user.Routes(application.UserStore, application.TokenService, authMiddleware))
	r.Mount("/token", token.Routes(application.TokenService, authMiddleware))
}
