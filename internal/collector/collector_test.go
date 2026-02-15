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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dirsigler/incidentio-exporter/internal/config"
	"github.com/dirsigler/incidentio-exporter/internal/incidentio"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	tests := []struct {
		name           string
		testType       string // "new", "describe", "collect_success", "collect_incidents_error", etc.
		setupServer    func() *httptest.Server
		runCollection  bool
		checkResult    func(*testing.T, *Collector, interface{})
	}{
		{
			name:     "new collector initialization",
			testType: "new",
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				require.NotNil(t, collector)
				assert.NotNil(t, collector.client)
				assert.NotNil(t, collector.totalCount)
				assert.NotNil(t, collector.severityCount)
				assert.NotNil(t, collector.statusCount)
			},
		},
		{
			name:     "describe returns metric descriptors",
			testType: "describe",
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				descs := data.([]*prometheus.Desc)
				assert.Len(t, descs, 15, "Expected 15 metric descriptors (5 incident + 3 user + 7 exporter)")
			},
		},
		{
			name:     "collect success with all metrics",
			testType: "collect",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/v2/incidents":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"incidents": [
								{"id": "inc-1", "severity": {"id": "sev-1", "name": "Critical"}, "incident_status": {"id": "status-1", "name": "Open"}},
								{"id": "inc-2", "severity": {"id": "sev-2", "name": "Minor"}, "incident_status": {"id": "status-2", "name": "Closed"}},
								{"id": "inc-3", "severity": {"id": "sev-1", "name": "Critical"}, "incident_status": {"id": "status-1", "name": "Open"}}
							],
							"pagination_meta": {"after": "", "page_size": 250, "total_record_count": 3}
						}`))
					case "/v1/severities":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"severities": [{"id": "sev-1", "name": "Critical"}, {"id": "sev-2", "name": "Minor"}]}`))
					case "/v1/incident_statuses":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"incident_statuses": [{"id": "status-1", "name": "Open"}, {"id": "status-2", "name": "Closed"}]}`))
					case "/v2/users":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"users": [{"id": "user-1", "name": "User 1", "base_role": {"id": "role-1", "name": "Owner"}, "custom_roles": []}], "pagination_meta": {"after": "", "page_size": 250}}`))
					}
				}))
			},
			runCollection: true,
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				metrics := data.([]prometheus.Metric)
				assert.GreaterOrEqual(t, len(metrics), 5, "Expected at least 5 metrics")
			},
		},
		{
			name:     "collect incidents API error",
			testType: "collect_error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			runCollection: true,
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				metrics := data.([]prometheus.Metric)
				// Should still have exporter metrics even on error
				assert.Len(t, metrics, 7, "Expected 7 exporter metrics on error")

				healthy, _, lastError := collector.HealthStatus()
				assert.False(t, healthy)
				assert.NotNil(t, lastError)
			},
		},
		{
			name:     "collect severities API error but other metrics succeed",
			testType: "collect_partial",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/v2/incidents":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"incidents": [{"id": "inc-1", "severity": {"id": "sev-1", "name": "Critical"}, "incident_status": {"id": "status-1", "name": "Open"}}],
							"pagination_meta": {"after": "", "page_size": 250, "total_record_count": 1}
						}`))
					case "/v1/severities":
						w.WriteHeader(http.StatusInternalServerError)
					case "/v1/incident_statuses":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"incident_statuses": [{"id": "status-1", "name": "Open"}]}`))
					case "/v2/users":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"users": [], "pagination_meta": {"after": "", "page_size": 250}}`))
					}
				}))
			},
			runCollection: true,
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				metrics := data.([]prometheus.Metric)
				assert.GreaterOrEqual(t, len(metrics), 2, "Expected at least total and status metrics")
			},
		},
		{
			name:     "health status - healthy after successful collection",
			testType: "health_success",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/v2/incidents":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"incidents": [], "pagination_meta": {"after": "", "page_size": 250, "total_record_count": 0}}`))
					case "/v1/severities":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"severities": []}`))
					case "/v1/incident_statuses":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"incident_statuses": []}`))
					}
				}))
			},
			runCollection: true,
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				healthy, lastSuccess, lastError := collector.HealthStatus()
				assert.True(t, healthy)
				assert.Nil(t, lastError)
				if !lastSuccess.IsZero() {
					assert.WithinDuration(t, time.Now(), lastSuccess, 5*time.Second)
				}
			},
		},
		{
			name:     "health status - unhealthy after failed collection",
			testType: "health_error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			runCollection: true,
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				healthy, _, lastError := collector.HealthStatus()
				assert.False(t, healthy)
				assert.NotNil(t, lastError)
			},
		},
		{
			name:     "health status - healthy initially without collection",
			testType: "health_initial",
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				healthy, _, lastError := collector.HealthStatus()
				assert.True(t, healthy)
				assert.Nil(t, lastError)
			},
		},
		{
			name:     "internal metrics registered",
			testType: "internal_metrics",
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				metrics := []prometheus.Collector{
					collector.collectionDuration,
					collector.apiCallDuration,
					collector.apiCallErrors,
					collector.lastCollectionSuccess,
					collector.lastCollectionAttempt,
					collector.collectionErrors,
				}

				for _, m := range metrics {
					assert.NotNil(t, m, "Internal metric should not be nil")
				}

				duration := testutil.ToFloat64(collector.collectionDuration)
				assert.GreaterOrEqual(t, duration, 0.0, "Collection duration should be non-negative")

				errors := testutil.ToFloat64(collector.collectionErrors)
				assert.GreaterOrEqual(t, errors, 0.0, "Collection errors should be non-negative")
			},
		},
		{
			name:     "internal metrics updated after collection",
			testType: "metrics_updated",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/v2/incidents":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"incidents": [], "pagination_meta": {"after": "", "page_size": 250, "total_record_count": 0}}`))
					case "/v1/severities":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"severities": []}`))
					case "/v1/incident_statuses":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"incident_statuses": []}`))
					}
				}))
			},
			runCollection: true,
			checkResult: func(t *testing.T, collector *Collector, data interface{}) {
				duration := testutil.ToFloat64(collector.collectionDuration)
				assert.Greater(t, duration, 0.0, "Collection duration should be recorded")

				lastSuccess := testutil.ToFloat64(collector.lastCollectionSuccess)
				assert.Greater(t, lastSuccess, 0.0, "Last collection success should be recorded")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = tt.setupServer()
				defer server.Close()
			}

			var cfg config.IncidentIOConfig
			if server != nil {
				cfg = config.IncidentIOConfig{
					Key: "test-key",
					URL: server.URL,
				}
			} else {
				cfg = config.IncidentIOConfig{
					Key: "test-key",
					URL: "http://localhost",
				}
			}
			client := incidentio.NewClient(cfg)
			collector := New(client, []string{}, []string{})

			var result interface{}

			switch tt.testType {
			case "describe":
				ch := make(chan *prometheus.Desc, 10)
				go func() {
					collector.Describe(ch)
					close(ch)
				}()
				var descs []*prometheus.Desc
				for desc := range ch {
					descs = append(descs, desc)
				}
				result = descs

			case "collect", "collect_error", "collect_partial":
				ch := make(chan prometheus.Metric, 10)
				go func() {
					collector.Collect(ch)
					close(ch)
				}()
				var metrics []prometheus.Metric
				for metric := range ch {
					metrics = append(metrics, metric)
				}
				result = metrics

			case "health_success", "health_error":
				// Trigger collection
				ch := make(chan prometheus.Metric, 10)
				go func() {
					collector.Collect(ch)
					close(ch)
				}()
				for range ch {
				}

			case "metrics_updated":
				// Initial collection
				ch := make(chan prometheus.Metric, 10)
				go func() {
					collector.Collect(ch)
					close(ch)
				}()
				for range ch {
				}
			}

			if tt.checkResult != nil {
				tt.checkResult(t, collector, result)
			}
		})
	}
}
