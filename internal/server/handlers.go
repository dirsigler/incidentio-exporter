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
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type healthChecker interface {
	HealthStatus() (healthy bool, lastSuccess time.Time, lastError error)
	CheckUpstreamHealth() error
}

// indexHandler serves the root HTML page with links to metrics.
func indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := `
<html>
	<head><title>Incident.io Prometheus Exporter</title></head>
	<body>
		<h1>Incident.io Prometheus Exporter</h1>
		<ul>
			<li><a href="/metrics">Metrics</a></li>
			<li><a href="/health">Health</a></li>
			<li><a href="/ready">Readiness</a></li>
		</ul>
	</body>
</html>`

		_, err := w.Write([]byte(page))
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}
	}
}

// healthHandler returns a health check that includes upstream API connectivity.
func healthHandler(checker healthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check upstream API health
		upstreamErr := checker.CheckUpstreamHealth()

		response := map[string]interface{}{
			"status": "healthy",
			"upstream": map[string]interface{}{
				"reachable": upstreamErr == nil,
			},
		}

		if upstreamErr != nil {
			response["status"] = "degraded"
			response["upstream"].(map[string]interface{})["error"] = upstreamErr.Error()
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to encode health response", "error", err)
		}
	}
}

// readyHandler checks if the exporter is ready to serve metrics.
func readyHandler(checker healthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		healthy, lastSuccess, lastError := checker.HealthStatus()

		w.Header().Set("Content-Type", "application/json")

		response := map[string]interface{}{
			"ready": healthy,
		}

		if !lastSuccess.IsZero() {
			response["last_success"] = lastSuccess.Format(time.RFC3339)
			response["last_success_seconds_ago"] = time.Since(lastSuccess).Seconds()
		}

		if lastError != nil {
			response["last_error"] = lastError.Error()
		}

		if healthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to encode ready response", "error", err)
		}
	}
}
