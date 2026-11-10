package services

import (
	"sync"

	"email-summary-tool/services/summary"
)

// DependencyService groups the shared dependencies that handlers and services
// use, apart from the database and JWT config, which stay on the server.
type DependencyService struct {
	// InboundDomain is the domain used for generated project email addresses.
	InboundDomain string

	// Summarizer writes the AI summaries. Tests replace it with a fake.
	Summarizer summary.Summarizer

	Wg *sync.WaitGroup
}

func NewDependencyService(inboundDomain string, summarizer summary.Summarizer) *DependencyService {
	return &DependencyService{
		InboundDomain: inboundDomain,
		Summarizer:    summarizer,
	}
}

// CreateWg starts tracking background tasks with a wait group, so callers can
// wait for them through GetWg. Without it, WgAdd and WgDone do nothing.
func (d *DependencyService) CreateWg() {
	d.Wg = &sync.WaitGroup{}
}

func (d *DependencyService) GetWg() *sync.WaitGroup {
	return d.Wg
}

func (d *DependencyService) WgAdd() {
	if d.Wg != nil {
		d.Wg.Add(1)
	}
}

func (d *DependencyService) WgDone() {
	if d.Wg != nil {
		d.Wg.Done()
	}
}
