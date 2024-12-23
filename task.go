package iocast

import (
	"context"
	"fmt"
)

type Task interface {
	Exec()
}

type Result[T any] struct {
	Out T
	Err error
}

type taskFn[T any] func(ctx context.Context, previousResult Result[T]) Result[T]

type task[T any] struct {
	ctx        context.Context
	taskFn     taskFn[T]
	resultChan chan Result[T]
	next       *task[T]
	maxRetries int
}

func NewTaskFunc[Arg, T any](args Arg, fn func(ctx context.Context, args Arg) (T, error)) taskFn[T] {
	return func(ctx context.Context, previous Result[T]) Result[T] {
		out, err := fn(ctx, args)
		return Result[T]{Out: out, Err: err}
	}
}

func NewTaskFuncWithPreviousResult[Arg, T any](args Arg, fn func(ctx context.Context, args Arg, previousResult Result[T]) (T, error)) taskFn[T] {
	return func(ctx context.Context, previous Result[T]) Result[T] {
		out, err := fn(ctx, args, previous)
		return Result[T]{Out: out, Err: err}
	}
}

func (t *task[T]) link(next *task[T]) {
	t.next = next
}

func (t *task[T]) retry(previous Result[T]) Result[T] {
	var result Result[T]
	for i := 0; i < t.maxRetries; i++ {
		result = t.taskFn(t.ctx, previous)
		if result.Err == nil {
			break
		}
	}
	return result
}

func (t *task[T]) Wait() <-chan Result[T] {
	return t.resultChan
}

func (t *task[T]) Exec() {
	var idx int = 1
	var result Result[T]

	result = t.retry(result)
	if result.Err != nil {
		result.Err = fmt.Errorf("error in task number %d: %w", idx, result.Err)
		t.resultChan <- result
		close(t.resultChan)
		return
	}
	for t.next != nil {
		idx++

		result = t.next.retry(result)
		if result.Err != nil {
			result.Err = fmt.Errorf("error in task number %d: %w", idx, result.Err)
			break
		}
		t.next = t.next.next
	}
	t.resultChan <- result
	close(t.resultChan)
}
