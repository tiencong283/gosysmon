package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
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
	// supported match operations
	OIs        = "is"
	OContains  = "contains"
	OBeginWith = "begin with"
	OEndWith   = "end with"
	OImage     = "image" // "Match an image path (full path or only image name). For example: lsass.exe will match c:\windows\system32\lsass.exe"
)

// rules are chained into a singly-linked list
type Rule struct {
	Next  *Rule
	Op    string
	Value string
	Name  string
}

type FieldFilter struct {
	IncludeRules map[string]*Rule
	ExcludeRules map[string]*Rule
}

type EventFilter map[string]*FieldFilter

// NewEventFilter returns new instance of EventFilter
func NewEventFilter() EventFilter {
	return make(EventFilter)
}

// NewEventFilterFrom returns new instance of EventFilter initialized with ruleFilePath
func NewEventFilterFrom(ruleFilePath string) (EventFilter, error) {
	filter := NewEventFilter()
	if err := filter.UpdateFrom(ruleFilePath); err != nil {
		return nil, err
	}
	return filter, nil
}

// UpdateFrom updates rules from ruleFilePath file
func (filter EventFilter) UpdateFrom(ruleFilePath string) error {
	mitreFilterFile, err := os.Open(ruleFilePath)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(mitreFilterFile)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(data, &filter)
	if err != nil {
		return err
	}
	return nil
}
// Dump prints content of the filter for debugging purposes
func (filter EventFilter) Dump(){
	log.Println("*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*")
	bytes, err := json.MarshalIndent(filter, "", " ")
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("\n%s", string(bytes))
	}
	log.Println("*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*")
}

// onEventFilter is called whenever encounter a event rule
func (filter EventFilter) onEventFilter(ruleName, onMatch string) {
	if _, ok := filter[ruleName]; !ok { // first seen event filter
		filter[ruleName] = new(FieldFilter)
	}
	fieldFilter := filter[ruleName]

	switch onMatch {
	case "include":
		if fieldFilter.IncludeRules == nil {
			fieldFilter.IncludeRules = make(map[string]*Rule)
		}
		break
	case "exclude":
		if fieldFilter.ExcludeRules == nil {
			fieldFilter.ExcludeRules = make(map[string]*Rule)
		}
		break
	}
}

// onFieldFilter updates a rule for a field of an event
func (filter EventFilter) onFieldFilter(ruleName, onMatch, fieldName, label, condition, value string) {
	var fieldFilter map[string]*Rule

	if onMatch == "include" {
		fieldFilter = filter[ruleName].IncludeRules
	} else {
		fieldFilter = filter[ruleName].ExcludeRules
	}
	newRule := &Rule{
		Op:    condition,
		Value: value,
		Name:  label,
	}
	if _, ok := fieldFilter[fieldName]; !ok { // first seen field filter
		fieldFilter[fieldName] = newRule
		return
	}
	// insert at last to preserve order
	head := fieldFilter[fieldName]
	cur := head
	for ; cur.Next != nil; cur = cur.Next { // traverse to last
	}
	cur.Next = newRule
}

// getAttribute returns value of the attribute name of the tag start
func getAttribute(start xml.StartElement, name string) string {
	for _, attr := range start.Attr {
		if strings.EqualFold(attr.Name.Local, name) {
			return attr.Value
		}
	}
	return ""
}

// hasAttribute returns true if the tag has attribute along with the value
func hasAttribute(start xml.StartElement, name, value string) bool {
	for _, attr := range start.Attr {
		if strings.EqualFold(attr.Name.Local, name) && strings.EqualFold(attr.Value, value) {
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
	return string(content)
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
	var ruleName = ""
	var onMatch = ""

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
				if strings.EqualFold(tagName, "RuleGroup") && hasAttribute(element, "groupRelation", "or") { // only support OR
					childOfRuleGroupTag = true
				}
			} else if !childOfEventTag { // then find event rules
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
				onMatch = getAttribute(element, "onMatch")
				switch onMatch {
				case "include", "exclude":
					filter.onEventFilter(ruleName, onMatch)
					childOfEventTag = true
					break
				default:
					if err := d.Skip(); err != nil {
						return err
					}
				}
			} else { // inner of event tag
				label := getAttribute(element, "name")
				cond := getAttribute(element, "condition")
				filter.onFieldFilter(ruleName, onMatch, tagName, label, cond, getContent(d))
			}
			break
		case xml.EndElement:
			element := token.(xml.EndElement)
			tagName := getTagName(element)

			if strings.EqualFold(tagName, ruleName) {
				childOfEventTag = false
				ruleName = ""
			} else if strings.EqualFold(tagName, "RuleGroup") {
				childOfRuleGroupTag = false
			}
		}
	}

	return nil
}

// isMatched returns the rule that matches the event
func (filter EventFilter) isMatched(event SysmonEvent, rules map[string]*Rule) *Rule {
	if rules == nil || len(rules) == 0 {
		return nil
	}

	for propName, rule := range rules { // check each field
		propValue := strings.ToLower(event.EventData[propName])
		for ; rule != nil; rule = rule.Next {
			switch rule.Op {
			case OIs:
				if propValue == rule.Value {
					return rule
				}
				break
			case OContains:
				if strings.Contains(propValue, rule.Value) {
					return rule
				}
				break
			case OBeginWith:
				if strings.HasPrefix(propValue, rule.Value) {
					return rule
				}
				break
			case OEndWith:
				if strings.HasSuffix(propValue, rule.Value) {
					return rule
				}
				break
			case OImage:
				if filepath.IsAbs(rule.Value) {
					if propValue == rule.Value {
						return rule
					}
				} else if filepath.Base(propValue) == rule.Value {
					return rule
				}
				break
			default:
				log.Warnf("Operation %s not supported\n", rule.Op)
			}
		}
	}
	// reach here means nothing matched
	return nil
}

// GetTechName returns the technique name matched with the event
func (filter EventFilter) GetTechName(event SysmonEvent) string {
	ruleName, ok := eventIDToRuleName[event.EventID]
	if !ok { // unsupported event
		return ""
	}
	fieldFilter, ok := filter[ruleName]
	if !ok {
		return ""
	}
	// exclude matches take precedence
	matched := filter.isMatched(event, fieldFilter.ExcludeRules)
	if matched != nil {
		return ""
	}
	matched = filter.isMatched(event, fieldFilter.IncludeRules)
	if matched == nil {
		return ""
	}
	return matched.Name
}

// IsFiltered returns true if the event is excluded
func (filter EventFilter) IsFiltered(event SysmonEvent) bool {
	ruleName, ok := eventIDToRuleName[event.EventID]
	if !ok { // unsupported event
		return false
	}
	fieldFilter, ok := filter[ruleName]
	if !ok {
		return false
	}
	matched := filter.isMatched(event, fieldFilter.ExcludeRules)
	return matched != nil
}