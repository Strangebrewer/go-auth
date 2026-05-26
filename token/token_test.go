package token_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcmongo "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

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

	mongoContainer, err := tcmongo.Run(ctx, "mongo:6")
	if err != nil {
		log.Fatalf("failed to start mongodb container: %v", err)
	}
	defer func() {
		if err := mongoContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}()

	connStr, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	client, err := mongo.Connect(options.Client().ApplyURI(connStr))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("auth_test")

	testUserStore = user.NewStore(db)
	if testPrivateKey != "" {
		testTokenService, err = token.NewService(token.NewStore(db), testPrivateKey, testPepper)
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
