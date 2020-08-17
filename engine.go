package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	LServer = iota
	LClient
)

type ActivityLog struct {
	Timestamp time.Time
	Type      int
	Message   string
}

func (actLog *ActivityLog) Save() error {
	jsonLog, err := json.MarshalToString(actLog)
	if err != nil {
		return err
	}
	if _, err = RedisConn.Do("LPUSH", "activity-logs", jsonLog); err != nil {
		return err
	}
	return nil
}

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
func (fe *FilterEngine) Broadcast(msg *Message) {
	for _, filter := range fe.Filters {
		if filter.IsSupported(msg) {
			filter.MessageCh() <- msg
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
		close(filter.MessageCh())
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
	ExtractorEngine *PreprocessorEngine
	TermChan        chan os.Signal
	EventRateHooker *EventRateHooker
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
	engine.ExtractorEngine = NewPreprocessorEngine()

	// get kafka offset
	var lastOffset int64
	if engine.Config.KafkaParseFrom == "last" {
		val, err := redis.Int64(RedisConn.Do("GET", "lastKafkaOffset"))
		if err != nil && err != redis.ErrNil {
			return nil, err
		}
		lastOffset = val
	}
	log.Infof("parsing events from offset %d\n", lastOffset)

	engine.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{engine.Config.KafkaBrokers},
		Topic:    engine.Config.KafkaTopic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	if err := engine.Reader.SetOffset(lastOffset); err != nil {
		return nil, fmt.Errorf("cannot set topic offset to %d, %s", lastOffset, err)
	}
	// global transformation
	engine.ExtractorEngine.Register(NewRegistryProcessor())

	// register Filters
	if err := engine.FilterEngine.Register(NewRuleFilter()); err != nil {
		return nil, err
	}
	if err := engine.FilterEngine.Register(NewIOCFilter()); err != nil {
		return nil, err
	}
	if err := engine.FilterEngine.Register(NewTimestompFilter()); err != nil {
		return nil, err
	}
	// signal handling
	engine.TermChan = make(chan os.Signal, 64)
	signal.Notify(engine.TermChan, os.Interrupt, syscall.SIGTERM)

	engine.EventRateHooker = NewEventRateHooker()
	return engine, nil
}

// SaveServerLogs saves server logs
func (engine *Engine) SaveServerLogs(action string) error {
	actLog := &ActivityLog{
		Timestamp: time.Now(),
		Type:      LServer,
	}
	switch action {
	case "started":
		actLog.Message = "Server started"
	case "stopped":
		actLog.Message = "Server stopped"
	}
	return actLog.Save()
}

// Start starts receiving messages and distribute to workers
func (engine *Engine) Start() error {
	if engine.Config.KafkaParseFrom == "start" {
		if err := PgConn.DeleteAll(); err != nil { // clean all previous db
			return fmt.Errorf("cannot delete process-related tables, %s", err)
		}
		if _, err := RedisConn.Do("DEL", "activity-logs"); err != nil {
			return fmt.Errorf("cannot clean db on redis, %s", err)
		}
	} else if err := engine.HostManager.LoadData(); err != nil {
		return err
	}

	go engine.HostManager.Start()
	go engine.FilterEngine.Start()
	engine.StartWebApp()
	if err := engine.SaveServerLogs("started"); err != nil {
		log.Warn("cannot log server activities, ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func(termChan chan os.Signal, cancel context.CancelFunc) {
		sig := <-termChan
		if sig == os.Interrupt {
			fmt.Println("")
		}
		log.Infof("got %v signal, please wait while shutting down the server\n", sig)
		cancel()
	}(engine.TermChan, cancel)

	go engine.EventRateHooker.Start()

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
		engine.EventRateHooker.MessageCh <- msg
		if err := engine.ExtractorEngine.Transform(msg); err != nil {
			log.Warn("cannot transform the event,", err)
		}
		engine.HostManager.MessageCh <- msg
		engine.FilterEngine.Broadcast(msg)
		lastOffset = rawMsg.Offset
		msg = new(Message)
	}
	if lastOffset > preOffset {
		if _, err := RedisConn.Do("SET", "lastKafkaOffset", lastOffset+1); err != nil {
			log.Warnf("cannot save last kafka offset, %s", err)
		}
	}

	close(engine.HostManager.MessageCh)
	engine.FilterEngine.CloseAll()
	// wait until exit
	<-engine.HostManager.State

	if err := engine.SaveServerLogs("stopped"); err != nil {
		log.Warn("cannot log server activities, ", err)
	}
	return nil
}

func (engine *Engine) StartWebApp() {
	gin.SetMode(gin.ReleaseMode)
	endpoint := engine.Config.ServerHost + ":" + engine.Config.ServerPort
	log.Infoln("Starting the server at", endpoint)
	router := gin.Default()
	// allows all origins for debugging
	router.Use(cors.Default())
	// index.html
	router.StaticFile("", "./client/build/index.html")
	// static middleware
	router.Use(static.Serve("/", static.LocalFile("./client/build", false)))

	apiGroup := router.Group("api")
	apiGroup.GET("host", engine.HostManager.AllHostHandler)
	apiGroup.GET("ioc", engine.HostManager.AllIOCHandler)
	apiGroup.GET("alert", engine.HostManager.AllAlertHandler)
	apiGroup.POST("process", engine.HostManager.ProcessHandler)
	apiGroup.POST("process-tree", engine.HostManager.ProcessTreeHandler)
	apiGroup.GET("activity-log", engine.AllLogHandler)
	apiGroup.POST("process-activities", engine.HostManager.ProcessActivityHandler)
	apiGroup.GET("technique-stats", engine.TechniqueStatsHandler)

	// websockets
	router.Any("/ws/event-processing-rate", engine.EventRateHooker.ServeEventRateWs)
	go func() {
		if err := router.Run(endpoint); err != nil {
			log.Fatal(err)
		}
	}()
}

// Close cleans up any resources
func (engine *Engine) Close() {
	if err := engine.Reader.Close(); err != nil {
		log.Warn("cannot close Kafka Reader, ", err)
	}
	if err := RedisConn.Close(); err != nil {
		log.Warn("cannot close Redis Connection, ", err)
	}
	if err := PgConn.Close(); err != nil {
		log.Warn("cannot close Postgres Connection, ", err)
	}
}

// request handler for "/api/activity-log"
func (engine *Engine) AllLogHandler(c *gin.Context) {
	actLogs := make([]*ActivityLogView, 0)
	jsonActLogs, err := redis.Strings(RedisConn.Do("LRANGE", "activity-logs", 0, -1))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	for _, jsonActLog := range jsonActLogs {
		actLog := new(ActivityLog)
		if err := json.Unmarshal([]byte(jsonActLog), actLog); err != nil {
			log.Warn("cannot unmarshal jsonActLog, ", err)
		}
		actLogs = append(actLogs, NewActivityLogView(actLog))
	}
	c.JSON(http.StatusOK, actLogs)
}

func (engine *Engine) TechniqueStatsHandler(c *gin.Context) {
	stats, err := PgConn.GetTechniqueStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}
