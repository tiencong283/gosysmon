package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
)

const (
	MsgChBufSize   = 100000
	AlertChBufSize = 1000
)

type ResultId struct {
	HostId      string
	ProcessGuid string
}

func NewResultId(msg *Message) ResultId {
	return ResultId{
		HostId:      msg.Agent.ID,
		ProcessGuid: msg.Event.getProcessGUID(),
	}
}

// MitreATTCKResult represents an alert/feature
type MitreATTCKResult struct {
	ResultId
	Timestamp time.Time
	IsAlert   bool `json:"-"`
	Context   map[string]interface{}
	Message   string
	Technique *MitreTechnique
}

func NewMitreATTCKResult(isAlert bool, techID, message string, msg *Message, mergeEventContext bool) *MitreATTCKResult {
	alert := &MitreATTCKResult{
		Timestamp: msg.Event.getTimestamp(),
		Context:   make(map[string]interface{}),
		Message:   message,
		Technique: MitreTechniques[techID],
		ResultId:  NewResultId(msg),
		IsAlert:   isAlert,
	}
	if mergeEventContext {
		alert.MergeContext(msg.Event.EventData)
		alert.AddContext("EventID", msg.Event.EventID)
		alert.AddContext("RecordID", msg.Event.RecordID)
	}
	delete(alert.Context, "RuleName") // RuleName only for rule-based filter
	return alert
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
	IsSupported(msg *Message) bool
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
		messageCh: make(chan *Message, MsgChBufSize),
		logger:    log.WithField("FilterId", name),
	}
}

// working with ATT&CK https://github.com/mitre/cti
type RawAttackPattern struct {
	Type                 string `json:"type"`
	XMitreIsSubtechnique bool   `json:"x_mitre_is_subtechnique"`
	XMitreVersion        string `json:"x_mitre_version"`
	Name                 string `json:"name"`
	KillChainPhases      []struct {
		KillChainName string `json:"kill_chain_name"`
		PhaseName     string `json:"phase_name"`
	} `json:"kill_chain_phases"`
	ExternalReferences []struct {
		ExternalID string `json:"external_id,omitempty"`
		SourceName string `json:"source_name"`
		URL        string `json:"url"`
	} `json:"external_references"`
}

type MitreTechnique struct {
	Id, Url, Name  string
	Tactics        []string
	IsSubTechnique bool
}

var MitreTechniques = make(map[string]*MitreTechnique)

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
		if object.Type != "attack-pattern" || object.XMitreVersion == "" {
			continue
		}
		id := object.ExternalReferences[0].ExternalID
		MitreTechniques[id] = &MitreTechnique{
			Id:             id,
			Url:            object.ExternalReferences[0].URL,
			Name:           object.Name,
			Tactics:        make([]string, 0),
			IsSubTechnique: object.XMitreIsSubtechnique,
		}
		for _, kc := range object.KillChainPhases {
			if kc.KillChainName == "mitre-attack" {
				MitreTechniques[id].Tactics = append(MitreTechniques[id].Tactics, kc.PhaseName)
			}
		}
	}
}
