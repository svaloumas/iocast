package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/svaloumas/iocast"
)

func main() {
	// create the worker pool
	p := iocast.NewWorkerPool(4, 8)
	p.Start(context.Background())
	defer p.Stop()

	// create a task func
	args := &Args{addr: "http://somewhere.net", id: 1}
	taskFn := iocast.NewTaskFunc(args, DownloadContent)

	// create a wrapper task
	t := iocast.TaskBuilder("uuid", taskFn).Context(context.Background()).MaxRetries(3).Build()
	t2 := iocast.TaskBuilder("uuid2", taskFn).Context(context.Background()).MaxRetries(3).Build()

	// schedule the task
	m := &sync.Map{}
	db := iocast.NewScheduleMemDB(m)
	s := iocast.NewScheduler(db, p, time.Second)
	defer s.Stop()

	s.Dispatch()

	err := s.ScheduleRun(t, time.Now().Add(time.Minute))
	if err != nil {
		log.Printf("err: %v", err)
	}

	err = s.Schedule(t2, iocast.Schedule{
		Days:     []time.Weekday{0, 1},
		Interval: 2 * time.Hour,
	})
	if err != nil {
		log.Printf("err: %v", err)
	}

	// wait for the result
	result := <-t.Wait()
	log.Printf("result: %+v\n", result)
}
