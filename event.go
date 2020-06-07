package main

/*
	These event structures are quickly derived from json using https://github.com/ChimeraCoder/gojson
	Only choose used, useful features so it's not complete
*/

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

// the metadata of all sysmon events
type EventMetadata struct {
	ProviderGUID string `json:"provider_guid"`
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
}

// sysmon event representation
type SysmonEvent struct {
	EventMetadata
	EventData    map[string]string `json:"event_data"`
}

// isProcessEvent returns true if the event caused by a process. In other words, the event must contain  information to identify that process
func (event *SysmonEvent) isProcessEvent() bool {
	switch event.EventID {
	case EServiceStateChange, EDriverLoad, EConfigStateChange, EWmiEventFilter, EWmiEventConsumer, EWmiEventBinding, ESysmonError:
		return false
	}
	return true
}

// event message which wraps event along with other information+
type Message struct {
	Winlog SysmonEvent `json: "winlog"`
}