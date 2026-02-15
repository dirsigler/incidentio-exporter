// Copyright 2026 Dennis Irsigler
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/dirsigler/incidentio-exporter/internal/incidentio"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace         = "incidentio_incidents"
	usersNamespace    = "incidentio_users"
	exporterNamespace = "incidentio_exporter"
)

// Collector implements the Prometheus Collector interface for incident.io metrics.
type Collector struct {
	// Incident metrics
	totalCount        *prometheus.Desc
	severityCount     *prometheus.Desc
	statusCount       *prometheus.Desc
	customFieldCount  *prometheus.Desc
	catalogEntryCount *prometheus.Desc

	// User metrics
	usersTotal           *prometheus.Desc
	usersBaseRoleCount   *prometheus.Desc
	usersCustomRoleCount *prometheus.Desc

	// Exporter metrics
	up                    prometheus.Gauge
	collectionDuration    prometheus.Gauge
	apiCallDuration       *prometheus.GaugeVec
	apiCallErrors         *prometheus.CounterVec
	lastCollectionSuccess prometheus.Gauge
	lastCollectionAttempt prometheus.Gauge
	collectionErrors      prometheus.Counter

	client         *incidentio.Client
	customFieldIDs []string
	catalogTypeIDs []string
	mu             sync.RWMutex
	lastSuccess    time.Time
	lastError      error
}

// New creates a new incident.io Prometheus collector.
func New(client *incidentio.Client, customFieldIDs, catalogTypeIDs []string) *Collector {
	return &Collector{
		// Incident metrics
		totalCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "total", "count"),
			"The total number of incidents.",
			nil,
			nil,
		),
		severityCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "severity", "count"),
			"The number of incidents by severity.",
			[]string{"severity"},
			nil,
		),
		statusCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", "count"),
			"The number of incidents by status.",
			[]string{"status"},
			nil,
		),
		customFieldCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "custom_field", "count"),
			"The number of incidents by custom field value.",
			[]string{"custom_field_id", "custom_field_name", "value"},
			nil,
		),
		catalogEntryCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "catalog_entry", "count"),
			"The number of incidents by catalog entry.",
			[]string{"catalog_type_id", "catalog_type_name", "entry_id", "entry_name"},
			nil,
		),
		// User metrics
		usersTotal: prometheus.NewDesc(
			prometheus.BuildFQName(usersNamespace, "total", "count"),
			"The total number of users.",
			nil,
			nil,
		),
		usersBaseRoleCount: prometheus.NewDesc(
			prometheus.BuildFQName(usersNamespace, "base_role", "count"),
			"The number of users by base role.",
			[]string{"role_id", "role_name"},
			nil,
		),
		usersCustomRoleCount: prometheus.NewDesc(
			prometheus.BuildFQName(usersNamespace, "custom_role", "count"),
			"The number of users with each custom role.",
			[]string{"role_id", "role_name"},
			nil,
		),
		// Exporter metrics
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: exporterNamespace,
			Name:      "up",
			Help:      "Was the last scrape of the incident.io API successful.",
		}),
		collectionDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: exporterNamespace,
			Name:      "collection_duration_seconds",
			Help:      "Duration of the last collection cycle in seconds.",
		}),
		apiCallDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: exporterNamespace,
			Name:      "api_call_duration_seconds",
			Help:      "Duration of API calls in seconds.",
		}, []string{"endpoint"}),
		apiCallErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: exporterNamespace,
			Name:      "api_call_errors_total",
			Help:      "Total number of API call errors.",
		}, []string{"endpoint"}),
		lastCollectionSuccess: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: exporterNamespace,
			Name:      "last_collection_success_timestamp_seconds",
			Help:      "Timestamp of the last successful collection.",
		}),
		lastCollectionAttempt: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: exporterNamespace,
			Name:      "last_collection_attempt_timestamp_seconds",
			Help:      "Timestamp of the last collection attempt.",
		}),
		collectionErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: exporterNamespace,
			Name:      "collection_errors_total",
			Help:      "Total number of collection errors.",
		}),
		client:         client,
		customFieldIDs: customFieldIDs,
		catalogTypeIDs: catalogTypeIDs,
	}
}

// Describe sends the super-set of all possible descriptors to Prometheus.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	// Incident metrics
	ch <- c.totalCount
	ch <- c.severityCount
	ch <- c.statusCount
	ch <- c.customFieldCount
	ch <- c.catalogEntryCount

	// User metrics
	ch <- c.usersTotal
	ch <- c.usersBaseRoleCount
	ch <- c.usersCustomRoleCount

	// Exporter metrics
	c.up.Describe(ch)
	c.collectionDuration.Describe(ch)
	c.apiCallDuration.Describe(ch)
	c.apiCallErrors.Describe(ch)
	c.lastCollectionSuccess.Describe(ch)
	c.lastCollectionAttempt.Describe(ch)
	c.collectionErrors.Describe(ch)
}

// Collect fetches metrics from incident.io and sends them to Prometheus.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	slog.Info("starting to collect metrics")
	start := time.Now()
	c.lastCollectionAttempt.SetToCurrentTime()

	// Create context with timeout for all API calls
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Fetch all incidents once with pagination
	apiStart := time.Now()
	incidents, err := c.client.GetIncidents(ctx)
	c.apiCallDuration.WithLabelValues("incidents").Set(time.Since(apiStart).Seconds())

	if err != nil {
		slog.Error("failed to get incidents, aborting collection",
			"error", err,
			"component", "collector",
		)
		c.apiCallErrors.WithLabelValues("incidents").Inc()
		c.collectionErrors.Inc()
		c.up.Set(0)
		c.mu.Lock()
		c.lastError = err
		c.mu.Unlock()

		// Collect exporter metrics even on failure
		c.up.Collect(ch)
		c.collectionDuration.Collect(ch)
		c.apiCallDuration.Collect(ch)
		c.apiCallErrors.Collect(ch)
		c.lastCollectionSuccess.Collect(ch)
		c.lastCollectionAttempt.Collect(ch)
		c.collectionErrors.Collect(ch)
		return
	}
	slog.Debug("incidents retrieved",
		"total_count", len(incidents.Incidents),
	)

	// Emit total count metric
	ch <- prometheus.MustNewConstMetric(
		c.totalCount,
		prometheus.GaugeValue,
		float64(len(incidents.Incidents)),
	)

	// Get severities and statuses metadata
	var wg sync.WaitGroup
	var severities *incidentio.SeveritiesResponse
	var statuses *incidentio.StatusesResponse
	var severitiesErr, statusesErr error

	wg.Add(2)

	// Fetch severities concurrently
	go func() {
		defer wg.Done()
		apiStart := time.Now()
		severities, severitiesErr = c.client.GetSeverities(ctx)
		c.apiCallDuration.WithLabelValues("severities").Set(time.Since(apiStart).Seconds())
		if severitiesErr != nil {
			c.apiCallErrors.WithLabelValues("severities").Inc()
		}
	}()

	// Fetch statuses concurrently
	go func() {
		defer wg.Done()
		apiStart := time.Now()
		statuses, statusesErr = c.client.GetStatuses(ctx)
		c.apiCallDuration.WithLabelValues("statuses").Set(time.Since(apiStart).Seconds())
		if statusesErr != nil {
			c.apiCallErrors.WithLabelValues("statuses").Inc()
		}
	}()

	wg.Wait()

	// Count incidents by severity locally
	if severitiesErr != nil {
		slog.Error("failed to get severities, skipping severity metrics",
			"error", severitiesErr,
			"component", "collector",
		)
	} else {
		slog.Debug("severities retrieved", "count", len(severities.Severities))
		severityCounts := make(map[string]int)

		// Count incidents for each severity
		for _, incident := range incidents.Incidents {
			if incident.Severity.ID != "" {
				severityCounts[incident.Severity.Name]++
			}
		}

		// Emit severity metrics
		for _, severity := range severities.Severities {
			count := severityCounts[severity.Name]
			slog.Debug("emitting severity metric",
				"name", severity.Name,
				"count", count,
			)
			ch <- prometheus.MustNewConstMetric(
				c.severityCount,
				prometheus.GaugeValue,
				float64(count),
				severity.Name,
			)
		}
	}

	// Count incidents by status locally
	if statusesErr != nil {
		slog.Error("failed to get statuses, skipping status metrics",
			"error", statusesErr,
			"component", "collector",
		)
	} else {
		slog.Debug("statuses retrieved", "count", len(statuses.IncidentStatuses))
		statusCounts := make(map[string]int)

		// Count incidents for each status
		for _, incident := range incidents.Incidents {
			if incident.IncidentStatus.ID != "" {
				statusCounts[incident.IncidentStatus.Name]++
			}
		}

		// Emit status metrics
		for _, status := range statuses.IncidentStatuses {
			count := statusCounts[status.Name]
			slog.Debug("emitting status metric",
				"name", status.Name,
				"count", count,
			)
			ch <- prometheus.MustNewConstMetric(
				c.statusCount,
				prometheus.GaugeValue,
				float64(count),
				status.Name,
			)
		}
	}

	// Fetch and process users
	apiStart = time.Now()
	users, usersErr := c.client.GetUsers(ctx)
	c.apiCallDuration.WithLabelValues("users").Set(time.Since(apiStart).Seconds())

	if usersErr != nil {
		slog.Error("failed to get users, skipping user metrics",
			"error", usersErr,
			"component", "collector",
		)
		c.apiCallErrors.WithLabelValues("users").Inc()
	} else {
		slog.Debug("users retrieved", "count", len(users))

		// Emit total users count
		ch <- prometheus.MustNewConstMetric(
			c.usersTotal,
			prometheus.GaugeValue,
			float64(len(users)),
		)

		// Count users by base role
		baseRoleCounts := make(map[string]map[string]int) // map[roleID]map["name"|"count"]
		customRoleCounts := make(map[string]map[string]int)

		for _, user := range users {
			// Count base role
			if user.BaseRole.ID != "" {
				if _, exists := baseRoleCounts[user.BaseRole.ID]; !exists {
					baseRoleCounts[user.BaseRole.ID] = map[string]int{"count": 0}
				}
				baseRoleCounts[user.BaseRole.ID]["count"]++
				// Store name (will overwrite with same value, but that's fine)
				baseRoleCounts[user.BaseRole.ID]["name"] = len(user.BaseRole.Name)
			}

			// Count custom roles
			for _, customRole := range user.CustomRoles {
				if customRole.ID != "" {
					if _, exists := customRoleCounts[customRole.ID]; !exists {
						customRoleCounts[customRole.ID] = map[string]int{"count": 0}
					}
					customRoleCounts[customRole.ID]["count"]++
					customRoleCounts[customRole.ID]["name"] = len(customRole.Name)
				}
			}
		}

		// Emit base role metrics
		// We need to track role names separately
		baseRoleNames := make(map[string]string)
		for _, user := range users {
			if user.BaseRole.ID != "" {
				baseRoleNames[user.BaseRole.ID] = user.BaseRole.Name
			}
		}

		for roleID, counts := range baseRoleCounts {
			roleName := baseRoleNames[roleID]
			count := counts["count"]
			slog.Debug("emitting base role metric",
				"role_id", roleID,
				"role_name", roleName,
				"count", count,
			)
			ch <- prometheus.MustNewConstMetric(
				c.usersBaseRoleCount,
				prometheus.GaugeValue,
				float64(count),
				roleID,
				roleName,
			)
		}

		// Emit custom role metrics
		customRoleNames := make(map[string]string)
		for _, user := range users {
			for _, customRole := range user.CustomRoles {
				if customRole.ID != "" {
					customRoleNames[customRole.ID] = customRole.Name
				}
			}
		}

		for roleID, counts := range customRoleCounts {
			roleName := customRoleNames[roleID]
			count := counts["count"]
			slog.Debug("emitting custom role metric",
				"role_id", roleID,
				"role_name", roleName,
				"count", count,
			)
			ch <- prometheus.MustNewConstMetric(
				c.usersCustomRoleCount,
				prometheus.GaugeValue,
				float64(count),
				roleID,
				roleName,
			)
		}
	}

	// Process custom fields if configured
	if len(c.customFieldIDs) > 0 {
		slog.Debug("processing custom fields", "count", len(c.customFieldIDs))
		for _, customFieldID := range c.customFieldIDs {
			apiStart := time.Now()
			customFieldOptions, err := c.client.GetCustomFieldOptions(ctx, customFieldID)
			c.apiCallDuration.WithLabelValues("custom_field_options").Set(time.Since(apiStart).Seconds())

			if err != nil {
				slog.Error("failed to get custom field options",
					"error", err,
					"custom_field_id", customFieldID,
				)
				c.apiCallErrors.WithLabelValues("custom_field_options").Inc()
				continue
			}

			// Count incidents per custom field option
			customFieldCounts := make(map[string]int)
			customFieldNames := make(map[string]string)

			for _, incident := range incidents.Incidents {
				for _, customFieldEntry := range incident.CustomFieldEntries {
					if customFieldEntry.CustomField.ID == customFieldID {
						customFieldNames[customFieldID] = customFieldEntry.CustomField.Name
						for _, value := range customFieldEntry.Values {
							if value.ValueOption != nil && value.ValueOption.Value != "" {
								customFieldCounts[value.ValueOption.Value]++
							} else if value.ValueText != "" {
								customFieldCounts[value.ValueText]++
							}
						}
					}
				}
			}

			// Emit metrics for each custom field option
			customFieldName := customFieldNames[customFieldID]
			if customFieldName == "" {
				customFieldName = customFieldID
			}

			for _, option := range customFieldOptions.CustomFieldOptions {
				count := customFieldCounts[option.Value]
				slog.Debug("emitting custom field metric",
					"custom_field_id", customFieldID,
					"custom_field_name", customFieldName,
					"value", option.Value,
					"count", count,
				)
				ch <- prometheus.MustNewConstMetric(
					c.customFieldCount,
					prometheus.GaugeValue,
					float64(count),
					customFieldID,
					customFieldName,
					option.Value,
				)
			}
		}
	}

	// Process catalog types if configured
	if len(c.catalogTypeIDs) > 0 {
		slog.Debug("processing catalog types", "count", len(c.catalogTypeIDs))
		for _, catalogTypeID := range c.catalogTypeIDs {
			apiStart := time.Now()
			catalogEntries, err := c.client.GetCatalogEntries(ctx, catalogTypeID)
			c.apiCallDuration.WithLabelValues("catalog_entries").Set(time.Since(apiStart).Seconds())

			if err != nil {
				slog.Error("failed to get catalog entries",
					"error", err,
					"catalog_type_id", catalogTypeID,
				)
				c.apiCallErrors.WithLabelValues("catalog_entries").Inc()
				continue
			}

			if len(catalogEntries.CatalogEntries) == 0 {
				slog.Warn("no catalog entries found", "catalog_type_id", catalogTypeID)
				continue
			}

			catalogTypeName := catalogEntries.CatalogType.Name

			// Count incidents per catalog entry
			// This checks if any custom field references the catalog entry
			for _, entry := range catalogEntries.CatalogEntries {
				incidentCount := 0

				for _, incident := range incidents.Incidents {
					for _, customFieldEntry := range incident.CustomFieldEntries {
						for _, value := range customFieldEntry.Values {
							if value.ValueCatalogEntry != nil && value.ValueCatalogEntry.ID == entry.ID {
								incidentCount++
								break
							}
						}
					}
				}

				slog.Debug("emitting catalog entry metric",
					"catalog_type_id", catalogTypeID,
					"catalog_type_name", catalogTypeName,
					"entry_id", entry.ID,
					"entry_name", entry.Name,
					"count", incidentCount,
				)
				ch <- prometheus.MustNewConstMetric(
					c.catalogEntryCount,
					prometheus.GaugeValue,
					float64(incidentCount),
					catalogTypeID,
					catalogTypeName,
					entry.ID,
					entry.Name,
				)
			}
		}
	}

	duration := time.Since(start)
	c.collectionDuration.Set(duration.Seconds())
	c.lastCollectionSuccess.SetToCurrentTime()
	c.up.Set(1)

	c.mu.Lock()
	c.lastSuccess = time.Now()
	c.lastError = nil
	c.mu.Unlock()

	// Collect exporter metrics
	c.up.Collect(ch)
	c.collectionDuration.Collect(ch)
	c.apiCallDuration.Collect(ch)
	c.apiCallErrors.Collect(ch)
	c.lastCollectionSuccess.Collect(ch)
	c.lastCollectionAttempt.Collect(ch)
	c.collectionErrors.Collect(ch)

	slog.Info("finished collecting metrics",
		"duration_seconds", duration.Seconds(),
		"component", "collector",
	)
}

// HealthStatus returns the health status of the collector.
func (c *Collector) HealthStatus() (healthy bool, lastSuccess time.Time, lastError error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Consider healthy if we've had a successful collection or no error yet
	healthy = c.lastError == nil
	return healthy, c.lastSuccess, c.lastError
}

// CheckUpstreamHealth performs a lightweight health check against the incident.io API.
func (c *Collector) CheckUpstreamHealth() error {
	return c.client.HealthCheck()
}
