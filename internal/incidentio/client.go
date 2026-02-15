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
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/dirsigler/incidentio-exporter/internal/config"
)

const (
	// HTTP client timeouts
	defaultDialTimeout     = 10 * time.Second
	defaultRequestTimeout  = 30 * time.Second
	defaultIdleConnTimeout = 90 * time.Second
	defaultMaxIdleConns    = 100
	defaultMaxConnsPerHost = 10
)

// Client represents an incident.io API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new incident.io API client with configured timeouts.
func NewClient(cfg config.IncidentIOConfig) *Client {
	transport := &http.Transport{
		DialContext:         (&http.Transport{}).DialContext,
		MaxIdleConns:        defaultMaxIdleConns,
		MaxConnsPerHost:     defaultMaxConnsPerHost,
		IdleConnTimeout:     defaultIdleConnTimeout,
		TLSHandshakeTimeout: defaultDialTimeout,
	}

	slog.Info("initializing incident.io client",
		"api_url", cfg.URL,
		"api_key", redactAPIKey(cfg.Key),
	)

	return &Client{
		apiKey:  cfg.Key,
		baseURL: cfg.URL,
		httpClient: &http.Client{
			Timeout:   defaultRequestTimeout,
			Transport: transport,
		},
	}
}

// doRequest performs an HTTP GET request to the incident.io API and decodes the response.
func (c *Client) doRequest(ctx context.Context, path string, target interface{}) error {
	url := c.baseURL + path

	slog.Debug("api request starting",
		"method", "GET",
		"path", path,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Error("failed to create HTTP request",
			"error", err,
			"path", path,
		)
		return &NetworkError{URL: path, Err: err}
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("HTTP request failed",
			"error", err,
			"path", path,
		)
		return &NetworkError{URL: path, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("unexpected HTTP status code",
			"status_code", resp.StatusCode,
			"path", path,
			"status", resp.Status,
		)
		return &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   path,
			Status:     resp.Status,
			Message:    "unexpected HTTP status",
		}
	}

	if err = json.NewDecoder(resp.Body).Decode(&target); err != nil {
		slog.Error("failed to decode JSON response",
			"error", err,
			"path", path,
		)
		return &DecodeError{URL: path, Err: err}
	}

	slog.Debug("api request completed",
		"path", path,
		"status_code", resp.StatusCode,
	)

	return nil
}

// HealthCheck performs a lightweight API call to verify connectivity.
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Use the severities endpoint as it's lightweight and stable
	var result SeveritiesResponse
	return c.doRequest(ctx, "/v1/severities", &result)
}
