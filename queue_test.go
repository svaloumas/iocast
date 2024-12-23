package iocast

import (
	"context"
	"testing"
)

func testTaskFn(ctx context.Context, args string) (string, error) {
	return args, nil
}

func TestEnqueue(t *testing.T) {
	q := NewQueue(1, 1)
	q.Start(context.Background())
	defer q.Stop()

	q2 := NewQueue(0, 0) // will refuse to enqueue
	q2.Start(context.Background())
	defer q2.Stop()

	args := "test"
	taskFn := NewTaskFunc(args, testTaskFn)
	task := TaskBuilder(taskFn).Build()
	task2 := TaskBuilder(taskFn).Build()

	tests := []struct {
		name     string
		expected string
		q        *queue
	}{
		{
			"ok",
			args,
			q,
		},
		{
			"full queue",
			args,
			q2,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if i == 0 {
				ok := tt.q.Enqueue(task)
				if !ok {
					t.Errorf("unexpected full queue")
				}
			} else {
				ok := tt.q.Enqueue(task2)
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
