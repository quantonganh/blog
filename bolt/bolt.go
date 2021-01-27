package bolt

import (
	"context"

	"github.com/asdine/storm/v3"
)

type DB struct {
	path   string
	db     *storm.DB
	ctx    context.Context
	cancel func()
}

func NewDB(path string) *DB {
	db := &DB{
		path: path,
	}

	db.ctx, db.cancel = context.WithCancel(context.Background())

	return db
}

func (db *DB) Open() error {
	stormDB, err := storm.Open(db.path)
	if err != nil {
		return err
	}
	db.db = stormDB

	return nil
}

func (db *DB) Close() error {
	db.cancel()

	if db.db != nil {
		return db.db.Close()
	}

	return nil
}
