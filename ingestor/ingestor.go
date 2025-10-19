package ingestor

import (
	"sync"
	"sync/atomic"

	"github.com/miguelhrocha/otel-collector/config"
)

type Record struct {
	AttrValue string
	TimeUnix  uint64
	ObsUnix   uint64
	Severity  int32
	Body      string
	TraceID   string
	SpanID    string
}

type Ingestor struct {
	q  chan Record
	wg sync.WaitGroup

	aggregator   *Aggregator
	deduplicator *Deduplicator

	stopped atomic.Bool
}

func NewIngestor(cfg config.Config, a *Aggregator, d *Deduplicator) *Ingestor {
	in := &Ingestor{
		q:            make(chan Record, cfg.QueueSize),
		aggregator:   a,
		deduplicator: d,
	}

	for i := 0; i < cfg.Workers; i++ {
		in.wg.Add(1)
		go func() {
			defer in.wg.Done()
			for r := range in.q {
				in.process(r)
			}
		}()
	}

	return in
}

func (i *Ingestor) TryEnqueue(r Record) bool {
	if i.stopped.Load() {
		return false
	}

	select {
	case i.q <- r:
		return true
	default:
		return false
	}
}

func (i *Ingestor) Stop() {
	i.stopped.Store(true)
	close(i.q)
	i.wg.Wait()
}

func (i *Ingestor) process(r Record) {
	if i.deduplicator.IsNew(r) {
		i.aggregator.Inc(r.AttrValue)
	}
}
