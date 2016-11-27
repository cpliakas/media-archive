package main

import (
	"fmt"
	"io/ioutil"
	"sync"
)

// ArchiveMedia reads the files send through all of the in channels and
// sends them to the configured backends.
func ArchiveMedia(in ...<-chan string) {
	go func() {
		out := merge(in...)
		for path := range out {

			dat, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err) // TODO Don't panic.
			}

			// TODO This is where we send the data to S3.
			fmt.Println(hash(dat), path)
		}
	}()
}

// merge implements the fan-in pattern and merges the channels passed to it
// into a single channel that can be processed.
//
// https://blog.golang.org/pipelines
func merge(cs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	// Start an output goroutine for each input channel in cs. output copies
	// values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan string) {
		for p := range c {
			out <- p
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
