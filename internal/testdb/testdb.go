package testdb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open creates a migrated PostgreSQL test schema and returns a pool scoped to it.
func Open(t testing.TB) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	schema := newSchemaName(t)
	if _, err := adminPool.Exec(ctx, `SELECT pg_advisory_lock(424242)`); err != nil {
		adminPool.Close()
		t.Fatalf("lock pgcrypto extension setup: %v", err)
	}
	if _, err := adminPool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA public`); err != nil {
		_, _ = adminPool.Exec(ctx, `SELECT pg_advisory_unlock(424242)`)
		adminPool.Close()
		t.Fatalf("create pgcrypto extension: %v", err)
	}
	if _, err := adminPool.Exec(ctx, `SELECT pg_advisory_unlock(424242)`); err != nil {
		adminPool.Close()
		t.Fatalf("unlock pgcrypto extension setup: %v", err)
	}
	if _, err := adminPool.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		adminPool.Close()
		t.Fatalf("create test schema: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		adminPool.Close()
		t.Fatalf("parse test database URL: %v", err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ",public"

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		adminPool.Close()
		t.Fatalf("connect to test schema: %v", err)
	}

	if err := runMigrations(ctx, pool); err != nil {
		pool.Close()
		adminPool.Close()
		t.Fatalf("run migrations: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()

		pool.Close()
		if _, err := adminPool.Exec(cleanupCtx, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
		adminPool.Close()
	})

	return pool
}

func newSchemaName(t testing.TB) string {
	t.Helper()

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("generate schema name: %v", err)
	}

	return "test_" + hex.EncodeToString(buf)
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(root, "migrations", "*.up.sql"))
	if err != nil {
		return fmt.Errorf("find migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filepath.Base(file), err)
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("apply migration %s: %w", filepath.Base(file), err)
		}
	}

	return nil
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir || strings.TrimSpace(parent) == "" {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}
