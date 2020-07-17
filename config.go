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
	DefPgHost      = "localost"
	DefPgPort      = 5432
	DefPgDatabase  = "gosysmon"
	DefPgUser      = "tiencong283"
	DefPgPassword  = "gosysmon"
	PgConUrlFormat = "host=%s port=%d database=%s user=%s password=%s" // DSN string

	// default Redis
	DefRHost          = "localhost"
	DefRPort          = 6379
	DefRUsername      = "default"
	DefRPassword      = "gosysmon"
	RedisConUrlFormat = "redis://%s:%s@%s:%d" // redis://user:secret@localhost:6379, https://www.iana.org/assignments/uri-schemes/prov/redis

	// default app behavior
	DefSaveOnExit = true

	// default endpoint address
	DefServerHost = "0.0.0.0"
	DefServerPort = "9090"
)

// Config is the application configuration
type Config struct {
	KafkaBrokers string
	KafkaTopic   string
	PgConUrl     string
	RedisConUrl  string
	SaveOnExit   bool
	ServerHost   string
	ServerPort   string
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
	viper.SetDefault("redis.host", DefPgHost)
	viper.SetDefault("redis.port", DefPgPort)
	viper.SetDefault("redis.user", DefPgUser)
	viper.SetDefault("redis.password", DefPgPassword)
	viper.SetDefault("server.host", DefServerHost)
	viper.SetDefault("server.port", DefServerPort)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warnf("the config file '%s' not found, use the default configuration\n", configFilePath)
		} else {
			return err
		}
	}
	config.KafkaBrokers = viper.GetString("kafka.brokers")
	config.KafkaTopic = viper.GetString("kafka.topic")

	config.PgConUrl = fmt.Sprintf(PgConUrlFormat, viper.GetString("postgresql.host"), viper.GetInt("postgresql.port"),
		viper.GetString("postgresql.database"), viper.GetString("postgresql.user"),
		viper.GetString("postgresql.password"))

	config.RedisConUrl = fmt.Sprintf(RedisConUrlFormat, viper.GetString("redis.user"), viper.GetString("redis.password"),
		viper.GetString("redis.host"), viper.GetInt("redis.port"))

	config.SaveOnExit = viper.GetBool("save-on-exit")
	config.ServerHost = viper.GetString("server.host")
	config.ServerPort = viper.GetString("server.port")
	return nil
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	// setup the logger
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
}
