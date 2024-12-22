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

type Args1 struct {
	name string
	age  int
}

type Out struct {
	message  string
	totalAge int
}

func main() {
	q := iocast.NewQueue(4, 8)
	q.StartWithContext(context.Background())
	defer q.Stop()

	args := &Args{name: "bob", age: 9}
	taskFunc := func(ctx context.Context, args *Args) (Out, error) {
		// do magic
		out := Out{
			message: fmt.Sprintf("beautiful %s", args.name), totalAge: args.age,
		}
		return out, nil
	}
	task := iocast.NewTask(context.Background(), args, taskFunc)

	args1 := &Args1{name: "alice", age: 11}
	taskFunc1 := func(ctx context.Context, args1 *Args1, previous iocast.Result[Out]) (Out, error) {
		// do magic
		out := Out{
			message:  fmt.Sprintf("the most %s and %s", previous.Out.message, args1.name),
			totalAge: previous.Out.totalAge + args1.age,
		}
		return out, nil
	}
	task1 := iocast.NewTaskWithPreviousResult(context.Background(), args1, taskFunc1)

	j := iocast.NewJob(task)
	j1 := iocast.NewJob(task1)

	// Link the jobs together.
	j.Link(j1)

	// Enqueue only the first one.
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
