package main

import (
	"sync"
	"time"
)

type Process struct {
}

// client representation
type Host struct {
	ComputerName string
	FirstSeen    time.Time
	IsActive     bool
	ProcessList  map[string]*Process
}

// NewHost returns new instance of HostManager
func NewHost(computerName string) *Host {
	return &Host{
		ComputerName: computerName,
		IsActive:     true,
		ProcessList:  make(map[string]*Process, 10000),
	}
}

// getNumberOfProcesses returns number of processes
func (host *Host) getNumberOfProcesses() int {
	return len(host.ProcessList)
}

// client manager, the key is ProviderGuid which is the identity of the application or service (sysmon) that logged the record
// so it can be used relatively to represent a computer
type HostManager struct {
	Hosts map[string]*Host
	Mux   sync.Mutex
}

// NewHostManager returns new instance of HostManager
func NewHostManager() *HostManager {
	return &HostManager{
		Hosts: make(map[string]*Host),
	}
}

// AddHost adds new host
func (hm *HostManager) AddHost(providerGuid string, host *Host) {
	hm.Mux.Lock()
	hm.Hosts[providerGuid] = host
	hm.Mux.Unlock()
}

// GetHost returns the host with providerGuid
func (hm *HostManager) GetHost(providerGuid string) *Host {
	return hm.Hosts[providerGuid]
}

// GetNumOfHosts returns number of hosts
func (hm *HostManager) GetNumOfHosts() int {
	return len(hm.Hosts)
}

// OnEvent processes each event for updating the model and any type of checking
func (hm *HostManager) OnEvent(event *SysmonEvent) {

}
