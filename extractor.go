package main

import (
	"strings"
)

type Extractorer interface {
	IsSupported(msg *Message) bool
	Transform(msg *Message) error
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

func (ee *ExtractorEngine) Transform(msg *Message) error {
	for _, e := range ee.extractors {
		if e.IsSupported(msg) {
			if err := e.Transform(msg); err != nil {
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

func (e *RegistryExtractor) IsSupported(msg *Message) bool {
	switch msg.Event.EventID {
	case ERegistryEventAdd, ERegistryEventSet, ERegistryEventRename:
		return true
	}
	return false
}

func (e *RegistryExtractor) Transform(msg *Message) error {
	regTarget := msg.Event.get("TargetObject")

	if strings.HasPrefix(regTarget, "HKU\\") {
		tokens := strings.SplitN(regTarget, "\\", 3)
		if len(tokens) >= 3 {
			tokens[1] = tokens[0]
			transformed := strings.Join(tokens[1:], "\\")
			msg.Event.set("TargetObject", transformed)
		}
	}
	return nil
}
