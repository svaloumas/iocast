package main

import (
	"context"
	"log"

	"github.com/svaloumas/iocast"
)

func main() {
	// create the queue
	q := iocast.NewQueue(4, 8)
	q.Start(context.Background())
	defer q.Stop()

	// create a task func
	args := &Args{addr: "http://somewhere.net", id: 1}
	taskFn := iocast.NewTaskFunc(args, DownloadContent)

	// create a wrapper task
	t := iocast.NewTask(context.Background(), taskFn)

	// enqueue the task
	ok := q.Enqueue(t)
	if !ok {
		log.Fatal("queue is full")
	}

	// wait for the result
	result := <-t.Wait()
	log.Printf("result: %+v\n", result)
}
