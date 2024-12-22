package iocast

import (
	"context"
	"fmt"
)

type Job interface {
	Exec()
}

type Result[T any] struct {
	out T
	err error
}

type task[Current, Previous any] func(previousResult Result[Previous]) Result[Current]

type job[Current, Previous any] struct {
	task       task[Current, Previous]
	resultChan chan Result[Current]
	next       *job[Current, Previous]
}

func NewTask[Arg, Current, Previous any](ctx context.Context, args Arg, fn func(ctx context.Context, args Arg, previousResult result[Previous]) (Current, error)) task[Current, Previous] {
	return func(previous Result[Previous]) Result[Current] {
		out, err := fn(ctx, args, previous)
		return Result[Current]{out: out, err: err}
	}
}

func NewJob[Current, Previous any](task task[Current, Previous]) *job[Current, Previous] {
	return &job[Current, Previous]{
		task:       task,
		resultChan: make(chan Result[Current], 1),
	}
}

func (j *job[Current, Previous]) Link(next *job[Current, Previous]) {
	j.next = next
}

func (j *job[Current, Previous]) Wait() <-chan Result[Current] {
	return j.resultChan
}

func (j *job[Current, Previous]) Exec() {
	var idx int = 1
	var previous Result[Previous]
	var current Result[Current]

	current = j.task(previous)
	if current.err != nil {
		current.err = fmt.Errorf("error in job number %d: %w", idx, current.err)
		j.resultChan <- current
		close(j.resultChan)
		return
	}
	for j.next != nil {
		idx++
		out, ok := any(current.out).(Previous)
		if !ok {
			panic("damn...")
		}
		previous = Result[Previous]{out: out, err: current.err}

		current = j.next.task(previous)
		if current.err != nil {
			current.err = fmt.Errorf("error in job number %d: %w", idx, current.err)
			break
		}
		j.next = j.next.next
	}
	j.resultChan <- current
	close(j.resultChan)
}
