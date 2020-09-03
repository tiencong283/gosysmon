package main

import (
	"fmt"
	"strings"
)

/*
	References:
		https://attack.mitre.org/techniques/T1569/002/
		https://attack.mitre.org/techniques/T1543/003/
	List:
		Detect service execution via system utility sc.exe
		For service installation and modification. There are two options:
			1. We can look for changes in service Registry entries. In this way,we can capture almost every cases
				no matter how the adversary do it,but it does not hint about the original process that do that action
			2. The old one, monitor processes and commandline arguments for utilities such as sc.exe and Reg
	Todo:
		cover reg.exe for service modification
	Testing:
		https://github.com/redcanaryco/atomic-red-team/blob/master/atomics/T1569.002/T1569.002.md (#1)
		https://github.com/redcanaryco/atomic-red-team/blob/master/atomics/T1543.003/T1543.003.md (#1, #2)
*/

// SCExeFilter filter behaviors related to sc.exe process
type SCExeFilter struct {
	CommonFilterer
}

func NewSCExeFilter() *SCExeFilter {
	return &SCExeFilter{
		CommonFilterer: NewCommonFilterer("SC.EXE Filter"),
	}
}

func (filter *SCExeFilter) IsSupported(msg *Message) bool {
	return msg.Event.EventID == EProcessCreate && strings.EqualFold(msg.Event.getImage(), "C:\\Windows\\System32\\sc.exe")
}

func (filter *SCExeFilter) Init() error {
	return nil
}

func (filter *SCExeFilter) MessageCh() chan *Message {
	return filter.messageCh
}

func (filter *SCExeFilter) StateCh() chan int {
	return filter.State
}

func (filter *SCExeFilter) SetAlertCh(alertCh chan interface{}) {
	filter.AlertCh = alertCh
}

func (filter *SCExeFilter) Start() {
	for msg := range filter.messageCh {
		event := msg.Event
		/*
			Usage: sc <server> [command] [service name] <option1> <option2>...
				server: \\ServerName
			common cases:
				Starts a service running:
					sc <server> start [service name] <arg1> <arg2> ...
				Modify a service:
					sc <server> config [service name] <option1> <option2>...
		*/
		args := commandLineToArgv(event.get("CommandLine"))
		remoteHost := "localhost"
		if len(args) >= 3 {
			if args[1][0] == '\\' && args[1][1] == '\\' {
				remoteHost = args[1]
				args = append(args[0:1], args[2:]...)
			}
			switch action := strings.ToLower(args[1]); action {
			case "start":
				if len(args) >= 3 {
					serviceName := args[2]
					alertMsg := fmt.Sprintf("An attempt to start service '%s'", serviceName)
					if remoteHost != "localhost" {
						alertMsg += " on host " + remoteHost
					}
					alert := NewMitreATTCKResult(true, "T1569.002", alertMsg, msg, true)
					alert.AddContext("ServiceName", serviceName)
					filter.AlertCh <- alert
				}
			case "create", "config":
				var alertMsg string
				if len(args) >= 4 { // at least binPath= <BinaryPathName to the .exe file>
					serviceName := args[2]
					serviceOptions := args[3:]
					if action == "create" {
						alertMsg = fmt.Sprintf("An attempt to create service '%s'", serviceName)
					} else {
						alertMsg = fmt.Sprintf("An attempt to modify service '%s'", serviceName)
					}
					if remoteHost != "localhost" {
						alertMsg += " on host " + remoteHost
					}
					alert := NewMitreATTCKResult(true, "T1543.003", alertMsg, msg, true)
					alert.AddContext("ServiceName", serviceName)
					alert.AddContext("ServiceOptions", serviceOptions)
					filter.AlertCh <- alert
				}
			}
		}
	}
	filter.State <- 1
}
