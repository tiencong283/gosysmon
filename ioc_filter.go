package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type IOCResult struct {
	ResultId
	Timestamp   *time.Time
	IOCType     int
	Indicator   string
	Message     string
	ExternalUrl string
}

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

func NewIOCFilter() *IOCFilter {
	return &IOCFilter{
		Client:         new(http.Client),
		VirustotalAPI:  viper.GetString("virustotal-api"),
		CommonFilterer: NewCommonFilterer("IOC Filter"),
	}
}

func (filter *IOCFilter) IsSupported(msg *Message) bool {
	switch msg.Event.EventID {
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
	filter.logger.Infoln("initializing")
	return nil
}

func (filter *IOCFilter) MessageCh() chan *Message {
	return filter.messageCh
}

func (filter *IOCFilter) StateCh() chan int {
	return filter.State
}

func (filter *IOCFilter) SetAlertCh(alertCh chan interface{}) {
	filter.AlertCh = alertCh
}

func (filter *IOCFilter) CheckIOC(indicator string, iocType int) (bool, error) {
	// check in cache first
	var key string
	switch iocType {
	case IOCHash:
		key = "ioc:hash:" + indicator
	case IOCDomain:
		key = "ioc:domain:" + indicator
	case IOCIp:
		key = "ioc:ip_address:" + indicator
	}
	ok, err := redis.Bool(RedisConn.Do("GET", key))
	if err != redis.ErrNil {
		if err != nil {
			return false, err
		}
		return ok, nil
	}

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
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("x-apikey", filter.VirustotalAPI)

	// querying
	resp, err := filter.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		analysis := new(Analysis)
		if err := json.Unmarshal(bytes, analysis); err != nil {
			return false, nil
		}
		ans := analysis.Data.Attributes.LastAnalysisStats.Malicious > 1
		if err = RedisConn.Send("SET", key, ans); err != nil { // cached in redis
			return false, err
		}
		_ = RedisConn.Flush()
		return ans, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	}
	return false, fmt.Errorf("virustotal response code %d", resp.StatusCode)
}

func (filter *IOCFilter) filterQueryName(dnsName string) bool {
	if strings.IndexByte(dnsName, '.') < 0 {
		return true
	}
	// todo: filter special and reserved dns names
	return false
}

func (filter *IOCFilter) Start() {
	var indicator string
	var iocType int

	for msg := range filter.messageCh {
		event := msg.Event
		switch event.EventID {
		case EProcessCreate, EFileDelete:
			if event.EventID == EProcessCreate && strings.HasPrefix(event.get("Image"), "C:\\Windows\\System32\\") {
				continue
			}
			if event.EventID == EFileDelete && !event.getBool("IsExecutable") {
				continue
			}
			indicator, iocType = GetKeyFrom(event.get("Hashes"), "MD5"), IOCHash
		case EDnsQuery:
			queryStatus, err := event.getInt("QueryStatus")
			if err != nil || queryStatus != 0 {
				continue
			}
			queryName := event.get("QueryName")
			if filter.filterQueryName(queryName) {
				continue
			}
			indicator, iocType = queryName, IOCDomain
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
		malicious, err := filter.CheckIOC(indicator, iocType)
		if err != nil {
			log.Warnf("cannot check %s: %v", indicator, err)
			continue
		}
		if malicious {
			var alertMsg, externalUrl string

			switch iocType {
			case IOCIp:
				alertMsg = fmt.Sprintf("An connection made to the malicious IP '%s'", indicator)
				externalUrl = fmt.Sprintf("https://www.virustotal.com/gui/ip-address/%s/detection", indicator)
			case IOCDomain:
				alertMsg = fmt.Sprintf("An DNS query to malicious domain '%s'", indicator)
				externalUrl = fmt.Sprintf("https://www.virustotal.com/gui/domain/%s/detection", indicator)
			case IOCHash:
				if event.EventID == EProcessCreate {
					alertMsg = fmt.Sprintf("Malicious process '%s' is executed", GetImageName(event.get("Image")))
				}
				externalUrl = fmt.Sprintf("https://www.virustotal.com/gui/file/%s/detection", indicator)
			}
			report := &IOCResult{
				Timestamp:   event.timestamp(),
				IOCType:     iocType,
				Indicator:   indicator,
				Message:     alertMsg,
				ExternalUrl: externalUrl,
				ResultId:    NewResultId(msg),
			}
			filter.AlertCh <- report
		}
	}
	filter.State <- 1
}
