package main

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "incidentio_incidents"
)

type Collector struct {
	TotalCount    *prometheus.Desc
	SeverityCount *prometheus.Desc
	StatusCount   *prometheus.Desc
	Application   *application
}

func NewIncidentCollector(app *application) Collector {
	collector := Collector{
		TotalCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "total", "count"),
			"The total number of incidents.",
			nil,
			nil,
		),
		SeverityCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "severity", "count"),
			"The number of incidents by severity.",
			[]string{"severity"},
			nil,
		),
		StatusCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "status", "count"),
			"The number of incidents by status.",
			[]string{"status"},
			nil,
		),
		Application: app,
	}

	return collector
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.TotalCount
	ch <- c.SeverityCount
	ch <- c.StatusCount
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.Application.logger.Info("starting to collect metrics")
	start := time.Now()

	var wg sync.WaitGroup

	// Here we retrieve ALL available Incidents in https://incident.io.
	// Afterwards we feed the total count to Prometheus.
	incidents := c.Application.getIncidents()
	ch <- prometheus.MustNewConstMetric(
		c.TotalCount,
		prometheus.CounterValue,
		float64(incidents.PaginationMeta.TotalRecordCount),
	)

	// Here we retrieve all available Severities in https://incident.io.
	// Afterwards for each available Severity we collect the count of incidents and feed them to Prometheus.
	severityResponse := c.Application.getSeverities()
	for _, severity := range severityResponse.Severities {
		c.Application.logger.Debug("found Severity", "name", severity.Name, "id", severity.ID)
		severityIncidents := c.Application.getSeverityIncidents(severity.ID)

		ch <- prometheus.MustNewConstMetric(
			c.SeverityCount,
			prometheus.CounterValue,
			float64(severityIncidents.PaginationMeta.TotalRecordCount),
			severity.Name,
		)
	}

	// Here we retrieve all available Statuses in https://incident.io.
	// Afterwards for each available Status we collect the count of incidents and feed them to Prometheus.
	statusResponse := c.Application.getStatuses()
	for _, status := range statusResponse.IncidentStatuses {
		c.Application.logger.Debug("found Status", "name", status.Name, "id", status.ID)
		statusIncidents := c.Application.getStatusIncidents(status.ID)

		ch <- prometheus.MustNewConstMetric(
			c.StatusCount,
			prometheus.CounterValue,
			float64(statusIncidents.PaginationMeta.TotalRecordCount),
			status.Name,
		)
	}

	wg.Wait()

	// Finish the run of the Collect function and end the application.
	// Total expected time to query all incidents for the Collector is 5 seconds.
	duration := time.Since(start)
	c.Application.logger.Info("finished collecting metrics", "duration", duration.Seconds())
}
