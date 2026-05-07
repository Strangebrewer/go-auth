package token

import (
	"net/http"

	"github.com/Strangebrewer/go-auth/tracer"
	"github.com/go-chi/chi/v5"
)

func Routes(svc *Service, tc *tracer.Client, authMiddleware func(http.Handler) http.Handler) chi.Router {
	h := NewHandler(svc, tc)
	r := chi.NewRouter()

	r.Post("/exchange", h.Refresh)
	r.With(authMiddleware).Post("/logout", h.Logout)

	return r
}
