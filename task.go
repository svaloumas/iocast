package iocast

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	StatusPending = "PENDING"
	StatusRunning = "RUNNING"
	StatusFailed  = "FAILED"
	StatusSuccess = "SUCCESS"
)

// Job represents a task to be executed.
type Job interface {
	ID() string
	Exec()
	Write() error
	Metadata() Metadata
}

type status string

type Metadata struct {
	CreatetAt time.Time     `json:"created_at"`
	StartedAt time.Time     `json:"started_at"`
	Elapsed   time.Duration `json:"elapsed"`
	Status    status        `json:"status"`
}

// Result is the output of a task's execution.
type Result[T any] struct {
	Out      T        `json:"out"`
	Err      error    `json:"err"`
	Metadata Metadata `json:"metadata"`
}

type TaskFn[T any] func(ctx context.Context, previousResult Result[T]) Result[T]

type Task[T any] struct {
	mu         sync.RWMutex
	id         string
	ctx        context.Context
	taskFn     TaskFn[T]
	resultChan chan Result[T]
	next       *Task[T]
	maxRetries int
	db         DB
	metadata   Metadata
}

// NewTaskFunc initializes and returns a new task func.
func NewTaskFunc[Arg, T any](
	args Arg,
	fn func(ctx context.Context, args Arg) (T, error)) TaskFn[T] {
	return func(ctx context.Context, _ Result[T]) Result[T] {
		out, err := fn(ctx, args)
		return Result[T]{Out: out, Err: err}
	}
}

// NewTaskFuncWithPreviousResult initializes and returns a new task func that can use the precious task's result.
func NewTaskFuncWithPreviousResult[Arg, T any](
	args Arg,
	fn func(ctx context.Context, args Arg, previousResult Result[T]) (T, error)) TaskFn[T] {
	return func(ctx context.Context, previous Result[T]) Result[T] {
		out, err := fn(ctx, args, previous)
		return Result[T]{Out: out, Err: err}
	}
}

func (t *Task[T]) link(next *Task[T]) {
	t.next = next
}

func (t *Task[T]) markRunning() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.metadata.StartedAt = time.Now().UTC()
	t.metadata.Status = StatusRunning
}

func (t *Task[T]) markFailed() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.metadata.Elapsed = time.Since(t.metadata.StartedAt)
	t.metadata.Status = StatusFailed
}

func (t *Task[T]) markSuccess() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.metadata.Elapsed = time.Since(t.metadata.StartedAt)
	t.metadata.Status = StatusSuccess
}

func (t *Task[T]) retry(previous Result[T]) Result[T] {
	var result Result[T]

	t.markRunning()
	for _ = range t.maxRetries {
		result = t.taskFn(t.ctx, previous)
		if result.Err == nil {
			t.markSuccess()
			break
		}
	}
	return result
}

// Wait blocks on the result channel of the task until it is ready.
func (t *Task[T]) Wait() <-chan Result[T] {
	return t.resultChan
}

// ID is an ID geter.
func (t *Task[T]) ID() string {
	return t.id
}

// Wait blocks on the result channel if there's a writer and writes the result when ready.
func (t *Task[T]) Write() error {
	if t.db != nil {
		select {
		case result := <-t.resultChan:
			return t.db.Write(t.id, Result[any]{
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
func (t *Task[T]) Exec() {
	idx := 1
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

// Metadata is a metadata getter.
func (t *Task[T]) Metadata() Metadata {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.metadata
}
