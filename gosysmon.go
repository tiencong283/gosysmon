package main

import (
	"context"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/segmentio/kafka-go"
	_ "github.com/segmentio/kafka-go/gzip"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	// setup the logger
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: time.RFC3339,
	})
}

func mainx() {
	// reading configuration
	viper.SetConfigName(ConfigFilePath)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("kafka.brokers", DefKafkaBrokers)
	viper.SetDefault("kafka.topic", DefKafkaTopic)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warnf("The config file '%s' not found, use the default configuration\n", ConfigFilePath)
		} else {
			log.Fatal(err)
		}
	}
	config := new(Config)
	config.KafkaBrokers = viper.GetStringSlice("kafka.brokers")
	config.KafkaTopic = viper.GetString("kafka.topic")

	// reading events
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.KafkaBrokers,
		Topic:       config.KafkaTopic,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: 0,
	})
	defer r.Close()

	hostManager := NewHostManager()
	var msg Message
	for {
		rawMsg, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		if err := json.Unmarshal(rawMsg.Value, &msg); err != nil {
			log.Warn(err)
		}
		// process events
		hostManager.OnEvent(&msg.Winlog)
	}
}
func main(){
	eventFilter, err := NewEventFilterFrom("rules/T1060_registry-run-keys-startup-folder.xml")
	if err!=nil {
		log.Fatal(err)
	}
	eventFilter.Dump()
}
