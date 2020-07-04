package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

const (
	ConfigFilePath = "config.yml"

	// default Kafka connection
	DefKafkaBrokers = "localhost:9092"
	DefKafkaTopic   = "winsysmon"

	// default Postgresql connection
	DefPgHost       = "localost"
	DefPgPort       = 5432
	DefPgDatabase   = "gosysmon"
	DefPgUser       = "tiencong283"
	DefPgPassword   = "@C0n9ht4"
	PgConnUrlFormat = "host=%s port=%d database=%s user=%s password=%s" // DSN string

	// default app behavior
	DefSaveOnExit = true
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Config is the application configuration
type Config struct {
	KafkaBrokers string
	KafkaTopic   string
	PgConUrl     string
	SaveOnExit   bool
}

// InitFrom reads configuration from file configFilePath
func (config *Config) InitFrom(configFilePath string) error {
	// reading configuration
	viper.SetConfigName(configFilePath)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("kafka.brokers", DefKafkaBrokers)
	viper.SetDefault("kafka.topic", DefKafkaTopic)
	viper.SetDefault("postgresql.host", DefPgHost)
	viper.SetDefault("postgresql.port", DefPgPort)
	viper.SetDefault("postgresql.database", DefPgDatabase)
	viper.SetDefault("postgresql.user", DefPgUser)
	viper.SetDefault("postgresql.password", DefPgPassword)
	viper.SetDefault("save-on-exit", DefSaveOnExit)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warnf("the config file '%s' not found, use the default configuration\n", configFilePath)
		} else {
			return err
		}
	}
	config.KafkaBrokers = viper.GetString("kafka.brokers")
	config.KafkaTopic = viper.GetString("kafka.topic")
	config.SaveOnExit = viper.GetBool("save-on-exit")

	host := viper.GetString("postgresql.host")
	port := viper.GetInt("postgresql.port")
	database := viper.GetString("postgresql.database")
	user := viper.GetString("postgresql.user")
	password := viper.GetString("postgresql.password")
	config.PgConUrl = fmt.Sprintf(PgConnUrlFormat, host, port, database, user, password)

	return nil
}

func init() {
	// setup the logger
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
}
