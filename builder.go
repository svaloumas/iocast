package iocast

import (
	"context"
)

type taskBuilder[T any] struct {
	id         string
	ctx        context.Context
	taskFn     taskFn[T]
	resultChan chan Result[T]
	next       *task[T]
	maxRetries int
	writer     ResultWriter
}

// TaskBuilder creates and returns a new TaskBuilder instance.
func TaskBuilder[T any](id string, fn taskFn[T]) *taskBuilder[T] {
	t := &taskBuilder[T]{
		id:         id,
		taskFn:     fn,
		resultChan: make(chan Result[T], 1),
		maxRetries: 1,
		ctx:        context.Background(),
	}
	return t
}

// Context passes a context to the task builder.
func (b *taskBuilder[T]) Context(ctx context.Context) *taskBuilder[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	b.ctx = ctx
	return b
}

// Context passes a number of max retries to the task builder.
func (b *taskBuilder[T]) MaxRetries(maxRetries int) *taskBuilder[T] {
	if maxRetries < 1 {
		maxRetries = 1
	}
	b.maxRetries = maxRetries
	return b
}

func (b *taskBuilder[T]) ResultWriter(w ResultWriter) *taskBuilder[T] {
	b.writer = w
	return b
}

// Build builds a new task instance and returns it.
func (b *taskBuilder[T]) Build() *task[T] {
	return &task[T]{
		id:         b.id,
		ctx:        b.ctx,
		taskFn:     b.taskFn,
		resultChan: b.resultChan,
		maxRetries: b.maxRetries,
		next:       b.next,
		writer:     b.writer,
	}
}
