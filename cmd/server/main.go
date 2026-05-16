package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Strangebrewer/go-auth/app"
	"github.com/Strangebrewer/go-auth/config"
	"github.com/Strangebrewer/go-auth/db_connection"
	"github.com/Strangebrewer/go-auth/middleware"
	"github.com/Strangebrewer/go-auth/server"
	"github.com/Strangebrewer/go-auth/token"
	"github.com/Strangebrewer/go-auth/tracer"
	"github.com/Strangebrewer/go-auth/user"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx := context.Background()
	mongoClient, db, err := db_connection.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			slog.Error("failed to disconnect from database", "error", err)
		}
	}()

	authMiddleware, err := middleware.RequireAuth(cfg.JWTPublicKey)
	if err != nil {
		slog.Error("failed to parse JWT public key", "error", err)
		os.Exit(1)
	}

	tokenStore := token.NewStore(db)
	tokenService, err := token.NewService(tokenStore, cfg.JWTPrivateKey, cfg.RefreshTokenPepper)
	if err != nil {
		slog.Error("failed to initialize token service", "error", err)
		os.Exit(1)
	}

	application := &app.Application{
		UserStore:       user.NewStore(db),
		TokenService:    tokenService,
		Tracer:          tracer.NewClient(cfg.TracerURL, cfg.TracerKey, "go-auth"),
		RubeOwidNextURL: cfg.RubeOwidNextURL,
	}

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	srv := server.New(":"+port, cfg.AllowedOrigins, application, authMiddleware)

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.HTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.HTTPServer.Shutdown(shutCtx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
