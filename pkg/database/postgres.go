package database

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Connect establishes a connection to PostgreSQL database
func Connect(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Transaction executes a function within a database transaction
func Transaction(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// sanitizeString removes null bytes from a string to ensure PostgreSQL UTF-8 compatibility
func sanitizeString(s string) string {
	// PostgreSQL UTF-8 encoding doesn't allow null bytes (0x00)
	// Remove them to prevent "invalid byte sequence for encoding" errors
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != 0 {
			result = append(result, s[i])
		}
	}
	return string(result)
}

// NullString converts a string pointer to sql.NullString
// It sanitizes the string by removing null bytes to ensure PostgreSQL compatibility
func NullString(s *string) sql.NullString {
	if s != nil {
		sanitized := sanitizeString(*s)
		return sql.NullString{String: sanitized, Valid: true}
	}
	return sql.NullString{Valid: false}
}

// StringPtr converts sql.NullString to string pointer
func StringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// NullInt64 converts an int64 pointer to sql.NullInt64
func NullInt64(i *int64) sql.NullInt64 {
	if i != nil {
		return sql.NullInt64{Int64: *i, Valid: true}
	}
	return sql.NullInt64{Valid: false}
}

// Int64Ptr converts sql.NullInt64 to int64 pointer
func Int64Ptr(ni sql.NullInt64) *int64 {
	if ni.Valid {
		return &ni.Int64
	}
	return nil
}

// NullFloat64 converts a float64 pointer to sql.NullFloat64
func NullFloat64(f *float64) sql.NullFloat64 {
	if f != nil {
		return sql.NullFloat64{Float64: *f, Valid: true}
	}
	return sql.NullFloat64{Valid: false}
}

// Float64Ptr converts sql.NullFloat64 to float64 pointer
func Float64Ptr(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}

// NullUUID converts a UUID pointer to sql.NullString for database storage
func NullUUID(u *uuid.UUID) sql.NullString {
	if u != nil {
		return sql.NullString{String: u.String(), Valid: true}
	}
	return sql.NullString{Valid: false}
}

// UUIDPtr converts sql.NullString to UUID pointer
func UUIDPtr(ns sql.NullString) *uuid.UUID {
	if ns.Valid {
		if u, err := uuid.Parse(ns.String); err == nil {
			return &u
		}
	}
	return nil
}

// StringValue converts sql.NullString to string with empty string default
func StringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
