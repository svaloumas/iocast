package main

import (
	"context"
	"fmt"
	"log"

	"github.com/svaloumas/iocast"
)

type Args struct {
	name string
	age  int
}

type Out struct {
	out string
}

func main() {
	q := iocast.NewQueue(4, 8)
	q.StartWithContext(context.Background())
	defer q.Stop()

	args := &Args{name: "maria", age: 3}
	taskFunc := func(ctx context.Context, args *Args) (Out, error) {
		// do magic
		out := Out{
			out: fmt.Sprintf("%s is only %d", args.name, args.age),
		}
		return out, nil
	}
	task := iocast.NewTask(context.Background(), args, taskFunc)

	j := iocast.NewJob(task)

	ok := q.Enqueue(j)
	if !ok {
		log.Fatal("queue is full")
	}

	result := <-j.Wait()
	if result.Err != nil {
		log.Fatalf("got an error: %v", result.Err)
	}

	fmt.Printf("result: %+v", result.Out)
}
