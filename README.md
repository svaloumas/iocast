# iocast

A zero-dependency async task running library that aims to be simple, easy to use and flexible.

## install

```bash
go get github.com/svaloumas/iocast@v0.1.0
```

## usage

The module utilizes Go Generics internally, enabling the flexibility to define your custom structs to use as arguments and arbitrary result types in your tasks.

```go
func DownloadContent(ctx context.Context, args *Args) (string, error) {

	contentChan := make(chan []byte)
	go func() {
		contentChan <- fetchContent(args.addr, args.id)
		close(contentChan)
	}()
	select {
	case content := <-contentChan:
		return saveToDisk(content)
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	q := iocast.NewWorkerPool(4, 8)
	q.Start(context.Background())
	defer q.Stop()

	args := &Args{addr: "http://somewhere.net", id: 1}
	taskFn := iocast.NewTaskFunc(args, DownloadContent)

	t := iocast.TaskBuilder(taskFn).Context(context.Background()).MaxRetries(3).Build()
	q.Enqueue(t)

	m := t.Metadata()
	log.Printf("status: %s", m.Status)

	result := <-t.Wait()
}
```

See [examples](_example/) for a detailed illustration of how to run simple tasks and linked tasks as pipelines.

## features

* Generic Task Arguments. Pass any built-in or custom type as an argument to your tasks.
* Flexible Task Results. Return any type of value from your tasks.
* Context Awareness. Optionally include a context when running tasks.
* Retry Policy. Define the number of retry attempts for each task.
* Task Pipelines. Chain tasks to execute sequentially, with the option to pass the result of one task as the argument for the next.
* Database Interface. Use the built-in in-memory database or use custom drivers for other storage engines by implementing an one-func interface.
* Task Metadata. Retrieve metadata such as status, creation time, execution time, and elapsed time. Metadata is also stored with the task results.
* Scheduler: Schedule tasks to run at a specific timestamp. Scheduler uses a built-in in-memory database for the schedule entities, but you can implement a simple interface to use your own drivers for other storage engines.

## test

```bash
go test -v ./...
```
