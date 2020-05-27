package main

const (
	ConfigFilePath = "config.yml"
)

const (
	// default configuration
	DefKafkaBrokers = "localhost:9092"
	DefKafkaTopic = "winsysmon"
)

// global configuration
type Config struct {
	KafkaBrokers string
	KafkaTopic   string
}