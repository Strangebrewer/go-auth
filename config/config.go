package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTPrivateKey       string
	JWTPublicKey        string
	RefreshTokenPepper  string
	AllowedOrigins      []string
	TracerURL           string
	TracerKey           string
}

func parseOrigins(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func Load() *Config {
	if _, err := os.Stat(".env.local"); err == nil {
		_ = godotenv.Load(".env.local")
	}

	return &Config{
		Port:               os.Getenv("PORT"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTPrivateKey:      os.Getenv("JWT_PRIVATE_KEY"),
		JWTPublicKey:       os.Getenv("JWT_PUBLIC_KEY"),
		RefreshTokenPepper: os.Getenv("REFRESH_TOKEN_PEPPER"),
		AllowedOrigins:     parseOrigins(os.Getenv("ALLOWED_ORIGINS")),
		TracerURL:          os.Getenv("TRACER_SERVICE_URL"),
		TracerKey:          os.Getenv("TRACER_SERVICE_KEY"),
	}
}
