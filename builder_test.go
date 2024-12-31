package iocast

import (
	"context"
	"testing"
)

func TestTaskBuilder(t *testing.T) {
	args := "test"

	taskFn := NewTaskFunc(context.Background(), args, testTaskFn)
	task := TaskBuilder("id", taskFn).MaxRetries(3).Build()

	if task.maxRetries != 3 {
		t.Errorf("TaskBuilder set wrong task maxRetries: got %v want %v", task.maxRetries, 3)
	}
}
