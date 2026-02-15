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
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		params      LoadParams
		wantErr     bool
		errContains string
		checkFn     func(*testing.T, *Config)
	}{
		{
			name: "successful load with required params",
			params: LoadParams{
				Port:     9193,
				LogLevel: "info",
				APIKey:   "test-api-key-123",
			},
			wantErr: false,
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "test-api-key-123", cfg.IncidentIO.Key)
				assert.Equal(t, "https://api.incident.io", cfg.IncidentIO.URL)
				assert.Equal(t, 9193, cfg.Port)
				assert.Equal(t, slog.LevelInfo, cfg.Level)
			},
		},
		{
			name: "successful load with custom API URL",
			params: LoadParams{
				Port:     9193,
				LogLevel: "info",
				APIKey:   "test-api-key-123",
				APIURL:   "https://custom.incident.io",
			},
			wantErr: false,
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "test-api-key-123", cfg.IncidentIO.Key)
				assert.Equal(t, "https://custom.incident.io", cfg.IncidentIO.URL)
			},
		},
		{
			name: "missing required API key",
			params: LoadParams{
				Port:     9193,
				LogLevel: "info",
			},
			wantErr:     true,
			errContains: "API key is required",
		},
		{
			name: "custom port",
			params: LoadParams{
				Port:     8080,
				LogLevel: "info",
				APIKey:   "test-key",
			},
			wantErr: false,
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 8080, cfg.Port)
			},
		},
		{
			name: "debug log level",
			params: LoadParams{
				Port:     9193,
				LogLevel: "debug",
				APIKey:   "test-key",
			},
			wantErr: false,
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, slog.LevelDebug, cfg.Level)
			},
		},
		{
			name: "warn log level",
			params: LoadParams{
				Port:     9193,
				LogLevel: "warn",
				APIKey:   "test-key",
			},
			wantErr: false,
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, slog.LevelWarn, cfg.Level)
			},
		},
		{
			name: "error log level",
			params: LoadParams{
				Port:     9193,
				LogLevel: "error",
				APIKey:   "test-key",
			},
			wantErr: false,
			checkFn: func(t *testing.T, cfg *Config) {
				assert.Equal(t, slog.LevelError, cfg.Level)
			},
		},
		{
			name: "invalid log level",
			params: LoadParams{
				Port:     9193,
				LogLevel: "invalid",
				APIKey:   "test-key",
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "custom field IDs",
			params: LoadParams{
				Port:           9193,
				LogLevel:       "info",
				APIKey:         "test-key",
				CustomFieldIDs: []string{"field1", "field2", "field3"},
			},
			wantErr: false,
			checkFn: func(t *testing.T, cfg *Config) {
				require.Len(t, cfg.IncidentIO.CustomFieldIDs, 3)
				assert.Equal(t, "field1", cfg.IncidentIO.CustomFieldIDs[0])
				assert.Equal(t, "field2", cfg.IncidentIO.CustomFieldIDs[1])
				assert.Equal(t, "field3", cfg.IncidentIO.CustomFieldIDs[2])
			},
		},
		{
			name: "catalog type IDs",
			params: LoadParams{
				Port:           9193,
				LogLevel:       "info",
				APIKey:         "test-key",
				CatalogTypeIDs: []string{"catalog1", "catalog2"},
			},
			wantErr: false,
			checkFn: func(t *testing.T, cfg *Config) {
				require.Len(t, cfg.IncidentIO.CatalogTypeIDs, 2)
				assert.Equal(t, "catalog1", cfg.IncidentIO.CatalogTypeIDs[0])
				assert.Equal(t, "catalog2", cfg.IncidentIO.CatalogTypeIDs[1])
			},
		},
		{
			name: "both custom field and catalog type IDs",
			params: LoadParams{
				Port:           9193,
				LogLevel:       "info",
				APIKey:         "test-key",
				CustomFieldIDs: []string{"field1", "field2"},
				CatalogTypeIDs: []string{"catalog1", "catalog2"},
			},
			wantErr: false,
			checkFn: func(t *testing.T, cfg *Config) {
				require.Len(t, cfg.IncidentIO.CustomFieldIDs, 2)
				require.Len(t, cfg.IncidentIO.CatalogTypeIDs, 2)
				assert.Equal(t, "field1", cfg.IncidentIO.CustomFieldIDs[0])
				assert.Equal(t, "catalog1", cfg.IncidentIO.CatalogTypeIDs[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.checkFn != nil {
				tt.checkFn(t, cfg)
			}
		})
	}
}
