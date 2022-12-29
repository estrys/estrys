package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Meter interface {
	GetMetric(name string) prometheus.Counter
}

type meter struct {
	metrics map[string]prometheus.Counter
}

func NewMeter() *meter {
	return &meter{
		metrics: map[string]prometheus.Counter{
			PollerFetchTweetIterationsCounter: promauto.NewCounter(
				prometheus.CounterOpts{Name: PollerFetchTweetIterationsCounter},
			),
		},
	}
}

func (m *meter) GetMetric(name string) prometheus.Counter {
	return m.metrics[name]
}
