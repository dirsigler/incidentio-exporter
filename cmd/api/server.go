package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	ctxTimeout   = time.Second * 30
	idleTimeout  = time.Minute * 1
	readTimeout  = time.Second * 5
	writeTimeout = time.Second * 10
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.Port),
		Handler:      app.routes(),
		IdleTimeout:  idleTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("received signal to stop server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting incident.io exporter", "addr", srv.Addr, "level", app.config.Level.String())
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}
	app.logger.Info("stopped incident.io exporter", "addr", srv.Addr)

	return nil
}
