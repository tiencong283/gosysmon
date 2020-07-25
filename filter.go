package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
)

const (
	EventChBufSize = 100000
	AlertChBufSize = 1000
)

type ResultId struct {
	HostId      string
	ProcessGuid string
}

func NewResultId(msg *Message) ResultId {
	return ResultId{
		HostId:      msg.Agent.ID,
		ProcessGuid: msg.Event.get("ProcessGuid"),
	}
}

// MitreATTCKResult represents an alert/feature
type MitreATTCKResult struct {
	ResultId
	Timestamp *time.Time
	IsAlert   bool `json:"-"`
	Context   map[string]interface{}
	Message   string
	Technique *AttackPattern
}

func NewMitreATTCKResult(isAlert bool, techID, message string, msg *Message) *MitreATTCKResult {
	return &MitreATTCKResult{
		Timestamp: msg.Event.timestamp(),
		Context:   make(map[string]interface{}),
		Message:   message,
		Technique: Techniques[techID],
		ResultId:  NewResultId(msg),
		IsAlert:   isAlert,
	}
}

func (r *MitreATTCKResult) MergeContext(m map[string]string) {
	for k, v := range m {
		r.Context[k] = v
	}
}

func (r *MitreATTCKResult) AddContext(key string, val interface{}) {
	r.Context[key] = val
}

// ModelFilter is the filter that builds models of detector for abnormal detection
type MitreATTCKFilterer interface {
	IsSupported(event *Message) bool
	Init() error
	MessageCh() chan *Message
	StateCh() chan int
	SetAlertCh(alertCh chan interface{})
	Start()
}

// CommonFilterer is the common properties of MitreATTCKFilterers
type CommonFilterer struct {
	State     chan int
	Name      string
	messageCh chan *Message
	AlertCh   chan interface{}
	logger    *log.Entry
}

func NewCommonFilterer(name string) CommonFilterer {
	return CommonFilterer{
		State:     make(chan int),
		Name:      name,
		messageCh: make(chan *Message, EventChBufSize),
		logger:    log.WithField("FilterId", name),
	}
}

// working with ATT&CK https://github.com/mitre/cti
type RawAttackPattern struct {
	Type            string `json:"type"`
	Name            string `json:"name"`
	KillChainPhases []struct {
		KillChainName string `json:"kill_chain_name"`
		PhaseName     string `json:"phase_name"`
	} `json:"kill_chain_phases"`
	ExternalReferences []struct {
		ExternalID string `json:"external_id,omitempty"`
		SourceName string `json:"source_name"`
		URL        string `json:"url"`
	} `json:"external_references"`
}

type AttackPattern struct {
	Id, Url, Name string
	Tactics       []string
}

var Techniques = make(map[string]*AttackPattern)

func init() {
	// initialize
	file, err := os.Open("resources/enterprise-attack.json")
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	attckContent := struct {
		Objects []RawAttackPattern `json:"objects"`
	}{}
	if err := json.Unmarshal(bytes, &attckContent); err != nil {
		log.Fatal(err)
	}
	for _, object := range attckContent.Objects {
		if object.Type != "attack-pattern" {
			continue
		}
		id := object.ExternalReferences[0].ExternalID
		Techniques[id] = &AttackPattern{
			Id:      id,
			Url:     object.ExternalReferences[0].URL,
			Name:    object.Name,
			Tactics: make([]string, 0),
		}
		for _, kc := range object.KillChainPhases {
			if kc.KillChainName == "mitre-attack" {
				Techniques[id].Tactics = append(Techniques[id].Tactics, kc.PhaseName)
			}
		}
	}
}
