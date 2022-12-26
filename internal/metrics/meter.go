package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Meter interface {
	GetRegistry() *prometheus.Registry
}

type meter struct {
	registry *prometheus.Registry
}

func NewMeter() *meter {
	return &meter{
		registry: prometheus.NewRegistry(),
	}
}

func (m *meter) GetRegistry() *prometheus.Registry {
	return m.registry
}
