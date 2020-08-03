package main

import (
	"strings"
)

type Preprocessor interface {
	IsSupported(msg *Message) bool
	Process(msg *Message) error
}

type PreprocessorEngine struct {
	processors []Preprocessor
}

func NewPreprocessorEngine() *PreprocessorEngine {
	return &PreprocessorEngine{
		processors: make([]Preprocessor, 0),
	}
}

func (ee *PreprocessorEngine) Register(processor Preprocessor) {
	ee.processors = append(ee.processors, processor)
}

func (ee *PreprocessorEngine) Transform(msg *Message) error {
	for _, e := range ee.processors {
		if e.IsSupported(msg) {
			if err := e.Process(msg); err != nil {
				return err
			}
		}
	}
	return nil
}

// extractor implementations
type RegistryProcessor struct {
}

func NewRegistryProcessor() *RegistryProcessor {
	return new(RegistryProcessor)
}

func (e *RegistryProcessor) IsSupported(msg *Message) bool {
	switch msg.Event.EventID {
	case ERegistryEventAdd, ERegistryEventSet, ERegistryEventRename:
		return true
	}
	return false
}

func (e *RegistryProcessor) Process(msg *Message) error {
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
