package main

import (
	"fmt"
	"strings"
)

/*
	References:
		https://attack.mitre.org/techniques/T1053/005/
	List:
		cover execution of schtasks.exe
	Todo:
	Testing:
		https://github.com/redcanaryco/atomic-red-team/blob/master/atomics/T1053.005/T1053.005.md (#1, #2)
*/

// ScheduledTaskFilter filters the technique T1053.005
type ScheduledTaskFilter struct {
	CommonFilterer
	TechniqueId string
}

func (filter *ScheduledTaskFilter) IsSupported(msg *Message) bool {
	return msg.Event.EventID == EProcessCreate && strings.EqualFold(msg.Event.getImage(), "C:\\Windows\\System32\\schtasks.exe")
}

func (filter *ScheduledTaskFilter) MessageCh() chan *Message {
	return filter.messageCh
}

func NewScheduledTaskFilter() *ScheduledTaskFilter {
	return &ScheduledTaskFilter{
		CommonFilterer: NewCommonFilterer("Scheduled Task Filter"),
		TechniqueId:    "T1053.005",
	}
}

func (filter *ScheduledTaskFilter) Init() error {
	return nil
}

func (filter *ScheduledTaskFilter) EventCh() chan *Message {
	return filter.messageCh
}

func (filter *ScheduledTaskFilter) StateCh() chan int {
	return filter.State
}

func (filter *ScheduledTaskFilter) SetAlertCh(alertCh chan interface{}) {
	filter.AlertCh = alertCh
}

func (filter *ScheduledTaskFilter) Start() {
	for msg := range filter.messageCh {
		event := msg.Event
		/*
			Usage: see schtasks /?
				SCHTASKS /Create [/S system [/U username [/P [password]]]]
			    [/RU username [/RP password]] /SC schedule [/MO modifier] [/D day]
			    [/M months] [/I idletime] /TN taskname /TR taskrun [/ST starttime]
			    [/RI interval] [ {/ET endtime | /DU duration} [/K] [/XML xmlfile] [/V1]]
			    [/SD startdate] [/ED enddate] [/IT | /NP] [/Z] [/F] [/HRESULT] [/?]

			Mandatory:
				SCHTASKS /Create /SC schedule /TN taskname /TR taskrun

			Examples:
				schtasks /create /tn "T1053_005_OnLogon" /sc onlogon /tr "cmd.exe /c calc.exe"

		*/
		var taskName, schedule, taskRun string
		args := commandLineToArgv(event.get("CommandLine"))
		if len(args) >= 8 && strings.EqualFold(args[1], "/create") {
			for i := 0; i < len(args); i++ {
				argv := args[i]
				if argv[0] == '/' && i != len(args)-1 {
					switch strings.ToLower(argv) {
					case "/sc":
						schedule = args[i+1]
					case "/tn":
						taskName = args[i+1]
					case "/tr":
						taskRun = args[i+1]
					}
				}
			}
			alertMsg := fmt.Sprintf("Scheduled task named '%s' added", taskName)
			alert := NewMitreATTCKResult(true, filter.TechniqueId, alertMsg, msg, true)
			alert.AddContext("TaskName", taskName)
			alert.AddContext("TaskSchedule", schedule)
			alert.AddContext("TaskRun", taskRun)
			filter.AlertCh <- alert
		}
	}
	filter.State <- 1
}
