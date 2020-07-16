package main

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/segmentio/kafka-go/gzip"
	log "github.com/sirupsen/logrus"
)

// the entry point
func main() {
	engine, err := NewEngine(ConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if err := engine.Start(); err != nil {
		engine.Close()
		log.Fatal(err)
	}
	engine.Close()
}
