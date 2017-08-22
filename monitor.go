package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
	"strings"
)

// A monitor watches a file for changes. it will send new lines over the lineChannel
type Monitor struct {
	filename       string
	file           *os.File
	watcher        *fsnotify.Watcher
	readBufferSize uint32 // size in bytes of chunk to read from watched file
	lineChannel    chan []string
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
				lineChannel:    make(chan []string, 1024),
			}
			err := m.seekToEnd()
			return m, err
		}
	}
}

// seeks the position of the monitored file to the end
func (m *Monitor) seekToEnd() error {
	fileInfo, err := m.file.Stat()
	if err == nil {
		endPosition := fileInfo.Size()
		log.Println("seeking", m.filename, "to", endPosition)
		m.file.Seek(int64(endPosition), 0)
		return nil
	} else {
		return err
	}
}

func (m *Monitor) handle() {
	readBuffer := make([]byte, 64*1024)
	var buffer []byte
	for {
		select {
		case ev := <-m.watcher.Events:
			switch ev.Op {
			case fsnotify.Write:
				for {
					// probably can read directly into buffer?
					n, err := m.file.Read(readBuffer)
					if n > 0 {
						buffer = append(buffer, readBuffer[:n]...)
						stringRead := string(buffer)
						lines := strings.Split(stringRead, "\n")
						numLines := len(lines)
						// assume that any line not ending in \n isn't a fulli
						fullLines := lines[:numLines-1]
						m.lineChannel <- fullLines
						partialLine := lines[numLines-1]
						buffer = []byte(partialLine)
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

// Start watching for changes. Should only be called once.
// TODO: probably should raise (or return) a runtime error if it is already watching
func (m *Monitor) StartWatching() {
	go m.handle()
	err := m.watcher.Add(m.filename)
	if err != nil {
		log.Fatal("error watching", err)
	}
}

// Drain takes input from the monitors and merges it down to one unified channel
type Drain struct {
	monitors map[string]*Monitor // map of path to monitor
	merged   chan string
}

func NewDrain(monitors ...*Monitor) *Drain {
	drain := &Drain{
		monitors: make(map[string]*Monitor),
		merged:   make(chan string, 10000),
	}
	// start piping monitors in, if any
	for _, mon := range monitors {
		drain.Pipe(mon)
	}
	return drain
}

// consumes off of the unified merge channel. Currently just dumps to stdout
// TODO: other output options
// Blocks the calling thread
func (drain *Drain) Consume() {
	for line := range drain.merged {
		// TODO:
		fmt.Println(line)
	}
}

// pipes output of a monitor into the drain
func (drain *Drain) Pipe(monitor *Monitor) {
	drain.monitors[monitor.filename] = monitor
	go func() {
		for lines := range monitor.lineChannel {
			for _, line := range lines {
				drain.merged <- line
			}
		}
	}()
}
