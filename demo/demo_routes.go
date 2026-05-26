package demo

import (
	"github.com/Strangebrewer/go-auth/pubsub"
	"github.com/Strangebrewer/go-auth/token"
	"github.com/Strangebrewer/go-auth/user"
	"github.com/go-chi/chi/v5"
)

func Routes(userStore *user.Store, tokenService *token.Service, demoStore *Store, publisher *pubsub.Publisher, topicID string) chi.Router {
	h := NewHandler(userStore, tokenService, demoStore, publisher, topicID)
	r := chi.NewRouter()

	r.Post("/register", h.Register)

	return r
}
