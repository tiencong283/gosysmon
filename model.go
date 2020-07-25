package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"sync"
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
	Abandoned   bool   // true if the process not derived from event ProcessCreate
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
	ParentPGuid string   `json:"-"`
	Parent      *Process `json:"-"`
	Children    ListHead `json:"-"`
	Sibling     ListHead `json:"-"`

	Features []*MitreATTCKResult `json:"-"`
}

// client representation
type Host struct {
	HostId    string
	Name      string
	FirstSeen *time.Time
	Active    bool
	Procs     map[string]*Process `json:"-"`
	ProcsLock sync.Mutex
}

// NewHost returns new instance of Host
func NewHostFrom(msg *Message) *Host {
	return &Host{
		HostId:    msg.Agent.ID,
		Name:      msg.Event.ComputerName,
		FirstSeen: msg.Event.timestamp(),
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
func (host *Host) AddProcess(abandoned bool, pGuid, pid, image, cmd string) *Process {
	processId, _ := strconv.Atoi(pid)
	proc := &Process{
		Abandoned:   abandoned,
		ProcessGuid: pGuid,
		State:       PSRunning,
		ProcessId:   processId,
		Image:       image,
		CommandLine: cmd,
		Features:    make([]*MitreATTCKResult, 0, 32),
	}
	proc.Sibling.Process = proc
	proc.Children.Process = proc

	host.ProcsLock.Lock()
	defer host.ProcsLock.Unlock()
	host.Procs[pGuid] = proc
	return proc
}

// getNumberOfProcesses returns number of processes
func (host *Host) GetNumberOfProcesses() int {
	return len(host.Procs)
}

// SaveProc saves process into db
func (host *Host) SaveProc(proc *Process) error {
	return PgConn.SaveProc(host.HostId, proc)
}

// UpdateProcTerm updates process state to stopped and save into db
func (host *Host) UpdateProcTerm(timestamp *time.Time, proc *Process) error {
	proc.State = PSStopped
	proc.TerminatedAt = timestamp
	return PgConn.UpdateProcTerm(host.HostId, proc)
}

// HostManager manages sensor clients, the key is HostId which is the identity of the application or service (Sysmon) that logged the record
// so it can be used relatively to represent a computer
type HostManager struct {
	State      chan int // simulate ON/OFF state
	Hosts      map[string]*Host
	HostsLock  sync.Mutex
	MessageCh  chan *Message
	AlertCh    chan interface{}
	Alerts     []*MitreATTCKResult
	AlertsLock sync.Mutex
	IOCs       []*IOCResult
	IOCsLock   sync.Mutex
	logger     *log.Entry
}

// NewHostManager returns new instance of HostManager
func NewHostManager(alertCh chan interface{}) *HostManager {
	return &HostManager{
		State:     make(chan int),
		Hosts:     make(map[string]*Host),
		MessageCh: make(chan *Message, EventChBufSize),
		AlertCh:   alertCh,
		Alerts:    make([]*MitreATTCKResult, 0, AlertChBufSize*10),
		IOCs:      make([]*IOCResult, 0, AlertChBufSize*10),
		logger:    log.WithField("section", "HostManager"),
	}
}

// LoadData loads all data from database into the HostManager
func (hm *HostManager) LoadData() error {
	hosts, err := PgConn.GetAllHosts()
	if err != nil {
		return err
	}
	for _, host := range hosts { // load hosts
		hm.Hosts[host.HostId] = host
		procs, err := PgConn.GetProcessesByHost(host.HostId)
		if err != nil {
			return err
		}
		for _, proc := range procs { // load processes
			proc.Sibling.Process = proc
			proc.Children.Process = proc

			if proc.ParentPGuid != "" {
				parent := host.GetProcess(proc.ParentPGuid)
				if parent == nil {
					return errors.New("ERROR: malformed process database")
				}
				proc.Parent = parent
				parent.AddChildProc(proc)
			}
			host.Procs[proc.ProcessGuid] = proc

			// load features
			features, err := PgConn.GetFeaturesByProcess(host.HostId, proc.ProcessGuid)
			if err != nil {
				return err
			}
			if len(features) > 0 {
				proc.Features = features
			} else {
				proc.Features = make([]*MitreATTCKResult, 0, 32)
			}
		}
	}
	// load all IOCs
	iocs, err := PgConn.GetAllIOCs()
	if err != nil {
		return err
	}
	if len(iocs) > 0 {
		hm.IOCs = append(hm.IOCs, iocs...)
	}
	return nil
}

// Start is the thread entry point
func (hm *HostManager) Start() {
	eventChClosed := false
	alertChClosed := false
	for !eventChClosed || !alertChClosed {
		select {
		case msg, ok := <-hm.MessageCh:
			if !ok {
				eventChClosed = true
				continue
			}
			if msg.Event.isProcessEvent() {
				hm.OnProcessEvent(msg)
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
func (hm *HostManager) OnAlert(alert interface{}) {
	var resultId *ResultId
	var mar *MitreATTCKResult
	var ioc *IOCResult

	switch alert := alert.(type) {
	case *MitreATTCKResult:
		mar = alert
		resultId = &mar.ResultId
	case *IOCResult:
		ioc = alert
		resultId = &ioc.ResultId
	default:
		return
	}
	host := hm.GetHost(resultId.HostId)
	if host == nil {
		// todo: the channel can be closed
		hm.AlertCh <- alert
	} else {
		proc := host.GetProcess(resultId.ProcessGuid)
		if proc == nil { // in rare cases when the process is not mapped to the tree yet (some kind of delays)
			// todo: the channel can be closed
			hm.AlertCh <- alert
			return
		}
		// mapping to corresponding process and db
		switch alert.(type) {
		case *MitreATTCKResult:
			if mar != nil && mar.IsAlert {
				hm.Alerts = append(hm.Alerts, mar)
			}
			proc.Features = append(proc.Features, mar)
			if err := PgConn.SaveFeature(mar); err != nil {
				log.Warnf("cannot persist the feature, %s\n", err)
			}
		case *IOCResult:
			hm.IOCsLock.Lock()
			hm.IOCs = append(hm.IOCs, ioc)
			hm.IOCsLock.Unlock()

			if err := PgConn.SaveIOC(ioc); err != nil {
				log.Warnf("cannot persist the IOC, %s\n", err)
			}
		}
	}
}

// OnProcessEvent updates the process tree and its entry information
func (hm *HostManager) OnProcessEvent(msg *Message) {
	event := msg.Event
	host := hm.GetOrCreateHost(msg)
	processId := event.get("ProcessId")
	processGuid := event.get("ProcessGuid")

	switch event.EventID {
	case EProcessCreate:
		// for some reasons (like filtering), it's possible for parent processes not in processList yet
		ppGuid := event.get("ParentProcessGuid")
		parent := host.GetProcess(ppGuid)
		if parent == nil {
			parent = host.AddProcess(true, ppGuid, event.get("ParentProcessId"), event.get("ParentImage"), event.get("ParentCommandLine"))
			if err := host.SaveProc(parent); err != nil {
				log.Warnf("cannot persist the proc, %s\n", err)
			}
		}
		proc := host.AddProcess(false, processGuid, processId, event.get("Image"), event.get("CommandLine"))
		proc.CreatedAt = event.timestamp()
		proc.OriginalFileName = event.get("OriginalFileName")
		proc.CurrentDirectory = event.get("CurrentDirectory")
		proc.IntegrityLevel = event.get("IntegrityLevel")
		proc.Hashes = StringToMap(event.get("Hashes"))

		proc.FileVersion = event.get("FileVersion")
		proc.Description = event.get("Description")
		proc.Product = event.get("Product")
		proc.Company = event.get("Company")

		proc.Parent = parent
		parent.AddChildProc(proc)
		if err := host.SaveProc(proc); err != nil {
			log.Warnf("cannot persist the proc, %s\n", err)
		}

	case EProcessTerminate:
		if proc := host.GetProcess(processGuid); proc != nil {
			if err := host.UpdateProcTerm(event.timestamp(), proc); err != nil {
				log.Warnf("cannot update the proc state, %s\n", err)
			}
		}
	default:
		var proc *Process
		if proc = host.GetProcess(processGuid); proc == nil {
			proc = host.AddProcess(true, processGuid, processId, event.get("Image"), "")
			if err := host.SaveProc(proc); err != nil {
				log.Warnf("cannot persist the proc, %s\n", err)
			}
		}
	}
}

// AddHost adds new host
func (hm *HostManager) AddHost(hostId string, host *Host) {
	hm.HostsLock.Lock()
	hm.Hosts[hostId] = host
	hm.HostsLock.Unlock()
	_ = PgConn.SaveHost(hostId, host)
}

// GetHost return the host with corresponding hostId
func (hm *HostManager) GetHost(hostId string) *Host {
	return hm.Hosts[hostId]
}

// GetOrCreateHost return the existing host or create a new host for the event
func (hm *HostManager) GetOrCreateHost(msg *Message) *Host {
	hostId := msg.Agent.ID
	host := hm.GetHost(hostId)
	if host != nil {
		return host
	}
	host = NewHostFrom(msg)
	hm.AddHost(hostId, host)
	return host
}

// GetNumOfHosts returns number of hosts
func (hm *HostManager) GetNumOfHosts() int {
	return len(hm.Hosts)
}

// request handler for "/api/host"
func (hm *HostManager) AllHostHandler(c *gin.Context) {
	hosts := make([]*HostView, 0)

	hm.HostsLock.Lock()
	for _, host := range hm.Hosts {
		hosts = append(hosts, NewHostView(host))
	}
	hm.HostsLock.Unlock()

	c.JSON(http.StatusOK, hosts)
}

// request handler for "/api/ioc"
func (hm *HostManager) AllIOCHandler(c *gin.Context) {
	hm.IOCsLock.Lock()
	iocList := make([]*IOCView, len(hm.IOCs))
	for i, ioc := range hm.IOCs {
		iocList[i] = NewIOCView(ioc)
	}
	hm.IOCsLock.Unlock()

	c.JSON(http.StatusOK, iocList)
}

// request handler for "/api/alert"
func (hm *HostManager) AllAlertHandler(c *gin.Context) {
	hm.AlertsLock.Lock()
	alertViews := make([]*AlertView, len(hm.Alerts))
	for i, alert := range hm.Alerts {
		alertViews[i] = hm.NewAlertView(alert)
	}
	hm.AlertsLock.Unlock()

	c.JSON(http.StatusOK, alertViews)
}

// request handler for "/api/process" (process information)
func (hm *HostManager) ProcessHandler(c *gin.Context) {
	hostId, processGuid := c.PostForm("HostId"), c.PostForm("ProcessGuid")
	if hostId == "" || processGuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
		return
	}

	hm.HostsLock.Lock()
	host := hm.GetHost(hostId)
	hm.HostsLock.Unlock()
	if host == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown host id"})
		return
	}

	host.ProcsLock.Lock()
	proc := host.GetProcess(processGuid)
	host.ProcsLock.Unlock()
	if proc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown process id"})
		return
	}
	c.JSON(http.StatusOK, NewProcessView(proc))
}

type ProcessTree struct {
	Nodes []*ProcessView
	Links [][2]string
}

// request handler for "/api/process-tree" (process relationship information)
func (hm *HostManager) ProcessTreeHandler(c *gin.Context) {
	hostId, processGuid := c.PostForm("HostId"), c.PostForm("ProcessGuid")
	if hostId == "" || processGuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
		return
	}

	hm.HostsLock.Lock()
	host := hm.GetHost(hostId)
	hm.HostsLock.Unlock()
	if host == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown host id"})
		return
	}

	host.ProcsLock.Lock()
	proc := host.GetProcess(processGuid)
	host.ProcsLock.Unlock()
	if proc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown process id"})
		return
	}
	procRoot := proc
	for ; procRoot.Parent != nil; procRoot = procRoot.Parent { // traverse to its root
	}
	procTree := &ProcessTree{
		Nodes: make([]*ProcessView, 0),
		Links: make([][2]string, 0),
	}
	queue := []*Process{procRoot}

	var curProc *Process
	for len(queue) > 0 {
		curProc, queue = queue[0], queue[1:]
		procTree.Nodes = append(procTree.Nodes, NewProcessView(curProc))

		head := &curProc.Children
		for cur := head.Next; cur != nil && cur != head; cur = cur.Next {
			procTree.Links = append(procTree.Links, [2]string{curProc.ProcessGuid, cur.ProcessGuid})
			queue = append(queue, cur.Process)
		}
	}
	c.JSON(http.StatusOK, procTree)
}
