package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Strangebrewer/go-auth/token"
)

func Routes(store *Store, tokenService *token.Service, authMiddleware func(http.Handler) http.Handler) chi.Router {
	h := NewHandler(store, tokenService)
	r := chi.NewRouter()

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.With(authMiddleware).Get("/me", h.Me)

	return r
}
