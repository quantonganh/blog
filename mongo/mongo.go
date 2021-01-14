package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const connectionTimeout = 10 * time.Second

type DB struct {
	client *mongo.Client
	ctx    context.Context
	cancel func()

	dsn string
}

func NewDB(dsn string) *DB {
	db := &DB{
		dsn: dsn,
	}
	db.ctx, db.cancel = context.WithCancel(context.Background())
	return db
}

func (db *DB) Open() error {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(db.dsn))
	if err != nil {
		return err
	}

	db.client = client

	return nil
}

func (db *DB) Close() error {
	db.cancel()

	if db.client != nil {
		return db.client.Disconnect(db.ctx)
	}

	return nil
}

func (db *DB) NewDatabase(dbname string) *mongo.Database {
	return db.client.Database(dbname)
}
