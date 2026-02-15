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

package incidentio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dirsigler/incidentio-exporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIncidents(t *testing.T) {
	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		wantErr        bool
		errContains    string
		checkResult    func(*testing.T, *IncidentsResponse)
		checkCallCount func(*testing.T, int)
	}{
		{
			name: "single page response",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/v2/incidents", r.URL.Path)
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, "250", r.URL.Query().Get("page_size"))

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"incidents": [
							{
								"id": "incident-1",
								"severity": {"id": "sev-1", "name": "Critical"},
								"incident_status": {"id": "status-1", "name": "Open"}
							},
							{
								"id": "incident-2",
								"severity": {"id": "sev-2", "name": "Minor"},
								"incident_status": {"id": "status-2", "name": "Closed"}
							}
						],
						"pagination_meta": {
							"after": "",
							"page_size": 250,
							"total_record_count": 2
						}
					}`))
				}))
			},
			wantErr: false,
			checkResult: func(t *testing.T, incidents *IncidentsResponse) {
				assert.Equal(t, 2, len(incidents.Incidents))
				assert.Equal(t, 2, incidents.PaginationMeta.TotalRecordCount)
				assert.Equal(t, "incident-1", incidents.Incidents[0].ID)
				assert.Equal(t, "Critical", incidents.Incidents[0].Severity.Name)
				assert.Equal(t, "Open", incidents.Incidents[0].IncidentStatus.Name)
			},
		},
		{
			name: "multiple pages",
			setupServer: func() *httptest.Server {
				callCount := 0
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callCount++
					assert.Equal(t, "/v2/incidents", r.URL.Path)
					assert.Equal(t, http.MethodGet, r.Method)

					switch callCount {
					case 1:
						assert.Equal(t, "", r.URL.Query().Get("after"))
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"incidents": [
								{
									"id": "incident-1",
									"severity": {"id": "sev-1", "name": "Critical"},
									"incident_status": {"id": "status-1", "name": "Open"}
								}
							],
							"pagination_meta": {
								"after": "cursor-page-2",
								"page_size": 250,
								"total_record_count": 3
							}
						}`))
					case 2:
						assert.Equal(t, "cursor-page-2", r.URL.Query().Get("after"))
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"incidents": [
								{
									"id": "incident-2",
									"severity": {"id": "sev-2", "name": "Minor"},
									"incident_status": {"id": "status-2", "name": "Resolved"}
								},
								{
									"id": "incident-3",
									"severity": {"id": "sev-1", "name": "Critical"},
									"incident_status": {"id": "status-1", "name": "Open"}
								}
							],
							"pagination_meta": {
								"after": "",
								"page_size": 250,
								"total_record_count": 3
							}
						}`))
					}
				}))
			},
			wantErr: false,
			checkResult: func(t *testing.T, incidents *IncidentsResponse) {
				assert.Equal(t, 3, len(incidents.Incidents))
				assert.Equal(t, 3, incidents.PaginationMeta.TotalRecordCount)
				assert.Equal(t, "incident-1", incidents.Incidents[0].ID)
				assert.Equal(t, "incident-2", incidents.Incidents[1].ID)
				assert.Equal(t, "incident-3", incidents.Incidents[2].ID)
			},
		},
		{
			name: "URL encoding with spaces",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					afterParam := r.URL.Query().Get("after")
					if afterParam != "" {
						assert.Equal(t, "cursor with spaces", afterParam)
					}

					w.WriteHeader(http.StatusOK)
					if afterParam == "" {
						w.Write([]byte(`{
							"incidents": [{"id": "incident-1", "severity": {"id": "sev-1", "name": "Critical"}, "incident_status": {"id": "status-1", "name": "Open"}}],
							"pagination_meta": {
								"after": "cursor with spaces",
								"page_size": 250,
								"total_record_count": 2
							}
						}`))
					} else {
						w.Write([]byte(`{
							"incidents": [{"id": "incident-2", "severity": {"id": "sev-2", "name": "Minor"}, "incident_status": {"id": "status-2", "name": "Closed"}}],
							"pagination_meta": {
								"after": "",
								"page_size": 250,
								"total_record_count": 2
							}
						}`))
					}
				}))
			},
			wantErr: false,
			checkResult: func(t *testing.T, incidents *IncidentsResponse) {
				assert.Equal(t, 2, len(incidents.Incidents))
			},
		},
		{
			name: "empty response",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"incidents": [],
						"pagination_meta": {
							"after": "",
							"page_size": 250,
							"total_record_count": 0
						}
					}`))
				}))
			},
			wantErr: false,
			checkResult: func(t *testing.T, incidents *IncidentsResponse) {
				assert.Equal(t, 0, len(incidents.Incidents))
				assert.Equal(t, 0, incidents.PaginationMeta.TotalRecordCount)
			},
		},
		{
			name: "error on second page",
			setupServer: func() *httptest.Server {
				callCount := 0
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callCount++
					switch callCount {
					case 1:
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"incidents": [{"id": "incident-1", "severity": {"id": "sev-1", "name": "Critical"}, "incident_status": {"id": "status-1", "name": "Open"}}],
							"pagination_meta": {
								"after": "cursor-page-2",
								"page_size": 250,
								"total_record_count": 2
							}
						}`))
					default:
						w.WriteHeader(http.StatusInternalServerError)
					}
				}))
			},
			wantErr:     true,
			errContains: "failed to get incidents (page 2)",
		},
		{
			name: "invalid JSON on second page",
			setupServer: func() *httptest.Server {
				callCount := 0
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callCount++
					switch callCount {
					case 1:
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"incidents": [{"id": "incident-1", "severity": {"id": "sev-1", "name": "Critical"}, "incident_status": {"id": "status-1", "name": "Open"}}],
							"pagination_meta": {
								"after": "cursor-page-2",
								"page_size": 250,
								"total_record_count": 2
							}
						}`))
					default:
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{invalid json`))
					}
				}))
			},
			wantErr:     true,
			errContains: "failed to get incidents (page 2)",
		},
		{
			name: "server error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			wantErr:     true,
			errContains: "failed to get incidents",
		},
		{
			name: "unauthorized error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
				}))
			},
			wantErr:     true,
			errContains: "401",
		},
		{
			name: "invalid JSON",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{invalid json`))
				}))
			},
			wantErr:     true,
			errContains: "failed to decode JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			cfg := config.IncidentIOConfig{
				Key: "test-key",
				URL: server.URL,
			}
			client := NewClient(cfg)

			incidents, err := client.GetIncidents(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, incidents)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, incidents)
			if tt.checkResult != nil {
				tt.checkResult(t, incidents)
			}
		})
	}
}
