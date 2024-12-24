package iocast

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	STATUS_PENDING = "PENING"
	STATUS_RUNNING = "RUNNING"
	STATUS_FAILED  = "FAILED"
	STATUS_SUCCESS = "SUCCESS"
)

// Task represents a task to be executed.
type Task interface {
	Id() string
	Exec()
	Write() error
	Metadata() metadata
}

type status string

type metadata struct {
	CreatetAt time.Time
	StartedAt time.Time
	Elapsed   time.Duration
	Status    status
}

// Result is the output of a task's execution.
type Result[T any] struct {
	Out      T
	Err      error
	Metadata metadata
}

type taskFn[T any] func(ctx context.Context, previousResult Result[T]) Result[T]

type task[T any] struct {
	mu         sync.RWMutex
	id         string
	ctx        context.Context
	taskFn     taskFn[T]
	resultChan chan Result[T]
	next       *task[T]
	maxRetries int
	writer     ResultWriter
	metadata   metadata
}

// NewTaskFunc initializes and returns a new task func.
func NewTaskFunc[Arg, T any](args Arg, fn func(ctx context.Context, args Arg) (T, error)) taskFn[T] {
	return func(ctx context.Context, previous Result[T]) Result[T] {
		out, err := fn(ctx, args)
		return Result[T]{Out: out, Err: err}
	}
}

// NewTaskFuncWithPreviousResult initializes and returns a new task func that can use the precious task's result.
func NewTaskFuncWithPreviousResult[Arg, T any](args Arg, fn func(ctx context.Context, args Arg, previousResult Result[T]) (T, error)) taskFn[T] {
	return func(ctx context.Context, previous Result[T]) Result[T] {
		out, err := fn(ctx, args, previous)
		return Result[T]{Out: out, Err: err}
	}
}

func (t *task[T]) link(next *task[T]) {
	t.next = next
}

func (t *task[T]) markRunning() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.metadata.StartedAt = time.Now().UTC()
	t.metadata.Status = STATUS_RUNNING
}

func (t *task[T]) markFailed() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.metadata.Elapsed = time.Since(t.metadata.StartedAt)
	t.metadata.Status = STATUS_FAILED
}

func (t *task[T]) markSuccess() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.metadata.Elapsed = time.Since(t.metadata.StartedAt)
	t.metadata.Status = STATUS_SUCCESS
}

func (t *task[T]) retry(previous Result[T]) Result[T] {
	var result Result[T]

	t.markRunning()
	for i := 0; i < t.maxRetries; i++ {
		result = t.taskFn(t.ctx, previous)
		if result.Err == nil {
			t.markSuccess()
			break
		}
	}
	return result
}

// Wait blocks on the result channel of the task until it is ready.
func (t *task[T]) Wait() <-chan Result[T] {
	return t.resultChan
}

// Id is an ID geter.
func (t *task[T]) Id() string {
	return t.id
}

// Wait blocks on the result channel if there's a writer and writes the result when ready.
func (t *task[T]) Write() error {
	if t.writer != nil {
		select {
		case result := <-t.resultChan:
			return t.writer.Write(t.id, Result[any]{
				Out:      result.Out,
				Err:      result.Err,
				Metadata: result.Metadata,
			})
		case <-t.ctx.Done():
			return t.ctx.Err()
		}
	}
	return nil
}

// Exec executes the task.
func (t *task[T]) Exec() {
	var idx int = 1
	var result Result[T]

	result = t.retry(result)
	if result.Err != nil {
		// it's a pipeline so wrap the error
		if t.next != nil {
			result.Err = fmt.Errorf("error in task number %d: %w", idx, result.Err)
		}
		t.markFailed()
		t.resultChan <- result
		close(t.resultChan)
		return
	}
	for t.next != nil {
		idx++

		result = t.next.retry(result)
		if result.Err != nil {
			result.Err = fmt.Errorf("error in task number %d: %w", idx, result.Err)
			// mark the head of the pipeline
			t.markFailed()
			break
		}
		t.next = t.next.next
	}
	result.Metadata = t.metadata
	t.resultChan <- result
	close(t.resultChan)
}

func (t *task[T]) Metadata() metadata {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.metadata
}
