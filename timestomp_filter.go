package main

import (
	"fmt"
	"time"
)

/*
	References:
		https://attack.mitre.org/techniques/T1070/006/
	List:
		Detect a process changing a file creation time
	Todo:
		Detect last modified, last accessed timestamps
	Testing:
		https://github.com/redcanaryco/atomic-red-team/blob/master/atomics/T1070.006/T1070.006.md (#5)
*/

var FilteredImages = []string{
	"C:\\Windows\\System32\\wuauclt.exe",
	"C:\\Windows\\System32\\backgroundTaskHost.exe",
	"C:\\Windows\\System32\\RuntimeBroker.exe",
}

// TimestompFilter filters T1070.006 technique
type TimestompFilter struct {
	CommonFilterer
	TechniqueId string
}

func (filter *TimestompFilter) IsSupported(msg *Message) bool {
	return msg.Event.EventID == EFileCreateTime
}

func (filter *TimestompFilter) MessageCh() chan *Message {
	return filter.messageCh
}

func NewTimestompFilter() *TimestompFilter {
	return &TimestompFilter{
		CommonFilterer: NewCommonFilterer("Timestomp  Filter"),
		TechniqueId:    "T1099",
	}
}

func (filter *TimestompFilter) Init() error {
	return nil
}

func (filter *TimestompFilter) EventCh() chan *Message {
	return filter.messageCh
}

func (filter *TimestompFilter) StateCh() chan int {
	return filter.State
}

func (filter *TimestompFilter) SetAlertCh(alertCh chan interface{}) {
	filter.AlertCh = alertCh
}

func (filter *TimestompFilter) Start() {
	for msg := range filter.messageCh {
		event := msg.Event
		targetFileName := event.get("TargetFilename")
		// filtering
		if SliceContainsIgnoreCase(FilteredImages, event.getImage()) {
			continue
		}
		if !HasPrefixIgnoreCase(targetFileName, "C:\\Users\\") { // not so reliable
			continue
		}
		ts, _ := time.Parse(TimeFormat, event.get("CreationUtcTime"))
		previousTs, _ := time.Parse(TimeFormat, event.get("PreviousCreationUtcTime"))
		if ts.Before(previousTs) { // should we consider the opposite case ?
			alertMsg := fmt.Sprintf("Creation timestamp of %s has been changed to older", GetImageName(targetFileName))
			alert := NewMitreATTCKResult(true, filter.TechniqueId, alertMsg, msg, true)
			alert.MergeContext(event.EventData)
			alert.AddContext("EventID", event.EventID)
			alert.AddContext("RecordID", event.RecordID)
			filter.AlertCh <- alert
		}
	}
	filter.State <- 1
}
