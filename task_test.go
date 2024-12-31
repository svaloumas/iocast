package iocast

import (
	"context"
	"errors"
	"testing"
)

var (
	retries = 0
)

func testTaskFn(_ context.Context, args string) (string, error) {
	return args, nil
}

func testTaskFnWithContext(ctx context.Context, _ string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return "meant to be cancelled", nil
	}
}

func testFailingTaskFn(_ context.Context, _ string) (string, error) {
	retries++
	return "", errors.New("something went wrong")
}

func TestTask(t *testing.T) {
	args := "test"

	taskFn := NewTaskFunc(context.Background(), args, testTaskFn)
	task := TaskBuilder("simple", taskFn).Build()

	taskFnWithRetries := NewTaskFunc(context.Background(), args, testFailingTaskFn)
	taskWithRetries := TaskBuilder("retries", taskFnWithRetries).MaxRetries(3).Build()

	tests := []struct {
		name     string
		expected string
		job      Job
	}{
		{
			"simple task",
			args,
			task,
		},
		{
			"task with retries",
			args,
			taskWithRetries,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.job.Exec()

			switch tt.name {
			case "simple task":
				result := <-task.Wait()
				if result.Out != args {
					t.Errorf("Exec returned unexpected result output: got %v want %v", result.Out, args)
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
