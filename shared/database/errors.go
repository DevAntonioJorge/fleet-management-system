package database

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ErrNotFound indicates that a requested record was not found in the database.
type ErrNotFound struct {
	Err error
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("record not found: %v", e.Err)
}

func (e *ErrNotFound) Unwrap() error {
	return e.Err
}

// ErrDatabase indicates a general database error.
type ErrDatabase struct {
	Err error
}

func (e *ErrDatabase) Error() string {
	return fmt.Sprintf("database error: %v", e.Err)
}

func (e *ErrDatabase) Unwrap() error {
	return e.Err
}

// ErrConstraintViolation indicates a constraint violation (e.g., unique, foreign key).
type ErrConstraintViolation struct {
	Err error
}

func (e *ErrConstraintViolation) Error() string {
	return fmt.Sprintf("constraint violation: %v", e.Err)
}

func (e *ErrConstraintViolation) Unwrap() error {
	return e.Err
}

// IsNotFound checks if the error is a "not found" error.
func IsNotFound(err error) bool {
	var notFoundErr *ErrNotFound
	return errors.As(err, &notFoundErr)
}

// IsConstraintViolation checks if the error is a constraint violation.
func IsConstraintViolation(err error) bool {
	var constraintErr *ErrConstraintViolation
	return errors.As(err, &constraintErr)
}

// WrapError wraps a pgx error into a domain-specific error type.
func WrapError(err error) error {
	if err == nil {
		return nil
	}

	// Wrap pgx.ErrNoRows as NotFound
	if errors.Is(err, pgx.ErrNoRows) {
		return &ErrNotFound{Err: err}
	}

	// Wrap other pgx errors as Database errors
	return &ErrDatabase{Err: err}
}
