package main

import (
	"fmt"
)

// SeverityResponse holds the minimum API response from the https://api.incident.io/v1/severities endpoint.
type SeverityResponse struct {
	Severities []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"severities"`
}

// getSeverities gets all available Severities configured.
func (app *application) getSeverities() SeverityResponse {
	response := SeverityResponse{}

	url := fmt.Sprintf(app.config.IncidentIO.URL + "/v1/severities")
	err := doHTTP(url, app.config.IncidentIO.Key, &response)
	if err != nil {
		app.logger.Error("failed to get all severities", "error", err)
	}

	return response
}

// getSeverityIncidents takes the severity ID and responds with all available incidents of this severity.
func (app *application) getSeverityIncidents(severityID string) IncidentsResponse {
	response := IncidentsResponse{}

	url := fmt.Sprintf(app.config.IncidentIO.URL+"/v2/incidents?severity[one_of]=%s", severityID)
	err := doHTTP(url, app.config.IncidentIO.Key, &response)
	if err != nil {
		app.logger.Error("failed to get all incidents", "error", err)
	}

	return response
}
