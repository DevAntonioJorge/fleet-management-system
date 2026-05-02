package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPoolFactory creates a pgxpool.Pool connection pool that implements the SQLc DBTX interface.
// Used for production and load testing scenarios requiring connection pooling.
func NewPoolFactory(ctx context.Context, dsn string, maxConnections int32) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	config.MaxConns = maxConnections

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Health check: ping the pool to ensure connectivity
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// NewConnFactory creates a single pgx.Conn connection that implements the SQLc DBTX interface.
// Used for testing and debugging scenarios where transaction rollback is beneficial.
func NewConnFactory(ctx context.Context, dsn string) (pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return pgx.Conn{}, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Health check: ping the connection to ensure connectivity
	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return pgx.Conn{}, fmt.Errorf("failed to ping database: %w", err)
	}

	return *conn, nil
}
