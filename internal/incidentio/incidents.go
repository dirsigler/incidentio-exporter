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
	"log/slog"
	"net/url"
)

const (
	// API pagination limit
	defaultPageSize = 250
)

// Incident represents a single incident from the API.
type Incident struct {
	ID       string `json:"id"`
	Severity struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"severity"`
	IncidentStatus struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"incident_status"`
	CustomFieldEntries []CustomFieldEntry `json:"custom_field_entries"`
}

// CustomFieldEntry represents a custom field value in an incident.
type CustomFieldEntry struct {
	CustomField struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"custom_field"`
	Values []CustomFieldValue `json:"values"`
}

// CustomFieldValueOption represents a selected option value for a custom field.
type CustomFieldValueOption struct {
	ID            string `json:"id"`
	CustomFieldID string `json:"custom_field_id"`
	Value         string `json:"value"`
	SortKey       int    `json:"sort_key"`
}

// CustomFieldValue represents a single value in a custom field.
type CustomFieldValue struct {
	ValueText         string                  `json:"value_text"`
	ValueOption       *CustomFieldValueOption `json:"value_option,omitempty"`
	ValueCatalogEntry *CatalogEntryRef        `json:"value_catalog_entry,omitempty"`
	ValueLink         string                  `json:"value_link"`
	ValueNumeric      string                  `json:"value_numeric"`
}

// CatalogEntryRef represents a reference to a catalog entry.
type CatalogEntryRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IncidentsResponse holds the API response from the /v2/incidents endpoint.
type IncidentsResponse struct {
	Incidents      []Incident `json:"incidents"`
	PaginationMeta struct {
		After            string `json:"after"`
		PageSize         int    `json:"page_size"`
		TotalRecordCount int    `json:"total_record_count"`
	} `json:"pagination_meta"`
}

// GetIncidents retrieves all incidents from the incident.io API with pagination support.
func (c *Client) GetIncidents(ctx context.Context) (*IncidentsResponse, error) {
	allIncidents := []Incident{}
	var totalRecordCount int
	after := ""
	pageCount := 0

	for {
		pageCount++
		path := fmt.Sprintf("/v2/incidents?page_size=%d", defaultPageSize)
		if after != "" {
			path += "&after=" + url.QueryEscape(after)
		}

		slog.Debug("fetching incidents page",
			"page", pageCount,
			"after", after,
		)

		response := &IncidentsResponse{}
		if err := c.doRequest(ctx, path, response); err != nil {
			return nil, fmt.Errorf("failed to get incidents (page %d): %w", pageCount, err)
		}

		allIncidents = append(allIncidents, response.Incidents...)
		totalRecordCount = response.PaginationMeta.TotalRecordCount

		slog.Debug("fetched incidents page",
			"page", pageCount,
			"page_incidents", len(response.Incidents),
			"total_fetched", len(allIncidents),
			"total_available", totalRecordCount,
		)

		// Check if we have more pages
		if response.PaginationMeta.After == "" || len(response.Incidents) == 0 {
			break
		}

		after = response.PaginationMeta.After
	}

	slog.Info("completed fetching all incidents",
		"total_pages", pageCount,
		"total_incidents", len(allIncidents),
	)

	return &IncidentsResponse{
		Incidents: allIncidents,
		PaginationMeta: struct {
			After            string `json:"after"`
			PageSize         int    `json:"page_size"`
			TotalRecordCount int    `json:"total_record_count"`
		}{
			TotalRecordCount: totalRecordCount,
		},
	}, nil
}
