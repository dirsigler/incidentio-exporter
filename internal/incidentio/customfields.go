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

// CustomFieldOption represents a single option for a custom field.
type CustomFieldOption struct {
	ID            string `json:"id"`
	CustomFieldID string `json:"custom_field_id"`
	Value         string `json:"value"`
}

// CustomFieldOptionsResponse holds the API response from the /v1/custom_field_options endpoint.
type CustomFieldOptionsResponse struct {
	CustomFieldOptions []CustomFieldOption `json:"custom_field_options"`
	PaginationMeta     struct {
		After    string `json:"after"`
		PageSize int    `json:"page_size"`
	} `json:"pagination_meta"`
}

// GetCustomFieldOptions retrieves all custom field options for a given custom field ID.
func (c *Client) GetCustomFieldOptions(ctx context.Context, customFieldID string) (*CustomFieldOptionsResponse, error) {
	allOptions := []CustomFieldOption{}
	after := ""
	pageCount := 0

	for {
		pageCount++
		path := fmt.Sprintf("/v1/custom_field_options?custom_field_id=%s&page_size=%d",
			url.QueryEscape(customFieldID), defaultPageSize)
		if after != "" {
			path += "&after=" + url.QueryEscape(after)
		}

		slog.Debug("fetching custom field options page",
			"custom_field_id", customFieldID,
			"page", pageCount,
			"after", after,
		)

		response := &CustomFieldOptionsResponse{}
		if err := c.doRequest(ctx, path, response); err != nil {
			return nil, fmt.Errorf("failed to get custom field options (page %d): %w", pageCount, err)
		}

		allOptions = append(allOptions, response.CustomFieldOptions...)

		slog.Debug("fetched custom field options page",
			"custom_field_id", customFieldID,
			"page", pageCount,
			"page_options", len(response.CustomFieldOptions),
			"total_fetched", len(allOptions),
		)

		// Check if we have more pages
		if response.PaginationMeta.After == "" || len(response.CustomFieldOptions) == 0 {
			break
		}

		after = response.PaginationMeta.After
	}

	slog.Info("completed fetching all custom field options",
		"custom_field_id", customFieldID,
		"total_pages", pageCount,
		"total_options", len(allOptions),
	)

	return &CustomFieldOptionsResponse{
		CustomFieldOptions: allOptions,
	}, nil
}
