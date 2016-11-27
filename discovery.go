package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// A Discoverer finds media files and sends them to the returned channel.
type Discoverer func(ctx context.Context, root string) (<-chan string, <-chan error)

// DirectoryWatcher is a Discoverer that watches root for file modifications
// and sends the media files and errors to the returned channels.
//
// See http://stackoverflow.com/a/6612243 for filepath.Walk() technique.
func DirectoryWatcher(ctx context.Context, root string) (<-chan string, <-chan error) {
	out := make(chan string)
	errs := make(chan error)

	go func() {

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			errs <- err
			return
		}
		defer watcher.Close()

		go func() {
			for {
				select {
				case event := <-watcher.Events:
					out <- event.Name
				case err := <-watcher.Errors:
					errs <- err
				}
			}
		}()

		err = watcher.Add(root)
		if err != nil {
			errs <- err
			return
		}

		<-ctx.Done()
	}()

	return out, errs
}

// DirectoryScanner is a Discoverer that iterates over all files in root,
// sending the matched files and errors to the returned channels.
//
// See http://stackoverflow.com/a/6612243 for filepath.Walk() technique.
func DirectoryScanner(ctx context.Context, root string) (<-chan string, <-chan error) {
	out := make(chan string)
	errs := make(chan error)

	go func() {
		err := filepath.Walk(root, walkFunc(ctx, out))
		if err != nil {
			errs <- err
		}

		<-ctx.Done()
	}()

	return out, errs
}

// walkFunc sends media files to the out channel and no-ops the rest.
func walkFunc(ctx context.Context, out chan string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {

		// Only files, please.
		if info.IsDir() {
			return nil
		}

		// TODO Filter out non-media files

		out <- path
		return nil
	}
}
