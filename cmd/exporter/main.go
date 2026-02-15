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

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/dirsigler/incidentio-exporter/internal/collector"
	"github.com/dirsigler/incidentio-exporter/internal/config"
	"github.com/dirsigler/incidentio-exporter/internal/incidentio"
	"github.com/dirsigler/incidentio-exporter/internal/logger"
	"github.com/dirsigler/incidentio-exporter/internal/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v3"
)

var (
	// Version information - set via ldflags at build time
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	cmd := &cli.Command{
		Name:    "incidentio-exporter",
		Usage:   "Prometheus exporter for incident.io metrics",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate),
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   9193,
				Usage:   "Port to listen for HTTP requests",
				Sources: cli.EnvVars("SERVER_PORT"),
			},
			&cli.StringFlag{
				Name:    "log-level",
				Value:   "info",
				Usage:   "Log level (debug, info, warn, error)",
				Sources: cli.EnvVars("LOG_LEVEL"),
			},
			&cli.StringFlag{
				Name:     "api-key",
				Usage:    "incident.io API key (required)",
				Sources:  cli.EnvVars("INCIDENTIO_API_KEY"),
				Required: true,
			},
			&cli.StringFlag{
				Name:    "api-url",
				Value:   "https://api.incident.io",
				Usage:   "incident.io API base URL",
				Sources: cli.EnvVars("INCIDENTIO_API_URL"),
			},
			&cli.StringSliceFlag{
				Name:    "custom-field-ids",
				Usage:   "Custom field IDs to track (can be specified multiple times)",
				Sources: cli.EnvVars("INCIDENTIO_CUSTOM_FIELD_IDS"),
			},
			&cli.StringSliceFlag{
				Name:    "catalog-type-ids",
				Usage:   "Catalog type IDs to track (can be specified multiple times)",
				Sources: cli.EnvVars("INCIDENTIO_CATALOG_TYPE_IDS"),
			},
		},
		Action: run,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	cfg, err := config.Load(config.LoadParams{
		Port:           cmd.Int("port"),
		LogLevel:       cmd.String("log-level"),
		APIKey:         cmd.String("api-key"),
		APIURL:         cmd.String("api-url"),
		CustomFieldIDs: cmd.StringSlice("custom-field-ids"),
		CatalogTypeIDs: cmd.StringSlice("catalog-type-ids"),
	})
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Setup(cfg.Level)
	slog.Info("application starting",
		"version", version,
		"commit", commit,
		"build_date", buildDate,
		"log_level", cfg.Level.String(),
		"api_url", cfg.IncidentIO.URL,
		"custom_field_ids", cfg.IncidentIO.CustomFieldIDs,
		"catalog_type_ids", cfg.IncidentIO.CatalogTypeIDs,
	)

	client := incidentio.NewClient(cfg.IncidentIO)
	incidentioCollector := collector.New(client, cfg.IncidentIO.CustomFieldIDs, cfg.IncidentIO.CatalogTypeIDs)
	prometheus.MustRegister(incidentioCollector)
	slog.Info("prometheus collector registered")

	srv := server.New(cfg, incidentioCollector)
	if err := srv.Run(); err != nil {
		return fmt.Errorf("server terminated with error: %w", err)
	}

	slog.Info("application shutdown complete")
	return nil
}
