package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

// FilterEngine is the detector engine
type FilterEngine struct {
	Filters []MitreATTCKFilterer
	AlertCh chan interface{}
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
	TermChan        chan os.Signal
}

func NewFilterEngine(alertCh chan interface{}) *FilterEngine {
	return &FilterEngine{
		Filters: make([]MitreATTCKFilterer, 0),
		AlertCh: alertCh,
	}
}

// NewEngine returns a new NewEngine with configFilePath as the configuration file
func NewEngine(configFilePath string) (*Engine, error) {
	engine := new(Engine)
	// parsing configuration
	if err := engine.Config.InitFrom(configFilePath); err != nil {
		return nil, err
	}
	// database
	if err := InitPg("pgx", engine.Config.PgConUrl); err != nil {
		return nil, err
	}
	if err := InitRedis(engine.Config.RedisConUrl); err != nil {
		return nil, err
	}
	// pipeline
	AlertCh := make(chan interface{}, AlertChBufSize)
	engine.HostManager = NewHostManager(AlertCh)
	engine.FilterEngine = NewFilterEngine(AlertCh)
	engine.ExtractorEngine = NewExtractorEngine()
	// kafka
	var lastOffset int64
	if engine.Config.SaveOnExit { // not parsing from the start
		lastOffset = PgConn.GetPreKafkaOffset()
	}
	engine.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{engine.Config.KafkaBrokers},
		Topic:    engine.Config.KafkaTopic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	if err := engine.Reader.SetOffset(lastOffset); err != nil {
		return nil, err // todo: should reset to 0
	}
	// global transformation
	engine.ExtractorEngine.InitDefault()
	// register Filters
	if err := engine.FilterEngine.Register(NewRuleFilter()); err != nil {
		return nil, err
	}
	if err := engine.FilterEngine.Register(NewIOCFilter()); err != nil {
		return nil, err
	}
	// signal handling
	engine.TermChan = make(chan os.Signal, 64)
	signal.Notify(engine.TermChan, os.Interrupt, syscall.SIGTERM)

	return engine, nil
}

// Start starts receiving messages and distribute to workers
func (engine *Engine) Start() error {
	if !engine.Config.SaveOnExit {
		if err := PgConn.DeleteAll(); err != nil { // clean all previous db
			return err
		}
	} else {
		if err := engine.HostManager.LoadData(); err != nil {
			return err
		}
	}

	go engine.HostManager.Start()
	go engine.FilterEngine.Start()
	engine.StartWebApp()

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
	preOffset := engine.Reader.Offset()
	lastOffset := preOffset
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
	if engine.Config.SaveOnExit && lastOffset > preOffset {
		if err := PgConn.SaveKafkaOffset(lastOffset + 1); err != nil {
			log.Warnf("cannot save kafka offset, %s", err)
		}
	}
	close(engine.HostManager.EventCh)
	engine.FilterEngine.CloseAll()

	// wait until exit
	<-engine.HostManager.State
	return nil
}

func (engine *Engine) StartWebApp() {
	gin.SetMode(gin.ReleaseMode)
	endpoint := engine.Config.ServerHost + ":" + engine.Config.ServerPort
	log.Infoln("Starting the server at", endpoint)
	router := gin.Default()
	// index.html
	router.StaticFile("", "./client/build/index.html")
	// static middleware
	router.Use(static.Serve("/", static.LocalFile("./client/build", false)))

	apiGroup := router.Group("api")
	apiGroup.GET("host", engine.HostManager.AllHostHandler)
	apiGroup.GET("ioc", engine.HostManager.AllIOCHandler)

	go func() {
		if err := router.Run(endpoint); err != nil {
			log.Fatal(err)
		}
	}()
}

// Close cleans up any resources
func (engine *Engine) Close() {
	_ = engine.Reader.Close()
	_ = RedisConn.Close()
	PgConn.Close()
}
