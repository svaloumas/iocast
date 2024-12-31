package iocast

import (
	"context"
	"log"
	"sync"
)

type WorkerPool struct {
	queue   chan Job
	workers int
	wg      *sync.WaitGroup
}

// NewWorkerPool initializes and returns new workerpool instance.
func NewWorkerPool(workers, capacity int) *WorkerPool {
	return &WorkerPool{
		queue:   make(chan Job, capacity),
		workers: workers,
		wg:      &sync.WaitGroup{},
	}
}

// Enqueue pushes a task to the queue.
func (p WorkerPool) Enqueue(t Job) bool {
	select {
	case p.queue <- t:
		return true
	default:
		return false
	}
}

// Start starts the worker pool pattern.
func (p WorkerPool) Start(ctx context.Context) {
	for _ = range p.workers {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case j, ok := <-p.queue:
					if !ok {
						return
					}
					go func() {
						err := j.Write()
						if err != nil {
							log.Printf("error writing the result of task %s: %v", j.ID(), err)
						}
					}()
					j.Exec()
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// Stop closes the queue and the worker pool gracefully.
func (p WorkerPool) Stop() {
	close(p.queue)
	// Wait for the workers to run their last tasks.
	p.wg.Wait()
}
