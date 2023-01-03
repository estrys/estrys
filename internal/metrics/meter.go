package metrics

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Meter interface {
	GetMetric(string) (any, bool)
	Register(string) error
}

type metricRegistry struct {
	metrics map[string]any
}

func (r *metricRegistry) Register(name string) error {
	r.metrics = make(map[string]any)
	if strings.HasSuffix(name, "_counter") {
		counter := promauto.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: name,
		})
		r.metrics[name] = counter
		return nil
	}
	if strings.HasSuffix(name, "_gauge") {
		gauge := promauto.NewGauge(prometheus.GaugeOpts{
			Name: "bla",
			Help: name,
		})
		r.metrics[name] = gauge
		return nil
	}
	return fmt.Errorf("cannot register metrics, unknown type for given name: %s", name)
}

func (r *metricRegistry) GetMetric(key string) (any, bool) {
	value, ok := r.metrics[key]
	return value, ok
}

func NewRegistry() *metricRegistry {
	mr := &metricRegistry{}
	for _, name := range []string{"bla_counter", "foo_gauge", "iteration_counter"} {
		err := mr.Register(name)
		if err != nil {
			panic(err)
		}
	}
	return mr
}
