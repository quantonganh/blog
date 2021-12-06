package bolt

import (
	"context"

	"github.com/asdine/storm/v3"
)

// DB represents a database
type DB struct {
	path    string
	stormDB *storm.DB
	ctx     context.Context
	cancel  func()
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
func (db *DB) Open() error {
	stormDB, err := storm.Open(db.path)
	if err != nil {
		return err
	}
	db.stormDB = stormDB

	return nil
}

// Close closes database connection
func (db *DB) Close() error {
	db.cancel()

	if db.stormDB != nil {
		return db.stormDB.Close()
	}

	return nil
}
