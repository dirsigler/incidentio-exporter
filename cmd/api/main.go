package main

import (
	"flag"
	"log/slog"
	"os"
	"sync"

	"github.com/caarlos0/env/v9"
	"github.com/prometheus/client_golang/prometheus"
)

// config type which holds the whole configuration of the application.
type config struct {
	IncidentIO struct {
		Key string `env:"INCIDENTIO_API_KEY"     required:"true"`
		URL string `env:"INCIDENTIO_API_URL"     envDefault:"https://api.incident.io"`
	}
	Port  int
	Level slog.Level
}

// application type which holds all dependencies and the configuration of the application.
type application struct {
	logger *slog.Logger
	config config
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	// Parse environment variables to the config struct.
	if err := env.Parse(&cfg); err != nil {
		return
	}

	flag.IntVar(&cfg.Port, "server.addr", 9193, "Address to listen for requests.")
	flag.TextVar(&cfg.Level, "log.level", slog.LevelInfo, "Configured Log level.")
	flag.Parse()

	// Initialize new default logger passed to application
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level}))

	// Implement an application
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Implement new collector
	incidentioCollector := NewIncidentCollector(app)
	prometheus.MustRegister(&incidentioCollector)

	// Serve HTTP application
	if err := app.serve(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
