package main

import (
	"context"
	"time"

	"github.com/svaloumas/iocast"
)

type DownloadArgs struct {
	addr string
	id   int
}

type ProcessArgs struct {
	mode string
}

type UploadArgs struct {
	addr string
}

type Output struct {
	path          string
	processedPath string
	uploadedPath  string
}

func fetchContent(addr string, id int) []byte {
	// do some heavy work
	time.Sleep(300 * time.Millisecond)
	return []byte("content")
}

func saveToDisk(content []byte) (string, error) {
	return "path/to/content", nil
}

func processContent(path, mode string) string {
	// do some heavy work
	time.Sleep(300 * time.Millisecond)
	return "path/to/processed/content"
}

func uploadContent(processedPath string) string {
	// do some heavy work
	time.Sleep(300 * time.Millisecond)
	return "path/to/uploaded/content"
}

func DownloadContent(ctx context.Context, args *DownloadArgs) (*Output, error) {

	contentChan := make(chan []byte)
	go func() {
		contentChan <- fetchContent(args.addr, args.id)
		close(contentChan)
	}()

	select {
	case content := <-contentChan:
		path, err := saveToDisk(content)
		if err != nil {
			return nil, err
		}
		return &Output{path: path}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func ProcessContent(ctx context.Context, args *ProcessArgs, previous iocast.Result[*Output]) (*Output, error) {

	pathChan := make(chan string)
	go func() {
		pathChan <- processContent(previous.Out.path, args.mode)
		close(pathChan)
	}()

	select {
	case processedPath := <-pathChan:
		return &Output{
			path:          previous.Out.path,
			processedPath: processedPath,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func UploadContent(ctx context.Context, args *UploadArgs, previous iocast.Result[*Output]) (*Output, error) {

	pathChan := make(chan string)
	go func() {
		pathChan <- uploadContent(previous.Out.processedPath)
		close(pathChan)
	}()

	select {
	case uploadedPath := <-pathChan:
		return &Output{
			path:          previous.Out.path,
			processedPath: previous.Out.processedPath,
			uploadedPath:  uploadedPath,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
