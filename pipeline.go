package iocast

import (
	"context"
	"errors"
)

const (
	minTasksNum = 2
)

type Pipeline[T any] struct {
	id         string
	head       *Task[T]
	resultChan chan Result[T]
}

// NewPipeline links tasks together to execute them in order, returns a pipeline instance.
func NewPipeline[T any](id string, tasks ...*Task[T]) (*Pipeline[T], error) {
	if len(tasks) < minTasksNum {
		return nil, errors.New("at least two tasks must be linked to create a pipeline")
	}
	head := tasks[0]
	for i, t := range tasks {
		if i < len(tasks)-1 {
			t.link(tasks[i+1])
		}
	}
	return &Pipeline[T]{
		id:         id,
		head:       head,
		resultChan: head.resultChan,
	}, nil
}

// Wait awaits for the final result of the pipeline (last task in the order).
func (p *Pipeline[T]) Wait() <-chan Result[T] {
	return p.head.resultChan
}

// Exec executes the linked tasks of the pipeline.
func (p *Pipeline[T]) Exec(ctx context.Context) {
	p.head.Exec(ctx)
}

// Write stores the results of the pipeline (head's result) to the database.
func (p *Pipeline[T]) Write() error {
	return p.head.Write()
}

// ID is an ID geter.
func (p *Pipeline[T]) ID() string {
	return p.id
}

// Metadata is a metadata getter.
func (p *Pipeline[T]) Metadata() Metadata {
	p.head.mu.Lock()
	defer p.head.mu.Unlock()
	return p.head.metadata
}
