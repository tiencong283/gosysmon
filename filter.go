package main

import log "github.com/sirupsen/logrus"

const (
	// IOC types
	IOCHash = iota
	IOCIp
	IOCHostname

	// Result types
	RNormal
	RAbnormal
)

// result context
type RContext struct {
	Context map[string]string
	Result  int
}

// IOCFilter is the filter that deals with IOCs
type IOCFilterer interface {
	// Search searches in database for an IOC and returns its context
	Search(indicator string, iocType int) *RContext
}

// RuleFilter is the filter using rules for abnormal detection
type RuleFilterer interface {
	// GetTechName returns the label of matched rule
	GetTechName(event *SysmonEvent) string
}

// ModelFilter is the filter that builds models of detector for abnormal detection
type ModelFilterer interface {
	// IsSupported return true if if support event type
	IsSupported(event *SysmonEvent) bool
	// Init initializes the model
	Init() error
	// TrainAndPredict do updates the profile and report any detection
	TrainAndPredict(event *SysmonEvent) *RContext
}

// FilterEngine is the detector engine
type FilterEngine struct {
	IOCFilter   IOCFilterer
	RuleFilter  RuleFilterer
	ModelFilter ModelFilterer
}

func NewFilterEninge() *FilterEngine {
	return new(FilterEngine)
}

func (fe *FilterEngine) Init(ruleDirPath string) {
	mitreATTCKFilter := NewEventFilter()
	if err := mitreATTCKFilter.UpdateFromDir(ruleDirPath); err != nil {
		fe.RuleFilter = nil
		log.Warn(err)
	} else {
		fe.RuleFilter = mitreATTCKFilter
	}
}
