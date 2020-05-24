package main

const (
	ConfigFilePath = "config.yml"
)

var (
	// default configuration
	DefKafkaBrokers = []string{
		"localhost:9092",
	}
	DefKafkaTopic = "winsysmon"
)

// global configuration
type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
}
