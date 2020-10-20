package localMetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	MetricsList = map[string]prometheus.Collector{
		"account_success": prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aws_account_shredder_accounts_success",
			Help: "Count of accounts that have been shredded successfully",
		}),
		"account_fail": prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aws_account_shredder_accounts_failed",
			Help: "Count of accounts that have failed to shred",
		}),
		"resource_success": prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "aws_account_shredder_resources_success",
			Help: "Count of specific AWS Resources that have been shredded successfully",
		}, []string{"resource_type"}),
		"resource_fail": prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "aws_account_shredder_resources_failed",
			Help: "Count of specific AWS Resources that have failed to shred",
		}, []string{"resource_type"}),
		"duration_seconds": prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "aws_account_shredder_duration_seconds",
			Help:    "Distribution of the number of seconds a AWS Shred operation takes",
			Buckets: []float64{0.001, 0.01, 0.1, 1, 5, 10, 20},
		}),
	}
)
