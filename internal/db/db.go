package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	// enable WAL mode and foreign keys for SQLite
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", dbPath)
	sqliteDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	sqliteDB.SetMaxOpenConns(1) // SQLite single-writer safe
	sqliteDB.SetMaxIdleConns(1)
	sqliteDB.SetConnMaxLifetime(time.Hour)

	db := &DB{DB: sqliteDB}
	if err := db.initSchema(context.Background()); err != nil {
		sqliteDB.Close()
		return nil, fmt.Errorf("init sqlite schema: %w", err)
	}

	return db, nil
}

func (db *DB) initSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS diagrams (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		code TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_diagrams_updated_at ON diagrams(updated_at DESC);
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return err
	}

	// Best-effort cleanup of legacy schema objects (sort_order column and the
	// unused title index). Errors are ignored: they just mean the object does
	// not exist (e.g. freshly created databases).
	_, _ = db.ExecContext(ctx, "ALTER TABLE diagrams DROP COLUMN sort_order")
	_, _ = db.ExecContext(ctx, "DROP INDEX IF EXISTS idx_diagrams_title")
	return nil
}
