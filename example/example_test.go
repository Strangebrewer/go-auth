package example_test

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

	"github.com/Strangebrewer/go-auth/example"
)

var testStore *example.Store

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

	testStore = example.NewStore(client.Database("example_test"))

	os.Exit(m.Run())
}

func TestExampleStore_Create(t *testing.T) {
	t.Skip("implement store methods before enabling")

	ctx := context.Background()

	req := example.CreateExampleRequest{Name: "test example"}
	result, err := testStore.Create(ctx, "test-user-id", req)

	require.NoError(t, err)
	assert.Equal(t, "test example", result.Name)
	assert.NotEmpty(t, result.ID)
}

func TestExampleStore_GetAll(t *testing.T) {
	t.Skip("implement store methods before enabling")

	ctx := context.Background()

	results, err := testStore.GetAll(ctx, "test-user-id")

	require.NoError(t, err)
	assert.NotNil(t, results)
}
