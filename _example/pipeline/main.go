package main

import (
	"context"
	"log"

	"github.com/svaloumas/iocast"
)

func main() {
	// create the worker pool
	q := iocast.NewWorkerPool(4, 8)
	q.Start(context.Background())
	defer q.Stop()

	// create the task funcs
	downloadArgs := &DownloadArgs{addr: "http://somewhere.net", id: 1}
	downloadFn := iocast.NewTaskFunc(downloadArgs, DownloadContent)

	processArgs := &ProcessArgs{mode: "MODE_1"}
	processFn := iocast.NewTaskFuncWithPreviousResult(processArgs, ProcessContent)

	uploadArgs := &UploadArgs{addr: "http://storage.net/path/to/file"}
	uploadFn := iocast.NewTaskFuncWithPreviousResult(uploadArgs, UploadContent)

	// create the wrapper tasks
	downloadTask := iocast.TaskBuilder("download", downloadFn).Context(context.Background()).MaxRetries(5).Build()
	processTask := iocast.TaskBuilder("process", processFn).Context(context.Background()).MaxRetries(4).Build()
	uploadTask := iocast.TaskBuilder("upload", uploadFn).Context(context.Background()).MaxRetries(3).Build()

	// create the pipeline
	p, err := iocast.NewPipeline(downloadTask, processTask, uploadTask)
	if err != nil {
		log.Fatalf("error creating a pipeine: %s", err)
	}

	// enqueue the pipeline
	ok := q.Enqueue(p)
	if !ok {
		log.Fatal("queue is full")
	}

	// wait for the result
	result := <-p.Wait()
	log.Printf("result out: %+v", *result.Out)
}
