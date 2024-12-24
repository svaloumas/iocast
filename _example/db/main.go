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
	m := &sync.Map{}
	db := iocast.NewMemDB(m)
	t := iocast.TaskBuilder("uuid", taskFn).Database(db).Build()

	// enqueue the task
	ok := p.Enqueue(t)
	if !ok {
		log.Fatal("queue is full")
	}

	// wait for the result
	time.Sleep(500 * time.Millisecond)
	data, ok := m.Load("uuid")
	if !ok {
		log.Fatal("result not written")
	}
	bytes := data.([]byte)
	log.Printf("result: %+v\n", string(bytes))
}
