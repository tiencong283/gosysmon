package main

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"os"
)

// FilterEngine is the detector engine
type FilterEngine struct {
	IOCFilter   IOCFilterer
	RuleFilter  RuleFilterer
	ModelFilter ModelFilterer
}

func (fe *FilterEngine) initDefault() {
	ruleFilter, err := NewEventFilterFrom("rules/T1060_registry-run-keys-startup-folder.xml")
	if err != nil {
		fe.RuleFilter = nil
		return
	}
	fe.RuleFilter = ruleFilter
}

// the app engine
type Engine struct {
	Config       Config
	Reader       *kafka.Reader
	NumOfWorkers int
	WorkerChans  []chan *SysmonEvent
	HostManager  *HostManager
	FilterEngine *FilterEngine
}

// NewEngine returns new instance of Engine initialized
func NewEngine(configFilePath string, numOfWorkers int) (*Engine, error) {
	if len(configFilePath) == 0 && numOfWorkers <= 0 {
		return nil, errors.New("invalid parameters")
	}

	engine := &Engine{
		NumOfWorkers: numOfWorkers,
		HostManager:  NewHostManager(),
		FilterEngine: new(FilterEngine),
	}
	engine.WorkerChans = make([]chan *SysmonEvent, engine.NumOfWorkers)

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

	engine.FilterEngine.initDefault()
	return engine, nil
}

// Start starts receiving messages and distribute to workers
func (engine *Engine) Start() error {
	for i := 0; i < engine.NumOfWorkers; i++ {
		engine.WorkerChans[i] = make(chan *SysmonEvent, MsgBufSize)
		go engine.WorkerHandler(i)
	}
	nextIdx := 0
	hostToChanIdx := make(map[string]int, 0)

	var msg Message
	for {
		rawMsg, err := engine.Reader.ReadMessage(context.Background())
		if err != nil {
			return err
		}
		if err := json.Unmarshal(rawMsg.Value, &msg); err != nil {
			log.Warn(err)
		}
		// distribute the message to worker
		chanIdx, ok := hostToChanIdx[msg.Winlog.ComputerName]
		if !ok {
			chanIdx = nextIdx
			hostToChanIdx[msg.Winlog.ComputerName] = nextIdx
			nextIdx = (nextIdx + 1) % engine.NumOfWorkers
		}
		// pass msg to chan will result in map concurrency error
		clone := &SysmonEvent{
			EventMetadata: EventMetadata{
				EventID:      msg.Winlog.EventID,
			},
			EventData: make(map[string]string),
		}
		for k, v := range msg.Winlog.EventData {
			clone.EventData[k] = v
		}
		engine.WorkerChans[chanIdx] <- clone
	}
}

// WorkerHandler is the thread for processing incoming events
func (engine *Engine) WorkerHandler(chanIdx int) {
	for event := range engine.WorkerChans[chanIdx] {
		label := engine.FilterEngine.RuleFilter.GetTechName(event)
		if len(label) > 0 {
			log.Println(label)
			os.Exit(0)
		}
	}
}

// Close cleans up any resources
func (engine *Engine) Close() {
	if engine.Reader != nil {
		engine.Reader.Close()
	}
}