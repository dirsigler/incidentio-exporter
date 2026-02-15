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

func TestGetSeverities(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		statusCode     int
		wantErr        bool
		errContains    string
		checkResult    func(*testing.T, *SeveritiesResponse)
	}{
		{
			name: "success with multiple severities",
			responseBody: `{
				"severities": [
					{"id": "sev-1", "name": "Critical"},
					{"id": "sev-2", "name": "High"},
					{"id": "sev-3", "name": "Medium"},
					{"id": "sev-4", "name": "Low"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			checkResult: func(t *testing.T, severities *SeveritiesResponse) {
				assert.Len(t, severities.Severities, 4)
				assert.Equal(t, "sev-1", severities.Severities[0].ID)
				assert.Equal(t, "Critical", severities.Severities[0].Name)
				assert.Equal(t, "sev-2", severities.Severities[1].ID)
				assert.Equal(t, "High", severities.Severities[1].Name)
			},
		},
		{
			name:         "empty list",
			responseBody: `{"severities": []}`,
			statusCode:   http.StatusOK,
			wantErr:      false,
			checkResult: func(t *testing.T, severities *SeveritiesResponse) {
				assert.Len(t, severities.Severities, 0)
			},
		},
		{
			name:         "server error",
			responseBody: "",
			statusCode:   http.StatusInternalServerError,
			wantErr:      true,
			errContains:  "failed to get all severities",
		},
		{
			name:         "invalid JSON",
			responseBody: `{invalid json}`,
			statusCode:   http.StatusOK,
			wantErr:      true,
			errContains:  "failed to decode JSON",
		},
		{
			name:         "unauthorized",
			responseBody: "",
			statusCode:   http.StatusUnauthorized,
			wantErr:      true,
			errContains:  "401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/severities", r.URL.Path)
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

			severities, err := client.GetSeverities(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, severities)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, severities)
			if tt.checkResult != nil {
				tt.checkResult(t, severities)
			}
		})
	}
}
