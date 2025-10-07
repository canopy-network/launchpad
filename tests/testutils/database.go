package testutils

import (
	"context"
	"testing"

	"github.com/enielson/launchpad/pkg/database"
	"github.com/jmoiron/sqlx"
)

const TestDatabaseURL = "postgres://launchpad:launchpad123@localhost:5432/launchpad?sslmode=disable"

// WithTestDB runs a test with a database connection and automatic cleanup
func WithTestDB(t *testing.T, fn func(*sqlx.DB)) {
	t.Helper()
	db, err := database.Connect(TestDatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	fn(db)
}

// WithTestTransaction runs a test in a transaction that automatically rolls back
// This ensures test isolation - all changes are reverted after the test completes
func WithTestTransaction(t *testing.T, fn func(context.Context, *sqlx.Tx)) {
	t.Helper()
	db, err := database.Connect(TestDatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // Always rollback for test isolation

	ctx := context.Background()
	fn(ctx, tx)
}

// TestDBQuerier is an interface that both *sqlx.DB and *sqlx.Tx implement
// This allows fixtures and repositories to work with either
type TestDBQuerier interface {
	sqlx.Queryer
	sqlx.Execer
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (interface{}, error)
}

// TruncateTables truncates specified tables for cleanup (use sparingly in integration tests)
func TruncateTables(t *testing.T, db *sqlx.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		_, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE")
		if err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// CleanupTestData removes test data by ID (for tests that don't use transactions)
func CleanupTestData(t *testing.T, db *sqlx.DB, table string, id interface{}) {
	t.Helper()
	_, err := db.Exec("DELETE FROM "+table+" WHERE id = $1", id)
	if err != nil {
		t.Logf("Warning: Failed to cleanup %s with id %v: %v", table, id, err)
	}
}
