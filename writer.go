package iocast

import (
	"encoding/json"
	"sync"
)

type ResultWriter[T any] interface {
	Write(string, Result[T]) error
}

type memWriter[T any] struct {
	db *sync.Map
}

func NewMemWriter[T any](db *sync.Map) *memWriter[T] {
	return &memWriter[T]{
		db: db,
	}
}

func (w *memWriter[T]) Write(id string, r Result[T]) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	w.db.Store(id, data)
	return nil
}
