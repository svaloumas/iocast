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
	backoff    []time.Duration
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
			Status:    TaskStatusPending,
		},
	}
	return t
}

// MaxRetries passes a number of max retries to the task builder.
func (b *taskBuilder[T]) MaxRetries(maxRetries int) *taskBuilder[T] {
	if maxRetries < 1 {
		maxRetries = 1
	}
	b.maxRetries = maxRetries
	return b
}

// BackOff passes backoff intervals between retires to the task builder.
func (b *taskBuilder[T]) BackOff(backoff []time.Duration) *taskBuilder[T] {
	if len(backoff) > b.maxRetries-1 {
		b.backoff = backoff[:b.maxRetries-1]
	} else if len(backoff) == b.maxRetries-1 {
		b.backoff = backoff
	}
	return b
}

// Database passes a database implementation to the task builder.
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
		backoff:    b.backoff,
		next:       b.next,
		db:         b.db,
		metadata:   b.metadata,
	}
}
