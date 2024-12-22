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

type OtherArgs struct {
	foo []byte
	bar int8
}

func main() {
	q := iocast.NewQueue(4, 8)
	q.StartWithContext(context.Background())
	defer q.Stop()

	myArgs := &Args{name: "maria", age: 3}
	taskFunc := func(ctx context.Context, args *Args) (string, error) {
		// do magic
		return fmt.Sprintf("%s is only %d", args.name, args.age), nil
	}
	task := iocast.NewTask(context.Background(), myArgs, taskFunc)

	j := iocast.NewJob(task)

	ok := q.Enqueue(j)
	if !ok {
		log.Fatal("queue is full")
	}

	otherArgs := &OtherArgs{foo: []byte("foo"), bar: 3}
	taskFunc1 := func(ctx context.Context, args *OtherArgs) (int, error) {
		// do magic
		return len(args.foo) + int(args.bar), nil
	}
	task1 := iocast.NewTask(context.Background(), otherArgs, taskFunc1)

	j1 := iocast.NewJob(task1)

	ok = q.Enqueue(j1)
	if !ok {
		log.Fatal("queue is full")
	}

	result := <-j.Wait()
	fmt.Printf("result: %+v\n", result)

	result1 := <-j1.Wait()
	fmt.Printf("result1: %+v\n", result1)
}
