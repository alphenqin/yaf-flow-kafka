package main

import (
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatchDir 监听目录，文件写完后回调处理。
func WatchDir(dir string, onFileReady func(file string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(dir); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// 在跨平台场景下使用 Write/Create 事件，再通过大小稳定性确认写完。
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					if isFileComplete(event.Name) {
						onFileReady(event.Name)
					}
				}
			case err := <-watcher.Errors:
				log.Println("fsnotify error:", err)
			}
		}
	}()

	return nil
}

// 通过多次检测文件大小稳定性，判断是否写完整。
func isFileComplete(path string) bool {
	var prev int64 = -1
	for i := 0; i < 3; i++ {
		info, err := os.Stat(path)
		if err != nil {
			return false
		}
		size := info.Size()
		if size == prev {
			return true
		}
		prev = size
		time.Sleep(500 * time.Millisecond)
	}
	return false
}
