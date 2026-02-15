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

package incidentio

import (
	"context"
	"fmt"
)

// Status represents an incident status.
type Status struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// StatusesResponse holds the API response from the /v1/incident_statuses endpoint.
type StatusesResponse struct {
	IncidentStatuses []Status `json:"incident_statuses"`
}

// GetStatuses retrieves all configured statuses from the incident.io API.
func (c *Client) GetStatuses(ctx context.Context) (*StatusesResponse, error) {
	response := &StatusesResponse{}
	if err := c.doRequest(ctx, "/v1/incident_statuses", response); err != nil {
		return nil, fmt.Errorf("failed to get all statuses: %w", err)
	}

	return response, nil
}
