package main

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

// the app engine
type Engine struct {
	Config          Config
	Reader          *kafka.Reader
	HostManager     *HostManager
	FilterEngine    *FilterEngine
	ExtractorEngine *ExtractorEngine
}

// NewEngine returns new instance of Engine initialized
func NewEngine(configFilePath string, numOfWorkers int) (*Engine, error) {
	if len(configFilePath) == 0 && numOfWorkers <= 0 {
		return nil, errors.New("invalid parameters")
	}

	engine := &Engine{
		HostManager:     NewHostManager(),
		FilterEngine:    NewFilterEninge(),
		ExtractorEngine: NewExtractorEngine(),
	}

	if err := engine.Config.init(configFilePath); err != nil {
		return nil, err
	}
	engine.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{engine.Config.KafkaBrokers},
		Topic:       engine.Config.KafkaTopic,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: 0,
	})

	if err := engine.FilterEngine.Init(engine.Config.RuleDirPath); err != nil {
		return nil, err
	}
	engine.ExtractorEngine.InitDefault()

	return engine, nil
}

// Start starts receiving messages and distribute to workers
func (engine *Engine) Start() error {
	go engine.HostManager.Start()

	var msg = new(Message)
	for {
		rawMsg, err := engine.Reader.ReadMessage(context.Background())
		if err != nil {
			return err
		}
		if err := json.Unmarshal(rawMsg.Value, &msg); err != nil {
			log.Warn(err)
		}
		event := &msg.Winlog
		engine.HostManager.EventCh <- event
		msg = new(Message)
	}
}

// Close cleans up any resources
func (engine *Engine) Close() {
	if engine.Reader != nil {
		_ = engine.Reader.Close()
	}
}
