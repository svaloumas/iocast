package iocast

import (
	"context"
	"fmt"
)

type Job interface {
	Exec()
}

type Result[T any] struct {
	Out T
	Err error
}

type task[T any] func(previousResult Result[T]) Result[T]

type job[T any] struct {
	task       task[T]
	resultChan chan Result[T]
	next       *job[T]
}

func NewTask[Arg, T any](ctx context.Context, args Arg, fn func(ctx context.Context, args Arg) (T, error)) task[T] {
	return func(previous Result[T]) Result[T] {
		out, err := fn(ctx, args)
		return Result[T]{Out: out, Err: err}
	}
}

func NewTaskWithPreviousResult[Arg, T any](ctx context.Context, args Arg, fn func(ctx context.Context, args Arg, previousResult Result[T]) (T, error)) task[T] {
	return func(previous Result[T]) Result[T] {
		out, err := fn(ctx, args, previous)
		return Result[T]{Out: out, Err: err}
	}
}

func NewJob[T any](task task[T]) *job[T] {
	return &job[T]{
		task:       task,
		resultChan: make(chan Result[T], 1),
	}
}

func (j *job[T]) Link(next *job[T]) {
	j.next = next
}

func (j *job[T]) Wait() <-chan Result[T] {
	return j.resultChan
}

func (j *job[T]) Exec() {
	var idx int = 1
	var result Result[T]

	result = j.task(result)
	if result.Err != nil {
		result.Err = fmt.Errorf("error in job number %d: %w", idx, result.Err)
		j.resultChan <- result
		close(j.resultChan)
		return
	}
	for j.next != nil {
		idx++

		result = j.next.task(result)
		if result.Err != nil {
			result.Err = fmt.Errorf("error in job number %d: %w", idx, result.Err)
			break
		}
		j.next = j.next.next
	}
	j.resultChan <- result
	close(j.resultChan)
}
