package app

import (
	"github.com/Strangebrewer/go-auth/demo"
	"github.com/Strangebrewer/go-auth/pubsub"
	"github.com/Strangebrewer/go-auth/token"
	"github.com/Strangebrewer/go-auth/tracer"
	"github.com/Strangebrewer/go-auth/user"
)

type Application struct {
	UserStore               *user.Store
	TokenService            *token.Service
	Tracer                  *tracer.Client
	RubeOwidNextURL         string
	Publisher               *pubsub.Publisher
	DemoStore               *demo.Store
	DemoRegisteredTopicID   string
}
