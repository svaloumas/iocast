# iocast

A zero-dependency async task running library that aims to be simple, easy to use and flexible.

## install

```bash
go get github.com/svaloumas/iocast
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

* Generic argument type. Pass any builtin or custom type as argument to your tasks.
* Generic return value. Return any type of value from your tasks.
* Optionally pass context to your tasks.
* Set retry-policy to each task.
* Create pipelines to execute tasks sequentially. 
* Optionally pass the result of the task to the next one as argument in a pipeline.
* Result writer interface. Memory DB writer is provided but you can optionally implement writers for arbitrary storage engines.
* Request for task metadata like status, time of creation, execution, elapsed time, etc. Metadata also get written also to the storage as part of the result.
* Scheduler. Pass either a specified timestamp for the task to be executed or a crontab for periodic execution. (WIP)

## test

```bash
go test -v ./...
```
