package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// FilterEngine is the detector engine
type FilterEngine struct {
	Filters []MitreATTCKFilterer
	AlertCh chan *RContext
}

// Register add a new Filter
func (fe *FilterEngine) Register(newFilter MitreATTCKFilterer) error {
	if err := newFilter.Init(); err != nil {
		return err
	}
	newFilter.SetAlertCh(fe.AlertCh)
	fe.Filters = append(fe.Filters, newFilter)
	return nil
}

// Broadcast send an event to all Filters for processing
func (fe *FilterEngine) Broadcast(event *SysmonEvent) {
	for _, filter := range fe.Filters {
		if filter.IsSupported(event) {
			filter.EventCh() <- event
		}
	}
}

// Start starts all ATT&CK MITRE Filters
func (fe *FilterEngine) Start() {
	for _, filter := range fe.Filters {
		go filter.Start()
	}
}

// CloseAll closes all Filters
func (fe *FilterEngine) CloseAll() {
	for _, filter := range fe.Filters {
		close(filter.EventCh())
		<-filter.StateCh()
	}
	close(fe.AlertCh)
}

// Engine is the application engine
type Engine struct {
	Config          Config
	Reader          *kafka.Reader
	HostManager     *HostManager
	FilterEngine    *FilterEngine
	ExtractorEngine *ExtractorEngine
	DBConn          *DBConn
	TermChan        chan os.Signal
}

func NewFilterEngine(alertCh chan *RContext) *FilterEngine {
	return &FilterEngine{
		Filters: make([]MitreATTCKFilterer, 0),
		AlertCh: alertCh,
	}
}

// NewEngine returns a new NewEngine with configFilePath as the configuration file
func NewEngine(configFilePath string) (*Engine, error) {
	engine := new(Engine)

	configFilePath = strings.TrimSpace(configFilePath)
	if len(configFilePath) == 0 {
		return nil, errors.New("invalid parameters")
	}
	if err := engine.Config.InitFrom(configFilePath); err != nil {
		return nil, err
	}
	dbConn, err := NewDBConn("pgx", engine.Config.PgConUrl)
	if err != nil {
		return nil, err
	}
	engine.DBConn = dbConn

	// pipeline
	AlertCh := make(chan *RContext, AlertChBufSize)
	engine.HostManager = NewHostManager(AlertCh, dbConn)
	engine.FilterEngine = NewFilterEngine(AlertCh)
	engine.ExtractorEngine = NewExtractorEngine()

	var lastOffset int64
	if engine.Config.SaveOnExit { // not parsing from the start
		lastOffset = engine.DBConn.GetPreKafkaOffset()
	}
	log.Infoln("offset ", lastOffset)
	engine.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{engine.Config.KafkaBrokers},
		Topic:    engine.Config.KafkaTopic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	if err := engine.Reader.SetOffset(lastOffset); err != nil {
		return nil, err // todo: should reset to 0
	}

	engine.ExtractorEngine.InitDefault()
	if err := engine.FilterEngine.Register(NewRuleFilter()); err != nil {
		return nil, err
	}
	engine.TermChan = make(chan os.Signal, 64)
	signal.Notify(engine.TermChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	return engine, nil
}

// Start starts receiving messages and distribute to workers
func (engine *Engine) Start() error {
	if !engine.Config.SaveOnExit { // parsing from start
		if err := engine.DBConn.DeleteAll(); err != nil { // clean all previous db
			return err
		}
	}

	go engine.HostManager.Start()
	go engine.FilterEngine.Start()

	ctx, cancel := context.WithCancel(context.Background())
	go func(termChan chan os.Signal, cancel context.CancelFunc) {
		sig := <-termChan
		if sig == os.Interrupt {
			fmt.Println("")
		}
		log.Infof("got %v signal, please wait while shutting down the server\n", sig)
		cancel()
	}(engine.TermChan, cancel)

	var msg = new(Message)
	lastOffset := int64(0)
	for {
		rawMsg, err := engine.Reader.ReadMessage(ctx)
		if err == context.Canceled {
			break
		}
		if err != nil {
			return err
		}
		if err := json.Unmarshal(rawMsg.Value, &msg); err != nil {
			log.Warn(err)
		}
		event := &msg.Winlog
		engine.HostManager.EventCh <- event
		engine.FilterEngine.Broadcast(event)
		lastOffset = rawMsg.Offset
		msg = new(Message)
	}
	if engine.Config.SaveOnExit {
		_ = engine.DBConn.SaveKafkaOffset(lastOffset)
	}
	close(engine.HostManager.EventCh)
	engine.FilterEngine.CloseAll()

	// wait until exit
	<-engine.HostManager.State
	return nil
}

// Close cleans up any resources
func (engine *Engine) Close() {
	_ = engine.Reader.Close()
	engine.DBConn.Close()
}
