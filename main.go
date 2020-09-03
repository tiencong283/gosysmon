package main

import (
	"flag"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/segmentio/kafka-go/gzip"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	applicationName = "Gosysmon"
	version         = "v1.0.0"
	helpMsg         = `Usage of %s:
  -c configuration, -config configuration
    	application configuration (default "%s")
  -v	see application version
  -h	see help
`
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), helpMsg, applicationName, ConfigFilePath)
	}
}

// the entry point
func main() {
	configFilePath := ""
	shouldPrintVersion := false

	// parsing arguments
	flag.StringVar(&configFilePath, "config", ConfigFilePath, "application `configuration`")
	flag.StringVar(&configFilePath, "c", ConfigFilePath, "application `configuration`")
	flag.BoolVar(&shouldPrintVersion, "v", false, "see application version")
	flag.Parse()

	if shouldPrintVersion {
		fmt.Printf("Gosysmon %s\n", version)
		os.Exit(0)
	}

	engine, err := NewEngine(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if err := engine.Start(); err != nil {
		engine.Close()
		log.Fatal(err)
	}
	engine.Close()
}
