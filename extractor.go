package main

import (
	"strings"
)

type Extractorer interface {
	IsSupported(event *SysmonEvent) bool
	Transform(event *SysmonEvent) error
}

type ExtractorEngine struct {
	extractors []Extractorer
}

func NewExtractorEngine() *ExtractorEngine {
	return &ExtractorEngine{
		extractors: make([]Extractorer, 0),
	}
}

func (ee *ExtractorEngine) Register(extractor Extractorer) {
	ee.extractors = append(ee.extractors, extractor)
}

func (ee *ExtractorEngine) Transform(event *SysmonEvent) error {
	for _, e := range ee.extractors {
		if e.IsSupported(event) {
			if err := e.Transform(event); err != nil {
				return err
			}
		}
	}
	return nil
}

// extractor implementations
type RegistryExtractor struct {
}

func NewRegistryExtractor() *RegistryExtractor {
	return new(RegistryExtractor)
}

func (e *RegistryExtractor) IsSupported(event *SysmonEvent) bool {
	switch event.EventID {
	case ERegistryEventAdd, ERegistryEventSet, ERegistryEventRename:
		return true
	}
	return false
}

func (e *RegistryExtractor) Transform(event *SysmonEvent) error {
	regTarget := event.get("TargetObject")

	if strings.HasPrefix(regTarget, "HKU\\") {
		tokens := strings.SplitN(regTarget, "\\", 3)
		if len(tokens) >= 3 {
			tokens[1] = tokens[0]
			transformed := strings.Join(tokens[1:], "\\")
			event.set("TargetObject", transformed)
		}
	}
	return nil
}
