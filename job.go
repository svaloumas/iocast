package main

import (
	"context"
	"fmt"
)

type Job interface {
	Exec()
}

type result[T any] struct {
	out T
	err error
}

type task[Current, Previous any] func(previousResult result[Previous]) result[Current]

type job[Current, Previous any] struct {
	task       task[Current, Previous]
	resultChan chan result[Current]
	next       *job[Current, Previous]
}

func NewTask[Arg, Current, Previous any](ctx context.Context, args Arg, fn func(ctx context.Context, args Arg, previousResult result[Previous]) (Current, error)) task[Current, Previous] {
	return func(previous result[Previous]) result[Current] {
		out, err := fn(ctx, args, previous)
		return result[Current]{out: out, err: err}
	}
}

func NewJob[Current, Previous any](task task[Current, Previous]) *job[Current, Previous] {
	return &job[Current, Previous]{
		task:       task,
		resultChan: make(chan result[Current], 1),
	}
}

func (j *job[Current, Previous]) Link(next *job[Current, Previous]) {
	j.next = next
}

func (j *job[Current, Previous]) Wait() <-chan result[Current] {
	return j.resultChan
}

func (j *job[Current, Previous]) Exec() {
	var idx int = 1
	var previous result[Previous]
	var current result[Current]

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
		previous = result[Previous]{out: out, err: current.err}

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
