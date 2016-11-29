package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// MediaWatcher recursively watches directories for media files.
type MediaWatcher struct {
	errs    chan error
	media   chan string
	root    string
	watcher *fsnotify.Watcher
}

// NewMediaWatcher returns a MediaWatcher that recursively watches root.
func NewMediaWatcher(root string) (*MediaWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watcher := &MediaWatcher{
		errs:    make(chan error),
		media:   make(chan string),
		root:    strings.TrimRight(root, "/"),
		watcher: w,
	}

	// Start the event handler in a goroutine to act on file creation and
	// modification. New directories are recursively watched, and new files
	// are send to the media channel for processing.
	go watcher.eventHandler()

	// Run this in a goroutine so that we can start listening for errors and
	// media files in the channels contained within the returned struct.
	go watcher.Add(root)

	return watcher, nil
}

// Add watches a directory and scans for subdirectories to watch and
// existing files. Basically this eliminates race conditions and implementes
// recursive directory watching.
func (w *MediaWatcher) Add(path string) {
	// Start watching the directory.
	err := w.watcher.Add(path)
	if err != nil {
		w.errs <- err
		return
	}

	// Scan for directories so we can watch them, and scan for files that
	// were added after the directory was created and before it was watched.
	files, err := ioutil.ReadDir(path)
	if err != nil {
		w.errs <- err
		return
	}
	basedir := strings.TrimRight(path, "/")
	for _, f := range files {
		filepath := fmt.Sprintf("%s/%s", basedir, f.Name())
		if f.IsDir() {
			w.Add(filepath)
		} else {
			w.media <- filepath
		}
	}
}

// Close closes the watcher.
func (w *MediaWatcher) Close() {
	w.watcher.Close()
}

// Media returns a channel that contains discovered media files.
func (w *MediaWatcher) Media() <-chan string {
	return w.media
}

// Errors returns a channel that contains errors encountered watching and
// discovering media files.
func (w *MediaWatcher) Errors() <-chan error {
	return w.errs
}

// RelativePath normalizes path to it's relative representation, i.e. the
// root directory prefix is stripped off.
func (w *MediaWatcher) RelativePath(path string) string {
	return strings.TrimLeft(strings.TrimPrefix(path, w.root), "/")
}

// eventHandler listens for events; It watches new directories and sends
// discovered media files to the media channel.
func (w *MediaWatcher) eventHandler() {
	for {
		select {
		case event := <-w.watcher.Events:
			switch event.Op {
			case fsnotify.Create, fsnotify.Write:
				if w.isDir(event.Name) {
					w.Add(event.Name)
				} else {
					w.media <- event.Name
				}
			}

			// No-op fsnotify.Rename, fsnotify.Remove and fsnotify.Chmod.
			// TODO How do we handle fsnotify.Rename?

		case err := <-w.watcher.Errors:
			w.errs <- err
		}
	}
}

// isDir checks whether path is a directory.
func (w *MediaWatcher) isDir(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		w.errs <- err
		return false
	}

	stat, err := file.Stat()
	if err != nil {
		w.errs <- err
		return false
	}

	return stat.IsDir()
}
