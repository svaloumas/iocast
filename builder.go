package iocast

import (
	"context"
)

type taskBuilder[T any] struct {
	ctx        context.Context
	taskFn     taskFn[T]
	resultChan chan Result[T]
	next       *task[T]
	maxRetries int
}

// TaskBuilder creates and returns a new TaskBuilder instance.
func TaskBuilder[T any](fn taskFn[T]) *taskBuilder[T] {
	t := &taskBuilder[T]{
		taskFn:     fn,
		resultChan: make(chan Result[T], 1),
		maxRetries: 1,
	}
	return t
}

// Context passes a context to the task builder.
func (b *taskBuilder[T]) Context(ctx context.Context) *taskBuilder[T] {
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

// Build builds a new task instance and returns it.
func (b *taskBuilder[T]) Build() *task[T] {
	return &task[T]{
		ctx:        b.ctx,
		taskFn:     b.taskFn,
		resultChan: b.resultChan,
		maxRetries: b.maxRetries,
		next:       b.next,
	}
}