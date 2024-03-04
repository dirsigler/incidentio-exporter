package main

import "fmt"

// IncidentsResponse holds the minimum API response from the https://api.incident.io/v2/incidents endpoint.
type IncidentsResponse struct {
	PaginationMeta struct {
		TotalRecordCount int `json:"total_record_count"`
	} `json:"pagination_meta"`
}

// getIncidents gets ALL incidents available and returns an object of type IncidentsResponse.
func (app *application) getIncidents() IncidentsResponse {
	response := IncidentsResponse{}

	url := fmt.Sprintf(app.config.IncidentIO.URL + "/v2/incidents")

	err := doHTTP(url, app.config.IncidentIO.Key, &response)
	if err != nil {
		app.logger.Error("failed to get all incidents", "error", err)
	}

	return response
}
