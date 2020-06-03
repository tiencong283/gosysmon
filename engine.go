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
	NumOfWorkers    int
	WorkerChans     []chan *SysmonEvent
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
		NumOfWorkers:    numOfWorkers,
		HostManager:     NewHostManager(),
		FilterEngine:    NewFilterEninge(),
		ExtractorEngine: NewExtractorEngine(),
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

	engine.FilterEngine.Init(engine.Config.RuleDirPath)
	engine.ExtractorEngine.InitDefault()

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
		// pass msg directly to chan will result in map concurrency error
		clone := &SysmonEvent{
			EventMetadata: EventMetadata{
				EventID: msg.Winlog.EventID,
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
		_ = engine.ExtractorEngine.Transform(event)
		label := engine.FilterEngine.RuleFilter.GetTechName(event)
		if len(label) > 0 {
			log.Println(ToJson(event))
			log.Println(label)
		}
	}
}

// Close cleans up any resources
func (engine *Engine) Close() {
	if engine.Reader != nil {
		_ = engine.Reader.Close()
	}
}
