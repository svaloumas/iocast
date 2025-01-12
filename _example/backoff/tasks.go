package main

import (
	"context"
	"errors"
	"fmt"
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
	return "", errors.New("some error")
}

func DownloadContent(ctx context.Context, args *Args) (string, error) {
	fmt.Println("attempting to download content...")

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
