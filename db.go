package iocast

import (
	"encoding/json"
	"sync"
)

type DB interface {
	Write(string, Result[any]) error
}

type memDB struct {
	db *sync.Map
}

func NewMemDB(db *sync.Map) DB {
	return &memDB{
		db: db,
	}
}

func (w *memDB) Write(id string, r Result[any]) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	w.db.Store(id, data)
	return nil
}
