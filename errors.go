package main

import (
	"log"
	"sync"
)

// HandleErrors logs all errors sent through the passed channels.
// TODO COme up with a better logging strategy.
func HandleErrors(in ...<-chan error) {
	go func() {
		out := merge(in...)
		for err := range out {
			log.Println("ERROR", err)
		}
	}()
}

// merge implements the fan-in pattern and merges the passed error channels
// into a single channel that can be processed.
//
// https://blog.golang.org/pipelines
func merge(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)

	// Start an output goroutine for each input channel in cs. output copies
	// values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan error) {
		for err := range c {
			out <- err
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
