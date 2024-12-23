package iocast

import (
	"context"
	"errors"
	"testing"
)

var (
	retries = 0
)

func testTaskFn(ctx context.Context, args string) (string, error) {
	return args, nil
}

func testTaskFnWithContext(ctx context.Context, args string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return "meant to be cancelled", nil
	}
}

func testFailingTaskFn(ctx context.Context, args string) (string, error) {
	retries++
	return "", errors.New("something went wrong")
}

func TestTask(t *testing.T) {
	args := "test"

	taskFn := NewTaskFunc(args, testTaskFn)
	task := TaskBuilder("simple", taskFn).Build()

	taskFnWithContext := NewTaskFunc(args, testTaskFnWithContext)
	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	taskWithContext := TaskBuilder("context", taskFnWithContext).Context(ctx).Build()

	taskFnWithRetries := NewTaskFunc(args, testFailingTaskFn)
	taskWithRetries := TaskBuilder("retries", taskFnWithRetries).MaxRetries(3).Build()

	tests := []struct {
		name     string
		expected string
		task     Task
	}{
		{
			"simple task",
			args,
			task,
		},
		{
			"task with context",
			args,
			taskWithContext,
		},
		{
			"task with retries",
			args,
			taskWithRetries,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.task.Exec()

			switch tt.name {
			case "simple task":
				result := <-task.Wait()
				if result.Out != args {
					t.Errorf("Exec returned unexpected result output: got %v want %v", result.Out, args)
				}
			case "task with context":
				result := <-taskWithContext.Wait()
				if result.Err.Error() != ctx.Err().Error() {
					t.Errorf("Exec returned unexpected result error: got %v want %v", result.Err.Error(), ctx.Err().Error())
				}
			case "task with retries":
				result := <-taskWithRetries.Wait()
				expectedMsg := "something went wrong"
				if result.Err.Error() != expectedMsg {
					t.Errorf("Exec returned unexpected result error: got %v want %v", result.Err.Error(), expectedMsg)
				}
				if retries != 3 {
					t.Error("unexpected retry attempts made")
				}
			}
		})
	}
}
