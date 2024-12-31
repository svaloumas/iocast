package iocast

import (
	"time"
)

type taskBuilder[T any] struct {
	id         string
	taskFn     TaskFn[T]
	resultChan chan Result[T]
	next       *Task[T]
	maxRetries int
	db         DB
	metadata   Metadata
}

// TaskBuilder creates and returns a new TaskBuilder instance.
func TaskBuilder[T any](id string, fn TaskFn[T]) *taskBuilder[T] {
	t := &taskBuilder[T]{
		id:         id,
		taskFn:     fn,
		resultChan: make(chan Result[T], 1),
		maxRetries: 1,
		metadata: Metadata{
			CreatetAt: time.Now().UTC(),
			Status:    StatusPending,
		},
	}
	return t
}

// Context passes a number of max retries to the task builder.
func (b *taskBuilder[T]) MaxRetries(maxRetries int) *taskBuilder[T] {
	if maxRetries < 1 {
		maxRetries = 1
	}
	b.maxRetries = maxRetries
	return b
}

// Context passes a database implementation to the task builder.
func (b *taskBuilder[T]) Database(db DB) *taskBuilder[T] {
	b.db = db
	return b
}

// Build initializes and returns a new task instance.
func (b *taskBuilder[T]) Build() *Task[T] {
	return &Task[T]{
		id:         b.id,
		taskFn:     b.taskFn,
		resultChan: b.resultChan,
		maxRetries: b.maxRetries,
		next:       b.next,
		db:         b.db,
		metadata:   b.metadata,
	}
}
