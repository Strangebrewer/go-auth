package token_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Strangebrewer/go-auth/token"
	"github.com/Strangebrewer/go-auth/user"
)

var (
	testTokenService *token.Service
	testUserStore    *user.Store
)

const (
	testPrivateKey = "" // populate with a test RSA private key PEM when enabling tests
	testPepper     = "test-pepper"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		log.Fatalf("failed to read schema: %v", err)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		log.Fatalf("failed to apply schema: %v", err)
	}

	testUserStore = user.NewStore(pool)
	if testPrivateKey != "" {
		testTokenService, err = token.NewService(token.NewStore(pool), testPrivateKey, testPepper)
		if err != nil {
			log.Fatalf("failed to create token service: %v", err)
		}
	}

	os.Exit(m.Run())
}

func TestTokenService_IssueForUser(t *testing.T) {
	t.Skip("implement when ready")

	ctx := context.Background()

	u, err := testUserStore.Create(ctx, "tokentest@example.com", "supersecretpassword")
	require.NoError(t, err)

	result, err := testTokenService.IssueForUser(ctx, u.ID)

	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
}

func TestTokenService_Exchange(t *testing.T) {
	t.Skip("implement when ready")

	ctx := context.Background()

	u, err := testUserStore.Create(ctx, "exchange@example.com", "supersecretpassword")
	require.NoError(t, err)

	first, err := testTokenService.IssueForUser(ctx, u.ID)
	require.NoError(t, err)

	second, err := testTokenService.Exchange(ctx, first.RefreshToken)

	require.NoError(t, err)
	assert.NotEmpty(t, second.AccessToken)
	assert.NotEqual(t, first.RefreshToken, second.RefreshToken)
}

func TestTokenService_Revoke(t *testing.T) {
	t.Skip("implement when ready")

	ctx := context.Background()

	u, err := testUserStore.Create(ctx, "revoke@example.com", "supersecretpassword")
	require.NoError(t, err)

	result, err := testTokenService.IssueForUser(ctx, u.ID)
	require.NoError(t, err)

	err = testTokenService.Revoke(ctx, result.RefreshToken)
	require.NoError(t, err)

	_, err = testTokenService.Exchange(ctx, result.RefreshToken)
	assert.ErrorIs(t, err, token.ErrInvalidToken)
}
