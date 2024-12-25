package iocast

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	p := NewWorkerPool(4, 8)
	p.Start(context.Background())
	defer p.Stop()

	taskFn := NewTaskFunc("args", testTaskFn)

	task := TaskBuilder("uuid", taskFn).Build()

	m := &sync.Map{}
	db := NewScheduleDB(m)
	s := NewScheduler(db, p, 50*time.Millisecond)
	defer s.Stop()

	s.Dispatch()

	err := s.Schedule(task, time.Now().Add(50*time.Millisecond))
	if err != nil {
		log.Printf("err: %v", err)
	}

	result := <-task.Wait()
	expected := "args"
	if result.Out != expected {
		t.Errorf("wrong result output: got %v want %v", result.Out, expected)
	}
}
