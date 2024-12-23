package iocast

import (
	"context"
	"testing"
)

func TestTaskBuilder(t *testing.T) {
	type key struct{}
	var someKey key

	args := "test"

	taskFn := NewTaskFunc(args, testTaskFn)
	ctx := context.WithValue(context.Background(), someKey, "val")
	task := TaskBuilder(taskFn).Context(ctx).MaxRetries(3).Build()

	if task.ctx.Value(someKey) != ctx.Value(someKey) {
		t.Errorf("TaskBuilder set wrong task context: got %v want %v", task.ctx.Value(someKey), ctx.Value(someKey))
	}
	if task.maxRetries != 3 {
		t.Errorf("TaskBuilder set wrong task maxRetries: got %v want %v", task.maxRetries, 3)
	}
}
