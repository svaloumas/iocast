package iocast

import (
	"encoding/json"
	"sync"
)

// DB represents a storage.
type DB interface {
	Write(string, Result[any]) error
}

type memDB struct {
	db *sync.Map
}

// NewMemDB creates and returns a new memDB instance.
func NewMemDB(db *sync.Map) DB {
	return &memDB{
		db: db,
	}
}

// Write stores the results to the database.
func (w *memDB) Write(id string, r Result[any]) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	w.db.Store(id, data)
	return nil
}
