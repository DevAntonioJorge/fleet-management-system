package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5"
)

// Connect creates a single pgx connection to the database at the given DSN.
func Connect(ctx context.Context, dsn string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return conn, nil
}

// resolveSchemaPath returns the absolute path to schema.sql relative to this package.
func resolveSchemaPath() string {
	_, currentFile, _, _ := runtime.Caller(0)
	testutilDir := filepath.Dir(currentFile)
	databaseDir := filepath.Dir(testutilDir)
	return filepath.Join(databaseDir, "sql", "schema.sql")
}

// ApplySchema reads schema.sql and executes it against the database at the given DSN.
func ApplySchema(dsn string) error {
	schemaPath := resolveSchemaPath()

	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	return nil
}
