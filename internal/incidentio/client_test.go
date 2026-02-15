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

func TestClient(t *testing.T) {
	tests := []struct {
		name        string
		testType    string // "newclient", "dorequest", "headers", "network"
		setupServer func() *httptest.Server
		path        string
		wantErr     bool
		errContains string
		checkResult func(*testing.T, interface{})
	}{
		{
			name:     "new client initialization",
			testType: "newclient",
			checkResult: func(t *testing.T, result interface{}) {
				client := result.(*Client)
				assert.NotNil(t, client)
				assert.Equal(t, "test-api-key", client.apiKey)
				assert.Equal(t, "https://test.incident.io", client.baseURL)
				assert.NotNil(t, client.httpClient)
			},
		},
		{
			name:     "doRequest success with valid response",
			testType: "dorequest",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"test": "data"}`))
				}))
			},
			path:    "/test",
			wantErr: false,
			checkResult: func(t *testing.T, result interface{}) {
				data := result.(map[string]interface{})
				assert.Equal(t, "data", data["test"])
			},
		},
		{
			name:     "doRequest with proper headers",
			testType: "headers",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
					assert.Equal(t, "application/json", r.Header.Get("Accept"))
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"test": "data"}`))
				}))
			},
			path:    "/test",
			wantErr: false,
		},
		{
			name:     "doRequest unauthorized error",
			testType: "dorequest",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error": "unauthorized"}`))
				}))
			},
			path:        "/test",
			wantErr:     true,
			errContains: "401",
		},
		{
			name:     "doRequest invalid JSON response",
			testType: "dorequest",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`not valid json`))
				}))
			},
			path:        "/test",
			wantErr:     true,
			errContains: "failed to decode JSON",
		},
		{
			name:     "doRequest server error",
			testType: "dorequest",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "internal error"}`))
				}))
			},
			path:        "/test",
			wantErr:     true,
			errContains: "500",
		},
		{
			name:     "doRequest forbidden error",
			testType: "dorequest",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusForbidden)
				}))
			},
			path:        "/test",
			wantErr:     true,
			errContains: "403",
		},
		{
			name:     "doRequest network error",
			testType: "network",
			path:     "/test",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg config.IncidentIOConfig
			var client *Client

			switch tt.testType {
			case "newclient":
				cfg = config.IncidentIOConfig{
					Key:      "test-api-key",
					URL:      "https://test.incident.io",
				}
				client = NewClient(cfg)
				if tt.checkResult != nil {
					tt.checkResult(t, client)
				}
				return

			case "network":
				cfg = config.IncidentIOConfig{
					Key:      "test-key",
					URL:      "http://localhost:1",
				}
				client = NewClient(cfg)

			default:
				server := tt.setupServer()
				defer server.Close()

				cfg = config.IncidentIOConfig{
					Key:      "test-api-key",
					URL:      server.URL,
				}
				client = NewClient(cfg)
			}

			var result map[string]interface{}
			err := client.doRequest(context.Background(), tt.path, &result)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
