package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
}

func NewWatcher(file string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(file)
	if err != nil {
		return nil, err
	}

	return &Watcher{watcher}, nil
}

func (w *Watcher) Watch() <-chan struct{} {
	fileChanged := make(chan struct{})
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					close(fileChanged)
					return
				}
				if event.Has(fsnotify.Write) {
					fileChanged <- struct{}{}
				}
			case err, ok := <-w.watcher.Errors:
				close(fileChanged)
				if !ok {
					return
				}
				log.Println("watcher error:", err)
			}
		}
	}()
	return fileChanged
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}
