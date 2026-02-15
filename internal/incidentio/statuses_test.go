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

func TestGetStatuses(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		statusCode     int
		wantErr        bool
		errContains    string
		checkResult    func(*testing.T, *StatusesResponse)
	}{
		{
			name: "success with multiple statuses",
			responseBody: `{
				"incident_statuses": [
					{"id": "status-1", "name": "Open"},
					{"id": "status-2", "name": "Investigating"},
					{"id": "status-3", "name": "Resolved"},
					{"id": "status-4", "name": "Closed"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			checkResult: func(t *testing.T, statuses *StatusesResponse) {
				assert.Len(t, statuses.IncidentStatuses, 4)
				assert.Equal(t, "status-1", statuses.IncidentStatuses[0].ID)
				assert.Equal(t, "Open", statuses.IncidentStatuses[0].Name)
				assert.Equal(t, "status-2", statuses.IncidentStatuses[1].ID)
				assert.Equal(t, "Investigating", statuses.IncidentStatuses[1].Name)
			},
		},
		{
			name:         "empty list",
			responseBody: `{"incident_statuses": []}`,
			statusCode:   http.StatusOK,
			wantErr:      false,
			checkResult: func(t *testing.T, statuses *StatusesResponse) {
				assert.Len(t, statuses.IncidentStatuses, 0)
			},
		},
		{
			name:         "forbidden error",
			responseBody: "",
			statusCode:   http.StatusForbidden,
			wantErr:      true,
			errContains:  "failed to get all statuses",
		},
		{
			name:         "invalid JSON",
			responseBody: `not valid json at all`,
			statusCode:   http.StatusOK,
			wantErr:      true,
			errContains:  "failed to decode JSON",
		},
		{
			name:         "server error",
			responseBody: "",
			statusCode:   http.StatusInternalServerError,
			wantErr:      true,
			errContains:  "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/incident_statuses", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			cfg := config.IncidentIOConfig{
				Key:      "test-key",
				URL:      server.URL,
			}
			client := NewClient(cfg)

			statuses, err := client.GetStatuses(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, statuses)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, statuses)
			if tt.checkResult != nil {
				tt.checkResult(t, statuses)
			}
		})
	}
}
