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

package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

// Config holds the entire application configuration.
type Config struct {
	IncidentIO IncidentIOConfig
	Port       int
	Level      slog.Level
}

// IncidentIOConfig holds incident.io API configuration.
type IncidentIOConfig struct {
	Key            string
	URL            string
	CustomFieldIDs []string
	CatalogTypeIDs []string
}

// LoadParams holds parameters for configuration loading.
type LoadParams struct {
	Port           int
	LogLevel       string
	APIKey         string
	APIURL         string
	CustomFieldIDs []string
	CatalogTypeIDs []string
}

// Load builds the configuration from provided parameters.
func Load(params LoadParams) (*Config, error) {
	var cfg Config

	// Parse log level
	var level slog.Level
	switch strings.ToLower(params.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", params.LogLevel)
	}
	cfg.Level = level
	cfg.Port = params.Port

	// Validate required fields
	if params.APIKey == "" {
		return nil, errors.New("API key is required")
	}
	cfg.IncidentIO.Key = params.APIKey

	// Set API URL
	if params.APIURL == "" {
		params.APIURL = "https://api.incident.io"
	}
	cfg.IncidentIO.URL = params.APIURL

	// Set custom field IDs
	cfg.IncidentIO.CustomFieldIDs = params.CustomFieldIDs

	// Set catalog type IDs
	cfg.IncidentIO.CatalogTypeIDs = params.CatalogTypeIDs

	return &cfg, nil
}
