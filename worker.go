package iocast

import (
	"context"
	"log"
	"sync"
)

type workerpool struct {
	queue   chan Task
	workers int
	wg      *sync.WaitGroup
}

// NewWorkerPool initializes and returns new workerpool instance.
func NewWorkerPool(workers, capacity int) *workerpool {
	return &workerpool{
		queue:   make(chan Task, capacity),
		workers: workers,
		wg:      &sync.WaitGroup{},
	}
}

// Enqueue pushes a task to the queue.
func (p workerpool) Enqueue(t Task) bool {
	select {
	case p.queue <- t:
		return true
	default:
		return false
	}
}

// Start starts the worker pool pattern.
func (p workerpool) Start(ctx context.Context) {
	for _ = range p.workers {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case t, ok := <-p.queue:
					if !ok {
						return
					}
					go func() {
						err := t.Write()
						if err != nil {
							log.Printf("error writing the result of task %s: %v", t.ID(), err)
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
func (p workerpool) Stop() {
	close(p.queue)
	// Wait for the workers to run their last tasks.
	p.wg.Wait()
}
