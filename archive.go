package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ArchiveMedia reads the files send through all of the in channels and
// sends them to the configured backends.
//
// TODO Implement storage adapters.
func ArchiveMedia(watcher *MediaWatcher, archive, bucket string) <-chan error {
	errs := make(chan error)

	go func() {

		sess, err := session.NewSession()
		if err != nil {
			errs <- err
			return
		}

		svc := s3.New(sess)

		for path := range watcher.Media() {

			// TODO check the cache

			dat, err := ioutil.ReadFile(path)
			if err != nil {
				errs <- err
				continue
			}

			params := &s3.PutObjectInput{
				Bucket:       aws.String(bucket),
				Key:          aws.String(fmt.Sprintf("%s/%s", archive, watcher.RelativePath(path))),
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
			log.Printf("file uploaded [filepath=%s]", path)
		}
	}()

	return errs
}
