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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dirsigler/incidentio-exporter/internal/config"
	"github.com/go-chi/chi/v5"
)

const (
	ctxTimeout   = time.Second * 30
	idleTimeout  = time.Minute * 1
	readTimeout  = time.Second * 5
	writeTimeout = time.Second * 10
)

// Server represents the HTTP server for the exporter.
type Server struct {
	config        *config.Config
	wg            sync.WaitGroup
	healthChecker healthChecker
}

// New creates a new HTTP server instance.
func New(cfg *config.Config, checker healthChecker) *Server {
	return &Server{
		config:        cfg,
		healthChecker: checker,
	}
}

// Run starts the HTTP server and handles graceful shutdown.
func (s *Server) Run() error {
	router := chi.NewRouter()
	setupRoutes(router, s.config.Level, s.healthChecker)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      router,
		IdleTimeout:  idleTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		slog.Info("shutdown signal received",
			"signal", sig.String(),
			"component", "server",
		)

		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
		defer cancel()

		slog.Info("initiating graceful shutdown",
			"timeout_seconds", ctxTimeout.Seconds(),
		)

		err := srv.Shutdown(ctx)
		if err != nil {
			slog.Error("error during shutdown",
				"error", err,
				"component", "server",
			)
			shutdownError <- err
			return
		}

		slog.Debug("waiting for background tasks to complete")
		s.wg.Wait()
		shutdownError <- nil
	}()

	slog.Info("http server starting",
		"addr", srv.Addr,
		"port", s.config.Port,
		"log_level", s.config.Level.String(),
		"component", "server",
	)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error("http server error",
			"error", err,
			"addr", srv.Addr,
			"component", "server",
		)
		return fmt.Errorf("server listen error: %w", err)
	}

	err = <-shutdownError
	if err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	slog.Info("http server stopped cleanly",
		"addr", srv.Addr,
		"component", "server",
	)

	return nil
}
