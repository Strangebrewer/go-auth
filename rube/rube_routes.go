package rube

import (
	"net/http"

	"github.com/Strangebrewer/go-auth/tracer"
	"github.com/go-chi/chi/v5"
)

func Routes(nextURL string, tc *tracer.Client, authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	h := NewHandler(nextURL, tc)
	r.With(authMiddleware).Post("/", h.Chain)
	return r
}
