package iocast

import (
	"context"
	"sync"
)

type Queue struct {
	queue   chan Task
	workers int
	wg      *sync.WaitGroup
}

func NewQueue(workers, capacity int) *Queue {
	return &Queue{
		queue:   make(chan Task, capacity),
		workers: workers,
		wg:      &sync.WaitGroup{},
	}
}

func (q Queue) Enqueue(t Task) bool {
	select {
	case q.queue <- t:
		return true
	default:
		return false
	}
}

func (q Queue) Start(ctx context.Context) {
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
					t.Exec()
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (q Queue) Stop() {
	close(q.queue)
	// Wait for the workers to run their last tasks.
	q.wg.Wait()
}
