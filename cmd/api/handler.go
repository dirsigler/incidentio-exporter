package main

import (
	"fmt"
	"net/http"
)

func (app *application) indexHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		app.logger.Info("serving index page.")

		page := `
<html>
	<head><title>Incident.io Prometheus Exporter</title></head>
	<body>
		<h1>Incident.io Prometheus Exporter</h1>
		<p><a href="/metrics">Metrics</a></p>
	</body>
</html>`

		_, err := w.Write([]byte(page))
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}
	}
}
