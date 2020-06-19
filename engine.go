package main

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"strings"
)

// FilterEngine is the detector engine
type FilterEngine struct {
	Filters []MitreATTCKFilterer
	AlertCh chan *RContext
}

func (fe *FilterEngine) Register(newFilter MitreATTCKFilterer) error {
	if err := newFilter.Init(); err != nil {
		return err
	}
	newFilter.SetAlertCh(fe.AlertCh)
	fe.Filters = append(fe.Filters, newFilter)
	return nil
}

func (fe *FilterEngine) Broadcast(event *SysmonEvent) {
	for _, filter := range fe.Filters {
		if filter.IsSupported(event) {
			filter.EventCh() <- event
		}
	}
}

func (fe *FilterEngine) Start() {
	for _, filter := range fe.Filters {
		go filter.Start()
	}
}

// the app engine
type Engine struct {
	Config          Config
	Reader          *kafka.Reader
	HostManager     *HostManager
	FilterEngine    *FilterEngine
	ExtractorEngine *ExtractorEngine
}

func NewFilterEninge(alertCh chan *RContext) *FilterEngine {
	return &FilterEngine{
		Filters: make([]MitreATTCKFilterer, 0),
		AlertCh: alertCh,
	}
}

// NewEngine returns a new NewEngine with configFilePath as the configuration file
func NewEngine(configFilePath string) (*Engine, error) {
	configFilePath = strings.TrimSpace(configFilePath)
	if len(configFilePath) == 0 {
		return nil, errors.New("invalid parameters")
	}

	// pipeline
	AlertCh := make(chan *RContext, AlertChBufSize)
	engine := &Engine{
		HostManager:     NewHostManager(AlertCh),
		FilterEngine:    NewFilterEninge(AlertCh),
		ExtractorEngine: NewExtractorEngine(),
	}

	if err := engine.Config.InitFrom(configFilePath); err != nil {
		return nil, err
	}

	engine.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{engine.Config.KafkaBrokers},
		Topic:       engine.Config.KafkaTopic,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: 0,
	})
	engine.ExtractorEngine.InitDefault()
	if err := engine.FilterEngine.Register(NewRuleFilter()); err != nil {
		return nil, err
	}
	return engine, nil
}

// Start starts receiving messages and distribute to workers
func (engine *Engine) Start() error {
	go engine.HostManager.Start()
	go engine.FilterEngine.Start()

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
		engine.FilterEngine.Broadcast(event)
		msg = new(Message)
	}
}

// Close cleans up any resources
func (engine *Engine) Close() {
	if engine.Reader != nil {
		_ = engine.Reader.Close()
	}
}
