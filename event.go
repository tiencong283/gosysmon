package main

/*
	These event structures are quickly derived from json using https://github.com/ChimeraCoder/gojson
	Only choose used, useful features so it's not complete
*/

// the metadata of all sysmon events
type EventMetadata struct {
	ProviderGUID string `json:"provider_guid"`
	ComputerName string `json:"computer_name"`
	RecordID     int64  `json:"record_id"`
	EventID      int64  `json:"event_id"`
	Opcode       string `json:"opcode"`
	Process      struct {
		Pid    int64 `json:"pid"`
		Thread struct {
			ID int64 `json:"id"`
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
	EventData map[string]string `json:"event_data"`
}
// event message which wraps event along with other information
type Message struct {
	Winlog SysmonEvent `json: "winlog"`
}