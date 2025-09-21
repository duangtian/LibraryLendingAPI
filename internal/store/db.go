package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &DB{Pool: pool}, nil
}

func (db *DB) Close() { db.Pool.Close() }

// RunMigrations executes SQL files in directory in lexical order if not yet applied.
func (db *DB) RunMigrations(ctx context.Context, dir string) error {
	if _, err := db.Pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil { return fmt.Errorf("read migrations dir: %w", err) }
	var files []string
	for _, e := range entries { if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") { files = append(files, e.Name()) } }
	sort.Strings(files)
	for _, f := range files {
		var exists bool
		if err := db.Pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename=$1)`, f).Scan(&exists); err != nil { return err }
		if exists { continue }
		path := filepath.Join(dir, f)
		b, err := os.ReadFile(path)
		if err != nil { return fmt.Errorf("read %s: %w", f, err) }
		if _, err := db.Pool.Exec(ctx, string(b)); err != nil { return fmt.Errorf("exec %s: %w", f, err) }
		if _, err := db.Pool.Exec(ctx, `INSERT INTO schema_migrations (filename) VALUES ($1)`, f); err != nil { return fmt.Errorf("record %s: %w", f, err) }
	}
	return nil
}
