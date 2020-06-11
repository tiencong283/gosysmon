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

	ChanBufSize = 1000000
	TimeFormat  = "2006-01-02 15:04:05.999999999"
)

type ListHead struct {
	Next, Prev *ListHead
	*Process
}

// Process is an process/task representation
type Process struct {
	// process creation and termination time
	CreatedAt, TerminatedAt *time.Time
	// process state
	State int
	// process info
	ProcessId        int
	Image            string
	OriginalFileName string
	CommandLine      string
	CurrentDirectory string
	IntegrityLevel   string
	Hashes           map[string]string
	// product information
	FileVersion, Description, Product, Company string
	// relationship
	Parent   *Process
	Children ListHead `json:"-"`
	Sibling  ListHead `json:"-"`
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
		State:       PSRunning,
		ProcessId:   processId,
		Image:       image,
		CommandLine: cmd,
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
	Hosts   map[string]*Host
	EventCh chan *SysmonEvent
	logger  *log.Entry
}

// NewHostManager returns new instance of HostManager
func NewHostManager() *HostManager {
	return &HostManager{
		Hosts:   make(map[string]*Host),
		EventCh: make(chan *SysmonEvent, ChanBufSize),
		logger:  log.WithField("section", "HostManager"),
	}
}

// Start is the thread entry point
func (hm *HostManager) Start() {
	for event := range hm.EventCh {
		if event.isProcessEvent() {
			hm.OnProcessEvent(event)
		}
	}
}

// OnProcessEvent updates the process tree and its entry information
func (hm *HostManager) OnProcessEvent(event *SysmonEvent) {
	host := hm.GetOrCreateHost(event)
	switch event.EventID {
	case EProcessCreate:
		log.Printf("new process (%s): %s\n", event.EventData["ProcessId"], event.EventData["Image"])
		log.Printf("parent process (%s): %s\n\n", event.EventData["ParentProcessId"], event.EventData["ParentImage"])

		// for some reasons (like filtering), it's possible for parent processes not in processList yet
		ppGuid := event.EventData["ParentProcessGuid"]
		parent := host.GetProcess(ppGuid)
		if parent == nil {
			parent = host.AddProcess(ppGuid, event.EventData["ParentProcessId"], event.EventData["ParentImage"], event.EventData["ParentCommandLine"])
		}
		process := host.AddProcess(event.EventData["ProcessGuid"], event.EventData["ProcessId"], event.EventData["Image"], event.EventData["CommandLine"])
		createdAt, _ := time.Parse(TimeFormat, event.EventData["UtcTime"])
		process.CreatedAt = &createdAt
		process.State = PSRunning
		process.OriginalFileName = event.EventData["OriginalFileName"]
		process.CurrentDirectory = event.EventData["CurrentDirectory"]
		process.IntegrityLevel = event.EventData["IntegrityLevel"]
		process.Hashes = StringToMap(event.EventData["Hashes"])

		process.FileVersion = event.EventData["FileVersion"]
		process.Description = event.EventData["Description"]
		process.Product = event.EventData["Product"]
		process.Company = event.EventData["Company"]

		process.Parent = parent
		parent.AddChildProc(process)

		if process.ProcessId == 8160 {
			hm.DumpProcess("{5770385f-c22a-43e0-bf4c-06f5698ffbd9}", "{1db83021-91f2-5ee2-4c01-000000000d00}")
		}
	case EProcessTerminate:
		if process := host.GetProcess(event.EventData["ProcessGuid"]); process != nil {
			process.State = PSStopped
			terminatedAt, _ := time.Parse(TimeFormat, event.EventData["UtcTime"])
			process.TerminatedAt = &terminatedAt
		}
	default:
		var process *Process
		if process = host.GetProcess(event.EventData["ProcessGuid"]); process == nil {
			process = host.AddProcess(event.EventData["ProcessGuid"], event.EventData["ProcessId"], event.EventData["Image"], "")
		}
	}
}

// AddHost adds new host
func (hm *HostManager) AddHost(providerGuid string, host *Host) {
	hm.Hosts[providerGuid] = host
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
