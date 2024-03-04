package main

import (
	"fmt"
)

// StatusResponse holds the minimum API response from the https://api.incident.io/v1/incident_statuses endpoint.
type StatusResponse struct {
	IncidentStatuses []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"incident_statuses"`
}

// getStatuses gets all configured Status in the custom field.
func (app *application) getStatuses() StatusResponse {
	response := StatusResponse{}

	url := fmt.Sprintf(app.config.IncidentIO.URL + "/v1/incident_statuses")
	err := doHTTP(url, app.config.IncidentIO.Key, &response)
	if err != nil {
		app.logger.Error("failed to get all incidents", "error", err)
	}

	return response
}

// getStatusIncidents takes the status ID and responds with all available incidents of this status.
func (app *application) getStatusIncidents(statusID string) IncidentsResponse {
	response := IncidentsResponse{}

	url := fmt.Sprintf(app.config.IncidentIO.URL+"/v2/incidents?status[one_of]=%s", statusID)
	err := doHTTP(url, app.config.IncidentIO.Key, &response)
	if err != nil {
		app.logger.Error("failed to get all incidents", "error", err)
	}

	return response
}
