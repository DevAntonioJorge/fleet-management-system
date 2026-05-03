package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultStartupTimeout = 30 * time.Second

// TestDB holds a test database container and its connection details.
type TestDB struct {
	Container *postgres.PostgresContainer
	DSN       string
}

// StartTestDB starts a PostgreSQL test container and returns a TestDB with the DSN.
func StartTestDB(ctx context.Context) (*TestDB, error) {
	dbName := "test_db"
	dbUser := "test_user"
	dbPassword := "test_password"

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(defaultStartupTimeout),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	return &TestDB{
		Container: container,
		DSN:       dsn,
	}, nil
}

// Terminate stops and removes the test database container.
func (tdb *TestDB) Terminate(ctx context.Context) error {
	if tdb.Container != nil {
		return tdb.Container.Terminate(ctx)
	}
	return nil
}
