package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
)

// ArchiveMedia reads the files send through all of the in channels and
// sends them to the configured backends.
//
// TODO Implement storage adapters.
func ArchiveMedia(in ...<-chan string) <-chan error {
	errs := make(chan error)

	go func() {

		sess, err := session.NewSession()
		if err != nil {
			errs <- err
			return
		}

		svc := s3.New(sess)

		// TODO Use dependency injection.
		dir := viper.GetString("root-dir")
		bucket := viper.GetString("aws-bucket")
		archive := viper.GetString("archive-name")

		// Makes the path to the file relative.
		normalizePath := func(p string) string {
			d := strings.TrimRight(dir, "/")
			return strings.TrimLeft(strings.TrimPrefix(p, d), "/")
		}

		out := merge(in...)
		for path := range out {

			// TODO check the cache

			dat, err := ioutil.ReadFile(path)
			if err != nil {
				errs <- err
				continue
			}

			params := &s3.PutObjectInput{
				Bucket:       aws.String(bucket),
				Key:          aws.String(fmt.Sprintf("%s/%s", archive, normalizePath(path))),
				Body:         bytes.NewReader(dat),
				StorageClass: aws.String(s3.TransitionStorageClassStandardIa),
			}

			_, err = svc.PutObject(params)
			if err != nil {
				errs <- err
				continue
			}

			// TODO write to cache

			// TODO Figure out a logging strategy.
			log.Println("uploaded file:", path)
		}
	}()

	return errs
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
