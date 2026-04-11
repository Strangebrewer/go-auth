package token

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type Service struct {
	store      *Store
	priv       *rsa.PrivateKey
	pepper     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(store *Store, privateKeyPEM, pepper string) (*Service, error) {
	priv, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	return &Service{
		store:      store,
		priv:       priv,
		pepper:     pepper,
		accessTTL:  15 * time.Minute,
		refreshTTL: 30 * 24 * time.Hour,
	}, nil
}

func (s *Service) IssueForUser(ctx context.Context, userID uuid.UUID) (*ExchangeResult, error) {
	access, err := s.mintAccessJWT(userID)
	if err != nil {
		return nil, fmt.Errorf("mint access token: %w", err)
	}

	plain, hash, err := s.mintRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("mint refresh token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(s.refreshTTL)
	if err := s.store.Create(ctx, userID, hash, expiresAt); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &ExchangeResult{AccessToken: access, RefreshToken: plain}, nil
}

// Exchange rotates the refresh token — revokes the old one and issues a new pair.
func (s *Service) Exchange(ctx context.Context, refreshPlain string) (*ExchangeResult, error) {
	hash := s.hashRefresh(refreshPlain)

	old, err := s.store.FindActiveByHash(ctx, hash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	_ = s.store.RevokeByID(ctx, old.ID)

	return s.IssueForUser(ctx, old.UserID)
}

func (s *Service) Revoke(ctx context.Context, refreshPlain string) error {
	hash := s.hashRefresh(refreshPlain)

	old, err := s.store.FindActiveByHash(ctx, hash)
	if err != nil {
		return ErrInvalidToken
	}

	return s.store.RevokeByID(ctx, old.ID)
}

func (s *Service) mintAccessJWT(userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	jti, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"sub": userID.String(),
		"typ": "access",
		"iat": now.Unix(),
		"exp": now.Add(s.accessTTL).Unix(),
		"jti": jti.String(),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return tok.SignedString(s.priv)
}

func (s *Service) mintRefreshToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, readErr := rand.Read(b); readErr != nil {
		return "", "", readErr
	}
	plain = base64.RawURLEncoding.EncodeToString(b)
	hash = s.hashRefresh(plain)
	return plain, hash, nil
}

func (s *Service) hashRefresh(plain string) string {
	sum := sha256.Sum256([]byte(s.pepper + ":" + plain))
	return hex.EncodeToString(sum[:])
}

func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, ErrInvalidToken
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrInvalidToken
	}
	key, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidToken
	}
	return key, nil
}
