package localMetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	DummyMetric = prometheus.NewGauge(prometheus.GaugeOpts{Name: "DummyMetric"})

	MetricsList = []prometheus.Collector{
		DummyMetric,
	}
)
