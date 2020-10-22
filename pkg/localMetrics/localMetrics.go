package localMetrics

import (
	"context"

	metricspkg "github.com/openshift/operator-custom-metrics/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var Metrics *MetricsStruct

// Resource Types to be used when reporting Metrics
const (
	EbsVolume           = "ebs_volume"
	EbsSnapshot         = "ebs_snapshot"
	Ec2Instance         = "ec2_instance"
	EfsVolume           = "efs_volume"
	Route53RecordSet    = "route53_record_set"
	Route53HostedZone   = "route53_hosted_zone"
	S3Bucket            = "s3_bucket"
	ElasticLoadBalancer = "elastic_loadbalancer"
	NatGateway          = "nat_gateway"
	NetworkLoadBalancer = "network_loadbalancer"
	NetworkInterface    = "network_interface"
	InternetGateway     = "internet_gateway"
	Subnet              = "subnet"
	RouteTable          = "route_table"
	NetworkACL          = "network_acl"
	SecurityGroup       = "security_group"
	VPC                 = "vpc"
	VpnConnection       = "vpn_connection"
	VpnGateway          = "vpn_gateway"
)

// Creates a Metrics struct
type MetricsStruct struct {
	AccountSuccess  prometheus.Counter
	AccountFail     prometheus.Counter
	ResourceSuccess *prometheus.CounterVec
	ResourceFail    *prometheus.CounterVec
	DurationSeconds prometheus.Histogram
}

// Intializes new Metrics Service
func Initialize(metricsPort string, metricsPath string) error {
	Metrics = &MetricsStruct{
		AccountSuccess: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aws_account_shredder_accounts_success",
			Help: "Count of accounts that have been shredded successfully",
		}),
		AccountFail: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aws_account_shredder_accounts_failed",
			Help: "Count of accounts that have failed to shred",
		}),
		ResourceSuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "aws_account_shredder_resources_success",
			Help: "Count of specific AWS Resources that have been shredded successfully",
		}, []string{"resource_type"}),
		ResourceFail: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "aws_account_shredder_resources_failed",
			Help: "Count of specific AWS Resources that have failed to shred",
		}, []string{"resource_type"}),
		DurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "aws_account_shredder_duration_seconds",
			Help:    "Distribution of the number of seconds a AWS Shred operation takes",
			Buckets: []float64{60, 120, 180, 240, 300, 360, 420, 480, 540, 600},
		}),
	}

	collectors := []prometheus.Collector{
		Metrics.AccountSuccess,
		Metrics.AccountFail,
		*Metrics.ResourceSuccess,
		*Metrics.ResourceFail,
		Metrics.DurationSeconds,
	}

	metricsServer := metricspkg.NewBuilder().WithPort(metricsPort).WithPath(metricsPath).
		WithCollectors(collectors).
		WithRoute().
		WithServiceName("aws-account-shredder").
		GetConfig()

	// Configure localMetrics if it errors log the error but continue
	return metricspkg.ConfigureMetrics(context.TODO(), *metricsServer)
}

func ResourceSuccess(resourceType string) {
	Metrics.ResourceSuccess.With(prometheus.Labels{"resource_type": resourceType}).Inc()
}
func ResourceFail(resourceType string) {
	Metrics.ResourceFail.With(prometheus.Labels{"resource_type": resourceType}).Inc()
}
