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
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHealthChecker is a mock implementation of healthChecker for testing
type mockHealthChecker struct {
	healthy         bool
	lastSuccess     time.Time
	lastError       error
	upstreamHealthy bool
	upstreamError   error
}

func (m *mockHealthChecker) HealthStatus() (bool, time.Time, error) {
	return m.healthy, m.lastSuccess, m.lastError
}

func (m *mockHealthChecker) CheckUpstreamHealth() error {
	if m.upstreamHealthy {
		return nil
	}
	if m.upstreamError != nil {
		return m.upstreamError
	}
	return nil
}

func TestSetupRoutes(t *testing.T) {
	tests := []struct {
		name           string
		logLevel       slog.Level
		method         string
		path           string
		expectedStatus int
		checkBody      func(*testing.T, string)
	}{
		// Basic routes tests
		{
			name:           "root path returns HTML",
			logLevel:       slog.LevelInfo,
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Incident.io Prometheus Exporter")
				assert.Contains(t, body, "/metrics")
			},
		},
		{
			name:           "metrics endpoint exists",
			logLevel:       slog.LevelInfo,
			method:         http.MethodGet,
			path:           "/metrics",
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				assert.NotEmpty(t, body)
			},
		},
		{
			name:           "health endpoint exists",
			logLevel:       slog.LevelInfo,
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "healthy")
			},
		},
		{
			name:           "ready endpoint exists",
			logLevel:       slog.LevelInfo,
			method:         http.MethodGet,
			path:           "/ready",
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "ready")
			},
		},
		// Debug mode tests - pprof endpoints
		{
			name:           "debug mode - pprof index accessible",
			logLevel:       slog.LevelDebug,
			method:         http.MethodGet,
			path:           "/debug/pprof/",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "debug mode - pprof cmdline accessible",
			logLevel:       slog.LevelDebug,
			method:         http.MethodGet,
			path:           "/debug/pprof/cmdline",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "debug mode - pprof symbol accessible",
			logLevel:       slog.LevelDebug,
			method:         http.MethodGet,
			path:           "/debug/pprof/symbol",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "debug mode - pprof trace accessible",
			logLevel:       slog.LevelDebug,
			method:         http.MethodGet,
			path:           "/debug/pprof/trace",
			expectedStatus: http.StatusOK,
		},
		// Non-debug mode test
		{
			name:           "non-debug mode - pprof not accessible",
			logLevel:       slog.LevelInfo,
			method:         http.MethodGet,
			path:           "/debug/pprof/",
			expectedStatus: http.StatusNotFound,
		},
		// Middleware test
		{
			name:           "middleware applied successfully",
			logLevel:       slog.LevelDebug,
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusOK,
		},
		// Not found route test
		{
			name:           "nonexistent route returns 404",
			logLevel:       slog.LevelInfo,
			method:         http.MethodGet,
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
		// Method not allowed test
		{
			name:           "POST to GET-only endpoint returns 405",
			logLevel:       slog.LevelInfo,
			method:         http.MethodPost,
			path:           "/",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			mockChecker := &mockHealthChecker{
				healthy:     true,
				lastSuccess: time.Now(),
			}
			setupRoutes(router, tt.logLevel, mockChecker)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.name == "middleware applied successfully" {
				req.Header.Set("User-Agent", "test-client")
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkBody != nil {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				tt.checkBody(t, string(body))
			}
		})
	}
}
