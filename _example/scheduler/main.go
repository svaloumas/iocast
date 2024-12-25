package main

import (
	"context"
	"log"
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
	t := iocast.TaskBuilder("uuid", taskFn).Build()

	// create the scheduler
	s := iocast.NewScheduler(p, 100*time.Millisecond)
	defer s.Stop()

	// run it
	s.Dispatch()

	// schedule the task
	err := s.Schedule(t, time.Now().Add(200*time.Millisecond))
	if err != nil {
		log.Printf("err: %v", err)
	}

	// wait for the result
	result := <-t.Wait()
	log.Printf("result: %+v\n", result)
}
