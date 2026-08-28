package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(dbPath string) (*DB, error) {
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
		updated_at DATETIME NOT NULL,
		sort_order INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_diagrams_updated_at ON diagrams(updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_diagrams_title ON diagrams(title);
	`
	_, err := db.ExecContext(ctx, query)
	return err
}
