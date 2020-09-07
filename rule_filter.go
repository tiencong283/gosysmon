package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	RuleDirPath       = "rules"
	SchemaDefFilePath = "sensor/sysmon_config_schema_definition.xml"
	// supported match operations (case insensitive by default)
	OIs          = "is"
	OIsNot       = "is not"
	OContains    = "contains"
	OContainsAny = "contains any"
	OContainsAll = "contains all"
	OExcludes    = "excludes"
	OExcludesAny = "excludes any"
	OExcludesAll = "excludes all"
	OBeginWith   = "begin with"
	OEndWith     = "end with"
	OImage       = "image"     // "Match an image path (full path or only image name). For example: lsass.exe will match c:\windows\system32\lsass.exe"
	ODir         = "directory" // use the all but last element of path, e.g the path's directory
)

type SchemaDef struct {
	Version           float64
	EventIDToRuleName map[int]string
	ValidProps        map[string]StringSet
	RuleNames         StringSet
}

// Rule is the individual rule for an attribute of an event
// For example: it represents <TargetObject condition="begin with">HKU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run</TargetObject>
type Rule struct {
	Name    string
	Cond    string
	Value   string
	CaseSen bool
}

// RuleGroup is a group of individual rules combined by Rel which either OR or AND.
// For example: it represents <Rule name="technique_id=T1060" groupRelation="or">...</Rule>
type RuleGroup struct {
	Label string
	Rel   string
	Rules []*Rule
}

// CombinedRule represents how RuleGroups are combined with each other which either OR or AND.
// For example: it represents  <RuleGroup name="" groupRelation="or">...</RuleGroup>
type CombinedRule struct {
	Next       *CombinedRule
	Label      string
	Rel        string
	RuleGroups []*RuleGroup
}

// RuleFilter contains filtering rules for some events. Two types of rules for each event are inclusion and exclusion:
// 'include' means only matched events are included while with 'exclude', the event will be included except if a rule match
type RuleFilter struct {
	SchemaDef *SchemaDef
	Filters   [2]map[string]*CombinedRule // index 0 for include and 1 for exclude
	CommonFilterer
}

func NewSchemaDef() *SchemaDef {
	return &SchemaDef{
		EventIDToRuleName: make(map[int]string),
		ValidProps:        make(map[string]StringSet),
		RuleNames:         NewStringSet(),
	}
}

func NewRuleFilter() *RuleFilter {
	return &RuleFilter{
		SchemaDef:      NewSchemaDef(),
		Filters:        [2]map[string]*CombinedRule{},
		CommonFilterer: NewCommonFilterer("Rule-based Filter"),
	}
}

func (mf *RuleFilter) IsSupported(msg *Message) bool {
	return msg.Event.isProcessEvent()
}

// Init loads rule's schema definition and rules
func (mf *RuleFilter) Init() error {
	mf.logger.Infoln("initializing")
	schemaDef := mf.SchemaDef

	// Load schema definition
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(SchemaDefFilePath); err != nil {
		return err
	}
	root := doc.Root()
	verAttr := root.SelectAttr("schemaversion")
	if root.Tag != "manifest" || verAttr == nil {
		return fmt.Errorf("manifest tag must be the root in configuration '%s'", SchemaDefFilePath)
	}
	schemaDef.Version, _ = strconv.ParseFloat(verAttr.Value, 64)
	events := root.SelectElement("events")
	if events == nil {
		return fmt.Errorf("failed to find events tag in configuration '%s'", SchemaDefFilePath)
	}
	for _, event := range events.ChildElements() {
		if event.Tag != "event" {
			return fmt.Errorf("element '%s' is unexpected in parent element 'events' in configuration '%s'", event.Tag, SchemaDefFilePath)
		}
		eventID, _ := strconv.Atoi(event.SelectAttrValue("value", ""))
		ruleName := event.SelectAttrValue("rulename", "")
		if eventID <= 0 || len(ruleName) == 0 {
			continue
		}
		validProps := NewStringSet()
		schemaDef.EventIDToRuleName[eventID] = ruleName
		schemaDef.ValidProps[ruleName] = validProps
		schemaDef.RuleNames.Add(ruleName)

		for _, data := range event.ChildElements() {
			if data.Tag != "data" {
				return fmt.Errorf("element '%s' is unexpected in parent element 'event' in configuration '%s'", event.Tag, SchemaDefFilePath)
			}
			propAttr := data.SelectAttr("name")
			if propAttr != nil {
				validProps.Add(propAttr.Value)
			}
		}
	}

	// Under each event tag, it's allowed to contain Rule tag
	for ruleName := range schemaDef.ValidProps {
		schemaDef.ValidProps[ruleName].Add("Rule")
	}

	schemaDef.ValidProps["Sysmon"] = NewStringSet()
	schemaDef.ValidProps["Sysmon"].Add("EventFiltering")
	schemaDef.ValidProps["EventFiltering"] = NewStringSet()
	schemaDef.ValidProps["EventFiltering"].Add("RuleGroup")
	schemaDef.ValidProps["EventFiltering"].AddFromSet(schemaDef.RuleNames)
	schemaDef.ValidProps["RuleGroup"] = NewStringSet()
	schemaDef.ValidProps["RuleGroup"].AddFromSet(schemaDef.RuleNames)

	if err := mf.LoadFromDir(RuleDirPath); err != nil {
		return err
	}
	return nil
}

func (mf *RuleFilter) MessageCh() chan *Message {
	return mf.messageCh
}

func (mf *RuleFilter) StateCh() chan int {
	return mf.State
}

func (mf *RuleFilter) SetAlertCh(alertCh chan interface{}) {
	mf.AlertCh = alertCh
}

func (mf *RuleFilter) Start() {
	for msg := range mf.messageCh {
		event := msg.Event
		if labels := mf.GetLabels(event); len(labels) > 0 {
			isAlert := true
			if labels["is_alert"] != "" {
				isAlert, _ = strconv.ParseBool(labels["is_alert"])
			}
			alert := NewMitreATTCKResult(isAlert, labels["technique_id"], "", msg, true)
			alert.MergeContext(event.EventData)
			alert.AddContext("RuleName", mf.SchemaDef.EventIDToRuleName[event.EventID])
			mf.AlertCh <- alert
		}
	}
	mf.State <- 1
}

// LoadFromDir updates rules from ruleDirPath directory
func (mf *RuleFilter) LoadFromDir(ruleDirPath string) error {
	return filepath.Walk(ruleDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".xml" {
			return nil
		}
		mf.logger.Infof("parsing rules from %s\n", path)
		return mf.UpdateFrom(path)
	})
}

// UpdateFrom updates rules in ruleFilePath file
func (mf *RuleFilter) UpdateFrom(ruleFilePath string) error {
	mitreFilterFile, err := os.Open(ruleFilePath)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(mitreFilterFile)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(data, mf)
	if err != nil {
		return err
	}
	return nil
}

// isMatched deals with each rule
func (rule *Rule) isMatched(event *SysmonEvent) bool {
	propValue := event.get(rule.Name)
	if !rule.CaseSen {
		propValue = strings.ToLower(propValue)
	}
	switch rule.Cond {
	case OIs:
		return propValue == rule.Value
	case OIsNot:
		return propValue != rule.Value
	case OContains:
		return strings.Contains(propValue, rule.Value)
	case OContainsAny:
		tokens := strings.Split(rule.Value, ";")
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if strings.Contains(propValue, token) {
				return true
			}
		}
		return false
	case OContainsAll:
		tokens := strings.Split(rule.Value, ";")
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if !strings.Contains(propValue, token) {
				return false
			}
		}
		return true
	case OExcludes:
		return !strings.Contains(propValue, rule.Value)
	case OExcludesAny:
		tokens := strings.Split(rule.Value, ";")
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if !strings.Contains(propValue, token) {
				return true
			}
		}
		return false
	case OExcludesAll:
		tokens := strings.Split(rule.Value, ";")
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if strings.Contains(propValue, token) {
				return false
			}
		}
		return true
	case OBeginWith:
		return strings.HasPrefix(propValue, rule.Value)
	case OEndWith:
		return strings.HasSuffix(propValue, rule.Value)
	case OImage:
		if WindowsIsAbs(rule.Value) {
			return propValue == rule.Value
		}
		return GetImageName(propValue) == rule.Value
	case ODir:
		return GetDir(propValue) == rule.Value
	}
	// should not reach here
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

// addRule adds new rule to each rule group
func (rg *RuleGroup) addRule(name, cond, value, caseSen string) error {
	isCaseSen, err := strconv.ParseBool(caseSen)
	if err != nil {
		return fmt.Errorf("invalid case attribute value '%s'", caseSen)
	}
	value = strings.TrimSpace(value)
	cond = strings.TrimSpace(cond)
	if !isCaseSen {
		value = strings.ToLower(value)
	}
	switch cond {
	case OIs, OIsNot, OContains, OBeginWith, OEndWith, OImage, ODir:
	default:
		return fmt.Errorf("invalid condition attribute value '%s'", cond)
	}
	if len(value) == 0 {
		return errors.New("the value for a rule cannot be empty")
	}
	rg.Rules = append(rg.Rules, &Rule{
		Name:    strings.TrimSpace(name),
		Cond:    cond,
		Value:   value,
		CaseSen: isCaseSen,
	})
	return nil
}

// onEventFilter is called whenever encountering RuleGroup/Event tags
func (mf *RuleFilter) onEventFilter(groupRelation, ruleName, onMatch, label string) (*CombinedRule, error) {
	if groupRelation != "or" && groupRelation != "and" {
		return nil, fmt.Errorf("invalid groupRelation attribute value '%s'", groupRelation)
	}
	if onMatch != "include" && onMatch != "exclude" {
		return nil, fmt.Errorf("invalid onMatch attribute value '%s'", groupRelation)
	}
	filterType := 0
	if onMatch == "exclude" {
		filterType = 1
	}
	filter := mf.Filters[filterType]
	if filter == nil { // first seen filter type
		filter = make(map[string]*CombinedRule)
		mf.Filters[filterType] = filter
	}
	newCombinedRule := &CombinedRule{
		Label:      strings.TrimSpace(label),
		Rel:        groupRelation,
		RuleGroups: make([]*RuleGroup, 0, 4),
	}
	if _, ok := filter[ruleName]; !ok { // first seen event type
		filter[ruleName] = newCombinedRule
		return newCombinedRule, nil
	}
	// insert at tail to preserve order
	cur := filter[ruleName]
	for ; cur.Next != nil; cur = cur.Next {
	}
	cur.Next = newCombinedRule
	return newCombinedRule, nil
}

// onRuleFilter is called whenever encountering Rule/Property tag
func (mf *RuleFilter) onRuleFilter(combinedRule *CombinedRule, label, groupRelation string) (*RuleGroup, error) {
	if combinedRule == nil {
		return nil, errors.New("invalid parameters in onRuleFilter")
	}
	if groupRelation != "or" && groupRelation != "and" {
		return nil, fmt.Errorf("invalid groupRelation attribute value '%s'", groupRelation)
	}
	newRg := &RuleGroup{
		Label: label,
		Rel:   groupRelation,
		Rules: make([]*Rule, 0, 4),
	}
	combinedRule.RuleGroups = append(combinedRule.RuleGroups, newRg)
	return newRg, nil
}

// getElementName returns the name of an element
func getElementName(element xml.Token) string {
	switch element := element.(type) {
	case xml.StartElement:
		return element.Name.Local
	case xml.EndElement:
		return element.Name.Local
	case xml.Attr:
		return element.Name.Local
	}
	return ""
}

// getAttribute returns value of the attribute name of the tag element
func getAttribute(element xml.StartElement, name string) (string, error) {
	for _, attr := range element.Attr {
		if getElementName(attr) == name {
			return strings.TrimSpace(attr.Value), nil
		}
	}
	return "", fmt.Errorf("failed to find attribute '%s' of tag %s", name, getElementName(element))
}

// getAttributeOr returns value of the attribute name of the tag element, if not found return the def value
func getAttributeOr(element xml.StartElement, name, def string) string {
	val, err := getAttribute(element, name)
	if err != nil {
		return def
	}
	return val
}

// hasAttVal returns true if the tag has attribute along with the value (value comparision is case-insensitive)
func hasAttVal(element xml.StartElement, name, value string) bool {
	for _, attr := range element.Attr {
		if getElementName(attr) == name && strings.EqualFold(attr.Value, value) {
			return true
		}
	}
	return false
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
func (mf *RuleFilter) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// check the signature of Sysmon config file
	if getElementName(start) == "Sysmon" {
		val, err := getAttribute(start, "schemaversion")
		if err != nil {
			return err
		}
		ver, err := strconv.ParseFloat(val, 64)
		if err != nil || ver != mf.SchemaDef.Version {
			return fmt.Errorf("incorrect or unsupported schema version")
		}
	} else {
		return fmt.Errorf("failed to find Sysmon tag in configuration")
	}

	tokStk := make([]string, 0, 64)
	var rgRel, rgLabel string
	var cr *CombinedRule
	var rg *RuleGroup

	tokStk = append(tokStk, getElementName(start))
	for { // traverse over nodes (preorder depth-first traversal)
		token, err := d.Token()
		if err != nil && err == io.EOF { // at the end of the input stream
			break
		}
		if token == nil {
			return err
		}
		elementName := getElementName(token)
		lastElementName := tokStk[len(tokStk)-1]
		switch token := token.(type) {
		case xml.StartElement:
			element := token
			// check if it's the valid relationship (parent-child)
			parent := lastElementName
			if parent == "Rule" {
				parent = tokStk[len(tokStk)-2] // upwards one level
			}
			if !mf.SchemaDef.ValidProps[parent].Has(elementName) {
				return fmt.Errorf("element %s is unexpected in parent element %s", elementName, parent)
			}
			tokStk = append(tokStk, elementName)

			switch lastElementName {
			case "Sysmon":
			case "EventFiltering":
				rgLabel = getAttributeOr(element, "name", "")
				if elementName == "RuleGroup" { // RuleGroup wraps Event tags
					rel, err := getAttribute(element, "groupRelation")
					if err != nil {
						return err
					}
					rgRel = rel
				} else { // Event tags directly
					onMatch, err := getAttribute(element, "onmatch")
					if err != nil {
						return err
					}
					cr, err = mf.onEventFilter("or", elementName, onMatch, rgLabel) // default is OR
					if err != nil {
						return err
					}
				}
			case "RuleGroup": // RuleGroup wraps Event tags
				onMatch, err := getAttribute(element, "onmatch")
				if err != nil {
					return err
				}
				cr, err = mf.onEventFilter(rgRel, elementName, onMatch, rgLabel)
				if err != nil {
					return err
				}
			case "Rule": // individual rules
				if rg != nil {
					cond, err := getAttribute(element, "condition")
					if err != nil {
						return err
					}
					if err := rg.addRule(elementName, cond, getContent(d), getAttributeOr(element, "case", "false")); err != nil {
						return err
					}
				}
			default:
				if mf.SchemaDef.RuleNames.Has(lastElementName) { // inner of Event tag
					rel := "or"                // default is OR
					if elementName == "Rule" { // Rule wraps individual rules
						if rel, err = getAttribute(element, "groupRelation"); err != nil {
							return err
						}
					}

					rg, err = mf.onRuleFilter(cr, getAttributeOr(element, "name", ""), rel)
					if err != nil {
						return err
					}

					if rg != nil && elementName != "Rule" {
						cond, err := getAttribute(element, "condition")
						if err != nil {
							return err
						}
						if err := rg.addRule(elementName, cond, getContent(d), getAttributeOr(element, "case", "false")); err != nil {
							return err
						}
					}
				}
			}
		case xml.EndElement:
			if lastElementName != elementName {
				return fmt.Errorf("unmatched opening %s and closing tag %s", lastElementName, elementName)
			}
			tokStk = tokStk[:len(tokStk)-1]
			switch elementName {
			case "Rule":
				rg = nil
			case "RuleGroup":
				cr = nil
			default:
				if mf.SchemaDef.RuleNames.Has(elementName) {
					cr = nil
				}
			}
		}
	}
	if len(tokStk) > 0 {
		return errors.New("unbalanced elements")
	}
	return nil
}

// isMatched returns the rule that matches the event
func (mf *RuleFilter) isMatched(event *SysmonEvent, ruleName string, filter map[string]*CombinedRule) (bool, string) {
	if filter == nil || filter[ruleName] == nil {
		return false, ""
	}
	for cr := filter[ruleName]; cr != nil; cr = cr.Next { // traverse over each CombinedRule
		switch cr.Rel {
		case "or":
			for _, rg := range cr.RuleGroups {
				if rg.isMatched(event) {
					return true, rg.Label
				}
			}
		case "and":
			matched := true
			for _, rg := range cr.RuleGroups {
				if !rg.isMatched(event) {
					matched = false
					break
				}
			}
			if matched {
				return true, cr.Label
			}
		}
	}
	return false, ""
}

// GetLabels returns the labels of matched rule
func (mf *RuleFilter) GetLabels(event *SysmonEvent) map[string]string {
	ruleName, ok := mf.SchemaDef.EventIDToRuleName[event.EventID]
	if !ok { // unsupported event
		return nil
	}
	// exclude matches take precedence
	matched, _ := mf.isMatched(event, ruleName, mf.Filters[1])
	if matched {
		return nil
	}
	matched, label := mf.isMatched(event, ruleName, mf.Filters[0])
	if !matched {
		return nil
	}
	return StringToMap(label)
}
