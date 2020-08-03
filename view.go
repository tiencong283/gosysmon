package main

import (
	"fmt"
)

const viewTimestampFormat = "2006-01-02 15:04:05.000"
const procRefUrlFormat = `/process?HostId=%s&ProcessGuid=%s`

// HostView is the view layer for Host object
type HostView struct {
	HostId    string
	Name      string
	FirstSeen string
	Active    bool
}

func NewHostView(host *Host) *HostView {
	return &HostView{
		HostId:    host.HostId,
		Name:      host.Name,
		FirstSeen: host.FirstSeen.Format(viewTimestampFormat),
		Active:    host.Active,
	}
}

// IOCView is the view layer for IOCResult object
type IOCView struct {
	ResultId
	Timestamp   string
	IOCType     string
	Indicator   string
	Message     string
	ExternalUrl string
	ProcRefUrl  string
}

func formatIOCType(iocType int) string {
	switch iocType {
	case IOCHash:
		return "Hash"
	case IOCDomain:
		return "Domain"
	case IOCIp:
		return "IP"
	default:
		return "Unknown"
	}
}

func NewIOCView(ioc *IOCResult) *IOCView {
	return &IOCView{
		ResultId:    ioc.ResultId,
		Timestamp:   ioc.Timestamp.Format(viewTimestampFormat),
		IOCType:     formatIOCType(ioc.IOCType),
		Indicator:   ioc.Indicator,
		Message:     ioc.Message,
		ExternalUrl: ioc.ExternalUrl,
		ProcRefUrl:  fmt.Sprintf(procRefUrlFormat, ioc.HostId, ioc.ProcessGuid),
	}
}

// AlertView is the view layer for alert object
type AlertView struct {
	ResultId
	Timestamp string
	Context   map[string]interface{}
	Message   string
	Technique *AttackPattern

	HostName     string
	ProcessImage string
	ProcessId    int
	ProcRefUrl   string
}

func (hm *HostManager) NewAlertView(alert *MitreATTCKResult) *AlertView {
	alertView := &AlertView{
		ResultId:   alert.ResultId,
		Timestamp:  alert.Timestamp.Format(viewTimestampFormat),
		Context:    alert.Context,
		Message:    alert.Message,
		Technique:  alert.Technique,
		ProcRefUrl: fmt.Sprintf(procRefUrlFormat, alert.HostId, alert.ProcessGuid),
	}
	if host := hm.GetHost(alert.HostId); host != nil {
		alertView.HostName = host.Name
		if proc := host.GetProcess(alert.ProcessGuid); proc != nil {
			alertView.ProcessImage = GetImageName(proc.Image)
			alertView.ProcessId = proc.ProcessId
		}
	}
	return alertView
}

// ProcessView is the view layer for Process object
type ProcessView struct {
	Abandoned   bool // true if the process not derived from event ProcessCreate
	ProcessGuid string
	// process creation and termination time
	CreatedAt    string
	TerminatedAt string
	// process state
	State string
	// process info
	ProcessId        int
	Image            string
	ImageName        string
	OriginalFileName string
	CommandLine      string
	CurrentDirectory string
	IntegrityLevel   string
	Hashes           map[string]string

	// product information
	FileVersion, Description, Product, Company string
}

func formatProcState(procState int) string {
	switch procState {
	case PSRunning:
		return "Running"
	case PSStopped:
		return "Stopped"
	}
	return "Unknown"
}

func NewProcessView(proc *Process) *ProcessView {
	procView := &ProcessView{
		Abandoned:        proc.Abandoned,
		ProcessGuid:      proc.ProcessGuid,
		State:            formatProcState(proc.State),
		ProcessId:        proc.ProcessId,
		Image:            proc.Image,
		ImageName:        GetImageName(proc.Image),
		OriginalFileName: proc.OriginalFileName,
		CommandLine:      proc.CommandLine,
		CurrentDirectory: proc.CurrentDirectory,
		IntegrityLevel:   proc.IntegrityLevel,
		Hashes:           proc.Hashes,
		FileVersion:      proc.FileVersion,
		Description:      proc.Description,
		Product:          proc.Product,
		Company:          proc.Company,
	}
	if proc.CreatedAt != nil {
		procView.CreatedAt = proc.CreatedAt.Format(viewTimestampFormat)
	}
	if proc.TerminatedAt != nil {
		procView.TerminatedAt = proc.TerminatedAt.Format(viewTimestampFormat)
	}
	return procView
}

// ActivityLogView is the view layer for ActivityLog object
type ActivityLogView struct {
	Timestamp string
	Type      string
	Message   string
}

func formatActLogType(logType int) string {
	switch logType {
	case LClient:
		return "Client"
	case LServer:
		return "Server"
	}
	return "Unknown"
}

func NewActivityLogView(actLog *ActivityLog) *ActivityLogView {
	return &ActivityLogView{
		Timestamp: actLog.Timestamp.Format(viewTimestampFormat),
		Type:      formatActLogType(actLog.Type),
		Message:   actLog.Message,
	}
}

// ProcessNodeView represents a node for using in network graph, Highcharts
type ProcessNodeView struct {
	ProcessGuid string
	Image       string
	ImageName   string
	ProcessId   int
	NodeType    string // one of 'focus', 'ancestor', 'child'
}

func NewProcessNodeView(proc *Process, nodeType string) *ProcessNodeView {
	return &ProcessNodeView{
		ProcessGuid: proc.ProcessGuid,
		Image:       proc.Image,
		ImageName:   GetImageName(proc.Image),
		ProcessId:   proc.ProcessId,
		NodeType:    nodeType,
	}
}
