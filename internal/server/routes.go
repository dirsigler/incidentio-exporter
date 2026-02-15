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
	"log/slog"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// setupRoutes configures all HTTP routes for the server.
func setupRoutes(router *chi.Mux, level slog.Level, checker healthChecker) {
	router.Use(logMiddleware(level))

	router.Get("/", indexHandler())
	router.Handle("/metrics", promhttp.Handler())
	router.Get("/health", healthHandler(checker))
	router.Get("/ready", readyHandler(checker))

	if level == slog.LevelDebug {
		slog.Warn("serving pprof path via '/debug/pprof'")
		router.Route("/debug/pprof", func(router chi.Router) {
			router.HandleFunc("/*", pprof.Index)
			router.HandleFunc("/cmdline", pprof.Cmdline)
			router.HandleFunc("/profile", pprof.Profile)
			router.HandleFunc("/symbol", pprof.Symbol)
			router.HandleFunc("/trace", pprof.Trace)
		})
	}
}
