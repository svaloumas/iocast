package iocast

import (
	"context"
	"testing"
)

func testTaskPipedFn(_ context.Context, args string, previous Result[string]) (string, error) {
	return args + previous.Out, nil
}

func TestPipeline(t *testing.T) {
	args := "test"

	taskFn := NewTaskFunc(args, testTaskFn)
	task := TaskBuilder("head", taskFn).Build()

	taskPipedFn := NewTaskFuncWithPreviousResult(args, testTaskPipedFn)
	pipedTask := TaskBuilder("second", taskPipedFn).Build()

	p, err := NewPipeline("id", task, pipedTask)
	if err != nil {
		t.Errorf("NewPipeline returned unexpected error: %v", err)
	}

	go p.Exec()

	result := <-p.Wait()
	if result.Err != nil {
		t.Errorf("Wait returned unexpected result error: %v", result.Err)
	}
	expected := args + args
	if result.Out != expected {
		t.Errorf("Wait returned unexpected result output: got %v want %v", result.Out, expected)
	}
}

func TestPipelineWithLessThanTwoTasks(t *testing.T) {
	args := "test"

	taskFn := NewTaskFunc(args, testTaskFn)
	task := TaskBuilder("head", taskFn).Build()

	_, err := NewPipeline("id", task)
	expectedMsg := "at least two tasks must be linked to create a pipeline"
	if err == nil {
		t.Errorf("NewPipeline did not return expected error: got nil want %v", expectedMsg)
	} else if err.Error() != expectedMsg {
		t.Errorf("NewPipeline returned unexpected error: got %v want %v", err.Error(), expectedMsg)
	}
}
