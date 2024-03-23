package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB represents the database connection.
type DB struct {
	sqlDB  *sql.DB
	ctx    context.Context
	cancel func()

	path string
}

// NewDB returns new database
func NewDB(path string) *DB {
	db := &DB{
		path: path,
	}

	db.ctx, db.cancel = context.WithCancel(context.Background())

	return db
}

// Open opens new database connection
func (db *DB) Open() (err error) {
	if db.path == "" {
		return errors.New("path required")
	}

	if db.sqlDB != nil {
		return nil
	}

	if db.sqlDB, err = sql.Open("sqlite3", db.path); err != nil {
		return err
	}

	if err := db.migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}

func (db *DB) migrate() error {
	if _, err := db.sqlDB.Exec(`CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);`); err != nil {
		return fmt.Errorf("cannot create migrations table: %w", err)
	}

	names, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	for _, name := range names {
		if err := db.migrateFile(name); err != nil {
			return fmt.Errorf("migration error: name=%q, err=%w", name, err)
		}
	}

	return nil
}

func (db *DB) migrateFile(name string) error {
	tx, err := db.sqlDB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = ?`, name).Scan(&n); err != nil {
		return err
	}
	if n != 0 {
		return nil
	}

	buf, err := fs.ReadFile(migrationsFS, name)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	if _, err := tx.Exec(`INSERT INTO migrations (name) VALUES (?)`, name); err != nil {
		return err
	}

	return tx.Commit()
}

// Close closes database connection
func (db *DB) Close() error {
	if db.sqlDB == nil {
		return nil
	}

	db.cancel()

	if err := db.sqlDB.Close(); err != nil {
		log.Println("Error closing database:", err)
	}

	return nil
}
