package main

import (
	"log"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]

	var files []string
	for _, arg := range args {
		// reserve the right to provide option flags in the format of --{KEY}={VALUE}
		if !strings.HasPrefix(arg, "--") {
			files = append(files, arg)
		}
	}

	log.Println("files:", files)

	drain := NewDrain()

	for _, file := range files {
		monitor, err := NewMonitor(file)
		if err != nil {
			log.Println("[ERROR] could not open for reading:", file, err)
		} else {
			drain.Pipe(monitor)
			monitor.StartWatching()
		}
	}

	drain.Consume()
}
