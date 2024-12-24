package iocast

import (
	"encoding/json"
	"sync"
)

type ResultWriter interface {
	Write(string, Result[any]) error
}

type memWriter struct {
	db *sync.Map
}

func NewMemWriter(db *sync.Map) *memWriter {
	return &memWriter{
		db: db,
	}
}

func (w *memWriter) Write(id string, r Result[any]) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	w.db.Store(id, data)
	return nil
}
