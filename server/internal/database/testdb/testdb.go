// Package testdb gives integration tests an isolated view of the test database.
//
// Every test gets its own transaction that is rolled back when the test ends, so tests
// can run in parallel against one database without seeing each other's rows.
//
// TEST_DATABASE_URL selects the database. When it is unset, database tests are skipped
// locally (so `go test ./...` works on a fresh machine) but fail in CI (CI is set).
package testdb

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bekbek22/portfolio/server/internal/database"
)

const envVar = "TEST_DATABASE_URL"

var (
	setupOnce sync.Once
	pool      *pgxpool.Pool
	setupErr  error
)

// Tx begins a transaction on the test database and rolls it back during test cleanup.
func Tx(t testing.TB) pgx.Tx {
	t.Helper()

	url := os.Getenv(envVar)
	if url == "" {
		if os.Getenv("CI") != "" {
			t.Fatalf("%s must be set in CI; database integration tests cannot be skipped", envVar)
		}
		t.Skipf("WARNING: %s is not set, skipping database integration test "+
			"(start Postgres with `docker compose up -d` and set %s)", envVar, envVar)
	}

	setupOnce.Do(func() { pool, setupErr = setup(url) })
	if setupErr != nil {
		t.Fatalf("test database setup: %v", setupErr)
	}

	tx, err := pool.Begin(t.Context())
	if err != nil {
		t.Fatalf("begin test transaction: %v", err)
	}
	t.Cleanup(func() {
		// t.Context() is already cancelled when cleanups run, so roll back with a fresh one.
		if err := tx.Rollback(context.Background()); err != nil {
			t.Errorf("roll back test transaction: %v", err)
		}
	})
	return tx
}

// setup migrates the test database to the latest schema and opens a shared pool.
// Concurrent test binaries are safe: golang-migrate holds an advisory lock while migrating.
func setup(url string) (*pgxpool.Pool, error) {
	mg, err := database.NewMigrator(url)
	if err != nil {
		return nil, err
	}
	if err := mg.Up(); err != nil {
		_ = mg.Close()
		return nil, err
	}
	if err := mg.Close(); err != nil {
		return nil, fmt.Errorf("close migrator: %w", err)
	}

	p, err := database.NewPool(context.Background(), url)
	if err != nil {
		return nil, err
	}
	if err := p.Ping(context.Background()); err != nil {
		p.Close()
		return nil, fmt.Errorf("ping test database: %w", err)
	}
	return p, nil
}
