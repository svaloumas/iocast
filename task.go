package iocast

import (
	"context"
	"errors"
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

type taskBuilder[T any] struct {
	ctx        context.Context
	taskFn     taskFn[T]
	resultChan chan Result[T]
	next       *task[T]
	maxRetries int
}

type task[T any] struct {
	ctx        context.Context
	taskFn     taskFn[T]
	resultChan chan Result[T]
	next       *task[T]
	maxRetries int
}

type pipeline[T any] struct {
	head       *task[T]
	resultChan chan Result[T]
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

func TaskBuilder[T any](fn taskFn[T]) *taskBuilder[T] {
	t := &taskBuilder[T]{
		taskFn:     fn,
		resultChan: make(chan Result[T], 1),
		maxRetries: 1,
	}
	return t
}

func (b *taskBuilder[T]) Context(ctx context.Context) *taskBuilder[T] {
	b.ctx = ctx
	return b
}

func (b *taskBuilder[T]) MaxRetries(maxRetries int) *taskBuilder[T] {
	if maxRetries < 1 {
		maxRetries = 1
	}
	b.maxRetries = maxRetries
	return b
}

func (b *taskBuilder[T]) Build() *task[T] {
	return &task[T]{
		ctx:        b.ctx,
		taskFn:     b.taskFn,
		resultChan: b.resultChan,
		maxRetries: b.maxRetries,
		next:       b.next,
	}
}

func NewPipeline[T any](tasks ...*task[T]) (*pipeline[T], error) {
	if len(tasks) < 2 {
		return nil, errors.New("at least two tasks must be linked to create a pipeline")
	}
	head := tasks[0]
	for i, t := range tasks {
		if i < len(tasks)-1 {
			t.link(tasks[i+1])
		}
	}
	return &pipeline[T]{
		head:       head,
		resultChan: head.resultChan,
	}, nil
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

func (p *pipeline[T]) Wait() <-chan Result[T] {
	return p.head.resultChan
}

func (p *pipeline[T]) Exec() {
	p.head.Exec()
}
