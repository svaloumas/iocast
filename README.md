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
	q := iocast.NewQueue(4, 8)
	q.Start(context.Background())
	defer q.Stop()

	args := &Args{addr: "http://somewhere.net", id: 1}
	taskFn := iocast.NewTaskFunc(args, DownloadContent)

	t := iocast.TaskBuilder(taskFn).Context(context.Background()).MaxRetries(3).Build()
	q.Enqueue(t)

	result := <-t.Wait()
}
```

See [examples](_example/) for a detailed illustration of how to run simple tasks and linked tasks as pipelines.
