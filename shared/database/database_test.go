package database

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/fms/fms/shared/database/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	testDB, err := testutil.StartTestDB(ctx)
	if err != nil {
		panic(err)
	}

	if err := testutil.ApplySchema(testDB.DSN); err != nil {
		testDB.Terminate(ctx)
		panic(err)
	}

	testPool, err = pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		testDB.Terminate(ctx)
		panic(err)
	}

	code := m.Run()

	testPool.Close()
	testDB.Terminate(ctx)

	os.Exit(code)
}

func TestNewPoolFactory_CreatesValidPool(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPoolFactory(ctx, testPool.Config().ConnString(), 5)
	if err != nil {
		t.Fatalf("expected pool creation to succeed, got: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("expected pool ping to succeed, got: %v", err)
	}
}

func TestNewConnFactory_CreatesValidConnection(t *testing.T) {
	ctx := context.Background()

	conn, err := NewConnFactory(ctx, testPool.Config().ConnString())
	if err != nil {
		t.Fatalf("expected connection creation to succeed, got: %v", err)
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		t.Fatalf("expected connection ping to succeed, got: %v", err)
	}
}

func TestErrorWrapping_NotFound(t *testing.T) {
	err := WrapError(pgx.ErrNoRows)

	if !IsNotFound(err) {
		t.Fatal("expected IsNotFound to return true")
	}

	var notFoundErr *ErrNotFound
	if !errors.As(err, &notFoundErr) {
		t.Fatal("expected error to be ErrNotFound")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("expected original pgx.ErrNoRows to be accessible via Unwrap")
	}
}

func TestErrorWrapping_ConstraintViolation(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"}
	err := WrapError(pgErr)

	if !IsConstraintViolation(err) {
		t.Fatal("expected IsConstraintViolation to return true")
	}

	var constraintErr *ErrConstraintViolation
	if !errors.As(err, &constraintErr) {
		t.Fatal("expected error to be ErrConstraintViolation")
	}
}

func TestErrorWrapping_GenericDatabase(t *testing.T) {
	genericErr := errors.New("some random error")
	err := WrapError(genericErr)

	if IsNotFound(err) {
		t.Fatal("expected IsNotFound to return false for generic error")
	}

	if IsConstraintViolation(err) {
		t.Fatal("expected IsConstraintViolation to return false for generic error")
	}

	var dbErr *ErrDatabase
	if !errors.As(err, &dbErr) {
		t.Fatal("expected error to be ErrDatabase")
	}
}
