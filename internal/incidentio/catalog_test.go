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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dirsigler/incidentio-exporter/internal/config"
)

func TestGetCatalogEntries(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.URL.Query().Get("catalog_type_id") != "test-catalog-id" {
			t.Errorf("unexpected catalog_type_id: %s", r.URL.Query().Get("catalog_type_id"))
		}

		// Return mock response matching actual API structure
		response := CatalogEntriesResponse{
			CatalogEntries: []CatalogEntry{
				{
					ID:            "entry-1",
					Name:          "Entry 1",
					CatalogTypeID: "test-catalog-id",
				},
				{
					ID:            "entry-2",
					Name:          "Entry 2",
					CatalogTypeID: "test-catalog-id",
				},
			},
			CatalogType: CatalogType{
				ID:   "test-catalog-id",
				Name: "Test Catalog",
			},
			PaginationMeta: struct {
				After            string `json:"after"`
				PageSize         int    `json:"page_size"`
				TotalRecordCount int    `json:"total_record_count"`
			}{
				After:            "",
				PageSize:         250,
				TotalRecordCount: 2,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server
	cfg := config.IncidentIOConfig{
		Key:      "test-key",
		URL:      server.URL,
	}
	client := NewClient(cfg)

	// Test GetCatalogEntries
	ctx := context.Background()
	result, err := client.GetCatalogEntries(ctx, "test-catalog-id")
	if err != nil {
		t.Fatalf("GetCatalogEntries failed: %v", err)
	}

	if len(result.CatalogEntries) != 2 {
		t.Errorf("expected 2 catalog entries, got %d", len(result.CatalogEntries))
	}

	if result.CatalogEntries[0].Name != "Entry 1" {
		t.Errorf("expected 'Entry 1', got '%s'", result.CatalogEntries[0].Name)
	}

	if result.PaginationMeta.TotalRecordCount != 2 {
		t.Errorf("expected total record count of 2, got %d", result.PaginationMeta.TotalRecordCount)
	}
}

func TestGetCatalogEntriesPagination(t *testing.T) {
	pageCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCount++

		var response CatalogEntriesResponse
		if pageCount == 1 {
			// First page
			response = CatalogEntriesResponse{
				CatalogEntries: []CatalogEntry{
					{
						ID:            "entry-1",
						Name:          "Entry 1",
						CatalogTypeID: "test-catalog-id",
					},
				},
				CatalogType: CatalogType{
					ID:   "test-catalog-id",
					Name: "Test Catalog",
				},
				PaginationMeta: struct {
					After            string `json:"after"`
					PageSize         int    `json:"page_size"`
					TotalRecordCount int    `json:"total_record_count"`
				}{
					After:            "page2",
					PageSize:         1,
					TotalRecordCount: 2,
				},
			}
		} else {
			// Second page
			response = CatalogEntriesResponse{
				CatalogEntries: []CatalogEntry{
					{
						ID:            "entry-2",
						Name:          "Entry 2",
						CatalogTypeID: "test-catalog-id",
					},
				},
				CatalogType: CatalogType{
					ID:   "test-catalog-id",
					Name: "Test Catalog",
				},
				PaginationMeta: struct {
					After            string `json:"after"`
					PageSize         int    `json:"page_size"`
					TotalRecordCount int    `json:"total_record_count"`
				}{
					After:            "",
					PageSize:         1,
					TotalRecordCount: 2,
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	cfg := config.IncidentIOConfig{
		Key:      "test-key",
		URL:      server.URL,
	}
	client := NewClient(cfg)

	ctx := context.Background()
	result, err := client.GetCatalogEntries(ctx, "test-catalog-id")
	if err != nil {
		t.Fatalf("GetCatalogEntries failed: %v", err)
	}

	if len(result.CatalogEntries) != 2 {
		t.Errorf("expected 2 catalog entries across pages, got %d", len(result.CatalogEntries))
	}

	if pageCount != 2 {
		t.Errorf("expected 2 pages to be fetched, got %d", pageCount)
	}
}
