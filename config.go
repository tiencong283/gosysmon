package main

import (
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

const (
	ConfigFilePath = "config.yml"
	MsgBufSize     = 100000

	// default configuration
	DefKafkaBrokers = "localhost:9092"
	DefKafkaTopic   = "winsysmon"
	DefRuleDirPath  = "rules"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// global configuration
type Config struct {
	KafkaBrokers string
	KafkaTopic   string
	RuleDirPath  string
}

func (config *Config) init(configFilePath string) error {
	// reading configuration
	viper.SetConfigName(configFilePath)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("kafka.brokers", DefKafkaBrokers)
	viper.SetDefault("kafka.topic", DefKafkaTopic)
	viper.SetDefault("rules.dirpath", DefRuleDirPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warnf("the config file '%s' not found, use the default configuration\n", ConfigFilePath)
		} else {
			return err
		}
	}
	config.KafkaBrokers = viper.GetString("kafka.brokers")
	config.KafkaTopic = viper.GetString("kafka.topic")
	config.RuleDirPath = viper.GetString("rules.dirpath")

	return nil
}

func init() {
	// setup the logger
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
}