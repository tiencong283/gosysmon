package main

import (
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// mapping from eventID to its rule name
var eventIDToRuleName = map[int]string{
	EProcessCreate:        "ProcessCreate",
	EFileCreateTime:       "FileCreateTime",
	ENetworkConnect:       "NetworkConnect",
	EProcessTerminate:     "ProcessTerminate",
	EDriverLoad:           "DriverLoad",
	EImageLoad:            "ImageLoad",
	ECreateRemoteThread:   "CreateRemoteThread",
	ERawAccessRead:        "RawAccessRead",
	EProcessAccess:        "ProcessAccess",
	EFileCreate:           "FileCreate",
	ERegistryEventAdd:     "RegistryEvent",
	ERegistryEventSet:     "RegistryEvent",
	ERegistryEventRename:  "RegistryEvent",
	EFileCreateStreamHash: "FileCreateStreamHash",
	EPipeEventCreate:      "PipeEvent",
	EPipeEventConnect:     "PipeEvent",
	EWmiEventFilter:       "WmiEvent",
	EWmiEventConsumer:     "WmiEvent",
	EWmiEventBinding:      "WmiEvent",
	EDnsQuery:             "DnsQuery",
	EFileDelete:           "FileDelete",
}

const (
	// the maximum number of rules per group
	MaxRulesPerGroup = 4
	// supported match operations (case insensitive)
	OIs        = "is"
	OContains  = "contains"
	OBeginWith = "begin with"
	OEndWith   = "end with"
	OImage     = "image" // "Match an image path (full path or only image name). For example: lsass.exe will match c:\windows\system32\lsass.exe"
)

// Rule is the individual rule for an attribute of an event
type Rule struct {
	Name    string
	Cond    string
	Value   string
	CaseSen bool
}

// RuleGroup is a group of individual rules combined by Rel which either OR or AND.
type RuleGroup struct {
	Next  *RuleGroup
	Label string
	Rel   string
	Rules []*Rule
}

// EventFilter contains filtering rules for all events. Two types of rules for each event are inclusion and exclusion:
// 'include' means only matched events are included while with 'exclude', the event will be included except if a rule match
type EventFilter struct {
	IncludeFilter map[string]*RuleGroup
	ExcludeFilter map[string]*RuleGroup
}

// isMatched deals with each rule
func (rule *Rule) isMatched(event *SysmonEvent) bool {
	propValue := event.EventData[rule.Name]
	if !rule.CaseSen {
		propValue = strings.ToLower(propValue)
	}
	switch rule.Cond {
	case OIs:
		if propValue == rule.Value {
			return true
		}
	case OContains:
		if strings.Contains(propValue, rule.Value) {
			return true
		}
	case OBeginWith:
		if strings.HasPrefix(propValue, rule.Value) {
			return true
		}
	case OEndWith:
		if strings.HasSuffix(propValue, rule.Value) {
			return true
		}
	case OImage:
		if filepath.IsAbs(rule.Value) {
			if propValue == rule.Value {
				return true
			}
		} else if filepath.Base(propValue) == rule.Value {
			return true
		}
		log.Warnf("Operation %s not supported\n", rule.Cond)
	}
	return false
}

// isMatched deals with each rule group
func (rg *RuleGroup) isMatched(event *SysmonEvent) bool {
	switch rg.Rel {
	case "or":
		for _, rule := range rg.Rules {
			if rule.isMatched(event) {
				return true
			}
		}
		return false
	case "and":
		for _, rule := range rg.Rules {
			if !rule.isMatched(event) {
				return false
			}
		}
		return true
	}
	return false
}

// NewEventFilter returns new instance of EventFilter
func NewEventFilter() *EventFilter {
	return new(EventFilter)
}

// NewEventFilterFrom returns new instance of EventFilter initialized with ruleFilePath
func NewEventFilterFrom(ruleFilePath string) (*EventFilter, error) {
	filter := NewEventFilter()
	if err := filter.UpdateFrom(ruleFilePath); err != nil {
		return nil, err
	}
	return filter, nil
}

// UpdateFrom updates rules from ruleDirPath directory
func (filter *EventFilter) UpdateFromDir(ruleDirPath string) error {
	return filepath.Walk(ruleDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".xml" {
			return nil
		}
		return filter.UpdateFrom(path)
	})
}

// UpdateFrom updates rules from ruleFilePath file
func (filter *EventFilter) UpdateFrom(ruleFilePath string) error {
	mitreFilterFile, err := os.Open(ruleFilePath)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(mitreFilterFile)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(data, filter)
	if err != nil {
		return err
	}
	return nil
}

// onEventFilter is called whenever encounter a event rule
func (filter *EventFilter) onEventFilter(ruleName, onMatch string) {
	switch onMatch {
	case "include":
		if filter.IncludeFilter == nil {
			filter.IncludeFilter = make(map[string]*RuleGroup)
		}
	case "exclude":
		if filter.ExcludeFilter == nil {
			filter.ExcludeFilter = make(map[string]*RuleGroup)
		}
	}
}

// onGroupFilter _
func (filter *EventFilter) onGroupFilter(ruleName, onMatch, label, rel string) *RuleGroup {
	rel = strings.ToLower(rel)
	if rel != "or" && rel != "and" {
		return nil
	}
	switch onMatch {
	case "include":
		head := filter.IncludeFilter[ruleName]
		newNode := &RuleGroup{
			Label: label,
			Rel:   rel,
			Rules: make([]*Rule, 0, MaxRulesPerGroup),
		}
		if head == nil { // insert at head
			filter.IncludeFilter[ruleName] = newNode
			return newNode
		}
		// insert at tail to preserve order
		cur := head
		for ; cur.Next != nil; cur = cur.Next { // traverse to the last node
		}
		cur.Next = newNode
		return newNode
	case "exclude":
		head := filter.ExcludeFilter[ruleName]
		newNode := &RuleGroup{
			Label: label,
			Rel:   rel,
			Rules: make([]*Rule, 0, MaxRulesPerGroup),
		}
		if head == nil { // insert at head
			filter.ExcludeFilter[ruleName] = newNode
			return newNode
		}
		// insert at tail to preserve order
		cur := head
		for ; cur.Next != nil; cur = cur.Next { // traverse to the last node
		}
		cur.Next = newNode
		return newNode
	}
	return nil
}

// getAttribute returns value of the attribute name of the tag start
func getAttribute(start xml.StartElement, name string) string {
	for _, attr := range start.Attr {
		if attr.Name.Local == name {
			return strings.TrimSpace(attr.Value)
		}
	}
	return ""
}

// hasAttribute returns true if the tag has attribute along with the value (value comparision is case-insensitive)
func hasAttribute(start xml.StartElement, name, value string) bool {
	for _, attr := range start.Attr {
		if attr.Name.Local == name && strings.EqualFold(attr.Value, value) {
			return true
		}
	}
	return false
}

// getTagName returns the tag name
func getTagName(element xml.Token) string {
	switch element.(type) {
	case xml.StartElement:
		return element.(xml.StartElement).Name.Local
	case xml.EndElement:
		return element.(xml.EndElement).Name.Local
	}
	return ""
}

// getContent return the content of tag
func getContent(d *xml.Decoder) string {
	token, err := d.Token()
	if err != nil {
		return ""
	}
	content, ok := token.(xml.CharData)
	if !ok {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// UnmarshalXML decodes the config in xml format and updates the filter rules
func (filter *EventFilter) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// check signature of sysmon config file
	if strings.EqualFold(start.Name.Local, "sysmon") {
		version, err := strconv.ParseFloat(getAttribute(start, "schemaversion"), 64)
		if err != nil {
			return fmt.Errorf("Invalid schemaversion version in xml file")
		}
		if version > 4.30 {
			return fmt.Errorf("Unsupported schemaversion version in xml file")
		}
	} else {
		return fmt.Errorf("Unsupported schemaversion version in xml file")
	}

	var childOfRuleGroupTag = false
	var childOfEventTag = false
	var childOfRuleGroup = false
	var ruleName = ""
	var onMatch = ""
	var ruleGroup *RuleGroup

	for { // traverse over nodes (preorder depth-first traversal)
		token, err := d.Token()
		if err != nil {
			if err == io.EOF { // end of tree
				break
			}
			return nil
		}
		switch token.(type) {
		case xml.StartElement:
			element := token.(xml.StartElement)
			tagName := getTagName(element)
			if !childOfRuleGroupTag { // find RuleGroup tag first
				if tagName == "RuleGroup" && hasAttribute(element, "groupRelation", "or") { // only support OR
					childOfRuleGroupTag = true
				}
			} else if !childOfEventTag { // then find event tag
				ok := false
				for _, ruleName := range eventIDToRuleName {
					if ruleName == tagName {
						ok = true
						break
					}
				}
				if !ok { // if the event cannot be filtered the skip that tag
					if err := d.Skip(); err != nil { // cannot jump to next tag
						return err
					}
				}
				ruleName = tagName
				onMatch = getAttribute(element, "onmatch")
				switch onMatch {
				case "include", "exclude":
					filter.onEventFilter(ruleName, onMatch)
					childOfEventTag = true
				default:
					if err := d.Skip(); err != nil {
						return err
					}
				}
			} else if !childOfRuleGroup { // inner of event tag
				label := getAttribute(element, "name")
				rel := "or" // default is OR
				if tagName == "Rule" {
					rel = getAttribute(element, "groupRelation")
					childOfRuleGroup = true
				}
				ruleGroup = filter.onGroupFilter(ruleName, onMatch, label, rel)
				if ruleGroup != nil && tagName != "Rule" {
					ruleGroup.Rules = append(ruleGroup.Rules, &Rule{
						Name:  tagName,
						Cond:  getAttribute(element, "condition"),
						Value: getContent(d),
					})
				}
			} else { // inner of Rule tag
				if ruleGroup == nil {
					break
				}
				caseSen := getAttribute(element, "case") == "true"
				content := getContent(d)
				if !caseSen { // speed up performance for comparing
					content = strings.ToLower(content)
				}
				ruleGroup.Rules = append(ruleGroup.Rules, &Rule{
					Name:    tagName,
					Cond:    getAttribute(element, "condition"),
					Value:   content,
					CaseSen: caseSen,
				})
			}
		case xml.EndElement:
			element := token.(xml.EndElement)
			tagName := getTagName(element)
			if tagName == ruleName {
				childOfEventTag = false
				ruleName = ""
			} else if tagName == "RuleGroup" {
				childOfRuleGroupTag = false
			} else if tagName == "Rule" {
				childOfRuleGroup = false
				ruleGroup = nil
			}
		}
	}
	return nil
}

// isMatched returns the rule that matches the event
func (filter *EventFilter) isMatched(event *SysmonEvent, ruleName string, rules map[string]*RuleGroup) *RuleGroup {
	if rules == nil || rules[ruleName] == nil {
		return nil
	}
	for rg := rules[ruleName]; rg != nil; rg = rg.Next { // traverse over each rule group
		if rg.isMatched(event) {
			return rg
		}
	}
	// reach here means nothing matched
	return nil
}

// GetTechName returns the technique name matched with the event
func (filter *EventFilter) GetTechName(event *SysmonEvent) string {
	ruleName, ok := eventIDToRuleName[event.EventID]
	if !ok { // unsupported event
		return ""
	}
	// exclude matches take precedence
	matched := filter.isMatched(event, ruleName, filter.ExcludeFilter)
	if matched != nil {
		return ""
	}

	matched = filter.isMatched(event, ruleName, filter.IncludeFilter)
	if matched == nil {
		return ""
	}
	return matched.Label
}
