package assessment

import (
	"encoding/json"
	"io"
	"sync"
)

type JSONPublisher struct {
	Writer io.Writer
}

func (p JSONPublisher) Publish(report Report) error {
	encoder := json.NewEncoder(p.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

type MemoryPublisher struct {
	mu      sync.Mutex
	reports []Report
}

func (p *MemoryPublisher) Publish(report Report) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reports = append(p.reports, report)
	return nil
}

func (p *MemoryPublisher) Reports() []Report {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]Report(nil), p.reports...)
}
