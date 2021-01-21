package utils

import (
	"github.com/prometheus/client_golang/prometheus"
)

var JobCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "jobs_performed",
		Help: "Number of jobs performed",
	},
	[]string{"job_name", "status"},
)

var JobTime = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "job_time",
		Help:    "The time taken to perform a job",
		Buckets: []float64{1, 5, 15, 30, 60, 60 * 5, 60 * 15, 60 * 30, 60 * 45, 60 * 60},
	},
	[]string{"job_name", "status"})

var InflightJob = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "inflight_job",
		Help: "The number of jobs in progress",
	}, []string{"job_name"})

func RegisterMetrics() {
	prometheus.MustRegister(JobTime, JobCount, InflightJob)
}
