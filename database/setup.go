package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Init opens a SQLite database at the given path and prepares connection settings.
func Init(ctx context.Context, path string) error {
	if DB != nil {
		return nil
	}
	// modernc.org/sqlite driver name is "sqlite"
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	// Pragmas to improve reliability and enforce FKs
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return fmt.Errorf("enable foreign_keys: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000;"); err != nil {
		_ = db.Close()
		return fmt.Errorf("set busy_timeout: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping: %w", err)
	}
	DB = db
	return nil
}

// Close closes the global DB connection.
func Close() error {
	if DB != nil {
		err := DB.Close()
		DB = nil
		return err
	}
	return nil
}
