package main

import (
	"context"
)

type Job interface {
	Exec()
}

type result[T any] struct {
	out T
	err error
}

type task[T any] func() result[T]

type job[T any] struct {
	task       task[T]
	resultChan chan result[T]
}

func NewTask[T, R any](ctx context.Context, args T, fn func(ctx context.Context, args T) (R, error)) task[R] {
	return func() result[R] {
		out, err := fn(ctx, args)
		return result[R]{out: out, err: err}
	}
}

func NewJob[T any](task task[T]) *job[T] {
	return &job[T]{
		task:       task,
		resultChan: make(chan result[T], 1),
	}
}

func (j *job[T]) Wait() <-chan result[T] {
	return j.resultChan
}

func (j *job[T]) Exec() {
	j.resultChan <- j.task()
	close(j.resultChan)
}
