package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

const (
	IOCHash = iota
	IOCIp
	IOCDomain
)

// An Analysis object represents an analysis of a URL or file submitted to VirusTotal
// https://developers.virustotal.com/v3.0/reference#analyses-object
type Analysis struct {
	Data struct {
		Attributes struct {
			LastAnalysisStats struct {
				ConfirmedTimeout int `json:"confirmed-timeout"`
				Failure          int `json:"failure"`
				Harmless         int `json:"harmless"`
				Malicious        int `json:"malicious"`
				Suspicious       int `json:"suspicious"`
				Timeout          int `json:"timeout"`
				TypeUnsupported  int `json:"type-unsupported"`
				Undetected       int `json:"undetected"`
			} `json:"last_analysis_stats"`
		} `json:"attributes"`
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"data"`
}

// IOCFilter is the filter for static indicators including Domains, IPs and Hashes
type IOCFilter struct {
	Client        *http.Client
	VirustotalAPI string
	Expired       bool
	CommonFilterer
}

func newIOCFilter() *IOCFilter {
	return &IOCFilter{
		Client:         new(http.Client),
		VirustotalAPI:  "e96afa0609dbc5a5111cee2039a203c14587e20c66360397280916edd6fc30ce",
		CommonFilterer: NewCommonFilterer("IOC Filter"),
	}
}

func (filter *IOCFilter) IsSupported(event *SysmonEvent) bool {
	switch event.EventID {
	case EProcessCreate: // Hashes, todo: EFileDelete, EDriverLoad, EImageLoad, EFileCreateStreamHash
		return true
	case EDnsQuery: // Domains
		return true
	case ENetworkConnect: // IPs
		return true
	}
	return false
}

func (filter *IOCFilter) Init() error {
	return nil
}

func (filter *IOCFilter) EventCh() chan *SysmonEvent {
	return filter.eventCh
}

func (filter *IOCFilter) StateCh() chan int {
	return filter.State
}

func (filter *IOCFilter) SetAlertCh(alertCh chan *RContext) {
	filter.AlertCh = alertCh
}

func (filter *IOCFilter) CheckIOC(indicator string, iocType int) bool {
	// build the endpoint url
	urlFormat := "http://www.virustotal.com/api/v3/%s/%s"
	var url string
	switch iocType {
	case IOCHash:
		url = fmt.Sprintf(urlFormat, "files", indicator)
	case IOCDomain:
		url = fmt.Sprintf(urlFormat, "domains", indicator)
	case IOCIp:
		url = fmt.Sprintf(urlFormat, "ip_addresses", indicator)
	default:
		return false
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("x-apikey", filter.VirustotalAPI)
	// querying
	resp, err := filter.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		analysis := new(Analysis)
		if err := json.Unmarshal(bytes, analysis); err != nil {
			log.Fatal(err)
		}
		if analysis.Data.Attributes.LastAnalysisStats.Malicious > 0 {
			return true
		}
	} else {
		log.Println(resp.Status, url)
	}
	return false
}

func (filter *IOCFilter) Start() {
	var indicator string
	var iocType int

	for event := range filter.eventCh {
		switch event.EventID {
		case EProcessCreate, EFileDelete:
			if event.EventID == EProcessCreate && strings.HasPrefix(event.get("Image"), "C:\\Windows\\System32\\") {
				continue
			}
			if event.EventID == EFileDelete && !event.getBool("IsExecutable") {
				continue
			}
			hashes := StringToMap(event.get("Hashes"))
			indicator, iocType = hashes["MD5"], IOCHash
		case EDnsQuery:
			queryStatus, err := event.getInt("QueryStatus")
			if err != nil || queryStatus != 0 {
				continue
			}
			indicator, iocType = event.get("QueryName"), IOCDomain
		case ENetworkConnect:
			if event.getBool("DestinationIsIpv6") {
				continue
			}
			targetIP := net.ParseIP(event.get("DestinationIp"))
			if targetIP == nil || !IsPublicGlobalUnicast(targetIP) {
				continue
			}
			indicator, iocType = targetIP.String(), IOCIp
		}
		if malicious := filter.CheckIOC(indicator, iocType); malicious {
			log.Printf("malicious '%s'\n", indicator)
		} else {
			log.Printf("not malicious '%s'\n", indicator)
		}
	}

	filter.State <- 1
}