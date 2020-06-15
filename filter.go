package main

import log "github.com/sirupsen/logrus"

const (
	EventChBufSize = 1000000
	AlertChBufSize = 1000
)

// RContext represents an alert
type RContext struct {
	Context map[string]string
	Message string
	TechID  string
}

// ModelFilter is the filter that builds models of detector for abnormal detection
type MitreATTCKFilterer interface {
	IsSupported(event *SysmonEvent) bool
	Init() error
	Start()
}

// CommonFilterer is the common properties of MitreATTCKFilterers
type CommonFilterer struct {
	Name    string
	EventCh chan *SysmonEvent
	AlertCh chan *RContext
	logger  *log.Entry
}

func NewCommonFilterer(name string) CommonFilterer {
	return CommonFilterer{
		Name:    name,
		EventCh: make(chan *SysmonEvent, EventChBufSize),
		AlertCh: make(chan *RContext, AlertChBufSize),
		logger:  log.WithField("FilterId", name),
	}
}
