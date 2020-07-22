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
	ProviderGuid string
	Name         string
	FirstSeen    time.Time
	Active       bool
	Procs        map[string]*Process `json:"-"`
	ProcsLock    sync.Mutex
}

// NewHost returns new instance of Host
func NewHostFrom(event *SysmonEvent) *Host {
	t, _ := time.Parse(TimeFormat, event.EventData["UtcTime"])
	return &Host{
		ProviderGuid: event.ProviderGUID,
		Name:         event.ComputerName,
		FirstSeen:    t,
		Active:       true,
		Procs:        make(map[string]*Process, 10000),
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

// HostManager manages sensor clients, the key is ProviderGuid which is the identity of the application or service (Sysmon) that logged the record
// so it can be used relatively to represent a computer
type HostManager struct {
	State      chan int // simulate ON/OFF state
	Hosts      map[string]*Host
	HostsLock  sync.Mutex
	EventCh    chan *SysmonEvent
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
		State:   make(chan int),
		Hosts:   make(map[string]*Host),
		EventCh: make(chan *SysmonEvent, EventChBufSize),
		AlertCh: alertCh,
		Alerts:  make([]*MitreATTCKResult, 0, AlertChBufSize*10),
		IOCs:    make([]*IOCResult, 0, AlertChBufSize*10),
		logger:  log.WithField("section", "HostManager"),
	}
}

// LoadData loads all data from database into the HostManager
func (hm *HostManager) LoadData() error {
	hosts, err := PgConn.GetAllHosts()
	if err != nil {
		return err
	}
	for _, host := range hosts { // load hosts
		hm.Hosts[host.ProviderGuid] = host
		procs, err := PgConn.GetProcessesByHost(host.ProviderGuid)
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
			features, err := PgConn.GetFeaturesByProcess(host.ProviderGuid, proc.ProcessGuid)
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
	host := hm.GetHost(resultId.ProviderGuid)
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
			parent = host.AddProcess(true, ppGuid, event.get("ParentProcessId"), event.get("ParentImage"), event.get("ParentCommandLine"))
			if err := PgConn.SaveProc(event.ProviderGUID, parent); err != nil {
				log.Warnf("cannot persist the process, %s\n", err)
			}
		}
		process := host.AddProcess(false, processGuid, processId, event.get("Image"), event.get("CommandLine"))
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
		if err := PgConn.SaveProc(event.ProviderGUID, process); err != nil {
			log.Warnf("cannot persist the process, %s\n", err)
		}

	case EProcessTerminate:
		if process := host.GetProcess(processGuid); process != nil {
			process.State = PSStopped
			process.TerminatedAt = event.timestamp()
			if err := PgConn.UpdateProcTerm(event.ProviderGUID, process); err != nil {
				log.Warnf("cannot update the process state, %s\n", err)
			}
		}
	default:
		var process *Process
		if process = host.GetProcess(processGuid); process == nil {
			process = host.AddProcess(true, processGuid, processId, event.get("Image"), "")
			if err := PgConn.SaveProc(event.ProviderGUID, process); err != nil {
				log.Warnf("cannot persist the process, %s\n", err)
			}
		}
	}
}

// AddHost adds new host
func (hm *HostManager) AddHost(providerGuid string, host *Host) {
	hm.HostsLock.Lock()
	hm.Hosts[providerGuid] = host
	hm.HostsLock.Unlock()
	_ = PgConn.SaveHost(providerGuid, host)
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

// request handler for "/api/host"
func (hm *HostManager) AllHostHandler(c *gin.Context) {
	hosts := make([]*Host, 0)

	hm.HostsLock.Lock()
	for _, host := range hm.Hosts {
		hosts = append(hosts, host)
	}
	hm.HostsLock.Unlock()

	c.JSON(http.StatusOK, hosts)
}

// request handler for "/api/ioc"
func (hm *HostManager) AllIOCHandler(c *gin.Context) {
	hm.IOCsLock.Lock()
	iocList := make([]*IOCResult, len(hm.IOCs))
	copy(iocList, hm.IOCs)
	hm.IOCsLock.Unlock()

	c.JSON(http.StatusOK, iocList)
}

// AlertView is the view layer for alert object
type AlertView struct {
	*MitreATTCKResult
	HostName     string
	ProcessImage string
	ProcessId    int
}

func (hm *HostManager) NewAlertView(alert *MitreATTCKResult) *AlertView {
	alertView := new(AlertView)
	alertView.MitreATTCKResult = alert
	if host := hm.GetHost(alert.ProviderGuid); host != nil {
		alertView.HostName = host.Name
		if proc := host.GetProcess(alert.ProcessGuid); proc != nil {
			alertView.ProcessImage = GetImageName(proc.Image)
			alertView.ProcessId = proc.ProcessId
		}
	}
	return alertView
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

// ProcessView is the view layer for Process object (due to some conflict on json tags)
type ProcessView struct {
	Abandoned   bool // true if the process not derived from event ProcessCreate
	ProcessGuid string
	// process creation and termination time
	CreatedAt    *time.Time
	TerminatedAt *time.Time
	// process state
	State int
	// process info
	ProcessId        int
	Image            string
	ImageName        string
	OriginalFileName string
	CommandLine      string
	CurrentDirectory string
	IntegrityLevel   string
	// workaround for errors in client
	MD5    string
	SHA256 string
	SHA1   string
	// product information
	FileVersion, Description, Product, Company string
}

func NewProcessView(proc *Process) *ProcessView {
	return &ProcessView{
		Abandoned:        proc.Abandoned,
		ProcessGuid:      proc.ProcessGuid,
		CreatedAt:        proc.CreatedAt,
		TerminatedAt:     proc.TerminatedAt,
		State:            proc.State,
		ProcessId:        proc.ProcessId,
		Image:            proc.Image,
		ImageName:        GetImageName(proc.Image),
		OriginalFileName: proc.OriginalFileName,
		CommandLine:      proc.CommandLine,
		CurrentDirectory: proc.CurrentDirectory,
		IntegrityLevel:   proc.IntegrityLevel,
		MD5:              proc.Hashes["MD5"],
		SHA256:           proc.Hashes["SHA256"],
		SHA1:             proc.Hashes["SHA1"],
		FileVersion:      proc.FileVersion,
		Description:      proc.Description,
		Product:          proc.Product,
		Company:          proc.Company,
	}
}

// request handler for "/api/process" (process information)
func (hm *HostManager) ProcessHandler(c *gin.Context) {
	providerGuid, processGuid := c.PostForm("ProviderGuid"), c.PostForm("ProcessGuid")
	if providerGuid == "" || processGuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
		return
	}

	hm.HostsLock.Lock()
	host := hm.GetHost(providerGuid)
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
	providerGuid, processGuid := c.PostForm("ProviderGuid"), c.PostForm("ProcessGuid")
	if providerGuid == "" || processGuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
		return
	}

	hm.HostsLock.Lock()
	host := hm.GetHost(providerGuid)
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
