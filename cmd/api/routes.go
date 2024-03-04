package main

import (
	"log/slog"
	"net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// routes serves the routes for the HTTP server.
func (app *application) routes() *chi.Mux {
	router := chi.NewRouter()

	router.Use(
		app.logHandlerMiddleware,
	)

	router.Get("/", app.indexHandler())

	router.Handle("/metrics", promhttp.Handler())

	// Serve pprof utilities if application runs in a Debug mode.
	if app.config.Level == slog.LevelDebug {
		app.logger.Warn("serving pprof path via '/debug/pprof'")
		router.Route("/debug/pprof", func(router chi.Router) {
			router.HandleFunc("/*", pprof.Index)
			router.HandleFunc("/cmdline", pprof.Cmdline)
			router.HandleFunc("/profile", pprof.Profile)
			router.HandleFunc("/symbol", pprof.Symbol)
			router.HandleFunc("/trace", pprof.Trace)
		})
	}

	return router
}
