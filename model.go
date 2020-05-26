package main

import "time"

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
type HostManager map[string]*Host

// NewHostManager returns new instance of HostManager
func NewHostManager() HostManager {
	return make(HostManager)
}

// AddHost adds new host
func (hosts HostManager) AddHost(providerGuid string, host *Host) {
	hosts[providerGuid] = host
}

// GetHost returns the host with providerGuid
func (hosts HostManager) GetHost(providerGuid string) *Host {
	return hosts[providerGuid]
}

// GetNumOfHosts returns number of hosts
func (hosts HostManager) GetNumOfHosts() int {
	return len(hosts)
}

// OnEvent processes each event for updating the model and any type of checking
func (hosts HostManager) OnEvent(event *SysmonEvent) {
	if event.isProcessEvent() { // process-related events

	}
}
