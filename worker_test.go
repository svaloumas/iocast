package iocast

import (
	"context"
	"testing"
)

func TestWorkerPool(t *testing.T) {
	p := NewWorkerPool(1, 1)
	p.Start(context.Background())
	defer p.Stop()

	p2 := NewWorkerPool(0, 0) // will refuse to enqueue
	p2.Start(context.Background())
	defer p2.Stop()

	args := "test"
	taskFn := NewTaskFunc(context.Background(), args, testTaskFn)
	task := TaskBuilder("ok", taskFn).Build()
	task2 := TaskBuilder("full queue", taskFn).Build()

	tests := []struct {
		name     string
		expected string
		p        *WorkerPool
	}{
		{
			"ok",
			args,
			p,
		},
		{
			"full queue",
			args,
			p2,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if i == 0 {
				ok := tt.p.Enqueue(task)
				if !ok {
					t.Errorf("unexpected full queue")
				}
			} else {
				ok := tt.p.Enqueue(task2)
				if ok {
					t.Errorf("unexpected not full queue")
				}
			}

			if i == 0 {
				result := <-task.Wait()
				if result.Out != tt.expected {
					t.Errorf("unexpected result out: got %v want %v", result.Out, tt.expected)
				}
			}
		})
	}
}
