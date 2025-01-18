package iocast

import (
	"context"
	"testing"
	"time"
)

func TestTaskBuilder(t *testing.T) {
	args := "test"

	taskFn := NewTaskFunc(context.Background(), args, testTaskFn)
	task := TaskBuilder("id", taskFn).
		MaxRetries(3).
		BackOff(
			[]time.Duration{
				1 * time.Second,
				2 * time.Second,
				3 * time.Second,
			},
		).Build()

	if task.maxRetries != 3 {
		t.Errorf("TaskBuilder set wrong task maxRetries: got %v want %v", task.maxRetries, 3)
	}
	if task.backoff[0] != 1*time.Second {
		t.Errorf("TaskBuilder set wrong task backoff interval: got %v want %v", task.backoff[0], 1*time.Second)
	}
	if task.backoff[1] != 2*time.Second {
		t.Errorf("TaskBuilder set wrong task backoff interval: got %v want %v", task.backoff[1], 2*time.Second)
	}
	if task.backoff[2] != 3*time.Second {
		t.Errorf("TaskBuilder set wrong task backoff interval: got %v want %v", task.backoff[2], 3*time.Second)
	}
}
