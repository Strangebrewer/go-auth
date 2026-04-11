package user_test

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

	"github.com/Strangebrewer/go-auth/user"
)

var testStore *user.Store

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

	testStore = user.NewStore(pool)

	os.Exit(m.Run())
}

func TestUserStore_Create(t *testing.T) {
	t.Skip("implement when ready")

	ctx := context.Background()

	u, err := testStore.Create(ctx, "test@example.com", "supersecretpassword")

	require.NoError(t, err)
	assert.NotEmpty(t, u.ID)
	assert.Equal(t, "test@example.com", u.Email)
}

func TestUserStore_FindByEmail(t *testing.T) {
	t.Skip("implement when ready")

	ctx := context.Background()

	u, err := testStore.FindByEmail(ctx, "test@example.com")

	require.NoError(t, err)
	assert.Equal(t, "test@example.com", u.Email)
}

func TestUserStore_DuplicateEmail(t *testing.T) {
	t.Skip("implement when ready")

	ctx := context.Background()

	_, err := testStore.Create(ctx, "test@example.com", "supersecretpassword")
	require.NoError(t, err)

	_, err = testStore.Create(ctx, "test@example.com", "supersecretpassword")
	assert.ErrorIs(t, err, user.ErrEmailExists)
}
