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
	taskFn := iocast.NewTaskFunc(context.Background(), args, DownloadContent)

	// create a wrapper task
	t := iocast.TaskBuilder("uuid", taskFn).
		MaxRetries(2).
		BackOff([]time.Duration{2 * time.Second, 5 * time.Second}).
		Build()

	// enqueue the task
	ok := p.Enqueue(t)
	if !ok {
		log.Fatal("queue is full")
	}

	m := t.Metadata()
	log.Printf("status: %s", m.Status)

	// wait for the result
	result := <-t.Wait()
	log.Printf("result: %+v\n", result)
}
