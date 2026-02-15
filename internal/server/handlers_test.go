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

package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers(t *testing.T) {
	tests := []struct {
		name           string
		handlerType    string // "health", "ready", "index"
		method         string
		checker        *mockHealthChecker
		expectedStatus int
		checkResponse  func(*testing.T, interface{})
	}{
		{
			name:           "health - GET request returns healthy",
			handlerType:    "health",
			method:         http.MethodGet,
			checker: &mockHealthChecker{
				upstreamHealthy: true,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp interface{}) {
				data := resp.(map[string]interface{})
				assert.Equal(t, "healthy", data["status"])
				assert.Contains(t, data, "upstream")
				upstream := data["upstream"].(map[string]interface{})
				assert.Equal(t, true, upstream["reachable"])
			},
		},
		{
			name:           "health - POST request returns healthy",
			handlerType:    "health",
			method:         http.MethodPost,
			checker: &mockHealthChecker{
				upstreamHealthy: true,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp interface{}) {
				data := resp.(map[string]interface{})
				assert.Equal(t, "healthy", data["status"])
			},
		},
		{
			name:        "health - degraded when upstream unreachable",
			handlerType: "health",
			method:      http.MethodGet,
			checker: &mockHealthChecker{
				upstreamHealthy: false,
				upstreamError:   errors.New("connection refused"),
			},
			expectedStatus: http.StatusServiceUnavailable,
			checkResponse: func(t *testing.T, resp interface{}) {
				data := resp.(map[string]interface{})
				assert.Equal(t, "degraded", data["status"])
				assert.Contains(t, data, "upstream")
				upstream := data["upstream"].(map[string]interface{})
				assert.Equal(t, false, upstream["reachable"])
				assert.Contains(t, upstream, "error")
			},
		},
		{
			name:        "ready - ready with successful collection",
			handlerType: "ready",
			method:      http.MethodGet,
			checker: &mockHealthChecker{
				healthy:     true,
				lastSuccess: time.Now().Add(-5 * time.Minute),
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp interface{}) {
				data := resp.(map[string]interface{})
				assert.Equal(t, true, data["ready"])
				assert.Contains(t, data, "last_success")
				assert.Contains(t, data, "last_success_seconds_ago")
				secondsAgo := data["last_success_seconds_ago"].(float64)
				assert.Greater(t, secondsAgo, 299.0) // ~5 minutes
			},
		},
		{
			name:        "ready - not ready with error",
			handlerType: "ready",
			method:      http.MethodGet,
			checker: &mockHealthChecker{
				healthy:     false,
				lastSuccess: time.Now().Add(-10 * time.Minute),
				lastError:   errors.New("API connection failed"),
			},
			expectedStatus: http.StatusServiceUnavailable,
			checkResponse: func(t *testing.T, resp interface{}) {
				data := resp.(map[string]interface{})
				assert.Equal(t, false, data["ready"])
				assert.Contains(t, data, "last_error")
				assert.Equal(t, "API connection failed", data["last_error"])
				assert.Contains(t, data, "last_success")
			},
		},
		{
			name:        "ready - ready but no successful collection yet",
			handlerType: "ready",
			method:      http.MethodGet,
			checker: &mockHealthChecker{
				healthy: true,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp interface{}) {
				data := resp.(map[string]interface{})
				assert.Equal(t, true, data["ready"])
				assert.NotContains(t, data, "last_success")
			},
		},
		{
			name:        "ready - not ready with no previous success",
			handlerType: "ready",
			method:      http.MethodGet,
			checker: &mockHealthChecker{
				healthy:   false,
				lastError: errors.New("initial connection failed"),
			},
			expectedStatus: http.StatusServiceUnavailable,
			checkResponse: func(t *testing.T, resp interface{}) {
				data := resp.(map[string]interface{})
				assert.Equal(t, false, data["ready"])
				assert.Contains(t, data, "last_error")
				assert.NotContains(t, data, "last_success")
			},
		},
		{
			name:           "index - GET request returns HTML with links",
			handlerType:    "index",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp interface{}) {
				body := resp.(string)
				assert.Contains(t, body, "Incident.io Prometheus Exporter")
				assert.Contains(t, body, "/metrics")
				assert.Contains(t, body, "/health")
				assert.Contains(t, body, "/ready")
			},
		},
		{
			name:           "index - POST request also works",
			handlerType:    "index",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp interface{}) {
				body := resp.(string)
				assert.Contains(t, body, "Incident.io Prometheus Exporter")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler http.HandlerFunc
			var url string

			switch tt.handlerType {
			case "health":
				handler = healthHandler(tt.checker)
				url = "/health"
			case "ready":
				handler = readyHandler(tt.checker)
				url = "/ready"
			case "index":
				handler = indexHandler()
				url = "/"
			}

			req := httptest.NewRequest(tt.method, url, nil)
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.handlerType == "index" {
				if tt.checkResponse != nil {
					body := w.Body.String()
					tt.checkResponse(t, body)
				}
			} else {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

				if tt.handlerType == "health" {
					var response map[string]interface{}
					err := json.NewDecoder(resp.Body).Decode(&response)
					require.NoError(t, err)
					if tt.checkResponse != nil {
						tt.checkResponse(t, response)
					}
				} else if tt.handlerType == "ready" {
					var response map[string]interface{}
					err := json.NewDecoder(resp.Body).Decode(&response)
					require.NoError(t, err)
					if tt.checkResponse != nil {
						tt.checkResponse(t, response)
					}
				}		}
	})
	}
}
