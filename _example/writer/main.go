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
	db := &sync.Map{}
	w := iocast.NewMemWriter(db)
	t := iocast.TaskBuilder("uuid", taskFn).ResultWriter(w).Build()

	// enqueue the task
	ok := p.Enqueue(t)
	if !ok {
		log.Fatal("queue is full")
	}

	// wait for the result
	time.Sleep(500 * time.Millisecond)
	data, ok := db.Load("uuid")
	if !ok {
		log.Fatal("result not written")
	}
	bytes := data.([]byte)
	log.Printf("result: %+v\n", string(bytes))
}
