//package monitors
package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
)

type Monitor struct {
	filename       string
	file           *os.File
	watcher        *fsnotify.Watcher
	readBufferSize uint32 // size in bytes of chunk to read from watched file
	position       int64  // really should be unsigned for correctness, os api
}

// Constructs a new monitor
func NewMonitor(pathToFile string) (*Monitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	} else {
		file, err := os.Open(pathToFile)
		if err != nil {
			return nil, err
		} else {
			m := &Monitor{
				filename:       pathToFile,
				file:           file,
				watcher:        watcher,
				readBufferSize: 64 * 1024,
				position:       0,
			}
			err := m.seekToEnd()
			return m, err
		}
	}
}

func (m *Monitor) seekToEnd() error {
	fileInfo, err := m.file.Stat()
	if err == nil {
		endPosition := fileInfo.Size()
		log.Println("seeking to", endPosition)
		m.file.Seek(int64(endPosition), 0)
		m.position = endPosition
		return nil
	} else {
		return err
	}
}

func (m *Monitor) handle() {
	buffer := make([]byte, 64*1024)
	for {
		select {
		case ev := <-m.watcher.Events:
			switch ev.Op {
			case fsnotify.Write:
				for {
					n, err := m.file.Read(buffer)
					if n > 0 {
						fmt.Print(string(buffer[:n]))
					}
					if err == io.EOF {
						break
					}
				}
			}
		case err := <-m.watcher.Errors:
			log.Fatal("error", err)
			break
		}
	}
}

func (m *Monitor) StartWatching() {
	go m.handle()
	err := m.watcher.Add(m.filename)
	if err != nil {
		log.Fatal("error watching", err)
	}
}

func main() {
	monitor, err := NewMonitor("/var/log/crap.out")
	if err != nil {
		log.Fatal(err)
	} else {
		monitor.StartWatching()

		done := make(chan bool)
		<-done
	}
}
