package user_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcmongo "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/Strangebrewer/go-auth/user"
)

var testStore *user.Store

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

	_, err = db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatalf("failed to create user indexes: %v", err)
	}

	testStore = user.NewStore(db)

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
