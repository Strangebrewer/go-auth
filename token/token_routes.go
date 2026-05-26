package token

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(svc *Service, authMiddleware func(http.Handler) http.Handler) chi.Router {
	h := NewHandler(svc)
	r := chi.NewRouter()

	r.Post("/exchange", h.Refresh)
	r.With(authMiddleware).Post("/logout", h.Logout)

	return r
}
