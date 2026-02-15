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

// CatalogType represents a catalog type from the API.
type CatalogType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CatalogEntry represents a single entry in a catalog type.
type CatalogEntry struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	CatalogTypeID  string `json:"catalog_type_id"`
}

// CatalogEntriesResponse holds the API response from the /v3/catalog_entries endpoint.
type CatalogEntriesResponse struct {
	CatalogEntries []CatalogEntry `json:"catalog_entries"`
	CatalogType    CatalogType    `json:"catalog_type"`
	PaginationMeta struct {
		After            string `json:"after"`
		PageSize         int    `json:"page_size"`
		TotalRecordCount int    `json:"total_record_count"`
	} `json:"pagination_meta"`
}

// GetCatalogEntries retrieves all catalog entries for a given catalog type ID using the V3 API.
func (c *Client) GetCatalogEntries(ctx context.Context, catalogTypeID string) (*CatalogEntriesResponse, error) {
	allEntries := []CatalogEntry{}
	var catalogType CatalogType
	var totalRecordCount int
	after := ""
	pageCount := 0

	for {
		pageCount++
		path := fmt.Sprintf("/v3/catalog_entries?catalog_type_id=%s&page_size=%d",
			url.QueryEscape(catalogTypeID), defaultPageSize)
		if after != "" {
			path += "&after=" + url.QueryEscape(after)
		}

		slog.Debug("fetching catalog entries page",
			"catalog_type_id", catalogTypeID,
			"page", pageCount,
			"after", after,
		)

		response := &CatalogEntriesResponse{}
		if err := c.doRequest(ctx, path, response); err != nil {
			return nil, fmt.Errorf("failed to get catalog entries (page %d): %w", pageCount, err)
		}

		allEntries = append(allEntries, response.CatalogEntries...)
		catalogType = response.CatalogType
		totalRecordCount = response.PaginationMeta.TotalRecordCount

		slog.Debug("fetched catalog entries page",
			"catalog_type_id", catalogTypeID,
			"page", pageCount,
			"page_entries", len(response.CatalogEntries),
			"total_fetched", len(allEntries),
			"total_available", totalRecordCount,
		)

		// Check if we have more pages
		if response.PaginationMeta.After == "" || len(response.CatalogEntries) == 0 {
			break
		}

		after = response.PaginationMeta.After
	}

	slog.Info("completed fetching all catalog entries",
		"catalog_type_id", catalogTypeID,
		"total_pages", pageCount,
		"total_entries", len(allEntries),
	)

	return &CatalogEntriesResponse{
		CatalogEntries: allEntries,
		CatalogType:    catalogType,
		PaginationMeta: struct {
			After            string `json:"after"`
			PageSize         int    `json:"page_size"`
			TotalRecordCount int    `json:"total_record_count"`
		}{
			TotalRecordCount: totalRecordCount,
		},
	}, nil
}
