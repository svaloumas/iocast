# iocast

An zero-dependency async task running library that aims to be simple, easy to use and flexible.

## install

```bash
go get github.com/svaloumas/iocast
```

## usage

It utilizes Go Generics internally, enabling the flexibility to define your custom structs to use as arguments and arbitrary result types in your tasks.

```go
	q := iocast.NewQueue(4, 8)
	q.Start(context.Background())
	defer q.Stop()

	myArgs := &Args{name: "maria", age: 3}
	taskFunc := func(ctx context.Context, args *Args) (string, error) {
		// do magic
		return fmt.Sprintf("%s is only %d", args.name, args.age), nil
	}
	task := iocast.NewTask(context.Background(), myArgs, taskFunc)

	j := iocast.NewJob(task)
	q.Enqueue(j)

	result := <-j.Wait()
```

See [examples](_example/) for a detailed illustration of how to run simple tasks and pipelines.
