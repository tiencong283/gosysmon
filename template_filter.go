package main

// TemplateFilter _
type TemplateFilter struct {
	CommonFilterer
}

func NewTemplateFilter() *IOCFilter {
	return &IOCFilter{
		CommonFilterer: NewCommonFilterer("Template Filter"),
	}
}

func (filter *TemplateFilter) IsSupported(event *SysmonEvent) bool {
	return false
}

func (filter *TemplateFilter) Init() error {
	return nil
}

func (filter *TemplateFilter) EventCh() chan *Message {
	return filter.messageCh
}

func (filter *TemplateFilter) StateCh() chan int {
	return filter.State
}

func (filter *TemplateFilter) SetAlertCh(alertCh chan interface{}) {
	filter.AlertCh = alertCh
}

func (filter *TemplateFilter) Start() {
	for _ = range filter.messageCh {
		// process events here
	}
	filter.State <- 1
}
