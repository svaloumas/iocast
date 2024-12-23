package main

import (
	"context"
	"time"
)

type Args struct {
	addr string
	id   int
}

func fetchContent(addr string, id int) []byte {
	// do some heavy work
	time.Sleep(200 * time.Millisecond)
	return []byte("content")
}

func saveToDisk(content []byte) (string, error) {
	return "path/to/content", nil
}

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
