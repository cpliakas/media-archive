package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	EventListener(cancel)

	root := "/Users/chris.pliakas/media-archive"

	out1, _ := DirectoryWatcher(ctx, root)
	out2, _ := DirectoryScanner(ctx, root)

	go func() {
		out := merge(out1, out2)
		for n := range out {
			fmt.Println(n)
		}
	}()

	// Wait for the cancel function to be called.
	<-ctx.Done()
}

// EventListener listens for shutdown signals and calls the context's cancel
// function when received.
func EventListener(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sig := range c {
			log.Printf("shutdown signal received [signal=%s]", sig)
			cancel()
			break
		}
	}()
}

// merge implements the fan-in pattern
// https://blog.golang.org/pipelines
func merge(cs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	// Start an output goroutine for each input channel in cs. output copies
	// values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan string) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done. This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
