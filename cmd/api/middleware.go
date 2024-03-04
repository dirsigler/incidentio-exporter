package main

import (
	"log/slog"
	"net/http"
)

// logHandlerMiddleware provides a middleware to log additional information from the HTTP request.
// It logs on the DEBUG Log level the requested HTTP path and remote address of the request initiator.
func (app *application) logHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.Level == slog.LevelDebug {
			app.logger.Debug("logged handler call", slog.Group("request",
				"path", r.URL.Path,
				"ip", r.RemoteAddr,
			))
		}

		next.ServeHTTP(w, r)
	})
}
