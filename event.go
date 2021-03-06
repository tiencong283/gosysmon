package main

import (
	"strconv"
	"time"
)

const (
	// Event timestamps are in UTC standard time
	TimeFormat = "2006-01-02 15:04:05.999999999"
)

// E stands for event
const (
	EProcessCreate        = 1
	EFileCreateTime       = 2
	ENetworkConnect       = 3
	EServiceStateChange   = 4
	EProcessTerminate     = 5
	EDriverLoad           = 6
	EImageLoad            = 7
	ECreateRemoteThread   = 8
	ERawAccessRead        = 9
	EProcessAccess        = 10
	EFileCreate           = 11
	ERegistryEventAdd     = 12
	ERegistryEventSet     = 13
	ERegistryEventRename  = 14
	EFileCreateStreamHash = 15
	EConfigStateChange    = 16
	EPipeEventCreate      = 17
	EPipeEventConnect     = 18
	EWmiEventFilter       = 19
	EWmiEventConsumer     = 20
	EWmiEventBinding      = 21
	EDnsQuery             = 22
	EFileDelete           = 23
	ESysmonError          = 255
)

/*
	These event structures are quickly derived from json using https://github.com/ChimeraCoder/gojson
	Only choose used, useful features so it's not complete
*/

// sysmon event representation
type SysmonEvent struct {
	// event metadata
	ProviderName string `json:"provider_name"`
	ProviderGUID string `json:"provider_guid"`
	Channel      string `json:"channel"`
	ComputerName string `json:"computer_name"`
	RecordID     int    `json:"record_id"`
	EventID      int    `json:"event_id"`
	Opcode       string `json:"opcode"`
	Process      struct {
		Pid    int `json:"pid"`
		Thread struct {
			ID int `json:"id"`
		} `json:"thread"`
	} `json:"process"`
	User struct {
		Domain     string `json:"domain"`
		Identifier string `json:"identifier"`
		Name       string `json:"name"`
		Type       string `json:"type"`
	} `json:"user"`

	// event data
	EventData map[string]string `json:"event_data"`
}

// event message which wraps real event data along with other information
type Message struct {
	Event *SysmonEvent `json:"winlog"`
	Agent struct {
		ID       string `json:"id"`
		Hostname string `json:"hostname"`
	} `json:"agent"`
}

// isSysmonEvent returns true if the event is about the Sysmon service
func (event *SysmonEvent) isSysmonEvent() bool {
	switch event.EventID {
	case EServiceStateChange, EConfigStateChange, ESysmonError:
		return true
	}
	return false
}

func (event *SysmonEvent) get(name string) string {
	return event.EventData[name]
}

func (event *SysmonEvent) getBool(name string) bool {
	v, err := strconv.ParseBool(event.get(name))
	if err != nil {
		return false
	}
	return v
}

func (event *SysmonEvent) getInt(name string) (int, error) {
	return strconv.Atoi(event.get(name))
}

func (event *SysmonEvent) set(name, value string) {
	event.EventData[name] = value
}

// getters for frequent used fields
func (event *SysmonEvent) getTimestamp() time.Time {
	t, _ := time.Parse(TimeFormat, event.EventData["UtcTime"]) // use zero value for any errors
	return t
}

// workaround for inconsistency between events, should it be done in preprocessor ?
func (event *SysmonEvent) getProcessId() string {
	switch event.EventID {
	case ECreateRemoteThread, EProcessAccess:
		return event.get("SourceProcessId")
	}
	return event.get("ProcessId")
}

func (event *SysmonEvent) getProcessGUID() string {
	switch event.EventID {
	case ECreateRemoteThread, EProcessAccess:
		return event.get("SourceProcessGUID")
	}
	return event.get("ProcessGuid")
}

func (event *SysmonEvent) getImage() string {
	switch event.EventID {
	case ECreateRemoteThread, EProcessAccess:
		return event.get("SourceImage")
	}
	return event.get("Image")
}

// isSystemEvent returns true if the event's scope is the system wide
func (event *SysmonEvent) isSystemEvent() bool {
	switch event.EventID {
	case EDriverLoad, EWmiEventFilter, EWmiEventConsumer, EWmiEventBinding:
		return true
	}
	return false
}

// isProcessEvent returns true if the event caused by a process. In other words, the event must contain  information to identify that process
func (event *SysmonEvent) isProcessEvent() bool {
	return !event.isSysmonEvent() && !event.isSystemEvent()
}
