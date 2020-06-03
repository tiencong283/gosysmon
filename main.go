package main

import (
	_ "github.com/segmentio/kafka-go/gzip"
	log "github.com/sirupsen/logrus"
	"runtime"
)

// the entry point
func main() {
	engine, err := NewEngine(ConfigFilePath, runtime.GOMAXPROCS(0))
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()
	if err := engine.Start(); err != nil {
		log.Fatal(err)
	}
}
