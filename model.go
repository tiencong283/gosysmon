package main

import (
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	// process running, process stopped
	PSRunning = iota
	PSStopped
)

type ListHead struct {
	Next, Prev *ListHead
	*Process
}

// Process is an process/task representation
type Process struct {
	ProcessGuid string `json:"-"`
	// process creation and termination time
	CreatedAt    *time.Time `json:"-"`
	TerminatedAt *time.Time `json:"-"`
	// process state
	State int `json:"-"`
	// process info
	ProcessId        int    `json:"-"`
	Image            string `json:"-"`
	OriginalFileName string
	CommandLine      string
	CurrentDirectory string
	IntegrityLevel   string
	Hashes           map[string]string
	// product information
	FileVersion, Description, Product, Company string
	// relationship
	Parent   *Process `json:"-"`
	Children ListHead `json:"-"`
	Sibling  ListHead `json:"-"`

	Alerts []*RContext `json:"-"`
}

// client representation
type Host struct {
	Name      string
	FirstSeen time.Time
	Active    bool
	Procs     map[string]*Process `json:"-"`
}

// NewHost returns new instance of Host
func NewHostFrom(event *SysmonEvent) *Host {
	t, _ := time.Parse(TimeFormat, event.EventData["UtcTime"])
	return &Host{
		Name:      event.ComputerName,
		FirstSeen: t,
		Active:    true,
		Procs:     make(map[string]*Process, 10000),
	}
}

// AddChildProc links a child process to the current process
func (current *Process) AddChildProc(proc *Process) {
	firstHead := current.Children.Next
	lastChild := current.Children.Prev

	current.Children.Prev = &proc.Sibling
	proc.Sibling.Next = &current.Children
	if firstHead == nil {
		current.Children.Next = &proc.Sibling
		proc.Sibling.Prev = &current.Children
	} else { // insert at tail
		proc.Sibling.Prev = lastChild
		lastChild.Sibling.Next = &proc.Sibling
	}
}

// getProcess returns the process with the corresponding pGuid
func (host *Host) GetProcess(pGuid string) *Process {
	return host.Procs[pGuid]
}

// AddProcess creates a new process for the event
func (host *Host) AddProcess(pGuid, pid, image, cmd string) *Process {
	processId, _ := strconv.Atoi(pid)
	proc := &Process{
		ProcessGuid: pGuid,
		State:       PSRunning,
		ProcessId:   processId,
		Image:       image,
		CommandLine: cmd,
		Alerts:      make([]*RContext, 0, 32),
	}
	proc.Sibling.Process = proc
	proc.Children.Process = proc
	host.Procs[pGuid] = proc
	return proc
}

// getNumberOfProcesses returns number of processes
func (host *Host) GetNumberOfProcesses() int {
	return len(host.Procs)
}

// HostManager manages sensor clients, the key is ProviderGuid which is the identity of the application or service (Sysmon) that logged the record
// so it can be used relatively to represent a computer
type HostManager struct {
	State   chan int // simulate ON/OFF state
	Hosts   map[string]*Host
	EventCh chan *SysmonEvent
	AlertCh chan *RContext
	Alerts  []*RContext
	logger  *log.Entry
	DBConn  *DBConn
}

// NewHostManager returns new instance of HostManager
func NewHostManager(alertCh chan *RContext, dbConn *DBConn) *HostManager {
	return &HostManager{
		State:   make(chan int),
		Hosts:   make(map[string]*Host),
		EventCh: make(chan *SysmonEvent, EventChBufSize),
		AlertCh: alertCh,
		Alerts:  make([]*RContext, 0, AlertChBufSize*10),
		logger:  log.WithField("section", "HostManager"),
		DBConn:  dbConn,
	}
}

// Start is the thread entry point
func (hm *HostManager) Start() {
	eventChClosed := false
	alertChClosed := false
	for !eventChClosed || !alertChClosed {
		select {
		case event, ok := <-hm.EventCh:
			if !ok {
				eventChClosed = true
				continue
			}
			if event.isProcessEvent() {
				hm.OnProcessEvent(event)
			}
		case alert, ok := <-hm.AlertCh:
			if !ok {
				alertChClosed = true
				continue
			}
			hm.OnAlert(alert)
		}
	}
	hm.State <- 0 // OFF
}

// OnAlert updates the alert into its process
func (hm *HostManager) OnAlert(alert *RContext) {
	host := hm.GetHost(alert.ProviderGUID)
	if host == nil {
		hm.AlertCh <- alert
	} else {
		proc := host.GetProcess(alert.ProcessGuid)
		if proc == nil { // in rare cases when the process is not mapped to the tree yet (some kind of delays)
			hm.AlertCh <- alert
			return
		}
		proc.Alerts = append(proc.Alerts, alert)
		log.Debug(ToJson(alert))
	}
}

// OnProcessEvent updates the process tree and its entry information
func (hm *HostManager) OnProcessEvent(event *SysmonEvent) {
	host := hm.GetOrCreateHost(event)
	processId := event.get("ProcessId")
	processGuid := event.get("ProcessGuid")

	switch event.EventID {
	case EProcessCreate:
		// for some reasons (like filtering), it's possible for parent processes not in processList yet
		ppGuid := event.get("ParentProcessGuid")
		parent := host.GetProcess(ppGuid)
		if parent == nil {
			parent = host.AddProcess(ppGuid, event.get("ParentProcessId"), event.get("ParentImage"), event.get("ParentCommandLine"))
			_ = hm.DBConn.SaveProc(host, parent)
		}
		process := host.AddProcess(processGuid, processId, event.get("Image"), event.get("CommandLine"))
		process.CreatedAt = event.timestamp()
		process.OriginalFileName = event.get("OriginalFileName")
		process.CurrentDirectory = event.get("CurrentDirectory")
		process.IntegrityLevel = event.get("IntegrityLevel")
		process.Hashes = StringToMap(event.get("Hashes"))

		process.FileVersion = event.get("FileVersion")
		process.Description = event.get("Description")
		process.Product = event.get("Product")
		process.Company = event.get("Company")

		process.Parent = parent
		parent.AddChildProc(process)
		_ = hm.DBConn.SaveProc(host, process)

	case EProcessTerminate:
		if process := host.GetProcess(processGuid); process != nil {
			process.State = PSStopped
			process.TerminatedAt = event.timestamp()
			_ = hm.DBConn.UpdateProcTerm(host, process)
		}
	default:
		var process *Process
		if process = host.GetProcess(processGuid); process == nil {
			process = host.AddProcess(processGuid, processId, event.get("Image"), "")
			_ = hm.DBConn.SaveProc(host, process)
		}
	}
}

// AddHost adds new host
func (hm *HostManager) AddHost(providerGuid string, host *Host) {
	hm.Hosts[providerGuid] = host
	hm.DBConn.SaveHost(host)
}

// GetHost return the host with corresponding providerGuid
func (hm *HostManager) GetHost(providerGuid string) *Host {
	return hm.Hosts[providerGuid]
}

// GetOrCreateHost return the existing host or create a new host for the event
func (hm *HostManager) GetOrCreateHost(event *SysmonEvent) *Host {
	providerGuid := event.ProviderGUID
	host := hm.Hosts[providerGuid]
	if host != nil {
		return host
	}
	host = NewHostFrom(event)
	hm.AddHost(providerGuid, host)
	return host
}

// GetNumOfHosts returns number of hosts
func (hm *HostManager) GetNumOfHosts() int {
	return len(hm.Hosts)
}

// *********************************************************************************************************************
func (hm *HostManager) DumpProcess(providerGuid, processGuid string) {
	if host := hm.GetHost(providerGuid); host != nil {
		log.Println(ToJson(host))
		log.Printf("Number of processes: %d\n", host.GetNumberOfProcesses())
		proc := host.GetProcess(processGuid)
		if proc == nil {
			return
		}
		log.Println(ToJson(proc))

		head := &proc.Children
		for cur := head.Next; cur != head; cur = cur.Next {
			log.Printf("pid: %d, image: %s, cmd: %s\n", cur.ProcessId, cur.Image, cur.CommandLine)
		}
	}
}
