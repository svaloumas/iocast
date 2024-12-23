package iocast

import (
	"context"
	"log"
	"sync"
)

type queue struct {
	queue   chan Task
	workers int
	wg      *sync.WaitGroup
}

// NewQueue initializes and returns new Queue instance.
func NewQueue(workers, capacity int) *queue {
	return &queue{
		queue:   make(chan Task, capacity),
		workers: workers,
		wg:      &sync.WaitGroup{},
	}
}

// Enqueue pushes a task to the queue.
func (q queue) Enqueue(t Task) bool {
	select {
	case q.queue <- t:
		return true
	default:
		return false
	}
}

// Start starts the worker pool pattern.
func (q queue) Start(ctx context.Context) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go func() {
			defer q.wg.Done()
			for {
				select {
				case t, ok := <-q.queue:
					if !ok {
						return
					}
					go func() {
						err := t.Write()
						if err != nil {
							log.Printf("error writing the result: %v", err)
						}
					}()
					t.Exec()
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// Stop closes the queue and the worker pool gracefully.
func (q queue) Stop() {
	close(q.queue)
	// Wait for the workers to run their last tasks.
	q.wg.Wait()
}
